//go:build !grpc

// Copyright (C) 2019-2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package p2p

import (
	"encoding/binary"
	"errors"
	"io"
	"sync"
)

// Message tags for ZAP encoding
const (
	tagCompressedZstd          = 2
	tagPing                    = 11
	tagPong                    = 12
	tagHandshake               = 13
	tagPeerList                = 14
	tagGetStateSummaryFrontier = 15
	tagStateSummaryFrontier    = 16
	tagGetAcceptedStateSummary = 17
	tagAcceptedStateSummary    = 18
	tagGetAcceptedFrontier     = 19
	tagAcceptedFrontier        = 20
	tagGetAccepted             = 21
	tagAccepted                = 22
	tagGetAncestors            = 23
	tagAncestors               = 24
	tagGet                     = 25
	tagPut                     = 26
	tagPushQuery               = 27
	tagPullQuery               = 28
	tagChits                   = 29
	tagAppRequest              = 30
	tagAppResponse             = 31
	tagAppGossip               = 32
	tagAppError                = 34
	tagGetPeerList             = 35
	tagSimplex                 = 36
)

// Simplex message tags
const (
	simplexBlockProposal       = 2
	simplexVote                = 3
	simplexEmptyVote           = 4
	simplexFinalizeVote        = 5
	simplexNotarization        = 6
	simplexEmptyNotarization   = 7
	simplexFinalization        = 8
	simplexReplicationRequest  = 9
	simplexReplicationResponse = 10
)

var (
	ErrInvalidMessage = errors.New("invalid wire message")
	ErrUnknownTag     = errors.New("unknown message tag")

	bufPool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 64*1024)
			return &b
		},
	}
)

// Buffer for zero-copy encoding
type Buffer struct {
	data   []byte
	offset int
}

func NewBuffer(size int) *Buffer {
	return &Buffer{data: make([]byte, size)}
}

func (b *Buffer) grow(n int) {
	if b.offset+n > len(b.data) {
		newData := make([]byte, (b.offset+n)*2)
		copy(newData, b.data[:b.offset])
		b.data = newData
	}
}

func (b *Buffer) WriteUint8(v uint8) {
	b.grow(1)
	b.data[b.offset] = v
	b.offset++
}

func (b *Buffer) WriteBool(v bool) {
	if v {
		b.WriteUint8(1)
	} else {
		b.WriteUint8(0)
	}
}

func (b *Buffer) WriteUint32(v uint32) {
	b.grow(4)
	binary.BigEndian.PutUint32(b.data[b.offset:], v)
	b.offset += 4
}

func (b *Buffer) WriteUint64(v uint64) {
	b.grow(8)
	binary.BigEndian.PutUint64(b.data[b.offset:], v)
	b.offset += 8
}

func (b *Buffer) WriteInt32(v int32) {
	b.WriteUint32(uint32(v))
}

func (b *Buffer) WriteBytes(data []byte) {
	b.WriteUint32(uint32(len(data)))
	b.grow(len(data))
	copy(b.data[b.offset:], data)
	b.offset += len(data)
}

func (b *Buffer) WriteString(s string) {
	b.WriteBytes([]byte(s))
}

func (b *Buffer) WriteBytesSlice(slices [][]byte) {
	b.WriteUint32(uint32(len(slices)))
	for _, s := range slices {
		b.WriteBytes(s)
	}
}

func (b *Buffer) WriteUint32Slice(vals []uint32) {
	b.WriteUint32(uint32(len(vals)))
	for _, v := range vals {
		b.WriteUint32(v)
	}
}

func (b *Buffer) WriteUint64Slice(vals []uint64) {
	b.WriteUint32(uint32(len(vals)))
	for _, v := range vals {
		b.WriteUint64(v)
	}
}

func (b *Buffer) Bytes() []byte {
	return b.data[:b.offset]
}

// Reader for zero-copy decoding
type Reader struct {
	data   []byte
	offset int
}

func NewReader(data []byte) *Reader {
	return &Reader{data: data}
}

func (r *Reader) ReadUint8() (uint8, error) {
	if r.offset+1 > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}
	v := r.data[r.offset]
	r.offset++
	return v, nil
}

func (r *Reader) ReadBool() (bool, error) {
	v, err := r.ReadUint8()
	return v != 0, err
}

func (r *Reader) ReadUint32() (uint32, error) {
	if r.offset+4 > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.BigEndian.Uint32(r.data[r.offset:])
	r.offset += 4
	return v, nil
}

func (r *Reader) ReadInt32() (int32, error) {
	v, err := r.ReadUint32()
	return int32(v), err
}

