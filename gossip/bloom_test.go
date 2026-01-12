// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package gossip

import (
	"testing"

	"github.com/luxfi/ids"
	"github.com/luxfi/metric"
	"github.com/stretchr/testify/require"
)

// testGossipable implements Gossipable for testing
type testGossipable struct {
	id ids.ID
}

func (t *testGossipable) GossipID() ids.ID {
	return t.id
}

func TestBloomFilterBasic(t *testing.T) {
	require := require.New(t)

	// Create bloom filter with reasonable parameters
	bf, err := NewBloomFilter(
		metric.NewRegistry(),
		"test",
		100,   // minTargetElements
		0.01,  // targetFalsePositiveRate
		0.001, // resetFalsePositiveRate
	)
	require.NoError(err)
	require.NotNil(bf)

	// Create test gossipables
	gossipable1 := &testGossipable{id: ids.GenerateTestID()}
	gossipable2 := &testGossipable{id: ids.GenerateTestID()}
	gossipable3 := &testGossipable{id: ids.GenerateTestID()}

	// Items should not be in filter initially
	require.False(bf.Has(gossipable1))
	require.False(bf.Has(gossipable2))
	require.False(bf.Has(gossipable3))

	// Add first item
	bf.Add(gossipable1)
	require.True(bf.Has(gossipable1))
	require.False(bf.Has(gossipable2))

	// Add second item
	bf.Add(gossipable2)
	require.True(bf.Has(gossipable1))
	require.True(bf.Has(gossipable2))
	require.False(bf.Has(gossipable3))
}

func TestBloomFilterMarshal(t *testing.T) {
	require := require.New(t)

	bf, err := NewBloomFilter(
		metric.NewRegistry(),
		"test",
		100,
		0.01,
		0.001,
	)
	require.NoError(err)

	// Add some items
	for i := 0; i < 10; i++ {
		bf.Add(&testGossipable{id: ids.GenerateTestID()})
	}

	// Marshal and verify we get data back
	bloomBytes, saltBytes := bf.Marshal()
	require.NotEmpty(bloomBytes)
	require.Len(saltBytes, ids.IDLen)
}

func TestBloomFilterSalt(t *testing.T) {
	require := require.New(t)

	bf, err := NewBloomFilter(
		metric.NewRegistry(),
		"test",
		100,
		0.01,
		0.001,
	)
	require.NoError(err)

	// Salt should not be empty
	salt := bf.Salt()
	require.NotEqual(ids.Empty, salt)
}

func TestResetBloomFilterIfNeeded(t *testing.T) {
	require := require.New(t)

	// Create bloom filter that will reset after just 1 element
	bf, err := NewBloomFilter(
		metric.NewRegistry(),
		"test",
		1,                  // minTargetElements
		0.01,               // targetFalsePositiveRate
		0.0000000000000001, // very low reset threshold
	)
	require.NoError(err)

	initialSalt := bf.Salt()

	// Add enough items to trigger reset threshold
	for i := 0; i < 100; i++ {
		bf.Add(&testGossipable{id: ids.GenerateTestID()})
	}

	// Check if reset is needed
	reset, err := ResetBloomFilterIfNeeded(bf, 100)
	require.NoError(err)
	require.True(reset)

	// Salt should have changed after reset
	newSalt := bf.Salt()
	require.NotEqual(initialSalt, newSalt)
}

func TestBloomFilterNoResetNotNeeded(t *testing.T) {
	require := require.New(t)

	bf, err := NewBloomFilter(
		metric.NewRegistry(),
		"test",
		1000, // large capacity
		0.01,
		0.001,
	)
	require.NoError(err)

	// Add just a few items (well under capacity)
	for i := 0; i < 10; i++ {
		bf.Add(&testGossipable{id: ids.GenerateTestID()})
	}

	// Reset should not be needed
	reset, err := ResetBloomFilterIfNeeded(bf, 100)
	require.NoError(err)
	require.False(reset)
}
