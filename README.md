# fixzero - Zero-Allocation FIX Protocol Library

**Ultra-high performance FIX protocol library for Go designed for HFT**

```
╔═══════════════════════════════════════════════════════════════════════════╗
║                        fixzero - ZERO-ALLOCATION                          ║
║                   The fastest FIX protocol library in Go                  ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

## ⚡ Características

- **Zero-Allocation** - 0 bytes allocations per message in critical path
- **Ultra-Low Latency** - 110ns parsing, 1.6ns field access
- **Zero-Copy** - Field views using unsafe pointer manipulation
- **Memory Pools** - sync.Pool for buffer and message reuse
- **Thread-Safe** - Fully concurrent operation
- **Full Protocol** - Complete FIX session management, validation, and connectors

## 🚀 Benchmarks

### Comparación con QuickFIX/Go

| Operación | fixzero | QuickFIX/Go | Mejora |
|-----------|---------|-------------|--------|
| **Parse** | 110 ns/op | 4,008 ns/op | **36x** |
| **Builder** | 156 ns/op | 8,654 ns/op | **55x** |
| **Field Access** | 1.6 ns/op | 384 ns/op | **244x** |
| **Serialize** | 639 ns/op | 2,689 ns/op | **4x** |

### Allocations por Operación

| Operación | fixzero | QuickFIX/Go |
|-----------|--------|-------------|
| Parse | **0 B, 0 allocs** | 2,112 B, 20 allocs |
| Field Access | **0 B, 0 allocs** | 64 B, 5 allocs |
| Serialize | 112 B, 1 alloc | 376 B, 10 allocs |
| SessionSeqNumIncr | **0 B, 0 allocs** | N/A |
| ValidateField | **0 B, 0 allocs** | N/A |

## 📦 Installation

```bash
go get github.com/fixzero/fixzero
```

## 🔧 Uso Básico

### Parse a Message

```go
package main

import (
    "fmt"
    fixzero "github.com/fixzero/fixzero"
)

func main() {
    // FIX message with SOH delimiter
    msgData := []byte("8=FIX.4.4\x019=0123\x0135=D\x0134=1\x0149=SENDER\x0156=TARGET\x0111=ORDER123\x0155=AAPL\x0154=1\x0140=2\x0138=100\x0144=150.50\x0110=001\x01")
    
    msg, err := fixzero.Parse(msgData)
    if err != nil {
        panic(err)
    }
    
    // Field access (zero-copy)
    fmt.Println("Symbol:", msg.GetString(fixzero.TagSymbol))    // AAPL
    fmt.Println("Price:", msg.GetFloat(fixzero.TagPrice))      // 150.50
    fmt.Println("Side:", msg.GetString(fixzero.TagSide))       // 1
    
    // Return to pool
    fixzero.PutMessage(msg)
}
```

### Build a Message

```go
msg := fixzero.NewBuilder().
    BeginString("FIX.4.4").
    MsgType("D").
    Set(fixzero.TagClOrdID, "ORDER123").
    Set(fixzero.TagSymbol, "AAPL").
    Set(fixzero.TagSide, fixzero.SideBuy).
    Set(fixzero.TagOrdType, fixzero.OrdTypeLimit).
    SetInt(fixzero.TagQuantity, 100).
    SetFloat(fixzero.TagPrice, 150.50).
    Build()

data := fixzero.Serialize(msg)
fixzero.PutMessage(msg)
```

### Field Access

```go
// String
symbol := msg.GetString(fixzero.TagSymbol)

// Integer  
qty, ok := msg.GetInt(fixzero.TagQuantity)

// Float
price, ok := msg.GetFloat(fixzero.TagPrice)

// Boolean
isBuy, ok := msg.GetBool(fixzero.TagSide)

// Verificar existencia
if msg.Has(fixzero.TagPrice) {
    // ...
}

// Iterate fields
msg.Iterate(func(tag fixzero.Tag, value string) bool {
    fmt.Printf("Tag %d = %s\n", tag, value)
    return true // continue iteration
})
```

---

## 🏢 Session Management

FIX session management with sequence tracking and heartbeat control.

```go
// Create session state
state := fixzero.GetSessionState()
state.SenderCompID = "SENDER"
state.TargetCompID = "TARGET"
state.BeginString = "FIX.4.4"
state.HeartBtInt = 30
state.OutMsgSeqNum = 1
state.InMsgSeqNum = 1