func (r *Reader) ReadUint64() (uint64, error) {
	if r.offset+8 > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.BigEndian.Uint64(r.data[r.offset:])
	r.offset += 8
	return v, nil
}

func (r *Reader) ReadBytes() ([]byte, error) {
	length, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	if r.offset+int(length) > len(r.data) {
		return nil, io.ErrUnexpectedEOF
	}
	data := r.data[r.offset : r.offset+int(length)]
	r.offset += int(length)
	return data, nil
}

func (r *Reader) ReadString() (string, error) {
	b, err := r.ReadBytes()
	return string(b), err
}

func (r *Reader) ReadBytesSlice() ([][]byte, error) {
	count, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	result := make([][]byte, count)
	for i := uint32(0); i < count; i++ {
		result[i], err = r.ReadBytes()
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (r *Reader) ReadUint32Slice() ([]uint32, error) {
	count, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	result := make([]uint32, count)
	for i := uint32(0); i < count; i++ {
		result[i], err = r.ReadUint32()
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (r *Reader) ReadUint64Slice() ([]uint64, error) {
	count, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	result := make([]uint64, count)
	for i := uint32(0); i < count; i++ {
		result[i], err = r.ReadUint64()
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Marshal encodes a Message to ZAP wire format
func Marshal(m *Message) ([]byte, error) {
	buf := NewBuffer(4096)

	switch {
	case m.CompressedZstd != nil:
		buf.WriteUint8(tagCompressedZstd)
		buf.WriteBytes(m.CompressedZstd)
	case m.Ping != nil:
		buf.WriteUint8(tagPing)
		marshalPing(buf, m.Ping)
	case m.Pong != nil:
		buf.WriteUint8(tagPong)
		marshalPong(buf, m.Pong)
	case m.Handshake != nil:
		buf.WriteUint8(tagHandshake)
		marshalHandshake(buf, m.Handshake)
	case m.GetPeerList != nil:
		buf.WriteUint8(tagGetPeerList)
		marshalGetPeerList(buf, m.GetPeerList)
	case m.PeerList != nil:
		buf.WriteUint8(tagPeerList)
		marshalPeerList(buf, m.PeerList)
	case m.GetStateSummaryFrontier != nil:
		buf.WriteUint8(tagGetStateSummaryFrontier)
		marshalGetStateSummaryFrontier(buf, m.GetStateSummaryFrontier)
	case m.StateSummaryFrontier != nil:
		buf.WriteUint8(tagStateSummaryFrontier)
		marshalStateSummaryFrontier(buf, m.StateSummaryFrontier)
	case m.GetAcceptedStateSummary != nil:
		buf.WriteUint8(tagGetAcceptedStateSummary)
		marshalGetAcceptedStateSummary(buf, m.GetAcceptedStateSummary)
	case m.AcceptedStateSummary != nil:
		buf.WriteUint8(tagAcceptedStateSummary)
		marshalAcceptedStateSummary(buf, m.AcceptedStateSummary)
	case m.GetAcceptedFrontier != nil:
		buf.WriteUint8(tagGetAcceptedFrontier)
		marshalGetAcceptedFrontier(buf, m.GetAcceptedFrontier)
	case m.AcceptedFrontier != nil:
		buf.WriteUint8(tagAcceptedFrontier)
		marshalAcceptedFrontier(buf, m.AcceptedFrontier)
	case m.GetAccepted != nil:
		buf.WriteUint8(tagGetAccepted)
		marshalGetAccepted(buf, m.GetAccepted)
	case m.Accepted != nil:
		buf.WriteUint8(tagAccepted)
		marshalAccepted(buf, m.Accepted)
	case m.GetAncestors != nil:
		buf.WriteUint8(tagGetAncestors)
		marshalGetAncestors(buf, m.GetAncestors)
	case m.Ancestors != nil:
		buf.WriteUint8(tagAncestors)
		marshalAncestors(buf, m.Ancestors)
	case m.Get != nil:
		buf.WriteUint8(tagGet)
		marshalGet(buf, m.Get)
	case m.Put != nil:
		buf.WriteUint8(tagPut)
		marshalPut(buf, m.Put)
	case m.PushQuery != nil:
		buf.WriteUint8(tagPushQuery)
		marshalPushQuery(buf, m.PushQuery)
	case m.PullQuery != nil:
		buf.WriteUint8(tagPullQuery)
		marshalPullQuery(buf, m.PullQuery)
	case m.Chits != nil:
		buf.WriteUint8(tagChits)
		marshalChits(buf, m.Chits)
	case m.AppRequest != nil:
		buf.WriteUint8(tagAppRequest)
		marshalAppRequest(buf, m.AppRequest)
	case m.AppResponse != nil:
		buf.WriteUint8(tagAppResponse)
		marshalAppResponse(buf, m.AppResponse)
	case m.AppGossip != nil:
		buf.WriteUint8(tagAppGossip)
		marshalAppGossip(buf, m.AppGossip)
	case m.AppError != nil:
		buf.WriteUint8(tagAppError)
		marshalAppError(buf, m.AppError)
	case m.Simplex != nil:
		buf.WriteUint8(tagSimplex)
		marshalSimplex(buf, m.Simplex)
	default:
		return nil, ErrInvalidMessage
	}

	return buf.Bytes(), nil
}

// Unmarshal decodes a Message from ZAP wire format
func Unmarshal(data []byte, m *Message) error {
	if len(data) < 1 {
		return ErrInvalidMessage
	}

	r := NewReader(data)
	tag, _ := r.ReadUint8()
	var err error

	switch tag {
	case tagCompressedZstd:
		m.CompressedZstd, err = r.ReadBytes()
	case tagPing:
		m.Ping, err = unmarshalPing(r)
	case tagPong:
		m.Pong, err = unmarshalPong(r)
	case tagHandshake:
		m.Handshake, err = unmarshalHandshake(r)
	case tagGetPeerList:
		m.GetPeerList, err = unmarshalGetPeerList(r)
	case tagPeerList:
		m.PeerList, err = unmarshalPeerList(r)
	case tagGetStateSummaryFrontier:
		m.GetStateSummaryFrontier, err = unmarshalGetStateSummaryFrontier(r)
	case tagStateSummaryFrontier:
		m.StateSummaryFrontier, err = unmarshalStateSummaryFrontier(r)
	case tagGetAcceptedStateSummary:
		m.GetAcceptedStateSummary, err = unmarshalGetAcceptedStateSummary(r)
	case tagAcceptedStateSummary:
		m.AcceptedStateSummary, err = unmarshalAcceptedStateSummary(r)
	case tagGetAcceptedFrontier:
		m.GetAcceptedFrontier, err = unmarshalGetAcceptedFrontier(r)
	case tagAcceptedFrontier:
		m.AcceptedFrontier, err = unmarshalAcceptedFrontier(r)
	case tagGetAccepted:
		m.GetAccepted, err = unmarshalGetAccepted(r)
	case tagAccepted:
		m.Accepted, err = unmarshalAccepted(r)
	case tagGetAncestors:
		m.GetAncestors, err = unmarshalGetAncestors(r)
	case tagAncestors:
		m.Ancestors, err = unmarshalAncestors(r)
	case tagGet:
		m.Get, err = unmarshalGet(r)
	case tagPut:
		m.Put, err = unmarshalPut(r)
	case tagPushQuery:
		m.PushQuery, err = unmarshalPushQuery(r)
	case tagPullQuery:
		m.PullQuery, err = unmarshalPullQuery(r)
	case tagChits:
		m.Chits, err = unmarshalChits(r)
	case tagAppRequest:
		m.AppRequest, err = unmarshalAppRequest(r)
	case tagAppResponse:
		m.AppResponse, err = unmarshalAppResponse(r)
	case tagAppGossip:
		m.AppGossip, err = unmarshalAppGossip(r)
	case tagAppError:
		m.AppError, err = unmarshalAppError(r)
	case tagSimplex:
		m.Simplex, err = unmarshalSimplex(r)
	default:
		return ErrUnknownTag
	}

	return err
}

// Size returns the encoded size of a message (estimate)
func Size(m *Message) int {
	return 4096
}

// Marshal helpers - each message type
func marshalPing(b *Buffer, m *Ping) {
	b.WriteUint32(m.Uptime)
}

func marshalPong(b *Buffer, m *Pong) {
	// Pong has no fields
}

func marshalHandshake(b *Buffer, m *Handshake) {
	b.WriteUint32(m.NetworkId)
	b.WriteUint64(m.MyTime)
	b.WriteBytes(m.IpAddr)
	b.WriteUint32(m.IpPort)
	b.WriteUint64(m.IpSigningTime)
	b.WriteBytes(m.IpNodeIdSig)
	b.WriteBytesSlice(m.TrackedNets)
	if m.Client != nil {
		b.WriteUint8(1)
		b.WriteString(m.Client.Name)
		b.WriteUint32(m.Client.Major)
		b.WriteUint32(m.Client.Minor)
		b.WriteUint32(m.Client.Patch)
	} else {
		b.WriteUint8(0)
	}
	b.WriteUint32Slice(m.SupportedLps)
	b.WriteUint32Slice(m.ObjectedLps)
	if m.KnownPeers != nil {
		b.WriteUint8(1)
		b.WriteBytes(m.KnownPeers.Filter)
		b.WriteBytes(m.KnownPeers.Salt)
	} else {
		b.WriteUint8(0)
	}
	b.WriteBytes(m.IpBlsSig)
	b.WriteBool(m.AllNets)
}

func marshalGetPeerList(b *Buffer, m *GetPeerList) {
	if m.KnownPeers != nil {
		b.WriteUint8(1)
		b.WriteBytes(m.KnownPeers.Filter)
		b.WriteBytes(m.KnownPeers.Salt)
	} else {
		b.WriteUint8(0)
	}
	b.WriteBool(m.AllNets)
}

func marshalPeerList(b *Buffer, m *PeerList) {
	b.WriteUint32(uint32(len(m.ClaimedIpPorts)))
	for _, p := range m.ClaimedIpPorts {
		b.WriteBytes(p.X509Certificate)
		b.WriteBytes(p.IpAddr)
		b.WriteUint32(p.IpPort)
		b.WriteUint64(p.Timestamp)
		b.WriteBytes(p.Signature)
		b.WriteBytes(p.TxId)
	}
}

func marshalGetStateSummaryFrontier(b *Buffer, m *GetStateSummaryFrontier) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteUint64(m.Deadline)
}

func marshalStateSummaryFrontier(b *Buffer, m *StateSummaryFrontier) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteBytes(m.Summary)
}

func marshalGetAcceptedStateSummary(b *Buffer, m *GetAcceptedStateSummary) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteUint64(m.Deadline)
	b.WriteUint64Slice(m.Heights)
}

func marshalAcceptedStateSummary(b *Buffer, m *AcceptedStateSummary) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteBytesSlice(m.SummaryIds)
}

func marshalGetAcceptedFrontier(b *Buffer, m *GetAcceptedFrontier) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteUint64(m.Deadline)
}

func marshalAcceptedFrontier(b *Buffer, m *AcceptedFrontier) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteBytes(m.ContainerId)
}

