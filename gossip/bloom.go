// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package gossip

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"sync"

	"github.com/luxfi/ids"
	"github.com/luxfi/metric"
)

const (
	minHashes      = 1
	maxHashes      = 16
	minEntries     = 1
	bitsPerByte    = 8
	bytesPerUint64 = 8
	hashRotation   = 17
	ln2Squared     = math.Ln2 * math.Ln2
)

var (
	errTooFewHashes  = errors.New("too few hashes")
	errTooManyHashes = errors.New("too many hashes")
	errTooFewEntries = errors.New("too few entries")
)

// BloomFilter is a probabilistic data structure for membership testing.
// It is safe for concurrent use after construction.
type BloomFilter struct {
	minTargetElements              int
	targetFalsePositiveProbability float64
	resetFalsePositiveProbability  float64

	metrics *bloomMetrics

	lock     sync.RWMutex
	maxCount int
	filter   *filter
	salt     ids.ID
}

// bloomMetrics tracks bloom filter statistics
type bloomMetrics struct {
	Count      metric.Gauge
	NumHashes  metric.Gauge
	NumEntries metric.Gauge
	MaxCount   metric.Gauge
	ResetCount metric.Counter
}

// filter is the internal bloom filter implementation
type filter struct {
	numBits   uint64
	hashSeeds []uint64
	entries   []byte
	count     int
}

// NewBloomFilter creates a new bloom filter for gossip membership tracking.
//
// Parameters:
//   - registerer: metrics registerer for tracking filter statistics
//   - namespace: metrics namespace prefix
//   - minTargetElements: minimum expected number of elements
//   - targetFalsePositiveProbability: desired false positive rate during normal operation
//   - resetFalsePositiveProbability: threshold at which the filter should be reset
//
// The filter is safe for concurrent reads after construction, but must not be
// reset concurrently with other operations.
func NewBloomFilter(
	registerer metric.Registerer,
	namespace string,
	minTargetElements int,
	targetFalsePositiveProbability,
	resetFalsePositiveProbability float64,
) (*BloomFilter, error) {
	metrics, err := newBloomMetrics(registerer, namespace)
	if err != nil {
		return nil, err
	}

	bf := &BloomFilter{
		minTargetElements:              minTargetElements,
		targetFalsePositiveProbability: targetFalsePositiveProbability,
		resetFalsePositiveProbability:  resetFalsePositiveProbability,
		metrics:                        metrics,
	}

	err = resetBloomFilter(
		bf,
		minTargetElements,
		targetFalsePositiveProbability,
		resetFalsePositiveProbability,
	)
	return bf, err
}

func newBloomMetrics(registerer metric.Registerer, namespace string) (*bloomMetrics, error) {
	registry, ok := registerer.(metric.Registry)
	if !ok {
		// Create metrics without registry if not available
		return &bloomMetrics{
			Count:      metric.NewGauge(metric.GaugeOpts{Namespace: namespace, Name: "bloom_count"}),
			NumHashes:  metric.NewGauge(metric.GaugeOpts{Namespace: namespace, Name: "bloom_hashes"}),
			NumEntries: metric.NewGauge(metric.GaugeOpts{Namespace: namespace, Name: "bloom_entries"}),
			MaxCount:   metric.NewGauge(metric.GaugeOpts{Namespace: namespace, Name: "bloom_max_count"}),
			ResetCount: metric.NewCounter(metric.CounterOpts{Namespace: namespace, Name: "bloom_reset_count"}),
		}, nil
	}

	metricsInstance := metric.NewWithRegistry(namespace, registry)
	return &bloomMetrics{
		Count:      metricsInstance.NewGauge("bloom_count", "Number of additions to the bloom filter"),
		NumHashes:  metricsInstance.NewGauge("bloom_hashes", "Number of hash functions"),
		NumEntries: metricsInstance.NewGauge("bloom_entries", "Number of bytes in the bloom filter"),
		MaxCount:   metricsInstance.NewGauge("bloom_max_count", "Maximum additions before reset"),
		ResetCount: metricsInstance.NewCounter("bloom_reset_count", "Number of filter resets"),
	}, nil
}

