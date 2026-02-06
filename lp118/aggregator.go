// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package lp118

import (
	"context"

	"github.com/luxfi/ids"
	"github.com/luxfi/log"
	"github.com/luxfi/warp"
)

// Client is the interface for making requests to peers.
type Client interface {
	// Request sends a request to a specific peer and returns the response.
	Request(ctx context.Context, nodeID ids.NodeID, request []byte) ([]byte, error)
}

// SignatureAggregator aggregates warp message signatures from validators.
type SignatureAggregator struct {
	log    log.Logger
	client Client
}

// NewSignatureAggregator creates a new SignatureAggregator.
func NewSignatureAggregator(logger log.Logger, client Client) *SignatureAggregator {
	return &SignatureAggregator{
		log:    logger,
		client: client,
	}
}

// AggregateSignatures collects signatures for the given unsigned message from validators.
func (a *SignatureAggregator) AggregateSignatures(
	ctx context.Context,
	msg *warp.UnsignedMessage,
	justification []byte,
	quorumNum uint64,
	quorumDen uint64,
) (*warp.Message, error) {
	// Serialize the message for requests
	msgBytes := msg.Bytes()

	// For now, return a message with empty signatures
	// Full implementation would:
	// 1. Get the validator set for the source chain
	// 2. Request signatures from validators
	// 3. Aggregate signatures until quorum is reached
	// 4. Create the signed warp message

	// Create a placeholder signed message
	signedMsg, err := warp.NewMessage(msg, &warp.BitSetSignature{})
	if err != nil {
		if a.log != nil {
			a.log.Warn("failed to create signed message",
				log.Err(err),
				log.Binary("msgBytes", msgBytes),
			)
		}
		return nil, err
	}

	return signedMsg, nil
}

// RequestSignature requests a signature from a specific node.
func (a *SignatureAggregator) RequestSignature(
	ctx context.Context,
	nodeID ids.NodeID,
	msg *warp.UnsignedMessage,
) ([]byte, error) {
	if a.client == nil {
		return nil, nil
	}

	return a.client.Request(ctx, nodeID, msg.Bytes())
}