func marshalGetAccepted(b *Buffer, m *GetAccepted) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteUint64(m.Deadline)
	b.WriteBytesSlice(m.ContainerIds)
}

func marshalAccepted(b *Buffer, m *Accepted) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteBytesSlice(m.ContainerIds)
}

func marshalGetAncestors(b *Buffer, m *GetAncestors) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteUint64(m.Deadline)
	b.WriteBytes(m.ContainerId)
	b.WriteUint32(uint32(m.EngineType))
}

func marshalAncestors(b *Buffer, m *Ancestors) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteBytesSlice(m.Containers)
}

func marshalGet(b *Buffer, m *Get) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteUint64(m.Deadline)
	b.WriteBytes(m.ContainerId)
}

func marshalPut(b *Buffer, m *Put) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteBytes(m.Container)
}

func marshalPushQuery(b *Buffer, m *PushQuery) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteUint64(m.Deadline)
	b.WriteBytes(m.Container)
	b.WriteUint64(m.RequestedHeight)
}

func marshalPullQuery(b *Buffer, m *PullQuery) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteUint64(m.Deadline)
	b.WriteBytes(m.ContainerId)
	b.WriteUint64(m.RequestedHeight)
}

func marshalChits(b *Buffer, m *Chits) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteBytes(m.PreferredId)
	b.WriteBytes(m.AcceptedId)
	b.WriteBytes(m.PreferredIdAtHeight)
	b.WriteUint64(m.AcceptedHeight)
}

