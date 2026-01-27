//go:build !grpc

// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package message

import (
	"errors"
	"fmt"

	"github.com/luxfi/math/set"
	"github.com/luxfi/p2p/proto/pb/p2p"
)

// Op is an opcode
type Op byte

// Types of messages that may be sent between nodes
// Note: If you add a new parseable Op below, you must add it to either
// [UnrequestedOps] or [FailedToResponseOps].
const (
	// Handshake:
	PingOp Op = iota
	PongOp
	HandshakeOp
	GetPeerListOp
	PeerListOp
	// State sync:
	GetStateSummaryFrontierOp
	GetStateSummaryFrontierFailedOp
	StateSummaryFrontierOp
	GetAcceptedStateSummaryOp
	GetAcceptedStateSummaryFailedOp
	AcceptedStateSummaryOp
	// Bootstrapping:
	GetAcceptedFrontierOp
	GetAcceptedFrontierFailedOp
	AcceptedFrontierOp
	GetAcceptedOp
	GetAcceptedFailedOp
	AcceptedOp
	GetAncestorsOp
	GetAncestorsFailedOp
	AncestorsOp
	// Consensus:
	GetOp
	GetFailedOp
	PutOp
	PushQueryOp
	PullQueryOp
	QueryFailedOp
	QbitOp // Authenticated preference signal (formerly ChitsOp)
	// Application:
	AppRequestOp
	AppErrorOp
	AppResponseOp
	AppGossipOp
	// Internal:
	ConnectedOp
	DisconnectedOp
	NotifyOp
	GossipRequestOp
	// Simplex
	SimplexOp
)

var (
	// UnrequestedOps are operations that are expected to be seen without having
	// requested them.
	UnrequestedOps = set.Of(
		GetAcceptedFrontierOp,
		GetAcceptedOp,
		GetAncestorsOp,
		GetOp,
		PushQueryOp,
		PullQueryOp,
		AppRequestOp,
		AppGossipOp,
		GetStateSummaryFrontierOp,
		GetAcceptedStateSummaryOp,
		SimplexOp,
	)
	// FailedToResponseOps maps response failure messages to their successful
	// counterparts.
	FailedToResponseOps = map[Op]Op{
		GetStateSummaryFrontierFailedOp: StateSummaryFrontierOp,
		GetAcceptedStateSummaryFailedOp: AcceptedStateSummaryOp,
		GetAcceptedFrontierFailedOp:     AcceptedFrontierOp,
		GetAcceptedFailedOp:             AcceptedOp,
		GetAncestorsFailedOp:            AncestorsOp,
		GetFailedOp:                     PutOp,
		QueryFailedOp:                   QbitOp,
		AppErrorOp:                      AppResponseOp,
	}

	errUnknownMessageType = errors.New("unknown message type")
)

func (op Op) String() string {
	switch op {
	// Handshake
	case PingOp:
		return "ping"
	case PongOp:
		return "pong"
	case HandshakeOp:
		return "handshake"
	case GetPeerListOp:
		return "get_peerlist"
	case PeerListOp:
		return "peerlist"
	// State sync
	case GetStateSummaryFrontierOp:
		return "get_state_summary_frontier"
	case GetStateSummaryFrontierFailedOp:
		return "get_state_summary_frontier_failed"
	case StateSummaryFrontierOp:
		return "state_summary_frontier"
	case GetAcceptedStateSummaryOp:
		return "get_accepted_state_summary"
	case GetAcceptedStateSummaryFailedOp:
		return "get_accepted_state_summary_failed"
	case AcceptedStateSummaryOp:
		return "accepted_state_summary"
	// Bootstrapping
	case GetAcceptedFrontierOp:
		return "get_accepted_frontier"
	case GetAcceptedFrontierFailedOp:
		return "get_accepted_frontier_failed"
	case AcceptedFrontierOp:
		return "accepted_frontier"
	case GetAcceptedOp:
		return "get_accepted"
	case GetAcceptedFailedOp:
		return "get_accepted_failed"
	case AcceptedOp:
		return "accepted"
	case GetAncestorsOp:
		return "get_ancestors"
	case GetAncestorsFailedOp:
		return "get_ancestors_failed"
	case AncestorsOp:
		return "ancestors"
	// Consensus
	case GetOp:
		return "get"
	case GetFailedOp:
		return "get_failed"
	case PutOp:
		return "put"
	case PushQueryOp:
		return "push_query"
	case PullQueryOp:
		return "pull_query"
	case QueryFailedOp:
		return "query_failed"
	case QbitOp:
		return "qbit"
	// Application
	case AppRequestOp:
		return "app_request"
	case AppErrorOp:
		return "app_error"
	case AppResponseOp:
		return "app_response"
	case AppGossipOp:
		return "app_gossip"
	// Internal
	case ConnectedOp:
		return "connected"
	case DisconnectedOp:
		return "disconnected"
	case NotifyOp:
		return "notify"
	case GossipRequestOp:
		return "gossip_request"
	// Simplex
	case SimplexOp:
		return "simplex"
	default:
		return "unknown"
	}
}