// Add adds a gossipable item to the bloom filter.
func (b *BloomFilter) Add(gossipable Gossipable) {
	h := gossipable.GossipID()

	b.lock.Lock()
	defer b.lock.Unlock()

	add(b.filter, h[:], b.salt[:])
	b.metrics.Count.Inc()
}

// Has returns true if the gossipable item may be in the bloom filter.
// False positives are possible; false negatives are not.
func (b *BloomFilter) Has(gossipable Gossipable) bool {
	h := gossipable.GossipID()

	b.lock.RLock()
	defer b.lock.RUnlock()

	return contains(b.filter, h[:], b.salt[:])
}

// Salt returns the current salt value used for hashing.
func (b *BloomFilter) Salt() ids.ID {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.salt
}

// Marshal returns the serialized bloom filter and salt.
func (b *BloomFilter) Marshal() ([]byte, []byte) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	bloomBytes := marshal(b.filter)
	salt := b.salt
	return bloomBytes, salt[:]
}

// ResetBloomFilterIfNeeded resets the bloom filter if it has exceeded the
// maximum allowed count. The targetElements parameter allows the filter to
// grow if the expected number of elements has increased.
//
// Returns true if the filter was reset, false otherwise.
func ResetBloomFilterIfNeeded(
	bloomFilter *BloomFilter,
	targetElements int,
) (bool, error) {
	bloomFilter.lock.Lock()
	defer bloomFilter.lock.Unlock()

	if bloomFilter.filter.count <= bloomFilter.maxCount {
		return false, nil
	}

	targetElements = max(bloomFilter.minTargetElements, targetElements)
	err := resetBloomFilter(
		bloomFilter,
		targetElements,
		bloomFilter.targetFalsePositiveProbability,
		bloomFilter.resetFalsePositiveProbability,
	)
	return err == nil, err
}

func resetBloomFilter(
	bf *BloomFilter,
	targetElements int,
	targetFalsePositiveProbability,
	resetFalsePositiveProbability float64,
) error {
	numHashes, numEntries := optimalParameters(targetElements, targetFalsePositiveProbability)

	newFilter, err := newFilter(numHashes, numEntries)
	if err != nil {
		return err
	}

	var newSalt ids.ID
	if _, err := rand.Read(newSalt[:]); err != nil {
		return err
	}

	maxCount := estimateCount(numHashes, numEntries, resetFalsePositiveProbability)

	bf.maxCount = maxCount
	bf.filter = newFilter
	bf.salt = newSalt

	if bf.metrics != nil {
		bf.metrics.Count.Set(0)
		bf.metrics.NumHashes.Set(float64(numHashes))
		bf.metrics.NumEntries.Set(float64(numEntries))
		bf.metrics.MaxCount.Set(float64(maxCount))
		bf.metrics.ResetCount.Inc()
	}

	return nil
}

// Internal filter implementation

func newFilter(numHashes, numEntries int) (*filter, error) {
	if numEntries < minEntries {
		return nil, errTooFewEntries
	}

	hashSeeds, err := newHashSeeds(numHashes)
	if err != nil {
		return nil, err
	}

	return &filter{
		numBits:   uint64(numEntries * bitsPerByte),
		hashSeeds: hashSeeds,
		entries:   make([]byte, numEntries),
		count:     0,
	}, nil
}

func newHashSeeds(count int) ([]uint64, error) {
	switch {
	case count < minHashes:
		return nil, fmt.Errorf("%w: %d < %d", errTooFewHashes, count, minHashes)
	case count > maxHashes:
		return nil, fmt.Errorf("%w: %d > %d", errTooManyHashes, count, maxHashes)
	}

	bytes := make([]byte, count*bytesPerUint64)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	seeds := make([]uint64, count)
	for i := range seeds {
		seeds[i] = binary.BigEndian.Uint64(bytes[i*bytesPerUint64:])
	}
	return seeds, nil
}