func marshalAppRequest(b *Buffer, m *AppRequest) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteUint64(m.Deadline)
	b.WriteBytes(m.AppBytes)
}

func marshalAppResponse(b *Buffer, m *AppResponse) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteBytes(m.AppBytes)
}

func marshalAppGossip(b *Buffer, m *AppGossip) {
	b.WriteBytes(m.ChainId)
	b.WriteBytes(m.AppBytes)
}

func marshalAppError(b *Buffer, m *AppError) {
	b.WriteBytes(m.ChainId)
	b.WriteUint32(m.RequestId)
	b.WriteInt32(m.ErrorCode)
	b.WriteString(m.ErrorMessage)
}

func marshalSimplex(b *Buffer, m *Simplex) {
	b.WriteBytes(m.ChainId)
	switch {
	case m.BlockProposal != nil:
		b.WriteUint8(simplexBlockProposal)
		marshalBlockProposal(b, m.BlockProposal)
	case m.Vote != nil:
		b.WriteUint8(simplexVote)
		marshalVote(b, m.Vote)
	case m.EmptyVote != nil:
		b.WriteUint8(simplexEmptyVote)
		marshalEmptyVote(b, m.EmptyVote)
	case m.FinalizeVote != nil:
		b.WriteUint8(simplexFinalizeVote)
		marshalVote(b, m.FinalizeVote)
	case m.Notarization != nil:
		b.WriteUint8(simplexNotarization)
		marshalQuorumCertificate(b, m.Notarization)
	case m.EmptyNotarization != nil:
		b.WriteUint8(simplexEmptyNotarization)
		marshalEmptyNotarization(b, m.EmptyNotarization)
	case m.Finalization != nil:
		b.WriteUint8(simplexFinalization)
		marshalQuorumCertificate(b, m.Finalization)
	case m.ReplicationRequest != nil:
		b.WriteUint8(simplexReplicationRequest)
		marshalReplicationRequest(b, m.ReplicationRequest)
	case m.ReplicationResponse != nil:
		b.WriteUint8(simplexReplicationResponse)
		marshalReplicationResponse(b, m.ReplicationResponse)
	default:
		b.WriteUint8(0)
	}
}

