// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package p2p

import (
	"container/heap"
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"

	consensusversion "github.com/luxfi/consensus/version"
	"github.com/luxfi/ids"
	log "github.com/luxfi/log"
	"github.com/luxfi/math/set"
	"github.com/luxfi/metric"

	safemath "github.com/luxfi/math"
)

const (
	bandwidthHalflife = 5 * time.Minute

	// controls how eagerly we connect to new peers vs. using peers with known
	// good response bandwidth.
	desiredMinResponsivePeers = 20
	newPeerConnectFactor      = 0.1

	// The probability that, when we select a peer, we select randomly rather
	// than based on their performance.
	randomPeerProbability = 0.2
)

// heapItem is an item in the bandwidth heap
type heapItem struct {
	nodeID    ids.NodeID
	bandwidth safemath.Averager
	index     int
}

// bandwidthHeap is a max-heap of peers ordered by bandwidth
type bandwidthHeap struct {
	items   []*heapItem
	nodeMap map[ids.NodeID]*heapItem
}

func newBandwidthHeap() *bandwidthHeap {
	return &bandwidthHeap{
		items:   make([]*heapItem, 0),
		nodeMap: make(map[ids.NodeID]*heapItem),
	}
}

func (h *bandwidthHeap) Len() int { return len(h.items) }

func (h *bandwidthHeap) Less(i, j int) bool {
	// Max heap: higher bandwidth first
	return h.items[i].bandwidth.Read() > h.items[j].bandwidth.Read()
}

func (h *bandwidthHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.items[i].index = i
	h.items[j].index = j
}

func (h *bandwidthHeap) Push(x interface{}) {
	item := x.(*heapItem)
	item.index = len(h.items)
	h.items = append(h.items, item)
	h.nodeMap[item.nodeID] = item
}

func (h *bandwidthHeap) Pop() interface{} {
	old := h.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	h.items = old[0 : n-1]
	delete(h.nodeMap, item.nodeID)
	item.index = -1
	return item
}

func (h *bandwidthHeap) Peek() (ids.NodeID, safemath.Averager, bool) {
	if len(h.items) == 0 {
		return ids.EmptyNodeID, nil, false
	}
	item := h.items[0]
	return item.nodeID, item.bandwidth, true
}

func (h *bandwidthHeap) PushOrUpdate(nodeID ids.NodeID, bandwidth safemath.Averager) {
	if item, ok := h.nodeMap[nodeID]; ok {
		item.bandwidth = bandwidth
		heap.Fix(h, item.index)
	} else {
		heap.Push(h, &heapItem{nodeID: nodeID, bandwidth: bandwidth})
	}
}

func (h *bandwidthHeap) Remove(nodeID ids.NodeID) {
	if item, ok := h.nodeMap[nodeID]; ok {
		heap.Remove(h, item.index)
	}
}

// Tracks the bandwidth of responses coming from peers,
// preferring to contact peers with known good bandwidth, connecting
// to new peers with an exponentially decaying probability.
type PeerTracker struct {
	// Lock to protect concurrent access to the peer tracker
	lock sync.RWMutex
	// Peers that we're connected to that we haven't sent a request to since we
	// most recently connected to them.
	untrackedPeers set.Set[ids.NodeID]
	// Peers that we're connected to that we've sent a request to since we most
	// recently connected to them.
	trackedPeers set.Set[ids.NodeID]
	// Peers that we're connected to that responded to the last request they
	// were sent.
	responsivePeers set.Set[ids.NodeID]
	// Bandwidth of peers that we have measured.
	peerBandwidth map[ids.NodeID]safemath.Averager
	// Max heap that contains the average bandwidth of peers that do not have an
	// outstanding request.
	bandwidthHeap *bandwidthHeap
	// Average bandwidth is only used for metric.
	averageBandwidth safemath.Averager

	// The below fields are assumed to be constant and are not protected by the
	// lock.
	log          log.Logger
	ignoredNodes set.Set[ids.NodeID]
	minVersion   *consensusversion.Application
	metrics      peerTrackerMetrics
}

type peerTrackerMetrics struct {
	numTrackedPeers    metric.Gauge
	numResponsivePeers metric.Gauge
	averageBandwidth   metric.Gauge
}

func NewPeerTracker(
	log log.Logger,
	metricsNamespace string,
	registerer metric.Registerer,
	ignoredNodes set.Set[ids.NodeID],
	minVersion *consensusversion.Application,
) (*PeerTracker, error) {
	t := &PeerTracker{
		untrackedPeers:   set.NewSet[ids.NodeID](0),
		trackedPeers:     set.NewSet[ids.NodeID](0),
		responsivePeers:  set.NewSet[ids.NodeID](0),
		peerBandwidth:    make(map[ids.NodeID]safemath.Averager),
		bandwidthHeap:    newBandwidthHeap(),
		averageBandwidth: safemath.NewAverager(0, bandwidthHalflife, time.Now()),
		log:              log,
		ignoredNodes:     ignoredNodes,
		minVersion:       minVersion,
		metrics: peerTrackerMetrics{
			numTrackedPeers: metric.NewGauge(
				metric.GaugeOpts{
					Namespace: metricsNamespace,
					Name:      "num_tracked_peers",
					Help:      "number of tracked peers",
				},
			),
			numResponsivePeers: metric.NewGauge(
				metric.GaugeOpts{
					Namespace: metricsNamespace,
					Name:      "num_responsive_peers",
					Help:      "number of responsive peers",
				},
			),
			averageBandwidth: metric.NewGauge(
				metric.GaugeOpts{
					Namespace: metricsNamespace,
					Name:      "average_bandwidth",
					Help:      "average sync bandwidth used by peers",
				},
			),
		},
	}

	err := errors.Join()
	return t, err
}