// Incrementar sequence numbers
state.IncrOutMsgSeqNum()
state.IncrInMsgSeqNum()

// Verificar heartbeat
if state.CheckHeartbeat() {
    fmt.Println("Heartbeat timeout!")
}

// Actualizar tiempos
state.UpdateLastSent()
state.UpdateLastRecv()

// Return to pool
fixzero.PutSessionState(state)
```

### Manejo de Resend Request

```go
// Generate messages for resend
msgs, err := state.HandleResendRequest(startSeq, endSeq, store)
for _, msg := range msgs {
    // Resend message
}
```

### Manejo de Sequence Reset

```go
// Ajustar secuencias
state.HandleSequenceReset(newSeqNum, gapFillFlag)
```

---

## 🔌 Network Connectors

### Initiator (Cliente)

```go
// Implementar Application
type MyApp struct{}

func (a *MyApp) OnCreate(sessionID fixzero.SessionID) {}
func (a *MyApp) OnLogon(sessionID fixzero.SessionID) {}
func (a *MyApp) OnLogout(sessionID fixzero.SessionID) {}
func (a *MyApp) ToAdmin(msg *fixzero.Message, sessionID fixzero.SessionID) {}
func (a *MyApp) ToApp(msg *fixzero.Message, sessionID fixzero.SessionID) error { return nil }
func (a *MyApp) FromAdmin(msg *fixzero.Message, sessionID fixzero.SessionID) fixzero.MessageRejectError { return nil }
func (a *MyApp) FromApp(msg *fixzero.Message, sessionID fixzero.SessionID) fixzero.MessageRejectError { return nil }

// Configure session
settings := &fixzero.SessionSettings{
    Host:             "localhost",
    Port:             "9876",
    SenderCompID:     "SENDER",
    TargetCompID:     "TARGET",
    BeginString:      "FIX.4.4",
    HeartBtInt:       30,
    ReconnectInterval: 5,
}

// Create initiator
initiator, err := fixzero.NewInitiator(&MyApp{}, settings)
if err != nil {
    panic(err)
}

// Start connections
err := initiator.Start()
// ...

// Detener
initiator.Stop()
```

### Acceptor (Servidor)

```go
// Create acceptor
acceptor, err := fixzero.NewAcceptor(&MyApp{}, settings)
if err != nil {
    panic(err)
}

// Iniciar servidor
err := acceptor.Start()
// ...

// Detener
acceptor.Stop()
```

### TLS Support

```go
settings := &fixzero.SessionSettings{
    Host: "localhost",
    Port: "9876",
    TLS: &tls.Config{
        MinVersion: tls.VersionTLS12,
        // ...
    },
}
```

---

## ✅ Message Validation

### Data Dictionary

```go
// Cargar especificación FIX
dd, err := fixzero.LoadDataDictionary("FIX44.xml")
if err != nil {
    panic(err)
}

// Validate message
msg, _ := fixzero.Parse(msgData)
errors := dd.Validate(msg)

if len(errors) > 0 {
    for _, err := range errors {
        fmt.Printf("Error: %s\n", err.Message)
    }
}
```

### Validación de Campos

```go
// Validate field type
err := fixzero.ValidateField(fixzero.TagPrice, "150.50", "DECIMAL")
// nil = válido

// Validate enumeration
err := fixzero.ValidateEnum("1", []string{"1", "2", "3", "4", "5"})
// nil = válido
```

### Tipos Soportados

```go
// STRING, CHAR, INT, UINT, FLOAT, BOOLEAN, DATE, TIME, DATETIME, DECIMAL
err := fixzero.ValidateField(tag, value, "INT")
```

---

## 📝 Message Types

### NewOrderSingle

```go
// Decode
msg, _ := fixzero.Parse(data)
nos := fixzero.NewOrderSingle{}
nos.Decode(msg)

// Access fields
fmt.Println(nos.ClOrdID)    // Order ID
fmt.Println(nos.Symbol)     // AAPL
fmt.Println(nos.Side)       // 1 (Buy)
fmt.Println(nos.OrdType)    // 2 (Limit)
fmt.Println(nos.Quantity)   // 100

