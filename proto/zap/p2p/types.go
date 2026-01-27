//go:build !grpc

// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package p2p provides ZAP-encoded p2p message types.
// These types are wire-compatible alternatives to protobuf messages.
package p2p

import "fmt"

// EngineType identifies the consensus engine for a message.
type EngineType int32

const (
	EngineType_ENGINE_TYPE_UNSPECIFIED  EngineType = 0
	EngineType_ENGINE_TYPE_LUX          EngineType = 1
	EngineType_ENGINE_TYPE_CONSENSUSMAN EngineType = 2
)

// Message is the top-level p2p message container.
// Only one field should be non-nil at a time.
type Message struct {
	CompressedZstd          []byte
	Ping                    *Ping
	Pong                    *Pong
	Handshake               *Handshake
	GetPeerList             *GetPeerList
	PeerList                *PeerList
	GetStateSummaryFrontier *GetStateSummaryFrontier
	StateSummaryFrontier    *StateSummaryFrontier
	GetAcceptedStateSummary *GetAcceptedStateSummary
	AcceptedStateSummary    *AcceptedStateSummary
	GetAcceptedFrontier     *GetAcceptedFrontier
	AcceptedFrontier        *AcceptedFrontier
	GetAccepted             *GetAccepted
	Accepted                *Accepted
	GetAncestors            *GetAncestors
	Ancestors               *Ancestors
	Get                     *Get
	Put                     *Put
	PushQuery               *PushQuery
	PullQuery               *PullQuery
	Chits                   *Chits
	AppRequest              *AppRequest
	AppResponse             *AppResponse
	AppGossip               *AppGossip
	AppError                *AppError
	Simplex                 *Simplex
}

// GetCompressedZstd returns compressed bytes if present.
func (m *Message) GetCompressedZstd() []byte {
	if m != nil {
		return m.CompressedZstd
	}
	return nil
}

// GetPing returns the Ping message if present.
func (m *Message) GetPing() *Ping {
	if m != nil {
		return m.Ping
	}
	return nil
}

// GetPong returns the Pong message if present.
func (m *Message) GetPong() *Pong {
	if m != nil {
		return m.Pong
	}
	return nil
}

// GetHandshake returns the Handshake message if present.
func (m *Message) GetHandshake() *Handshake {
	if m != nil {
		return m.Handshake
	}
	return nil
}

// GetGetPeerList returns the GetPeerList message if present.
func (m *Message) GetGetPeerList() *GetPeerList {
	if m != nil {
		return m.GetPeerList
	}
	return nil
}

// GetPeerList_ returns the PeerList message if present.
func (m *Message) GetPeerList_() *PeerList {
	if m != nil {
		return m.PeerList
	}
	return nil
}

// GetGetStateSummaryFrontier returns the message if present.
func (m *Message) GetGetStateSummaryFrontier() *GetStateSummaryFrontier {
	if m != nil {
		return m.GetStateSummaryFrontier
	}
	return nil
}

// GetStateSummaryFrontier_ returns the message if present.
func (m *Message) GetStateSummaryFrontier_() *StateSummaryFrontier {
	if m != nil {
		return m.StateSummaryFrontier
	}
	return nil
}

// GetGetAcceptedStateSummary returns the message if present.
func (m *Message) GetGetAcceptedStateSummary() *GetAcceptedStateSummary {
	if m != nil {
		return m.GetAcceptedStateSummary
	}
	return nil
}

// GetAcceptedStateSummary_ returns the message if present.
func (m *Message) GetAcceptedStateSummary_() *AcceptedStateSummary {
	if m != nil {
		return m.AcceptedStateSummary
	}
	return nil
}

// GetGetAcceptedFrontier returns the message if present.
func (m *Message) GetGetAcceptedFrontier() *GetAcceptedFrontier {
	if m != nil {
		return m.GetAcceptedFrontier
	}
	return nil
}

// GetAcceptedFrontier_ returns the message if present.
func (m *Message) GetAcceptedFrontier_() *AcceptedFrontier {
	if m != nil {
		return m.AcceptedFrontier
	}
	return nil
}

