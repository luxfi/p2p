// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/luxfi/ids"
)

func TestNoOpHandler(t *testing.T) {
	h := NoOpHandler{}
	h.Gossip(context.Background(), ids.EmptyNodeID, nil)
	resp, err := h.Request(context.Background(), ids.EmptyNodeID, time.Now(), nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Errorf("expected nil response, got: %v", resp)
	}
}

func TestThrottler(t *testing.T) {
	throttler := NewSlidingWindowThrottler(time.Second, 5)
	nodeID := ids.GenerateTestNodeID()

	// First 5 requests should pass
	for i := 0; i < 5; i++ {
		if !throttler.Handle(nodeID) {
			t.Errorf("request %d should have passed", i)
		}
	}

	// Next request should be throttled
	if throttler.Handle(nodeID) {
		t.Error("request should have been throttled")
	}
}

func TestPrefixMessage(t *testing.T) {
	prefix := []byte{0x01, 0x02}
	msg := []byte{0x03, 0x04, 0x05}

	result := PrefixMessage(prefix, msg)
	expected := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	if len(result) != len(expected) {
		t.Errorf("expected length %d, got %d", len(expected), len(result))
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("mismatch at position %d: expected %d, got %d", i, expected[i], result[i])
		}
	}
}

func TestParseMessage(t *testing.T) {
	// Handler ID 1 encoded as varint
	msg := append([]byte{0x01}, []byte("hello")...)

	handlerID, payload, ok := ParseMessage(msg)
	if !ok {
		t.Error("expected parsing to succeed")
	}
	if handlerID != 1 {
		t.Errorf("expected handler ID 1, got %d", handlerID)
	}
	if string(payload) != "hello" {
		t.Errorf("expected payload 'hello', got '%s'", string(payload))
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected int32
	}{
		{"ErrUnexpected", ErrUnexpected, -1},
		{"ErrUnregisteredHandler", ErrUnregisteredHandler, -2},
		{"ErrNotValidator", ErrNotValidator, -3},
		{"ErrThrottled", ErrThrottled, -4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.expected {
				t.Errorf("expected code %d, got %d", tt.expected, tt.err.Code)
			}
		})
	}
}
