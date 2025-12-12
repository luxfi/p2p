// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package p2p

import (
	"context"

	"github.com/luxfi/ids"
	"github.com/luxfi/math/set"
)

// SendConfig configures how a gossip message is sent
type SendConfig struct {
	// NodeIDs specifies specific nodes to send to
	NodeIDs set.Set[ids.NodeID]

	// Validators is the number of validators to sample and send to
	Validators int

	// NonValidators is the number of non-validators to sample and send to
	NonValidators int

	// Peers is the number of peers to sample and send to
	Peers int
}

// Sender sends messages to other nodes
type Sender interface {
	// SendRequest sends a request to the specified nodes
	SendRequest(ctx context.Context, nodeIDs set.Set[ids.NodeID], requestID uint32, request []byte) error

	// SendResponse sends a response to a previous request
	SendResponse(ctx context.Context, nodeID ids.NodeID, requestID uint32, response []byte) error

	// SendError sends an error response to a previous request
	SendError(ctx context.Context, nodeID ids.NodeID, requestID uint32, errorCode int32, errorMessage string) error

	// SendGossip sends a gossip message
	SendGossip(ctx context.Context, config SendConfig, msg []byte) error
}
