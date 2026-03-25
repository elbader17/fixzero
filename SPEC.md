# FIX Protocol Zero-Allocation Library

## Project Overview

- **Project Name**: fixzero
- **Project Type**: Go library for high-performance FIX protocol implementation
- **Core Functionality**: ZERO-ALLOCATION FIX message parsing, serialization, and validation for financial transactions
- **Target Users**: High-frequency trading systems, algorithmic trading platforms, financial exchanges

## Technical Architecture

### Zero-Allocation Strategy

1. **Stack-Only Processing**: All parsing/serialization uses stack-allocated buffers
2. **sync.Pool for Buffers**: Reusable byte slices for intermediate operations
3. **unsafe Pointers**: Direct memory manipulation without allocations
4. **Pre-allocated Message Pool**: Message structs pooled to avoid allocations
5. **Field Value Slicing**: Use unsafe to create string views without copying

### Memory Layout

```
Message Structure (Stack):
┌─────────────────────────────────────────────────────────┐
│ Header (24 bytes)                                       │
│ ├─ BeginString: [12]byte (FIX4.4, FIX5.0, etc)          │
│ ├─ BodyLength: uint32                                   │
│ └─ MsgType: [4]byte                                     │
├─────────────────────────────────────────────────────────┤
│ Fields (Inline, max 64 fields)                          │
│ ├─ Tag: uint16                                          │
│ ├─ Value: unsafe.String (zero-copy view)                │
│ └─ Length: uint32                                       │
├─────────────────────────────────────────────────────────┤
│ Trailer                                                  │
│ ├─ Checksum: [4]byte                                   │
│ └─ RawData (remaining bytes)                           │
└─────────────────────────────────────────────────────────┘
```

### Performance Targets

- **Parse Latency**: < 100ns per message
- **Serialize Latency**: < 150ns per message
- **Throughput**: > 10M messages/second
- **GC Pressure**: ZERO allocations per operation (sync.Pool only)

## Functionality Specification

### Core Components

1. **Buffer Pool** (`pool.go`)
   - `GetBuffer(size int) []byte` - Get pre-allocated buffer
   - `PutBuffer(buf []byte)` - Return to pool
   - Thread-safe sync.Pool implementation

2. **Message Parser** (`parser.go`)
   - `Parse(data []byte) (*Message, error)` - Parse FIX message
   - `ParseInto(data []byte, msg *Message) error` - Reuse existing message
   - Zero-copy field extraction using unsafe.String

3. **Message Serializer** (`serializer.go`)
   - `Serialize(msg *Message) []byte` - Serialize to bytes
   - `SerializeTo(msg *Message, buf []byte) []byte` - Use provided buffer
   - Pre-computed field lengths for speed

4. **Message Pool** (`msgpool.go`)
   - `GetMessage() *Message` - Get pooled message
   - `PutMessage(msg *Message)` - Return to pool
   - Pre-reserve field capacity

5. **Field Accessor** (`field.go`)
   - `GetString(msg *Message, tag Tag) (string, bool)`
   - `GetInt(msg *Message, tag Tag) (int64, bool)`
   - `GetFloat(msg *Message, tag Tag) (float64, bool)`
   - `GetBool(msg *Message, tag Tag) (bool, bool)`
   - All accessors use zero-copy views

6. **Message Builder** (`builder.go`)
   - Fluent API for building FIX messages
   - Pre-allocate field slice capacity
   - Pool messages on Build()

### Supported FIX Versions

- FIX 4.0
- FIX 4.1
- FIX 4.2
- FIX 4.3
- FIX 4.4
- FIX 5.0
- FIX 5.0 SP1
- FIX 5.0 SP2

### Supported Message Types

- **Admin**: Heartbeat, TestRequest, ResendRequest, Reject, SequenceReset, Logout
- **Trade**: NewOrderSingle, ExecutionReport, OrderCancelRequest, OrderCancelReplace, OrderStatusRequest
- **Quote**: QuoteRequest, QuoteResponse, QuoteCancel
- **Market Data**: MarketDataRequest, MarketDataSnapshotFullRefresh, MarketDataIncrementalRefresh

### FIX Tags Reference