func marshalBlockProposal(b *Buffer, m *BlockProposal) {
	b.WriteBytes(m.Block)
	if m.Vote != nil {
		b.WriteUint8(1)
		marshalVote(b, m.Vote)
	} else {
		b.WriteUint8(0)
	}
}

func marshalProtocolMetadata(b *Buffer, m *ProtocolMetadata) {
	b.WriteUint32(m.Version)
	b.WriteUint64(m.Epoch)
	b.WriteUint64(m.Round)
	b.WriteUint64(m.Seq)
	b.WriteBytes(m.Prev)
}

func marshalBlockHeader(b *Buffer, m *BlockHeader) {
	if m.Metadata != nil {
		b.WriteUint8(1)
		marshalProtocolMetadata(b, m.Metadata)
	} else {
		b.WriteUint8(0)
	}
	b.WriteBytes(m.Digest)
}

func marshalSignature(b *Buffer, m *Signature) {
	b.WriteBytes(m.Signer)
	b.WriteBytes(m.Value)
}

func marshalVote(b *Buffer, m *Vote) {
	if m.BlockHeader != nil {
		b.WriteUint8(1)
		marshalBlockHeader(b, m.BlockHeader)
	} else {
		b.WriteUint8(0)
	}
	if m.Signature != nil {
		b.WriteUint8(1)
		marshalSignature(b, m.Signature)
	} else {
		b.WriteUint8(0)
	}
}

func marshalEmptyVote(b *Buffer, m *EmptyVote) {
	if m.Metadata != nil {
		b.WriteUint8(1)
		marshalProtocolMetadata(b, m.Metadata)
	} else {
		b.WriteUint8(0)
	}
	if m.Signature != nil {
		b.WriteUint8(1)
		marshalSignature(b, m.Signature)
	} else {
		b.WriteUint8(0)
	}
}

func marshalQuorumCertificate(b *Buffer, m *QuorumCertificate) {
	if m.BlockHeader != nil {
		b.WriteUint8(1)
		marshalBlockHeader(b, m.BlockHeader)
	} else {
		b.WriteUint8(0)
	}
	b.WriteBytes(m.QuorumCertificate)
}

func marshalEmptyNotarization(b *Buffer, m *EmptyNotarization) {
	if m.Metadata != nil {
		b.WriteUint8(1)
		marshalProtocolMetadata(b, m.Metadata)
	} else {
		b.WriteUint8(0)
	}
	b.WriteBytes(m.QuorumCertificate)
}

func marshalReplicationRequest(b *Buffer, m *ReplicationRequest) {
	b.WriteUint64Slice(m.Seqs)
	b.WriteUint64(m.LatestRound)
}

func marshalReplicationResponse(b *Buffer, m *ReplicationResponse) {
	b.WriteUint32(uint32(len(m.Data)))
	for _, d := range m.Data {
		marshalQuorumRound(b, d)
	}
	if m.LatestRound != nil {
		b.WriteUint8(1)
		marshalQuorumRound(b, m.LatestRound)
	} else {
		b.WriteUint8(0)
	}
}

func marshalQuorumRound(b *Buffer, m *QuorumRound) {
	b.WriteBytes(m.Block)
	if m.Notarization != nil {
		b.WriteUint8(1)
		marshalQuorumCertificate(b, m.Notarization)
	} else {
		b.WriteUint8(0)
	}
	if m.EmptyNotarization != nil {
		b.WriteUint8(1)
		marshalEmptyNotarization(b, m.EmptyNotarization)
	} else {
		b.WriteUint8(0)
	}
	if m.Finalization != nil {
		b.WriteUint8(1)
		marshalQuorumCertificate(b, m.Finalization)
	} else {
		b.WriteUint8(0)
	}
}