// GetGetAccepted returns the message if present.
func (m *Message) GetGetAccepted() *GetAccepted {
	if m != nil {
		return m.GetAccepted
	}
	return nil
}

// GetAccepted_ returns the message if present.
func (m *Message) GetAccepted_() *Accepted {
	if m != nil {
		return m.Accepted
	}
	return nil
}

// GetGetAncestors returns the message if present.
func (m *Message) GetGetAncestors() *GetAncestors {
	if m != nil {
		return m.GetAncestors
	}
	return nil
}

// GetAncestors_ returns the message if present.
func (m *Message) GetAncestors_() *Ancestors {
	if m != nil {
		return m.Ancestors
	}
	return nil
}

// GetGet returns the message if present.
func (m *Message) GetGet() *Get {
	if m != nil {
		return m.Get
	}
	return nil
}

// GetPut returns the message if present.
func (m *Message) GetPut() *Put {
	if m != nil {
		return m.Put
	}
	return nil
}

// GetPushQuery returns the message if present.
func (m *Message) GetPushQuery() *PushQuery {
	if m != nil {
		return m.PushQuery
	}
	return nil
}

// GetPullQuery returns the message if present.
func (m *Message) GetPullQuery() *PullQuery {
	if m != nil {
		return m.PullQuery
	}
	return nil
}

// GetChits returns the message if present.
func (m *Message) GetChits() *Chits {
	if m != nil {
		return m.Chits
	}
	return nil
}

// GetAppRequest returns the message if present.
func (m *Message) GetAppRequest() *AppRequest {
	if m != nil {
		return m.AppRequest
	}
	return nil
}

// GetAppResponse returns the message if present.
func (m *Message) GetAppResponse() *AppResponse {
	if m != nil {
		return m.AppResponse
	}
	return nil
}

// GetAppGossip returns the message if present.
func (m *Message) GetAppGossip() *AppGossip {
	if m != nil {
		return m.AppGossip
	}
	return nil
}

// GetAppError returns the message if present.
func (m *Message) GetAppError() *AppError {
	if m != nil {
		return m.AppError
	}
	return nil
}

// GetSimplex returns the message if present.
func (m *Message) GetSimplex() *Simplex {
	if m != nil {
		return m.Simplex
	}
	return nil
}

// Ping reports a peer's perceived uptime percentage.
type Ping struct {
	Uptime uint32
}

func (m *Ping) String() string { return fmt.Sprintf("Ping{Uptime: %d}", m.Uptime) }

// Pong is sent in response to a Ping.
type Pong struct{}

func (m *Pong) String() string { return "Pong{}" }

// Handshake is the first outbound message for p2p handshake.
type Handshake struct {
	NetworkId     uint32
	MyTime        uint64
	IpAddr        []byte
	IpPort        uint32
	IpSigningTime uint64
	IpNodeIdSig   []byte
	TrackedNets   [][]byte
	Client        *Client
	SupportedLps  []uint32
	ObjectedLps   []uint32
	KnownPeers    *BloomFilter
	IpBlsSig      []byte
	AllNets       bool
}

func (m *Handshake) GetAllNets() bool {
	if m != nil {
		return m.AllNets
	}
	return false
}

func (m *Handshake) String() string {
	return fmt.Sprintf("Handshake{NetworkId: %d}", m.NetworkId)
}

// Client contains metadata about a peer's P2P client.
type Client struct {
	Name  string
	Major uint32
	Minor uint32
	Patch uint32
}

func (m *Client) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Client) GetMajor() uint32 {
	if m != nil {
		return m.Major
	}
	return 0
}

func (m *Client) GetMinor() uint32 {
	if m != nil {
		return m.Minor
	}
	return 0
}

func (m *Client) GetPatch() uint32 {
	if m != nil {
		return m.Patch
	}
	return 0
}

func (m *Client) String() string {
	return fmt.Sprintf("Client{Name: %s, Version: %d.%d.%d}", m.Name, m.Major, m.Minor, m.Patch)
}

// BloomFilter with a random salt to prevent consistent hash collisions.
type BloomFilter struct {
	Filter []byte
	Salt   []byte
}

