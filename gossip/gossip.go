// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package gossip

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	luxlog "github.com/luxfi/log"
	"github.com/luxfi/metric"

	"github.com/luxfi/ids"
	log "github.com/luxfi/log"
	"github.com/luxfi/math/set"
	"github.com/luxfi/p2p"
)

const (
	ioLabel    = "io"
	sentIO     = "sent"
	receivedIO = "received"

	typeLabel  = "type"
	pushType   = "push"
	pullType   = "pull"
	unsentType = "unsent"
	sentType   = "sent"

	defaultGossipableCount = 64
	defaultInitSize        = 32
)

var (
	_ Gossiper = (*ValidatorGossiper)(nil)
	_ Gossiper = (*PullGossiper[Gossipable])(nil)
	_ Gossiper = (*NoOpGossiper)(nil)

	_ Set[Gossipable] = (*FullSet[Gossipable])(nil)

	ErrInvalidNumValidators     = errors.New("num validators cannot be negative")
	ErrInvalidNumNonValidators  = errors.New("num non-validators cannot be negative")
	ErrInvalidNumPeers          = errors.New("num peers cannot be negative")
	ErrInvalidNumToGossip       = errors.New("must gossip to at least one peer")
	ErrInvalidDiscardedSize     = errors.New("discarded size cannot be negative")
	ErrInvalidTargetGossipSize  = errors.New("target gossip size cannot be negative")
	ErrInvalidRegossipFrequency = errors.New("re-gossip frequency cannot be negative")
)

// Gossiper gossips Gossipables to other nodes
type Gossiper interface {
	// Gossip runs a cycle of gossip. Returns an error if we failed to gossip.
	Gossip(ctx context.Context) error
}

// ValidatorGossiper only calls [Gossip] if the given node is a validator
type ValidatorGossiper struct {
	Gossiper

	NodeID     ids.NodeID
	Validators p2p.ValidatorSet
}

// Metrics that are tracked across a gossip protocol. A given protocol should
// only use a single instance of Metrics.
type Metrics struct {
	count                   metric.CounterVec
	bytes                   metric.CounterVec
	tracking                metric.GaugeVec
	trackingLifetimeAverage metric.Gauge
	topValidators           metric.GaugeVec
}

// NewMetrics returns a common set of metrics
func NewMetrics(
	metrics metric.Registerer,
	namespace string,
) (Metrics, error) {
	ioTypeLabels := []string{ioLabel, typeLabel}
	typeLabels := []string{typeLabel}

	m := Metrics{
		count: metric.NewCounterVec(
			metric.CounterOpts{
				Namespace: namespace,
				Name:      "gossip_count",
				Help:      "amount of gossip (n)",
			},
			ioTypeLabels,
		),
		bytes: metric.NewCounterVec(
			metric.CounterOpts{
				Namespace: namespace,
				Name:      "gossip_bytes",
				Help:      "amount of gossip (bytes)",
			},
			ioTypeLabels,
		),
		tracking: metric.NewGaugeVec(
			metric.GaugeOpts{
				Namespace: namespace,
				Name:      "gossip_tracking",
				Help:      "number of gossipables being tracked",
			},
			typeLabels,
		),
		trackingLifetimeAverage: metric.NewGauge(metric.GaugeOpts{
			Namespace: namespace,
			Name:      "gossip_tracking_lifetime_average",
			Help:      "average duration a gossipable has been tracked (ns)",
		}),
		topValidators: metric.NewGaugeVec(
			metric.GaugeOpts{
				Namespace: namespace,
				Name:      "top_validators",
				Help:      "number of validators gossipables are sent to due to stake",
			},
			typeLabels,
		),
	}
	err := errors.Join()
	return m, err
}

func (m *Metrics) observeMessage(labels map[string]string, count int, bytes int) {
	countMetric := m.count.With(labels)
	bytesMetric := m.bytes.With(labels)

	countMetric.Add(float64(count))
	bytesMetric.Add(float64(bytes))
}

func (v ValidatorGossiper) Gossip(ctx context.Context) error {
	if !v.Validators.Has(ctx, v.NodeID) {
		return nil
	}

	return v.Gossiper.Gossip(ctx)
}