func hash(key, salt []byte) uint64 {
	h := sha256.New()
	_, _ = h.Write(key)
	_, _ = h.Write(salt)
	output := make([]byte, 0, sha256.Size)
	return binary.BigEndian.Uint64(h.Sum(output))
}

func add(f *filter, key, salt []byte) {
	h := hash(key, salt)

	_ = 1 % f.numBits // hint to compiler that numBits is not 0
	for _, seed := range f.hashSeeds {
		h = bits.RotateLeft64(h, hashRotation) ^ seed
		index := h % f.numBits
		byteIndex := index / bitsPerByte
		bitIndex := index % bitsPerByte
		f.entries[byteIndex] |= 1 << bitIndex
	}
	f.count++
}

func contains(f *filter, key, salt []byte) bool {
	h := hash(key, salt)

	numBits := bitsPerByte * uint64(len(f.entries))
	_ = 1 % numBits // hint to compiler that numBits is not 0

	var accumulator byte = 1
	for seedIndex := 0; seedIndex < len(f.hashSeeds) && accumulator != 0; seedIndex++ {
		h = bits.RotateLeft64(h, hashRotation) ^ f.hashSeeds[seedIndex]
		index := h % numBits
		byteIndex := index / bitsPerByte
		bitIndex := index % bitsPerByte
		accumulator &= f.entries[byteIndex] >> bitIndex
	}
	return accumulator != 0
}

func marshal(f *filter) []byte {
	numHashes := len(f.hashSeeds)
	entriesOffset := 1 + numHashes*bytesPerUint64

	bytes := make([]byte, entriesOffset+len(f.entries))
	bytes[0] = byte(numHashes)
	for i, seed := range f.hashSeeds {
		binary.BigEndian.PutUint64(bytes[1+i*bytesPerUint64:], seed)
	}
	copy(bytes[entriesOffset:], f.entries)
	return bytes
}

// Optimal parameter calculations

func optimalParameters(count int, falsePositiveProbability float64) (int, int) {
	numEntries := optimalEntries(count, falsePositiveProbability)
	numHashes := optimalHashes(numEntries, count)
	return numHashes, numEntries
}

func optimalHashes(numEntries, count int) int {
	switch {
	case numEntries < minEntries:
		return minHashes
	case count <= 0:
		return maxHashes
	}

	numHashes := math.Ceil(float64(numEntries) * bitsPerByte * math.Ln2 / float64(count))
	if numHashes >= maxHashes {
		return maxHashes
	}
	return max(int(numHashes), minHashes)
}

func optimalEntries(count int, falsePositiveProbability float64) int {
	switch {
	case count <= 0:
		return minEntries
	case falsePositiveProbability >= 1:
		return minEntries
	case falsePositiveProbability <= 0:
		return math.MaxInt
	}

	entriesInBits := -float64(count) * math.Log(falsePositiveProbability) / ln2Squared
	entries := (entriesInBits + bitsPerByte - 1) / bitsPerByte
	if entries >= math.MaxInt {
		return math.MaxInt
	}
	return max(int(entries), minEntries)
}

func estimateCount(numHashes, numEntries int, falsePositiveProbability float64) int {
	switch {
	case numHashes < minHashes:
		return 0
	case numEntries < minEntries:
		return 0
	case falsePositiveProbability <= 0:
		return 0
	case falsePositiveProbability >= 1:
		return math.MaxInt
	}

	invNumHashes := 1 / float64(numHashes)
	numBits := float64(numEntries * 8)
	exp := 1 - math.Pow(falsePositiveProbability, invNumHashes)
	count := math.Ceil(-math.Log(exp) * numBits * invNumHashes)
	if count >= math.MaxInt {
		return math.MaxInt
	}
	return int(count)
}
