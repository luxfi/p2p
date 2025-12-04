// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package gossip

import (
	"encoding/binary"
	"fmt"

	"github.com/luxfi/ids"
)

// MarshalAppRequest marshals a pull gossip request to bytes
func MarshalAppRequest(filter, salt []byte) ([]byte, error) {
	// Format: filterLen(4) + filter + saltLen(4) + salt
	filterLen := len(filter)
	saltLen := len(salt)
	buf := make([]byte, 4+filterLen+4+saltLen)
	binary.BigEndian.PutUint32(buf[0:4], uint32(filterLen))
	copy(buf[4:4+filterLen], filter)
	binary.BigEndian.PutUint32(buf[4+filterLen:8+filterLen], uint32(saltLen))
	copy(buf[8+filterLen:], salt)
	return buf, nil
}

// ParseAppRequest parses bytes to a pull gossip request
func ParseAppRequest(bytes []byte) ([]byte, ids.ID, error) {
	if len(bytes) < 8 {
		return nil, ids.Empty, fmt.Errorf("data too short: %d", len(bytes))
	}
	filterLen := binary.BigEndian.Uint32(bytes[0:4])
	if len(bytes) < int(4+filterLen+4) {
		return nil, ids.Empty, fmt.Errorf("data too short for filter: %d", len(bytes))
	}
	saltLen := binary.BigEndian.Uint32(bytes[4+filterLen : 8+filterLen])
	if len(bytes) < int(8+filterLen+saltLen) {
		return nil, ids.Empty, fmt.Errorf("data too short for salt: %d", len(bytes))
	}

	filter := bytes[4 : 4+filterLen]
	saltBytes := bytes[8+filterLen : 8+filterLen+saltLen]

	salt, err := ids.ToID(saltBytes)
	if err != nil {
		return nil, ids.Empty, err
	}

	return filter, salt, nil
}

// MarshalAppResponse marshals gossip items to bytes
func MarshalAppResponse(gossip [][]byte) ([]byte, error) {
	// Format: count(4) + [len(4) + data]...
	totalLen := 4
	for _, g := range gossip {
		totalLen += 4 + len(g)
	}

	buf := make([]byte, totalLen)
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(gossip)))
	offset := 4
	for _, g := range gossip {
		binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(len(g)))
		copy(buf[offset+4:offset+4+len(g)], g)
		offset += 4 + len(g)
	}
	return buf, nil
}

// ParseAppResponse parses bytes to gossip items
func ParseAppResponse(bytes []byte) ([][]byte, error) {
	if len(bytes) < 4 {
		return nil, fmt.Errorf("data too short: %d", len(bytes))
	}
	count := binary.BigEndian.Uint32(bytes[0:4])
	gossip := make([][]byte, 0, count)
	offset := 4
	for i := uint32(0); i < count; i++ {
		if len(bytes) < offset+4 {
			return nil, fmt.Errorf("data too short for item %d length: %d", i, len(bytes))
		}
		itemLen := binary.BigEndian.Uint32(bytes[offset : offset+4])
		if len(bytes) < offset+4+int(itemLen) {
			return nil, fmt.Errorf("data too short for item %d: %d", i, len(bytes))
		}
		gossip = append(gossip, bytes[offset+4:offset+4+int(itemLen)])
		offset += 4 + int(itemLen)
	}
	return gossip, nil
}

// MarshalAppGossip marshals push gossip to bytes
func MarshalAppGossip(gossip [][]byte) ([]byte, error) {
	return MarshalAppResponse(gossip) // Same format
}

// ParseAppGossip parses bytes to push gossip
func ParseAppGossip(bytes []byte) ([][]byte, error) {
	return ParseAppResponse(bytes) // Same format
}