func NewPullGossiper[T Gossipable](
	log log.Logger,
	marshaller Marshaller[T],
	set Set[T],
	client *p2p.Client,
	metrics Metrics,
	pollSize int,
) *PullGossiper[T] {
	return &PullGossiper[T]{
		log:        log,
		marshaller: marshaller,
		set:        set,
		client:     client,
		metrics:    metrics,
		pollSize:   pollSize,
	}
}

type PullGossiper[T Gossipable] struct {
	log        log.Logger
	marshaller Marshaller[T]
	set        Set[T]
	client     *p2p.Client
	metrics    Metrics
	pollSize   int
}

func (p *PullGossiper[_]) Gossip(ctx context.Context) error {
	msgBytes, err := MarshalAppRequest(p.set.GetFilter())
	if err != nil {
		return err
	}

	for i := 0; i < p.pollSize; i++ {
		err := p.client.RequestAny(ctx, msgBytes, p.handleResponse)
		if err != nil && !errors.Is(err, p2p.ErrNoPeers) {
			return err
		}
	}

	return nil
}

func (p *PullGossiper[_]) handleResponse(
	_ context.Context,
	nodeID ids.NodeID,
	responseBytes []byte,
	err error,
) {
	if err != nil {
		p.log.Debug(
			"failed gossip request",
			luxlog.Stringer("nodeID", nodeID),
			luxlog.Reflect("error", err),
		)
		return
	}

	gossip, err := ParseAppResponse(responseBytes)
	if err != nil {
		p.log.Debug("failed to unmarshal gossip response", luxlog.Reflect("error", err))
		return
	}

	receivedBytes := 0
	for _, bytes := range gossip {
		receivedBytes += len(bytes)

		gossipable, err := p.marshaller.UnmarshalGossip(bytes)
		if err != nil {
			p.log.Debug(
				"failed to unmarshal gossip",
				luxlog.Stringer("nodeID", nodeID),
				luxlog.Reflect("error", err),
			)
			continue
		}

		gossipID := gossipable.GossipID()
		p.log.Debug(
			"received gossip",
			luxlog.Stringer("nodeID", nodeID),
			luxlog.Stringer("id", gossipID),
		)
		if err := p.set.Add(gossipable); err != nil {
			p.log.Debug(
				"failed to add gossip to the known set",
				luxlog.Stringer("nodeID", nodeID),
				luxlog.Stringer("id", gossipID),
				luxlog.Reflect("error", err),
			)
			continue
		}
	}

	p.metrics.observeMessage(receivedPullLabels, len(gossip), receivedBytes)
}

var receivedPullLabels = map[string]string{
	ioLabel:   receivedIO,
	typeLabel: pullType,
}

var sentPushLabels = map[string]string{
	ioLabel:   sentIO,
	typeLabel: pushType,
}

var unsentLabels = map[string]string{
	typeLabel: unsentType,
}

var sentLabels = map[string]string{
	typeLabel: sentType,
}

// Deque interface for buffer operations (internal use)
type Deque[T any] interface {
	PushRight(T) bool
	PushLeft(T) bool
	PopLeft() (T, bool)
	Len() int
}

// Cacher interface for cache operations (internal use)
type Cacher[K comparable, V any] interface {
	Get(key K) (V, bool)
	Put(key K, value V)
}

// EmptyCache is a no-op cache
type EmptyCache[K comparable, V any] struct{}

func (EmptyCache[K, V]) Get(key K) (V, bool) {
	var zero V
	return zero, false
}

func (EmptyCache[K, V]) Put(key K, value V) {}

// unboundedDeque is an unbounded double-ended queue implementation
type unboundedDeque[T any] struct {
	size, left, right int
	data              []T
}

// newUnboundedDeque creates a new unbounded deque with the given initial size
func newUnboundedDeque[T any](initSize int) *unboundedDeque[T] {
	if initSize < 2 {
		initSize = defaultInitSize
	}
	return &unboundedDeque[T]{
		data:  make([]T, initSize),
		right: 1,
	}
}

func (b *unboundedDeque[T]) PushRight(elt T) bool {
	b.data[b.right] = elt
	b.size++
	b.right++
	b.right %= len(b.data)
	b.resize()
	return true
}