func (m *BloomFilter) String() string {
	return fmt.Sprintf("BloomFilter{FilterLen: %d}", len(m.Filter))
}

// ClaimedIpPort contains metadata needed to connect to a peer.
type ClaimedIpPort struct {
	X509Certificate []byte
	IpAddr          []byte
	IpPort          uint32
	Timestamp       uint64
	Signature       []byte
	TxId            []byte
}

func (m *ClaimedIpPort) String() string {
	return fmt.Sprintf("ClaimedIpPort{Port: %d, Timestamp: %d}", m.IpPort, m.Timestamp)
}

// GetPeerList contains a bloom filter of currently known validator IPs.
type GetPeerList struct {
	KnownPeers *BloomFilter
	AllNets    bool
}

func (m *GetPeerList) GetKnownPeers() *BloomFilter {
	if m != nil {
		return m.KnownPeers
	}
	return nil
}

func (m *GetPeerList) GetAllNets() bool {
	if m != nil {
		return m.AllNets
	}
	return false
}

func (m *GetPeerList) String() string {
	return fmt.Sprintf("GetPeerList{AllNets: %v}", m.AllNets)
}

// PeerList contains network-level metadata for a set of validators.
type PeerList struct {
	ClaimedIpPorts []*ClaimedIpPort
}

func (m *PeerList) String() string {
	return fmt.Sprintf("PeerList{NumPeers: %d}", len(m.ClaimedIpPorts))
}

// GetStateSummaryFrontier requests a peer's most recently accepted state summary.
type GetStateSummaryFrontier struct {
	ChainId   []byte
	RequestId uint32
	Deadline  uint64
}