// Returns true if:
//   - We have not observed the desired minimum number of responsive peers.
//   - Randomly with the frequency decreasing as the number of responsive peers
//     increases.
//
// Assumes the read lock is held.
func (p *PeerTracker) shouldSelectUntrackedPeer() bool {
	numResponsivePeers := p.responsivePeers.Len()
	if numResponsivePeers < desiredMinResponsivePeers {
		return true
	}
	if p.untrackedPeers.Len() == 0 {
		return false // already tracking all peers
	}

	newPeerProbability := math.Exp(-float64(numResponsivePeers) * newPeerConnectFactor)
	return rand.Float64() < newPeerProbability // #nosec G404
}

// SelectPeer that we could send a request to.
func (p *PeerTracker) SelectPeer() (ids.NodeID, bool) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if p.shouldSelectUntrackedPeer() {
		if nodeID, ok := p.untrackedPeers.Peek(); ok {
			p.log.Debug("selecting peer",
				log.UserString("reason", "untracked"),
				log.Stringer("nodeID", nodeID),
				log.Int("trackedPeers", p.trackedPeers.Len()),
				log.Int("responsivePeers", p.responsivePeers.Len()),
			)
			return nodeID, true
		}
	}

	useBandwidthHeap := rand.Float64() > randomPeerProbability // #nosec G404
	if useBandwidthHeap {
		if nodeID, bandwidth, ok := p.bandwidthHeap.Peek(); ok {
			p.log.Debug("selecting peer",
				log.UserString("reason", "bandwidth"),
				log.Stringer("nodeID", nodeID),
				log.Float64("bandwidth", bandwidth.Read()),
			)
			return nodeID, true
		}
	} else {
		if nodeID, ok := p.responsivePeers.Peek(); ok {
			p.log.Debug("selecting peer",
				log.UserString("reason", "responsive"),
				log.Stringer("nodeID", nodeID),
			)
			return nodeID, true
		}
	}

	if nodeID, ok := p.trackedPeers.Peek(); ok {
		p.log.Debug("selecting peer",
			log.UserString("reason", "tracked"),
			log.Stringer("nodeID", nodeID),
			log.Bool("checkedBandwidthHeap", useBandwidthHeap),
		)
		return nodeID, true
	}

	// We're not connected to any peers.
	return ids.EmptyNodeID, false
}

// Record that we sent a request to [nodeID].
func (p *PeerTracker) RegisterRequest(nodeID ids.NodeID) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.untrackedPeers.Remove(nodeID)
	p.trackedPeers.Add(nodeID)
	p.bandwidthHeap.Remove(nodeID)

	p.metrics.numTrackedPeers.Set(float64(p.trackedPeers.Len()))
}

// Record that we observed that [nodeID]'s bandwidth is [bandwidth].
func (p *PeerTracker) RegisterResponse(nodeID ids.NodeID, bandwidth float64) {
	p.updateBandwidth(nodeID, bandwidth, true)
}

// Record that a request failed to [nodeID].
func (p *PeerTracker) RegisterFailure(nodeID ids.NodeID) {
	p.updateBandwidth(nodeID, 0, false)
}

func (p *PeerTracker) updateBandwidth(nodeID ids.NodeID, bandwidth float64, responsive bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if !p.trackedPeers.Contains(nodeID) {
		p.log.Debug("tracking bandwidth for untracked peer",
			log.Stringer("nodeID", nodeID),
		)
		return
	}

	now := time.Now()
	peerBandwidth, ok := p.peerBandwidth[nodeID]
	if ok {
		peerBandwidth.Observe(bandwidth, now)
	} else {
		peerBandwidth = safemath.NewAverager(bandwidth, bandwidthHalflife, now)
		p.peerBandwidth[nodeID] = peerBandwidth
	}
	p.bandwidthHeap.PushOrUpdate(nodeID, peerBandwidth)
	p.averageBandwidth.Observe(bandwidth, now)

	if responsive {
		p.responsivePeers.Add(nodeID)
	} else {
		p.responsivePeers.Remove(nodeID)
	}

	p.metrics.numResponsivePeers.Set(float64(p.responsivePeers.Len()))
	p.metrics.averageBandwidth.Set(p.averageBandwidth.Read())
}

// Connected should be called when [nodeID] connects to this node.
func (p *PeerTracker) Connected(nodeID ids.NodeID, nodeVersion *consensusversion.Application) {
	// If this peer should be ignored, don't mark it as connected.
	if p.ignoredNodes.Contains(nodeID) {
		return
	}
	// If minVersion is specified and peer's version is less, don't mark it as
	// connected.
	if p.minVersion != nil && nodeVersion.Compare(p.minVersion) < 0 {
		return
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	p.untrackedPeers.Add(nodeID)
}

// Disconnected should be called when [nodeID] disconnects from this node.
func (p *PeerTracker) Disconnected(nodeID ids.NodeID) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.untrackedPeers.Remove(nodeID)
	p.trackedPeers.Remove(nodeID)
	p.responsivePeers.Remove(nodeID)
	delete(p.peerBandwidth, nodeID)
	p.bandwidthHeap.Remove(nodeID)

	p.metrics.numTrackedPeers.Set(float64(p.trackedPeers.Len()))
	p.metrics.numResponsivePeers.Set(float64(p.responsivePeers.Len()))
}

// Returns the number of peers the node is connected to.
func (p *PeerTracker) Size() int {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.untrackedPeers.Len() + p.trackedPeers.Len()
}