```go
const (
    TagBeginString     Tag = 8   // FIX version
    TagBodyLength      Tag = 9   // Message length
    TagMsgType         Tag = 35  // Message type
    TagSenderCompID    Tag = 49  // Sender ID
    TagTargetCompID    Tag = 56  // Target ID
    TagMsgSeqNum       Tag = 34  // Sequence number
    TagSendingTime     Tag = 52  // Timestamp
    TagOrdID           Tag = 37   // Order ID
    TagClOrdID         Tag = 11   // Client Order ID
    TagSymbol          Tag = 55   // Instrument symbol
    TagSide            Tag = 54   // Buy/Sell
    TagOrdType         Tag = 40   // Order type
    TagPrice           Tag = 44   // Price
    TagQuantity        Tag = 38   // Quantity
    TagExecType        Tag = 150  // Execution type
    TagOrdStatus       Tag = 39   // Order status
    TagLeavesQty       Tag = 151  // Remaining qty
    TagCumQty          Tag = 14   // Filled qty
    TagAvgPx           Tag = 6    // Average price
    TagChecksum        Tag = 10   // Checksum (trailer)
)
```

## Implementation Details

### Zero-Copy Parsing Algorithm

```
1. Validate SOH positions (every 2nd character should be =)
2. For each field:
   a. Find = position
   b. Parse tag number (convert string to uint16)
   c. Use unsafe.String to create string view from value
   d. Store (tag, value_ptr, value_len) - NO COPY
3. Validate checksum (last field must be tag 10)
4. Return message with all field views
```

### Serializer Optimization

```
1. Pre-calculate total message length
2. Use stack buffer or pooled buffer
3. Copy strings directly (these are already in memory)
4. Write SOH and = bytes directly
5. Compute and append checksum
```

### Buffer Pool Implementation

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        buf := make([]byte, 0, 4096)  // Pre-allocate 4KB
        return &buf
    },
}

func GetBuffer(size int) []byte {
    b := bufferPool.Get().(*[]byte)
    if cap(*b) < size {
        *b = make([]byte, 0, size)  // Only allocate if needed
    }
    *b = (*b)[:0]
    return *b
}
```

## API Specification

```go
// Pool - Buffer management
func fixzero.GetBuffer(size int) []byte
func fixzero.PutBuffer(buf []byte)

// Parser - Parse FIX messages
func fixzero.Parse(data []byte) (*fixzero.Message, error)
func fixzero.ParseInto(data []byte, msg *fixzero.Message) error

// Serializer - Serialize FIX messages  
func fixzero.Serialize(msg *Message) []byte
func fixzero.SerializeTo(msg *Message, buf []byte) []byte
func fixzero.SerializeInto(msg *Message, buf *[]byte)

// Message Pool
func fixzero.GetMessage() *Message
func fixzero.PutMessage(msg *Message)

// Field Accessors
func (m *Message) GetString(tag Tag) (string, bool)
func (m *Message) GetInt(tag Tag) (int64, bool)
func (m *Message) GetFloat(tag Tag) (float64, bool)
func (m *Message) GetBool(tag Tag) (bool, bool)
func (m *Message) Has(tag Tag) bool
func (m *Message) Iterate(func(tag Tag, value string))

// Builder
func fixzero.NewBuilder() *Builder
func (b *Builder) Set(tag Tag, value string) *Builder
func (b *Builder) SetInt(tag Tag, value int64) *Builder
func (b *Builder) SetFloat(tag Tag, value float64) *Builder
func (b *Builder) Build() *Message
```

## Benchmarks

Target metrics:
- Parse: 50-100ns per message
- Serialize: 100-150ns per message
- Field access: < 10ns per field
- Memory: 0 heap allocations (except initial pool)

## Error Handling

- Parse errors return specific error codes:
  - `ErrInvalidFormat` - Malformed message
  - `ErrMissingRequired` - Required tag missing
  - `ErrInvalidTag` - Invalid tag number
  - `ErrInvalidChecksum` - Checksum mismatch
  - `ErrBufferOverflow` - Message too large

## Testing Requirements

1. Unit tests for all parsers/serializers
2. Fuzz testing for malformed input
3. Benchmarks comparing to baseline
4. Memory profiling to verify zero-allocation