func (m *GetStateSummaryFrontier) GetChainId() []byte   { return m.ChainId }
func (m *GetStateSummaryFrontier) GetRequestId() uint32 { return m.RequestId }
func (m *GetStateSummaryFrontier) GetDeadline() uint64  { return m.Deadline }
func (m *GetStateSummaryFrontier) String() string {
	return fmt.Sprintf("GetStateSummaryFrontier{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// StateSummaryFrontier is sent in response to GetStateSummaryFrontier.
type StateSummaryFrontier struct {
	ChainId   []byte
	RequestId uint32
	Summary   []byte
}

func (m *StateSummaryFrontier) GetChainId() []byte   { return m.ChainId }
func (m *StateSummaryFrontier) GetRequestId() uint32 { return m.RequestId }
func (m *StateSummaryFrontier) String() string {
	return fmt.Sprintf("StateSummaryFrontier{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// GetAcceptedStateSummary requests state summaries at a set of block heights.
type GetAcceptedStateSummary struct {
	ChainId   []byte
	RequestId uint32
	Deadline  uint64
	Heights   []uint64
}

func (m *GetAcceptedStateSummary) GetChainId() []byte   { return m.ChainId }
func (m *GetAcceptedStateSummary) GetRequestId() uint32 { return m.RequestId }
func (m *GetAcceptedStateSummary) GetDeadline() uint64  { return m.Deadline }
func (m *GetAcceptedStateSummary) String() string {
	return fmt.Sprintf("GetAcceptedStateSummary{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// AcceptedStateSummary is sent in response to GetAcceptedStateSummary.
type AcceptedStateSummary struct {
	ChainId    []byte
	RequestId  uint32
	SummaryIds [][]byte
}

func (m *AcceptedStateSummary) GetChainId() []byte   { return m.ChainId }
func (m *AcceptedStateSummary) GetRequestId() uint32 { return m.RequestId }
func (m *AcceptedStateSummary) String() string {
	return fmt.Sprintf("AcceptedStateSummary{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// GetAcceptedFrontier requests the accepted frontier from a peer.
type GetAcceptedFrontier struct {
	ChainId   []byte
	RequestId uint32
	Deadline  uint64
}

func (m *GetAcceptedFrontier) GetChainId() []byte   { return m.ChainId }
func (m *GetAcceptedFrontier) GetRequestId() uint32 { return m.RequestId }
func (m *GetAcceptedFrontier) GetDeadline() uint64  { return m.Deadline }
func (m *GetAcceptedFrontier) String() string {
	return fmt.Sprintf("GetAcceptedFrontier{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// AcceptedFrontier contains the remote peer's last accepted frontier.
type AcceptedFrontier struct {
	ChainId     []byte
	RequestId   uint32
	ContainerId []byte
}

func (m *AcceptedFrontier) GetChainId() []byte   { return m.ChainId }
func (m *AcceptedFrontier) GetRequestId() uint32 { return m.RequestId }
func (m *AcceptedFrontier) String() string {
	return fmt.Sprintf("AcceptedFrontier{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// GetAccepted sends a request with the sender's accepted frontier.
type GetAccepted struct {
	ChainId      []byte
	RequestId    uint32
	Deadline     uint64
	ContainerIds [][]byte
}

func (m *GetAccepted) GetChainId() []byte   { return m.ChainId }
func (m *GetAccepted) GetRequestId() uint32 { return m.RequestId }
func (m *GetAccepted) GetDeadline() uint64  { return m.Deadline }
func (m *GetAccepted) String() string {
	return fmt.Sprintf("GetAccepted{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// Accepted is sent in response to GetAccepted.
type Accepted struct {
	ChainId      []byte
	RequestId    uint32
	ContainerIds [][]byte
}

func (m *Accepted) GetChainId() []byte   { return m.ChainId }
func (m *Accepted) GetRequestId() uint32 { return m.RequestId }
func (m *Accepted) String() string {
	return fmt.Sprintf("Accepted{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// GetAncestors requests the ancestors for a given container.
type GetAncestors struct {
	ChainId     []byte
	RequestId   uint32
	Deadline    uint64
	ContainerId []byte
	EngineType  EngineType
}

func (m *GetAncestors) GetChainId() []byte        { return m.ChainId }
func (m *GetAncestors) GetRequestId() uint32      { return m.RequestId }
func (m *GetAncestors) GetDeadline() uint64       { return m.Deadline }
func (m *GetAncestors) GetEngineType() EngineType { return m.EngineType }
func (m *GetAncestors) String() string {
	return fmt.Sprintf("GetAncestors{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// Ancestors is sent in response to GetAncestors.
type Ancestors struct {
	ChainId    []byte
	RequestId  uint32
	Containers [][]byte
}

func (m *Ancestors) GetChainId() []byte   { return m.ChainId }
func (m *Ancestors) GetRequestId() uint32 { return m.RequestId }
func (m *Ancestors) String() string {
	return fmt.Sprintf("Ancestors{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// Get requests a container from a remote peer.
type Get struct {
	ChainId     []byte
	RequestId   uint32
	Deadline    uint64
	ContainerId []byte
}

func (m *Get) GetChainId() []byte   { return m.ChainId }
func (m *Get) GetRequestId() uint32 { return m.RequestId }
func (m *Get) GetDeadline() uint64  { return m.Deadline }
func (m *Get) String() string {
	return fmt.Sprintf("Get{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// Put is sent in response to Get with the requested block.
type Put struct {
	ChainId   []byte
	RequestId uint32
	Container []byte
}

func (m *Put) GetChainId() []byte   { return m.ChainId }
func (m *Put) GetRequestId() uint32 { return m.RequestId }
func (m *Put) String() string {
	return fmt.Sprintf("Put{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// PushQuery requests the preferences of a remote peer given a container.
type PushQuery struct {
	ChainId         []byte
	RequestId       uint32
	Deadline        uint64
	Container       []byte
	RequestedHeight uint64
}

func (m *PushQuery) GetChainId() []byte   { return m.ChainId }
func (m *PushQuery) GetRequestId() uint32 { return m.RequestId }
func (m *PushQuery) GetDeadline() uint64  { return m.Deadline }
func (m *PushQuery) String() string {
	return fmt.Sprintf("PushQuery{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// PullQuery requests the preferences of a remote peer given a container id.
type PullQuery struct {
	ChainId         []byte
	RequestId       uint32
	Deadline        uint64
	ContainerId     []byte
	RequestedHeight uint64
}

func (m *PullQuery) GetChainId() []byte   { return m.ChainId }
func (m *PullQuery) GetRequestId() uint32 { return m.RequestId }
func (m *PullQuery) GetDeadline() uint64  { return m.Deadline }
func (m *PullQuery) String() string {
	return fmt.Sprintf("PullQuery{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// Chits contains the preferences of a peer in response to PushQuery or PullQuery.
type Chits struct {
	ChainId             []byte
	RequestId           uint32
	PreferredId         []byte
	AcceptedId          []byte
	PreferredIdAtHeight []byte
	AcceptedHeight      uint64
}

func (m *Chits) GetChainId() []byte   { return m.ChainId }
func (m *Chits) GetRequestId() uint32 { return m.RequestId }
func (m *Chits) String() string {
	return fmt.Sprintf("Chits{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// AppRequest is a VM-defined request.
type AppRequest struct {
	ChainId   []byte
	RequestId uint32
	Deadline  uint64
	AppBytes  []byte
}

func (m *AppRequest) GetChainId() []byte   { return m.ChainId }
func (m *AppRequest) GetRequestId() uint32 { return m.RequestId }
func (m *AppRequest) GetDeadline() uint64  { return m.Deadline }
func (m *AppRequest) String() string {
	return fmt.Sprintf("AppRequest{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// AppResponse is a VM-defined response sent in response to AppRequest.
type AppResponse struct {
	ChainId   []byte
	RequestId uint32
	AppBytes  []byte
}

func (m *AppResponse) GetChainId() []byte   { return m.ChainId }
func (m *AppResponse) GetRequestId() uint32 { return m.RequestId }
func (m *AppResponse) String() string {
	return fmt.Sprintf("AppResponse{ChainId: %x, RequestId: %d}", m.ChainId, m.RequestId)
}

// AppError is a VM-defined error sent in response to AppRequest.
type AppError struct {
	ChainId      []byte
	RequestId    uint32
	ErrorCode    int32
	ErrorMessage string
}

func (m *AppError) GetChainId() []byte   { return m.ChainId }
func (m *AppError) GetRequestId() uint32 { return m.RequestId }
func (m *AppError) String() string {
	return fmt.Sprintf("AppError{ChainId: %x, RequestId: %d, Code: %d}", m.ChainId, m.RequestId, m.ErrorCode)
}

// AppGossip is a VM-defined message.
type AppGossip struct {
	ChainId  []byte
	AppBytes []byte
}

func (m *AppGossip) GetChainId() []byte { return m.ChainId }
func (m *AppGossip) String() string {
	return fmt.Sprintf("AppGossip{ChainId: %x}", m.ChainId)
}

// Simplex contains BFT consensus messages.
type Simplex struct {
	ChainId             []byte
	BlockProposal       *BlockProposal
	Vote                *Vote
	EmptyVote           *EmptyVote
	FinalizeVote        *Vote
	Notarization        *QuorumCertificate
	EmptyNotarization   *EmptyNotarization
	Finalization        *QuorumCertificate
	ReplicationRequest  *ReplicationRequest
	ReplicationResponse *ReplicationResponse
}

func (m *Simplex) GetChainId() []byte { return m.ChainId }
func (m *Simplex) String() string {
	return fmt.Sprintf("Simplex{ChainId: %x}", m.ChainId)
}

// GetBlockProposal returns the BlockProposal if present.
func (s *Simplex) GetBlockProposal() *BlockProposal {
	if s != nil {
		return s.BlockProposal
	}
	return nil
}

// GetVote returns the Vote if present.
func (s *Simplex) GetVote() *Vote {
	if s != nil {
		return s.Vote
	}
	return nil
}

// GetEmptyVote returns the EmptyVote if present.
func (s *Simplex) GetEmptyVote() *EmptyVote {
	if s != nil {
		return s.EmptyVote
	}
	return nil
}

// GetFinalizeVote returns the finalize Vote if present.
func (s *Simplex) GetFinalizeVote() *Vote {
	if s != nil {
		return s.FinalizeVote
	}
	return nil
}

// GetNotarization returns the Notarization if present.
func (s *Simplex) GetNotarization() *QuorumCertificate {
	if s != nil {
		return s.Notarization
	}
	return nil
}

// GetEmptyNotarization returns the EmptyNotarization if present.
func (s *Simplex) GetEmptyNotarization() *EmptyNotarization {
	if s != nil {
		return s.EmptyNotarization
	}
	return nil
}

// GetFinalization returns the Finalization if present.
func (s *Simplex) GetFinalization() *QuorumCertificate {
	if s != nil {
		return s.Finalization
	}
	return nil
}

// GetReplicationRequest returns the ReplicationRequest if present.
func (s *Simplex) GetReplicationRequest() *ReplicationRequest {
	if s != nil {
		return s.ReplicationRequest
	}
	return nil
}

// GetReplicationResponse returns the ReplicationResponse if present.
func (s *Simplex) GetReplicationResponse() *ReplicationResponse {
	if s != nil {
		return s.ReplicationResponse
	}
	return nil
}

// BlockProposal contains a block and its vote.
type BlockProposal struct {
	Block []byte
	Vote  *Vote
}

func (m *BlockProposal) String() string {
	return fmt.Sprintf("BlockProposal{BlockLen: %d}", len(m.Block))
}

// ProtocolMetadata contains protocol information for a block.
type ProtocolMetadata struct {
	Version uint32
	Epoch   uint64
	Round   uint64
	Seq     uint64
	Prev    []byte
}

func (m *ProtocolMetadata) String() string {
	return fmt.Sprintf("ProtocolMetadata{Version: %d, Epoch: %d, Round: %d, Seq: %d}", m.Version, m.Epoch, m.Round, m.Seq)
}

// BlockHeader contains block metadata and digest.
type BlockHeader struct {
	Metadata *ProtocolMetadata
	Digest   []byte
}

func (m *BlockHeader) String() string {
	return fmt.Sprintf("BlockHeader{Digest: %x}", m.Digest)
}

// Signature contains a signer and signature value.
type Signature struct {
	Signer []byte
	Value  []byte
}

func (m *Signature) String() string {
	return fmt.Sprintf("Signature{Signer: %x}", m.Signer)
}

// Vote contains a block header and signature.
type Vote struct {
	BlockHeader *BlockHeader
	Signature   *Signature
}

func (m *Vote) String() string {
	return "Vote{}"
}

// EmptyVote contains metadata and signature for an empty round.
type EmptyVote struct {
	Metadata  *ProtocolMetadata
	Signature *Signature
}

func (m *EmptyVote) String() string {
	return "EmptyVote{}"
}

// QuorumCertificate contains a block header and quorum certificate.
type QuorumCertificate struct {
	BlockHeader       *BlockHeader
	QuorumCertificate []byte
}

func (m *QuorumCertificate) String() string {
	return fmt.Sprintf("QuorumCertificate{CertLen: %d}", len(m.QuorumCertificate))
}

// EmptyNotarization contains metadata and certificate for an empty round.
type EmptyNotarization struct {
	Metadata          *ProtocolMetadata
	QuorumCertificate []byte
}

func (m *EmptyNotarization) String() string {
	return fmt.Sprintf("EmptyNotarization{CertLen: %d}", len(m.QuorumCertificate))
}

// ReplicationRequest requests data for specific sequences.
type ReplicationRequest struct {
	Seqs        []uint64
	LatestRound uint64
}

func (m *ReplicationRequest) String() string {
	return fmt.Sprintf("ReplicationRequest{NumSeqs: %d, LatestRound: %d}", len(m.Seqs), m.LatestRound)
}

// ReplicationResponse contains requested data.
type ReplicationResponse struct {
	Data        []*QuorumRound
	LatestRound *QuorumRound
}

func (m *ReplicationResponse) String() string {
	return fmt.Sprintf("ReplicationResponse{NumData: %d}", len(m.Data))
}

// QuorumRound represents a round that has achieved quorum.
type QuorumRound struct {
	Block             []byte
	Notarization      *QuorumCertificate
	EmptyNotarization *EmptyNotarization
	Finalization      *QuorumCertificate
}

func (m *QuorumRound) String() string {
	return fmt.Sprintf("QuorumRound{BlockLen: %d}", len(m.Block))
}