// Unmarshal helpers
func unmarshalPing(r *Reader) (*Ping, error) {
	m := &Ping{}
	var err error
	m.Uptime, err = r.ReadUint32()
	return m, err
}

func unmarshalPong(r *Reader) (*Pong, error) {
	return &Pong{}, nil
}

func unmarshalHandshake(r *Reader) (*Handshake, error) {
	m := &Handshake{}
	var err error
	m.NetworkId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.MyTime, err = r.ReadUint64()
	if err != nil {
		return nil, err
	}
	m.IpAddr, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.IpPort, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.IpSigningTime, err = r.ReadUint64()
	if err != nil {
		return nil, err
	}
	m.IpNodeIdSig, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.TrackedNets, err = r.ReadBytesSlice()
	if err != nil {
		return nil, err
	}
	hasClient, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasClient == 1 {
		m.Client = &Client{}
		m.Client.Name, err = r.ReadString()
		if err != nil {
			return nil, err
		}
		m.Client.Major, err = r.ReadUint32()
		if err != nil {
			return nil, err
		}
		m.Client.Minor, err = r.ReadUint32()
		if err != nil {
			return nil, err
		}
		m.Client.Patch, err = r.ReadUint32()
		if err != nil {
			return nil, err
		}
	}
	m.SupportedLps, err = r.ReadUint32Slice()
	if err != nil {
		return nil, err
	}
	m.ObjectedLps, err = r.ReadUint32Slice()
	if err != nil {
		return nil, err
	}
	hasKnownPeers, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasKnownPeers == 1 {
		m.KnownPeers = &BloomFilter{}
		m.KnownPeers.Filter, err = r.ReadBytes()
		if err != nil {
			return nil, err
		}
		m.KnownPeers.Salt, err = r.ReadBytes()
		if err != nil {
			return nil, err
		}
	}
	m.IpBlsSig, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.AllNets, err = r.ReadBool()
	return m, err
}

func unmarshalGetPeerList(r *Reader) (*GetPeerList, error) {
	m := &GetPeerList{}
	hasKnownPeers, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasKnownPeers == 1 {
		m.KnownPeers = &BloomFilter{}
		m.KnownPeers.Filter, err = r.ReadBytes()
		if err != nil {
			return nil, err
		}
		m.KnownPeers.Salt, err = r.ReadBytes()
		if err != nil {
			return nil, err
		}
	}
	m.AllNets, err = r.ReadBool()
	return m, err
}

func unmarshalPeerList(r *Reader) (*PeerList, error) {
	m := &PeerList{}
	count, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.ClaimedIpPorts = make([]*ClaimedIpPort, count)
	for i := uint32(0); i < count; i++ {
		p := &ClaimedIpPort{}
		p.X509Certificate, err = r.ReadBytes()
		if err != nil {
			return nil, err
		}
		p.IpAddr, err = r.ReadBytes()
		if err != nil {
			return nil, err
		}
		p.IpPort, err = r.ReadUint32()
		if err != nil {
			return nil, err
		}
		p.Timestamp, err = r.ReadUint64()
		if err != nil {
			return nil, err
		}
		p.Signature, err = r.ReadBytes()
		if err != nil {
			return nil, err
		}
		p.TxId, err = r.ReadBytes()
		if err != nil {
			return nil, err
		}
		m.ClaimedIpPorts[i] = p
	}
	return m, nil
}