func (b *unboundedDeque[T]) PushLeft(elt T) bool {
	b.data[b.left] = elt
	b.size++
	b.left--
	if b.left < 0 {
		b.left = len(b.data) - 1
	}
	b.resize()
	return true
}

func (b *unboundedDeque[T]) PopLeft() (T, bool) {
	if b.size == 0 {
		var zero T
		return zero, false
	}
	idx := b.leftmostEltIdx()
	elt := b.data[idx]
	var zero T
	b.data[idx] = zero
	b.size--
	b.left++
	b.left %= len(b.data)
	return elt, true
}

func (b *unboundedDeque[T]) Len() int {
	return b.size
}

func (b *unboundedDeque[T]) leftmostEltIdx() int {
	if b.left == len(b.data)-1 {
		return 0
	}
	return b.left + 1
}

func (b *unboundedDeque[T]) resize() {
	if b.size != len(b.data) {
		return
	}
	newData := make([]T, b.size*2)
	leftmostIdx := b.leftmostEltIdx()
	copy(newData, b.data[leftmostIdx:])
	numCopied := len(b.data) - leftmostIdx
	copy(newData[numCopied:], b.data[:b.right])
	b.data = newData
	b.left = len(b.data) - 1
	b.right = b.size
}

// lruCache is a simple LRU cache implementation
type lruCache[K comparable, V any] struct {
	maxSize int
	data    map[K]V
	order   []K
}

// newLRUCache creates a new LRU cache with the given maximum size
func newLRUCache[K comparable, V any](maxSize int) *lruCache[K, V] {
	if maxSize < 1 {
		maxSize = 1
	}
	return &lruCache[K, V]{
		maxSize: maxSize,
		data:    make(map[K]V),
		order:   make([]K, 0, maxSize),
	}
}

func (c *lruCache[K, V]) Get(key K) (V, bool) {
	val, ok := c.data[key]
	if ok {
		// Move to end (most recently used)
		c.moveToEnd(key)
	}
	return val, ok
}

func (c *lruCache[K, V]) Put(key K, value V) {
	if _, ok := c.data[key]; ok {
		c.data[key] = value
		c.moveToEnd(key)
		return
	}

	// Evict oldest if at capacity
	if len(c.data) >= c.maxSize && len(c.order) > 0 {
		oldest := c.order[0]
		delete(c.data, oldest)
		c.order = c.order[1:]
	}

	c.data[key] = value
	c.order = append(c.order, key)
}

func (c *lruCache[K, V]) moveToEnd(key K) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, key)
			return
		}
	}
}

// NewPushGossiper returns an instance of PushGossiper
func NewPushGossiper[T Gossipable](
	marshaller Marshaller[T],
	mempool Set[T],
	validators p2p.ValidatorSubset,
	client *p2p.Client,
	metrics Metrics,
	gossipParams BranchingFactor,
	regossipParams BranchingFactor,
	discardedSize int,
	targetGossipSize int,
	maxRegossipFrequency time.Duration,
) (*PushGossiper[T], error) {
	if err := gossipParams.Verify(); err != nil {
		return nil, fmt.Errorf("invalid gossip params: %w", err)
	}
	if err := regossipParams.Verify(); err != nil {
		return nil, fmt.Errorf("invalid regossip params: %w", err)
	}
	switch {
	case discardedSize < 0:
		return nil, ErrInvalidDiscardedSize
	case targetGossipSize < 0:
		return nil, ErrInvalidTargetGossipSize
	case maxRegossipFrequency < 0:
		return nil, ErrInvalidRegossipFrequency
	}

	return &PushGossiper[T]{
		marshaller:           marshaller,
		set:                  mempool,
		validators:           validators,
		client:               client,
		metrics:              metrics,
		gossipParams:         gossipParams,
		regossipParams:       regossipParams,
		targetGossipSize:     targetGossipSize,
		maxRegossipFrequency: maxRegossipFrequency,

		tracking:   make(map[ids.ID]*tracking),
		toGossip:   newUnboundedDeque[T](0),
		toRegossip: newUnboundedDeque[T](0),
		discarded:  newLRUCache[ids.ID, struct{}](discardedSize),
	}, nil
}

