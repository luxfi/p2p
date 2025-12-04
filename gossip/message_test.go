// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package gossip

import (
	"testing"
)

func TestMarshalAppRequest(t *testing.T) {
	filter := []byte{0x01, 0x02, 0x03}
	salt := make([]byte, 32)
	for i := range salt {
		salt[i] = byte(i)
	}

	data, err := MarshalAppRequest(filter, salt)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	decodedFilter, decodedSalt, err := ParseAppRequest(data)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if string(decodedFilter) != string(filter) {
		t.Errorf("filter mismatch")
	}
	for i := range salt {
		if decodedSalt[i] != salt[i] {
			t.Errorf("salt mismatch at position %d", i)
		}
	}
}

func TestMarshalAppResponse(t *testing.T) {
	gossip := [][]byte{
		[]byte("first item"),
		[]byte("second item"),
		[]byte("third item"),
	}

	data, err := MarshalAppResponse(gossip)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	decoded, err := ParseAppResponse(data)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(decoded) != len(gossip) {
		t.Errorf("count mismatch: expected %d, got %d", len(gossip), len(decoded))
	}
	for i := range gossip {
		if string(decoded[i]) != string(gossip[i]) {
			t.Errorf("item %d mismatch: expected %s, got %s", i, gossip[i], decoded[i])
		}
	}
}

func TestMarshalAppGossip(t *testing.T) {
	gossip := [][]byte{
		[]byte("push item 1"),
		[]byte("push item 2"),
	}

	data, err := MarshalAppGossip(gossip)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	decoded, err := ParseAppGossip(data)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(decoded) != len(gossip) {
		t.Errorf("count mismatch: expected %d, got %d", len(gossip), len(decoded))
	}
	for i := range gossip {
		if string(decoded[i]) != string(gossip[i]) {
			t.Errorf("item %d mismatch: expected %s, got %s", i, gossip[i], decoded[i])
		}
	}
}

func TestParseErrors(t *testing.T) {
	// Test empty data
	_, _, err := ParseAppRequest(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}

	_, err = ParseAppResponse(nil)
	if err == nil {
		t.Error("expected error for nil data")
	}

	_, err = ParseAppGossip([]byte{0, 0, 0})
	if err == nil {
		t.Error("expected error for short data")
	}
}