func unmarshalGetStateSummaryFrontier(r *Reader) (*GetStateSummaryFrontier, error) {
	m := &GetStateSummaryFrontier{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Deadline, err = r.ReadUint64()
	return m, err
}

func unmarshalStateSummaryFrontier(r *Reader) (*StateSummaryFrontier, error) {
	m := &StateSummaryFrontier{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Summary, err = r.ReadBytes()
	return m, err
}

func unmarshalGetAcceptedStateSummary(r *Reader) (*GetAcceptedStateSummary, error) {
	m := &GetAcceptedStateSummary{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Deadline, err = r.ReadUint64()
	if err != nil {
		return nil, err
	}
	m.Heights, err = r.ReadUint64Slice()
	return m, err
}

func unmarshalAcceptedStateSummary(r *Reader) (*AcceptedStateSummary, error) {
	m := &AcceptedStateSummary{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.SummaryIds, err = r.ReadBytesSlice()
	return m, err
}

func unmarshalGetAcceptedFrontier(r *Reader) (*GetAcceptedFrontier, error) {
	m := &GetAcceptedFrontier{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Deadline, err = r.ReadUint64()
	return m, err
}

func unmarshalAcceptedFrontier(r *Reader) (*AcceptedFrontier, error) {
	m := &AcceptedFrontier{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.ContainerId, err = r.ReadBytes()
	return m, err
}

func unmarshalGetAccepted(r *Reader) (*GetAccepted, error) {
	m := &GetAccepted{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Deadline, err = r.ReadUint64()
	if err != nil {
		return nil, err
	}
	m.ContainerIds, err = r.ReadBytesSlice()
	return m, err
}

func unmarshalAccepted(r *Reader) (*Accepted, error) {
	m := &Accepted{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.ContainerIds, err = r.ReadBytesSlice()
	return m, err
}

func unmarshalGetAncestors(r *Reader) (*GetAncestors, error) {
	m := &GetAncestors{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Deadline, err = r.ReadUint64()
	if err != nil {
		return nil, err
	}
	m.ContainerId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	et, err := r.ReadUint32()
	m.EngineType = EngineType(et)
	return m, err
}

func unmarshalAncestors(r *Reader) (*Ancestors, error) {
	m := &Ancestors{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Containers, err = r.ReadBytesSlice()
	return m, err
}

func unmarshalGet(r *Reader) (*Get, error) {
	m := &Get{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Deadline, err = r.ReadUint64()
	if err != nil {
		return nil, err
	}
	m.ContainerId, err = r.ReadBytes()
	return m, err
}

func unmarshalPut(r *Reader) (*Put, error) {
	m := &Put{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Container, err = r.ReadBytes()
	return m, err
}

func unmarshalPushQuery(r *Reader) (*PushQuery, error) {
	m := &PushQuery{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Deadline, err = r.ReadUint64()
	if err != nil {
		return nil, err
	}
	m.Container, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestedHeight, err = r.ReadUint64()
	return m, err
}

func unmarshalPullQuery(r *Reader) (*PullQuery, error) {
	m := &PullQuery{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Deadline, err = r.ReadUint64()
	if err != nil {
		return nil, err
	}
	m.ContainerId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestedHeight, err = r.ReadUint64()
	return m, err
}

func unmarshalChits(r *Reader) (*Chits, error) {
	m := &Chits{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.PreferredId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.AcceptedId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.PreferredIdAtHeight, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.AcceptedHeight, err = r.ReadUint64()
	return m, err
}

func unmarshalAppRequest(r *Reader) (*AppRequest, error) {
	m := &AppRequest{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Deadline, err = r.ReadUint64()
	if err != nil {
		return nil, err
	}
	m.AppBytes, err = r.ReadBytes()
	return m, err
}

func unmarshalAppResponse(r *Reader) (*AppResponse, error) {
	m := &AppResponse{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.AppBytes, err = r.ReadBytes()
	return m, err
}

func unmarshalAppGossip(r *Reader) (*AppGossip, error) {
	m := &AppGossip{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.AppBytes, err = r.ReadBytes()
	return m, err
}

func unmarshalAppError(r *Reader) (*AppError, error) {
	m := &AppError{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.RequestId, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.ErrorCode, err = r.ReadInt32()
	if err != nil {
		return nil, err
	}
	m.ErrorMessage, err = r.ReadString()
	return m, err
}

func unmarshalSimplex(r *Reader) (*Simplex, error) {
	m := &Simplex{}
	var err error
	m.ChainId, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	msgType, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	switch msgType {
	case simplexBlockProposal:
		m.BlockProposal, err = unmarshalBlockProposal(r)
	case simplexVote:
		m.Vote, err = unmarshalVote(r)
	case simplexEmptyVote:
		m.EmptyVote, err = unmarshalEmptyVote(r)
	case simplexFinalizeVote:
		m.FinalizeVote, err = unmarshalVote(r)
	case simplexNotarization:
		m.Notarization, err = unmarshalQuorumCertificate(r)
	case simplexEmptyNotarization:
		m.EmptyNotarization, err = unmarshalEmptyNotarization(r)
	case simplexFinalization:
		m.Finalization, err = unmarshalQuorumCertificate(r)
	case simplexReplicationRequest:
		m.ReplicationRequest, err = unmarshalReplicationRequest(r)
	case simplexReplicationResponse:
		m.ReplicationResponse, err = unmarshalReplicationResponse(r)
	}
	return m, err
}

func unmarshalBlockProposal(r *Reader) (*BlockProposal, error) {
	m := &BlockProposal{}
	var err error
	m.Block, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	hasVote, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasVote == 1 {
		m.Vote, err = unmarshalVote(r)
	}
	return m, err
}

func unmarshalProtocolMetadata(r *Reader) (*ProtocolMetadata, error) {
	m := &ProtocolMetadata{}
	var err error
	m.Version, err = r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Epoch, err = r.ReadUint64()
	if err != nil {
		return nil, err
	}
	m.Round, err = r.ReadUint64()
	if err != nil {
		return nil, err
	}
	m.Seq, err = r.ReadUint64()
	if err != nil {
		return nil, err
	}
	m.Prev, err = r.ReadBytes()
	return m, err
}

func unmarshalBlockHeader(r *Reader) (*BlockHeader, error) {
	m := &BlockHeader{}
	var err error
	hasMetadata, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasMetadata == 1 {
		m.Metadata, err = unmarshalProtocolMetadata(r)
		if err != nil {
			return nil, err
		}
	}
	m.Digest, err = r.ReadBytes()
	return m, err
}

func unmarshalSignature(r *Reader) (*Signature, error) {
	m := &Signature{}
	var err error
	m.Signer, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	m.Value, err = r.ReadBytes()
	return m, err
}

func unmarshalVote(r *Reader) (*Vote, error) {
	m := &Vote{}
	var err error
	hasBlockHeader, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasBlockHeader == 1 {
		m.BlockHeader, err = unmarshalBlockHeader(r)
		if err != nil {
			return nil, err
		}
	}
	hasSignature, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasSignature == 1 {
		m.Signature, err = unmarshalSignature(r)
	}
	return m, err
}

func unmarshalEmptyVote(r *Reader) (*EmptyVote, error) {
	m := &EmptyVote{}
	var err error
	hasMetadata, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasMetadata == 1 {
		m.Metadata, err = unmarshalProtocolMetadata(r)
		if err != nil {
			return nil, err
		}
	}
	hasSignature, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasSignature == 1 {
		m.Signature, err = unmarshalSignature(r)
	}
	return m, err
}

func unmarshalQuorumCertificate(r *Reader) (*QuorumCertificate, error) {
	m := &QuorumCertificate{}
	var err error
	hasBlockHeader, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasBlockHeader == 1 {
		m.BlockHeader, err = unmarshalBlockHeader(r)
		if err != nil {
			return nil, err
		}
	}
	m.QuorumCertificate, err = r.ReadBytes()
	return m, err
}

func unmarshalEmptyNotarization(r *Reader) (*EmptyNotarization, error) {
	m := &EmptyNotarization{}
	var err error
	hasMetadata, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasMetadata == 1 {
		m.Metadata, err = unmarshalProtocolMetadata(r)
		if err != nil {
			return nil, err
		}
	}
	m.QuorumCertificate, err = r.ReadBytes()
	return m, err
}

func unmarshalReplicationRequest(r *Reader) (*ReplicationRequest, error) {
	m := &ReplicationRequest{}
	var err error
	m.Seqs, err = r.ReadUint64Slice()
	if err != nil {
		return nil, err
	}
	m.LatestRound, err = r.ReadUint64()
	return m, err
}

func unmarshalReplicationResponse(r *Reader) (*ReplicationResponse, error) {
	m := &ReplicationResponse{}
	count, err := r.ReadUint32()
	if err != nil {
		return nil, err
	}
	m.Data = make([]*QuorumRound, count)
	for i := uint32(0); i < count; i++ {
		m.Data[i], err = unmarshalQuorumRound(r)
		if err != nil {
			return nil, err
		}
	}
	hasLatest, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasLatest == 1 {
		m.LatestRound, err = unmarshalQuorumRound(r)
	}
	return m, err
}

func unmarshalQuorumRound(r *Reader) (*QuorumRound, error) {
	m := &QuorumRound{}
	var err error
	m.Block, err = r.ReadBytes()
	if err != nil {
		return nil, err
	}
	hasNotarization, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasNotarization == 1 {
		m.Notarization, err = unmarshalQuorumCertificate(r)
		if err != nil {
			return nil, err
		}
	}
	hasEmptyNotarization, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasEmptyNotarization == 1 {
		m.EmptyNotarization, err = unmarshalEmptyNotarization(r)
		if err != nil {
			return nil, err
		}
	}
	hasFinalization, err := r.ReadUint8()
	if err != nil {
		return nil, err
	}
	if hasFinalization == 1 {
		m.Finalization, err = unmarshalQuorumCertificate(r)
	}
	return m, err
}