// PushGossiper broadcasts gossip to peers randomly in the network
type PushGossiper[T Gossipable] struct {
	marshaller Marshaller[T]
	set        Set[T]
	validators p2p.ValidatorSubset
	client     *p2p.Client
	metrics    Metrics

	gossipParams         BranchingFactor
	regossipParams       BranchingFactor
	targetGossipSize     int
	maxRegossipFrequency time.Duration

	lock         sync.Mutex
	tracking     map[ids.ID]*tracking
	addedTimeSum float64 // unix nanoseconds
	toGossip     Deque[T]
	toRegossip   Deque[T]
	discarded    Cacher[ids.ID, struct{}]
}

type BranchingFactor struct {
	// StakePercentage determines the percentage of stake that should have
	// gossip sent to based on the inverse CDF of stake weights. This value does
	// not account for the connectivity of the nodes.
	StakePercentage float64
	// Validators specifies the number of connected validators, in addition to
	// any validators sent from the StakePercentage parameter, to send gossip
	// to. These validators are sampled uniformly rather than by stake.
	Validators int
	// NonValidators specifies the number of connected non-validators to send
	// gossip to.
	NonValidators int
	// Peers specifies the number of connected validators or non-validators, in
	// addition to the number sent due to other configs, to send gossip to.
	Peers int
}

func (b *BranchingFactor) Verify() error {
	switch {
	case b.Validators < 0:
		return ErrInvalidNumValidators
	case b.NonValidators < 0:
		return ErrInvalidNumNonValidators
	case b.Peers < 0:
		return ErrInvalidNumPeers
	case max(b.Validators, b.NonValidators, b.Peers) == 0:
		return ErrInvalidNumToGossip
	default:
		return nil
	}
}

type tracking struct {
	addedTime    float64 // unix nanoseconds
	lastGossiped time.Time
}

// Gossip flushes any queued gossipables.
func (p *PushGossiper[T]) Gossip(ctx context.Context) error {
	var (
		now         = time.Now()
		nowUnixNano = float64(now.UnixNano())
	)

	p.lock.Lock()
	defer func() {
		p.updateMetrics(nowUnixNano)
		p.lock.Unlock()
	}()

	if len(p.tracking) == 0 {
		return nil
	}

	if err := p.gossip(
		ctx,
		now,
		p.gossipParams,
		p.toGossip,
		p.toRegossip,
		EmptyCache[ids.ID, struct{}]{}, // Don't mark dropped unsent transactions as discarded
		unsentLabels,
	); err != nil {
		return fmt.Errorf("unexpected error during gossip: %w", err)
	}

	if err := p.gossip(
		ctx,
		now,
		p.regossipParams,
		p.toRegossip,
		p.toRegossip,
		p.discarded, // Mark dropped sent transactions as discarded
		sentLabels,
	); err != nil {
		return fmt.Errorf("unexpected error during regossip: %w", err)
	}
	return nil
}

func (p *PushGossiper[T]) gossip(
	ctx context.Context,
	now time.Time,
	gossipParams BranchingFactor,
	toGossip Deque[T],
	toRegossip Deque[T],
	discarded Cacher[ids.ID, struct{}],
	metricsLabels map[string]string,
) error {
	var (
		sentBytes                   = 0
		gossip                      = make([][]byte, 0, defaultGossipableCount)
		maxLastGossipTimeToRegossip = now.Add(-p.maxRegossipFrequency)
	)

	for sentBytes < p.targetGossipSize {
		gossipable, ok := toGossip.PopLeft()
		if !ok {
			break
		}

		// Ensure item is still in the set before we gossip.
		gossipID := gossipable.GossipID()
		tracking := p.tracking[gossipID]
		if !p.set.Has(gossipID) {
			delete(p.tracking, gossipID)
			p.addedTimeSum -= tracking.addedTime
			discarded.Put(gossipID, struct{}{}) // Cache that the item was dropped
			continue
		}

		// Ensure we don't attempt to send a gossipable too frequently.
		if maxLastGossipTimeToRegossip.Before(tracking.lastGossiped) {
			// Put the gossipable on the front of the queue to keep items sorted
			// by last issuance time.
			toGossip.PushLeft(gossipable)
			break
		}

		bytes, err := p.marshaller.MarshalGossip(gossipable)
		if err != nil {
			delete(p.tracking, gossipID)
			p.addedTimeSum -= tracking.addedTime
			return err
		}

		gossip = append(gossip, bytes)
		sentBytes += len(bytes)
		toRegossip.PushRight(gossipable)
		tracking.lastGossiped = now
	}

	// If there is nothing to gossip, we can exit early.
	if len(gossip) == 0 {
		return nil
	}

	// Send gossipables to peers
	msgBytes, err := MarshalAppGossip(gossip)
	if err != nil {
		return err
	}

	p.metrics.observeMessage(sentPushLabels, len(gossip), sentBytes)

	topValidatorsMetric := p.metrics.topValidators.With(metricsLabels)

	validatorsByStake := p.validators.Top(ctx, gossipParams.StakePercentage)
	topValidatorsMetric.Set(float64(len(validatorsByStake)))

	// Convert []ids.NodeID to set for SendConfig
	nodeIDSet := set.NewSet[ids.NodeID](len(validatorsByStake))
	for _, nodeID := range validatorsByStake {
		nodeIDSet.Add(nodeID)
	}

	return p.client.Gossip(
		ctx,
		p2p.SendConfig{
			NodeIDs:       nodeIDSet,
			Validators:    gossipParams.Validators,
			NonValidators: gossipParams.NonValidators,
			Peers:         gossipParams.Peers,
		},
		msgBytes,
	)
}

