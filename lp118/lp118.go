// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package lp118 implements the Lux Protocol 118 (LP118) for warp message handling.
// This is the Lux-native equivalent of ACP118.
package lp118

import (
	"context"
	"time"

	"github.com/luxfi/consensus/core"
	"github.com/luxfi/ids"
	"github.com/luxfi/p2p"
	"github.com/luxfi/warp"
)

// HandlerID is the protocol identifier for LP118 warp message handlers.
// This corresponds to SignatureRequestHandlerID in the p2p package.
const HandlerID = p2p.SignatureRequestHandlerID

// Verifier verifies warp messages.
type Verifier interface {
	// Verify verifies the unsigned message with the given justification.
	// Returns nil if verification succeeds, or an AppError if it fails.
	Verify(ctx context.Context, msg *warp.UnsignedMessage, justification []byte) *core.AppError
}

// Signer signs warp messages.
type Signer interface {
	// Sign signs the unsigned message and returns the signature bytes.
	Sign(msg *warp.UnsignedMessage) ([]byte, error)
}

// Handler handles LP118 warp message requests.
type Handler interface {
	// Request handles an incoming application request.
	Request(ctx context.Context, nodeID ids.NodeID, deadline time.Time, requestBytes []byte) ([]byte, *p2p.Error)
	// Gossip handles gossip messages (no-op for LP118).
	Gossip(ctx context.Context, nodeID ids.NodeID, gossipBytes []byte)
}

// Cache provides caching functionality.
type Cache interface {
	Get(key []byte) ([]byte, bool)
	Put(key []byte, value []byte)
}

// CachedHandler is a handler that caches responses.
type CachedHandler struct {
	cache    Cache
	verifier Verifier
	signer   Signer
}

// NewCachedHandler creates a new cached handler.
func NewCachedHandler(cache Cache, verifier Verifier, signer Signer) *CachedHandler {
	return &CachedHandler{
		cache:    cache,
		verifier: verifier,
		signer:   signer,
	}
}

// Request handles an LP118 request.
func (h *CachedHandler) Request(ctx context.Context, nodeID ids.NodeID, deadline time.Time, requestBytes []byte) ([]byte, *p2p.Error) {
	// Check cache first
	if h.cache != nil {
		if cached, ok := h.cache.Get(requestBytes); ok {
			return cached, nil
		}
	}

	// Parse the request to get the unsigned message
	msg, err := warp.ParseUnsignedMessage(requestBytes)
	if err != nil {
		return nil, &p2p.Error{
			Code:    1,
			Message: "failed to parse warp message: " + err.Error(),
		}
	}

	// Verify the message
	if h.verifier != nil {
		if appErr := h.verifier.Verify(ctx, msg, nil); appErr != nil {
			return nil, &p2p.Error{
				Code:    int32(appErr.Code),
				Message: appErr.Message,
			}
		}
	}

	// Sign the message if we have a signer
	var response []byte
	if h.signer != nil {
		sig, err := h.signer.Sign(msg)
		if err != nil {
			return nil, &p2p.Error{
				Code:    2,
				Message: "failed to sign message: " + err.Error(),
			}
		}
		response = sig
	}

	// Cache the response
	if h.cache != nil && len(response) > 0 {
		h.cache.Put(requestBytes, response)
	}

	return response, nil
}

// Gossip handles gossip messages (no-op for LP118).
func (h *CachedHandler) Gossip(ctx context.Context, nodeID ids.NodeID, gossipBytes []byte) {
	// LP118 doesn't use gossip
}

// NewHandlerAdapter creates a p2p.Handler from an LP118 Handler.
// Since LP118 Handler already implements p2p.Handler interface, this is a simple cast.
func NewHandlerAdapter(handler Handler) p2p.Handler {
	return handler
}