// Encode
msg = nos.Encode()
data = fixzero.Serialize(msg)
```

### ExecutionReport

```go
er := fixzero.ExecutionReport{}
er.Decode(msg)
fmt.Println(er.OrderID)
fmt.Println(er.ExecID)
fmt.Println(er.OrdStatus)  // 0=New, 1=PartiallyFilled, 2=Filled, etc.
```

### Mensajes Soportados

| Tipo | Código | Descripción |
|------|--------|-------------|
| NewOrderSingle | D | Nueva orden |
| ExecutionReport | 8 | Reporte de ejecución |
| OrderCancelRequest | F | Solicitud de cancelación |
| OrderCancelReplace | G | Modificación de orden |
| OrderStatusRequest | H | Consulta de estado |
| OrderCancelReject | 9 | Rechazo de cancelación |
| Heartbeat | 0 | Heartbeat |
| TestRequest | 1 | Test request |
| ResendRequest | 2 | Resend messages |
| Reject | 3 | Rechazo |
| SequenceReset | 4 | Reset de secuencia |
| Logout | 5 | Logout |
| Logon | A | Logon |

---

## 💾 Message Store

### MemoryStore (En memoria)

```go
store, err := fixzero.NewMemoryStore()
if err != nil {
    panic(err)
}

// Save message
store.SaveMessage(1, msgData)

// Retrieve message
msg, err := store.GetMessage(1)

// Message range
msgs, err := store.GetRange(1, 100)

// Secuencias
store.SetNextSenderSeqNum(10)
senderSeq := store.GetNextSenderSeqNum()
store.IncrNextSenderSeqNum()

// Close
store.Close()
```

### FileStore (Archivos)

```go
store, err := fixzero.NewFileStore("./store")
if err != nil {
    panic(err)
}
// ... mismo API que MemoryStore
store.Close()
```

### SQLiteStore

```go
store, err := fixzero.NewSQLiteStore("./fixzero.db")
if err != nil {
    panic(err)
}
// ... mismo API que MemoryStore
store.Close()
```

---

## 📋 Logging

### NullLog (Descarta todo)

```go
log := fixzero.NewNullLog()
log.OnIncoming("8=FIX.4.4|35=D|")
log.OnOutgoing("8=FIX.4.4|35=8|")
log.OnEvent("Connected")
log.OnError("Error message")
```

### ScreenLog (Consola)

```go
log := fixzero.NewScreenLog(true) // true = with colors
log.OnIncoming("8=FIX.4.4|35=D|")
// Output: 2026-03-25 19:00:00 [INCOMING] 8=FIX.4.4|35=D|
```

### FileLog (Archivos)

```go
log, err := fixzero.NewFileLog("./logs")
if err != nil {
    panic(err)
}
log.OnIncoming("8=FIX.4.4|35=D|")
log.OnOutgoing("8=FIX.4.4|35=8|")
log.OnEvent("Event")
log.OnError("Error")
log.Close() // Close files
```

---

## 🔄 Repeating Groups

### Count Groups

```go
// Count NoPartyIDs groups (tag 453)
count := fixzero.CountGroups(msg, 453)
fmt.Printf("There are %d groups\n", count)
```

### Iterate Groups

```go
// Iterate all groups
fixzero.IterateGroups(msg, 453, func(idx int, fields []fixzero.Field) bool {
    fmt.Printf("Grupo %d:\n", idx)
    for _, f := range fields {
        fmt.Printf("  %d = %s\n", f.Tag, f.Value)
    }
    return true // continuar
})
```

### Get Specific Group

```go
// Get third group (index 2)
group := fixzero.GetGroup(msg, "453", 2)
```

### Validate Groups

```go
// Define group
groupDef := &fixzero.GroupDef{
    Tag:      453,
    NumField: 580, // NoPartyIDs
    Fields:   []int{448, 447, 452}, // PartyID, PartyIDSource, PartyRole
}

// Validate
errors := fixzero.ValidateGroup(msg, groupDef)
```

---

## 📚 API Reference

### Pool functions

```go
// Buffer pool
buf := fixzero.GetBuffer(1024)
fixzero.PutBuffer(buf)

// Message pool
msg := fixzero.GetMessage()
fixzero.PutMessage(msg)

// Session state pool
state := fixzero.GetSessionState()
fixzero.PutSessionState(state)
```

### Parser

```go
// Parse message
msg, err := fixzero.Parse(data)

// Parse into existing message (reuse)
err := fixzero.ParseInto(data, msg)

