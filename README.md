# p2p

The `p2p` package provides a low-level networking abstraction for peer-to-peer communication in the Lux ecosystem.

## Installation

```bash
go get github.com/luxfi/p2p
```

## Core Interfaces

### Handler

The `Handler` interface defines how incoming messages are processed:

```go
type Handler interface {
    // Gossip handles an incoming gossip message
    Gossip(ctx context.Context, nodeID ids.NodeID, gossipBytes []byte)
    // Request handles an incoming request and returns a response or error
    Request(ctx context.Context, nodeID ids.NodeID, deadline time.Time, requestBytes []byte) ([]byte, *Error)
}
```

### Sender

The `Sender` interface defines how outgoing messages are sent:

```go
type Sender interface {
    SendRequest(ctx context.Context, nodeIDs set.Set[ids.NodeID], requestID uint32, request []byte) error
    SendResponse(ctx context.Context, nodeID ids.NodeID, requestID uint32, response []byte) error
    SendError(ctx context.Context, nodeID ids.NodeID, requestID uint32, errorCode int32, errorMessage string) error
    SendGossip(ctx context.Context, config SendConfig, msg []byte) error
}
```

### Error

The `Error` type represents application-level errors:

```go
type Error struct {
    Code    int32
    Message string
}
```

Standard errors:
- `ErrUnexpected` (-1): Generic unexpected error
- `ErrUnregisteredHandler` (-2): No handler registered for protocol
- `ErrNotValidator` (-3): Requesting peer is not a validator
- `ErrThrottled` (-4): Request rate limit exceeded

## Sub-packages

### gossip

Implements gossip protocols for efficient data propagation:

```go
import "github.com/luxfi/p2p/gossip"

// Create a push gossiper
gossiper := gossip.NewPushGossiper[MyGossipable](...)

// Start gossiping
gossiper.Gossip(ctx)
```

### lp118

Implements LP-118 warp message signature handling:

```go
import "github.com/luxfi/p2p/lp118"

// Create a signature aggregator
aggregator := lp118.NewSignatureAggregator(log, client)

// Aggregate signatures for a warp message
msg, stake, total, err := aggregator.AggregateSignatures(ctx, message, justification, validators, quorumNum, quorumDen)
```

## Usage with warp

The `warp` package re-exports types from `p2p` for backward compatibility:

```go
import "github.com/luxfi/warp"

// warp.Error is an alias for p2p.Error
// warp.Sender is an alias for p2p.Sender
// warp.SendConfig is an alias for p2p.SendConfig
```

## License

See LICENSE file.
