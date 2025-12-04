// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package lp118 implements LP-118 warp message signature handling
package lp118

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/luxfi/ids"
	"github.com/luxfi/warp"
)

// HandlerID is the protocol ID for LP-118
const HandlerID = 0x12345678

// Handler handles LP-118 messages
type Handler interface {
	// AppRequest handles an incoming request
	AppRequest(ctx context.Context, nodeID ids.NodeID, deadline time.Time, request []byte) ([]byte, error)
}

// Signer signs warp messages and returns the signature bytes
type Signer interface {
	// Sign signs an unsigned warp message and returns the signature bytes
	Sign(msg *warp.UnsignedMessage) ([]byte, error)
}

// NoOpHandler is a no-op implementation of Handler
type NoOpHandler struct{}

// AppRequest returns an empty response
func (NoOpHandler) AppRequest(context.Context, ids.NodeID, time.Time, []byte) ([]byte, error) {
	return nil, nil
}

// Cacher provides caching for signature responses
type Cacher[K comparable, V any] interface {
	Get(key K) (V, bool)
	Put(key K, value V)
}

// CachedHandler implements a cached handler for LP-118
type CachedHandler struct {
	cache   Cacher[ids.ID, []byte]
	backend interface{}
	signer  Signer
}

// NewCachedHandler creates a new cached handler
func NewCachedHandler(cache Cacher[ids.ID, []byte], backend interface{}, signer Signer) Handler {
	return &CachedHandler{
		cache:   cache,
		backend: backend,
		signer:  signer,
	}
}

// AppRequest handles an incoming request with caching
func (h *CachedHandler) AppRequest(ctx context.Context, nodeID ids.NodeID, deadline time.Time, request []byte) ([]byte, error) {
	req, err := UnmarshalSignatureRequest(request)
	if err != nil {
		return nil, err
	}

	unsignedMessage, err := warp.ParseUnsignedMessage(req.Message)
	if err != nil {
		return nil, err
	}

	// Check cache - convert []byte ID to ids.ID
	idBytes := unsignedMessage.ID()
	var messageID ids.ID
	copy(messageID[:], idBytes)
	if signatureBytes, ok := h.cache.Get(messageID); ok {
		return MarshalSignatureResponse(signatureBytes)
	}

	// Verify if backend is a Verifier
	if verifier, ok := h.backend.(Verifier); ok {
		if appErr := verifier.Verify(ctx, unsignedMessage, req.Justification); appErr != nil {
			return nil, appErr
		}
	}

	// Sign the message
	if h.signer == nil {
		return nil, fmt.Errorf("signer is nil")
	}
	signatureBytes, err := h.signer.Sign(unsignedMessage)
	if err != nil {
		return nil, err
	}

	h.cache.Put(messageID, signatureBytes)

	return MarshalSignatureResponse(signatureBytes)
}

// SignatureRequest represents a request for a signature
type SignatureRequest struct {
	Message       []byte
	Justification []byte
}

// SignatureResponse represents a signature response
type SignatureResponse struct {
	Signature []byte
}

// MarshalSignatureRequest marshals a signature request to bytes
func MarshalSignatureRequest(req *SignatureRequest) ([]byte, error) {
	// Format: msgLen(4) + msg + justLen(4) + just
	msgLen := len(req.Message)
	justLen := len(req.Justification)
	buf := make([]byte, 4+msgLen+4+justLen)
	binary.BigEndian.PutUint32(buf[0:4], uint32(msgLen))
	copy(buf[4:4+msgLen], req.Message)
	binary.BigEndian.PutUint32(buf[4+msgLen:8+msgLen], uint32(justLen))
	copy(buf[8+msgLen:], req.Justification)
	return buf, nil
}

// UnmarshalSignatureRequest unmarshals bytes to a signature request
func UnmarshalSignatureRequest(data []byte) (*SignatureRequest, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data too short: %d", len(data))
	}
	msgLen := binary.BigEndian.Uint32(data[0:4])
	if len(data) < int(4+msgLen+4) {
		return nil, fmt.Errorf("data too short for message: %d", len(data))
	}
	justLen := binary.BigEndian.Uint32(data[4+msgLen : 8+msgLen])
	if len(data) < int(8+msgLen+justLen) {
		return nil, fmt.Errorf("data too short for justification: %d", len(data))
	}
	return &SignatureRequest{
		Message:       data[4 : 4+msgLen],
		Justification: data[8+msgLen : 8+msgLen+justLen],
	}, nil
}

// MarshalSignatureResponse marshals a signature response to bytes
func MarshalSignatureResponse(signature []byte) ([]byte, error) {
	// Format: sigLen(4) + sig
	buf := make([]byte, 4+len(signature))
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(signature)))
	copy(buf[4:], signature)
	return buf, nil
}

// UnmarshalSignatureResponse unmarshals bytes to a signature response
func UnmarshalSignatureResponse(data []byte) (*SignatureResponse, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short: %d", len(data))
	}
	sigLen := binary.BigEndian.Uint32(data[0:4])
	if len(data) < int(4+sigLen) {
		return nil, fmt.Errorf("data too short for signature: %d", len(data))
	}
	return &SignatureResponse{
		Signature: data[4 : 4+sigLen],
	}, nil
}