func Unwrap(m *p2p.Message) (fmt.Stringer, error) {
	switch {
	case m.Ping != nil:
		return m.Ping, nil
	case m.Pong != nil:
		return m.Pong, nil
	case m.Handshake != nil:
		return m.Handshake, nil
	case m.GetPeerList != nil:
		return m.GetPeerList, nil
	case m.PeerList != nil:
		return m.PeerList, nil
	case m.GetStateSummaryFrontier != nil:
		return m.GetStateSummaryFrontier, nil
	case m.StateSummaryFrontier != nil:
		return m.StateSummaryFrontier, nil
	case m.GetAcceptedStateSummary != nil:
		return m.GetAcceptedStateSummary, nil
	case m.AcceptedStateSummary != nil:
		return m.AcceptedStateSummary, nil
	case m.GetAcceptedFrontier != nil:
		return m.GetAcceptedFrontier, nil
	case m.AcceptedFrontier != nil:
		return m.AcceptedFrontier, nil
	case m.GetAccepted != nil:
		return m.GetAccepted, nil
	case m.Accepted != nil:
		return m.Accepted, nil
	case m.GetAncestors != nil:
		return m.GetAncestors, nil
	case m.Ancestors != nil:
		return m.Ancestors, nil
	case m.Get != nil:
		return m.Get, nil
	case m.Put != nil:
		return m.Put, nil
	case m.PushQuery != nil:
		return m.PushQuery, nil
	case m.PullQuery != nil:
		return m.PullQuery, nil
	case m.Chits != nil:
		return m.Chits, nil
	case m.AppRequest != nil:
		return m.AppRequest, nil
	case m.AppResponse != nil:
		return m.AppResponse, nil
	case m.AppError != nil:
		return m.AppError, nil
	case m.AppGossip != nil:
		return m.AppGossip, nil
	case m.Simplex != nil:
		return m.Simplex, nil
	default:
		return nil, errUnknownMessageType
	}
}

// ToConsensusOp maps message.Op to consensus router Op values
func ToConsensusOp(op Op) (byte, bool) {
	switch op {
	case GetAcceptedFrontierOp:
		return 0, true
	case AcceptedFrontierOp:
		return 1, true
	case GetAcceptedOp:
		return 2, true
	case AcceptedOp:
		return 3, true
	case GetOp:
		return 4, true
	case PutOp:
		return 5, true
	case PushQueryOp:
		return 6, true
	case PullQueryOp:
		return 7, true
	case QbitOp:
		return 8, true
	case GetAncestorsOp:
		return 9, true
	case AncestorsOp:
		return 10, true
	default:
		return 0, false
	}
}

// GetContainerBytes extracts the container/body bytes from various message types
func GetContainerBytes(msg fmt.Stringer) []byte {
	switch m := msg.(type) {
	case *p2p.Put:
		return m.Container
	case *p2p.PushQuery:
		return m.Container
	case *p2p.Ancestors:
		if len(m.Containers) > 0 {
			return m.Containers[0]
		}
		return nil
	case *p2p.AppGossip:
		return m.AppBytes
	case *p2p.AppRequest:
		return m.AppBytes
	case *p2p.AppResponse:
		return m.AppBytes
	case *p2p.GetAccepted:
		containerIDs := m.ContainerIds
		if len(containerIDs) == 0 {
			return nil
		}
		result := make([]byte, 0, len(containerIDs)*32)
		for _, id := range containerIDs {
			result = append(result, id...)
		}
		return result
	case *p2p.Get:
		return m.ContainerId
	case *p2p.GetAncestors:
		return m.ContainerId
	case *p2p.PullQuery:
		return m.ContainerId
	case *p2p.Chits:
		return m.PreferredId
	case *p2p.AcceptedFrontier:
		return m.ContainerId
	case *p2p.Accepted:
		containerIDs := m.ContainerIds
		if len(containerIDs) == 0 {
			return nil
		}
		result := make([]byte, 0, len(containerIDs)*32)
		for _, id := range containerIDs {
			result = append(result, id...)
		}
		return result
	default:
		return nil
	}
}

func ToOp(m *p2p.Message) (Op, error) {
	switch {
	case m.Ping != nil:
		return PingOp, nil
	case m.Pong != nil:
		return PongOp, nil
	case m.Handshake != nil:
		return HandshakeOp, nil
	case m.GetPeerList != nil:
		return GetPeerListOp, nil
	case m.PeerList != nil:
		return PeerListOp, nil
	case m.GetStateSummaryFrontier != nil:
		return GetStateSummaryFrontierOp, nil
	case m.StateSummaryFrontier != nil:
		return StateSummaryFrontierOp, nil
	case m.GetAcceptedStateSummary != nil:
		return GetAcceptedStateSummaryOp, nil
	case m.AcceptedStateSummary != nil:
		return AcceptedStateSummaryOp, nil
	case m.GetAcceptedFrontier != nil:
		return GetAcceptedFrontierOp, nil
	case m.AcceptedFrontier != nil:
		return AcceptedFrontierOp, nil
	case m.GetAccepted != nil:
		return GetAcceptedOp, nil
	case m.Accepted != nil:
		return AcceptedOp, nil
	case m.GetAncestors != nil:
		return GetAncestorsOp, nil
	case m.Ancestors != nil:
		return AncestorsOp, nil
	case m.Get != nil:
		return GetOp, nil
	case m.Put != nil:
		return PutOp, nil
	case m.PushQuery != nil:
		return PushQueryOp, nil
	case m.PullQuery != nil:
		return PullQueryOp, nil
	case m.Chits != nil:
		return QbitOp, nil
	case m.AppRequest != nil:
		return AppRequestOp, nil
	case m.AppResponse != nil:
		return AppResponseOp, nil
	case m.AppError != nil:
		return AppErrorOp, nil
	case m.AppGossip != nil:
		return AppGossipOp, nil
	case m.Simplex != nil:
		return SimplexOp, nil
	default:
		return 0, errUnknownMessageType
	}
}