// Add enqueues new gossipables to be pushed. If a gossipable is already tracked,
// it is not added again.
func (p *PushGossiper[T]) Add(gossipables ...T) {
	var (
		now         = time.Now()
		nowUnixNano = float64(now.UnixNano())
	)

	p.lock.Lock()
	defer func() {
		p.updateMetrics(nowUnixNano)
		p.lock.Unlock()
	}()

	// Add new gossipables to be sent.
	for _, gossipable := range gossipables {
		gossipID := gossipable.GossipID()
		if _, ok := p.tracking[gossipID]; ok {
			continue
		}

		tracking := &tracking{
			addedTime: nowUnixNano,
		}
		if _, ok := p.discarded.Get(gossipID); ok {
			// Pretend that recently discarded transactions were just gossiped.
			tracking.lastGossiped = now
			p.toRegossip.PushRight(gossipable)
		} else {
			p.toGossip.PushRight(gossipable)
		}
		p.tracking[gossipID] = tracking
		p.addedTimeSum += nowUnixNano
	}
}

func (p *PushGossiper[_]) updateMetrics(nowUnixNano float64) {
	var (
		numUnsent       = float64(p.toGossip.Len())
		numSent         = float64(p.toRegossip.Len())
		numTracking     = numUnsent + numSent
		averageLifetime float64
	)
	if numTracking != 0 {
		averageLifetime = nowUnixNano - p.addedTimeSum/numTracking
	}

	p.metrics.tracking.With(unsentLabels).Set(numUnsent)
	p.metrics.tracking.With(sentLabels).Set(numSent)
	p.metrics.trackingLifetimeAverage.Set(averageLifetime)
}

// Every calls [Gossip] every [frequency] amount of time.
func Every(ctx context.Context, log log.Logger, gossiper Gossiper, frequency time.Duration) {
	ticker := time.NewTicker(frequency)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := gossiper.Gossip(ctx); err != nil {
				log.Warn("failed to gossip",
					luxlog.Reflect("error", err),
				)
			}
		case <-ctx.Done():
			log.Debug("shutting down gossip")
			return
		}
	}
}

type NoOpGossiper struct{}

func (NoOpGossiper) Gossip(context.Context) error {
	return nil
}

type TestGossiper struct {
	GossipF func(ctx context.Context) error
}

func (t *TestGossiper) Gossip(ctx context.Context) error {
	return t.GossipF(ctx)
}

type FullSet[T Gossipable] struct{}

func (FullSet[_]) Gossip(context.Context) error {
	return nil
}

func (FullSet[T]) Add(T) error {
	return nil
}

func (FullSet[T]) Has(ids.ID) bool {
	return true
}

func (FullSet[T]) Iterate(func(gossipable T) bool) {}

func (FullSet[_]) GetFilter() ([]byte, []byte) {
	return nil, ids.Empty[:]
}