// Parser personalizado
parser := fixzero.NewParser()
msg, err := parser.Parse(data)
```

### Serializer

```go
// Serialize a nuevo buffer
data := fixzero.Serialize(msg)

// Serialize a buffer existente
data := fixzero.SerializeTo(msg, existingBuf)
```

### Builder

```go
// Create builder
b := fixzero.NewBuilder()

// Métodos chaining
msg := b.BeginString("FIX.4.4").
    MsgType("D").
    Set(tag, value).
    SetInt(tag, intValue).
    SetFloat(tag, floatValue).
    Build()

// Reuse builder
b.Reset()
```

### Field Accessors

```go
// Getters
msg.GetString(tag)      // string
msg.GetInt(tag)         // (int64, bool)
msg.GetUint(tag)        // (uint64, bool)  
msg.GetFloat(tag)       // (float64, bool)
msg.GetBool(tag)       // (bool, bool)
msg.GetBytes(tag)      // []byte (copy)

// Checkers
msg.Has(tag)            // bool

// Iteration
msg.Iterate(func(tag Tag, value string) bool)
```

---

## 🏷️ Constants

### Tags estándar

```go
fixzero.TagBeginString    // 8
fixzero.TagBodyLength    // 9
fixzero.TagMsgType       // 35
fixzero.TagSenderCompID  // 49
fixzero.TagTargetCompID  // 56
fixzero.TagMsgSeqNum     // 34
fixzero.TagSendingTime   // 52
fixzero.TagClOrdID       // 11
fixzero.TagSymbol        // 55
fixzero.TagSide         // 54
fixzero.TagOrdType      // 40
fixzero.TagQuantity     // 38
fixzero.TagPrice        // 44
fixzero.TagCheckSum     // 10
// ... and more in tags.go
```

### Message Types

```go
fixzero.MsgTypeNewOrderSingle    // D
fixzero.MsgTypeExecutionReport   // 8
fixzero.MsgTypeOrderCancelRequest // F
fixzero.MsgTypeOrderCancelReplace // G
fixzero.MsgTypeQuoteRequest      // R
fixzero.MsgTypeMarketDataRequest // V
fixzero.MsgTypeHeartbeat         // 0
fixzero.MsgTypeLogout            // 5
fixzero.MsgTypeLogon             // A
// ... y más
```

### Enums

```go
// Side
fixzero.SideBuy      // "1"
fixzero.SideSell     // "2"
fixzero.SideSellShort // "5"

// OrdType
fixzero.OrdTypeMarket    // "1"
fixzero.OrdTypeLimit    // "2"
fixzero.OrdTypeStop     // "3"
fixzero.OrdTypeStopLimit // "4"

// TimeInForce
fixzero.TIFF_Day     // "0"
fixzero.TIFF_IOC     // "1"
fixzero.TIFF_GTC     // "3"
fixzero.TIFF_GTD     // "4"
```

---

## 🎯 Applied Optimizations

1. **Array-based field indexing** - O(1) lookup sin map
2. **Unsafe pointer manipulation** - Zero-copy string views  
3. **Pre-allocated buffers** - Sin reallocations
4. **sync.Pool** - Reutilización de memoria
5. **Inline hints** - compiler optimization
6. **Hot path optimization** - Loop sin llamadas externas

---

## 📁 Project Structure

```
fixzero/
├── pool.go             # Buffer pools, Message struct
├── parser.go           # Zero-allocation parser
├── serializer.go       # Serializer optimizado
├── field.go            # Field accessors
├── msgpool.go          # Message pool + Builder
├── tags.go             # Tags, message types, enums
├── session.go          # Session management
├── connectors.go      # Initiator/Acceptor
├── validation.go       # Data dictionary, validation
├── msgtypes.go         # Message type structs
├── store.go            # MessageStore implementations
├── logging.go          # Log implementations
├── groups.go           # Repeating groups
├── fixzero_test.go     # Tests
└── compare/bench_test.go # Benchmarks
```

---

## ⚠️ Notes

- El delimitador de campos en FIX es SOH (0x01), no `|`
- Messages must keep reference to original while Message exists
- Usa `PutMessage()` para devolver al pool
- Todas las operaciones críticas mantienen zero-allocation

---

## 📜 License

PROPIETARY - Personal Use Only - Contact for Commercial License

---

**¿Necesitas más rendimiento?** La librería está optimizada para HFT con latencia ultra-baja.
