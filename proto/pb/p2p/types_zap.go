//go:build !grpc

// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package p2p re-exports ZAP types for default builds.
// For protobuf types, build with -tags=grpc.
package p2p

import (
	zap "github.com/luxfi/p2p/proto/zap/p2p"
)

// Type aliases for ZAP encoding types
type (
	EngineType              = zap.EngineType
	Message                 = zap.Message
	Ping                    = zap.Ping
	Pong                    = zap.Pong
	Handshake               = zap.Handshake
	Client                  = zap.Client
	BloomFilter             = zap.BloomFilter
	ClaimedIpPort           = zap.ClaimedIpPort
	GetPeerList             = zap.GetPeerList
	PeerList                = zap.PeerList
	GetStateSummaryFrontier = zap.GetStateSummaryFrontier
	StateSummaryFrontier    = zap.StateSummaryFrontier
	GetAcceptedStateSummary = zap.GetAcceptedStateSummary
	AcceptedStateSummary    = zap.AcceptedStateSummary
	GetAcceptedFrontier     = zap.GetAcceptedFrontier
	AcceptedFrontier        = zap.AcceptedFrontier
	GetAccepted             = zap.GetAccepted
	Accepted                = zap.Accepted
	GetAncestors            = zap.GetAncestors
	Ancestors               = zap.Ancestors
	Get                     = zap.Get
	Put                     = zap.Put
	PushQuery               = zap.PushQuery
	PullQuery               = zap.PullQuery
	Chits                   = zap.Chits
	AppRequest              = zap.AppRequest
	AppResponse             = zap.AppResponse
	AppGossip               = zap.AppGossip
	AppError                = zap.AppError
	Simplex                 = zap.Simplex
	BlockProposal           = zap.BlockProposal
	ProtocolMetadata        = zap.ProtocolMetadata
	BlockHeader             = zap.BlockHeader
	Signature               = zap.Signature
	Vote                    = zap.Vote
	EmptyVote               = zap.EmptyVote
	QuorumCertificate       = zap.QuorumCertificate
	EmptyNotarization       = zap.EmptyNotarization
	ReplicationRequest      = zap.ReplicationRequest
	ReplicationResponse     = zap.ReplicationResponse
	QuorumRound             = zap.QuorumRound
	Buffer                  = zap.Buffer
	Reader                  = zap.Reader
)

// Engine type constants
const (
	EngineType_ENGINE_TYPE_UNSPECIFIED  = zap.EngineType_ENGINE_TYPE_UNSPECIFIED
	EngineType_ENGINE_TYPE_LUX          = zap.EngineType_ENGINE_TYPE_LUX
	EngineType_ENGINE_TYPE_CONSENSUSMAN = zap.EngineType_ENGINE_TYPE_CONSENSUSMAN
)

// Error constants
var (
	ErrInvalidMessage = zap.ErrInvalidMessage
	ErrUnknownTag     = zap.ErrUnknownTag
)

// Buffer/Reader constructors
var (
	NewBuffer = zap.NewBuffer
	NewReader = zap.NewReader
)

// Marshal encodes a Message to ZAP wire format.
func Marshal(m *Message) ([]byte, error) {
	return zap.Marshal(m)
}

// Unmarshal decodes a Message from ZAP wire format.
func Unmarshal(data []byte, m *Message) error {
	return zap.Unmarshal(data, m)
}

// Size returns the encoded size estimate for a message.
func Size(m *Message) int {
	return zap.Size(m)
}
