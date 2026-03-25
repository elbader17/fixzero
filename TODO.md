# TODO - Funcionalidades para equiparar a QuickFIX/Go

## Visión General

fixzero es actualmente un parser/serializer de bajo nivel ultra-rápido. Para equiparar funcionalidad con QuickFIX/Go, se necesitan las siguientes adiciones.

---

## 🔴 Alta Prioridad (Sesión/Red)

### 1. Session Management

```go
// Estado requerido por sesión
type SessionState struct {
    SessionID     SessionID
    SenderCompID  string
    TargetCompID  string
    BeginString   string
    
    // Sequence numbers
    OutMsgSeqNum  int
    InMsgSeqNum   int
    
    // Heartbeat
    LastSentTime  time.Time
    LastRecvTime  time.Time
    HeartBtInt    int // segundos
    
    // Estado
    LogonTime     time.Time
    IsLogon       bool
}
```

**Features necesarios:**
- [ ] Sequence number increment/decrement
- [ ] Sequence number validation (check for gaps, duplicates)
- [ ] ResendRequest handling
- [ ] SequenceReset message handling
- [ ] Gap fill detection
- [ ] Heartbeat timeout detection

### 2. Network Connectors

```go
// Initiator - Cliente
type Initiator struct {
    sessions    map[SessionID]*Session
    settings    *SessionSettings
    dialer      *net.Dialer
}

// Acceptor - Servidor  
type Acceptor struct {
    sessions    map[SessionID]*Session
    listener    net.Listener
}
```

**Features necesarios:**
- [ ] TCP connection management
- [ ] Auto reconnection
- [ ] TLS support
- [ ] Connection timeout handling
- [ ] Message framing (reading until SOH or length)

---

## 🟡 Media Prioridad (Validación)

### 3. Message Validation

```go
// Data Dictionary
type DataDictionary struct {
    messageTypes map[string]*MessageDef
    fields      map[int]*FieldDef
}

type MessageDef struct {
    name        string
    required    []int
    repeatingGroups []*RepeatingGroup
}
```

**Features necesarios:**
- [ ] Field presence validation
- [ ] Type validation (string, int, float, etc.)
- [ ] Enum validation
- [ ] Required field checking
- [ ] Repeating group validation

### 4. Message Type Definitions

```go
// Código generado para cada mensaje
type NewOrderSingle struct {
    Header  Header
    Body    NewOrderSingleBody
    Trailer Trailer
}

type NewOrderSingleBody struct {
    ClOrdID      string
    Symbol       string  
    Side         string
    TransactTime string
    OrdType      string
    // ... campos
}
```

**Mensajes a soportar:**
- [ ] NewOrderSingle (D)
- [ ] ExecutionReport (8)
- [ ] OrderCancelRequest (F)
- [ ] OrderCancelReplace (G)
- [ ] OrderStatusRequest (H)
- [ ] OrderCancelReject (9)
- [ ] Heartbeat (0)
- [ ] TestRequest (1)
- [ ] ResendRequest (2)
- [ ] Reject (3)
- [ ] SequenceReset (4)
- [ ] Logout (5)
- [ ] Logon (A)

---

## 🟢 Baja Prioridad (Almacenamiento/Logging)

### 5. Message Store

```go
type MessageStore interface {
    SaveMessage(seqNum int, msg []byte) error
    GetMessage(seqNum int) ([]byte, error)
    GetRange(start, end int) ([][]byte, error)
    SetNextSenderSeqNum(num int)
    SetNextTargetSeqNum(num int)
    GetNextSenderSeqNum() int
    GetNextTargetSeqNum() int
    IncrNextSenderSeqNum()
    IncrNextTargetSeqNum()
    Refresh() error
    Close() error
}
```

**Implementaciones:**
- [ ] MemoryStore (in-memory)
- [ ] FileStore (disk-based)
- [ ] SQLiteStore

### 6. Logging

```go
type Log interface {
    OnIncoming(string)
    OnOutgoing(string)
    OnEvent(string)
    OnError(string)
}
```

**Implementaciones:**
- [x] NullLog
- [x] ScreenLog
- [x] FileLog

### 7. Repeating Groups

```go
type RepeatingGroup struct {
    tag         int
    numField    int
    fields      [][]int
}
```

**Features:**
- [ ] Group counting
- [ ] Group iteration
- [ ] Group validation

---

## 📋 Estado Actual

```
Funcionalidad              Status
──────────────────────────────────────
Parse/Serialize            ✅ Completo
Field Access               ✅ Completo
Zero-Allocation            ✅ Completo
Memory Pools               ✅ Completo
Session Management        🔴 Pendiente
Network Connectors        🔴 Pendiente
Message Validation        🟡 Parcial
Message Store             🟢 Pendiente
Logging                   🟢 Pendiente
Repeating Groups          🟢 Pendiente
```

---

## 🎯 Estrategia de Implementación

### Fase 1: Core (Ya hecho)
- Parser/Serializer
- Field access
- Memory pools

### Fase 2: Session (Alta prioridad)
- Session state management
- Sequence handling
- Network connectors

### Fase 3: Validation (Media)
- Data dictionary
- Field validation
- Message type definitions

### Fase 4: Storage (Baja)
- Message stores
- Logging
- Repeating groups

---

## 🔧 API Propuesta para Session

```go
// Session handling
type Session struct {
    state    *SessionState
    store    MessageStore
    log      Log
    encode   *Encoder
    decode   *Decoder
}

func NewSession(cfg SessionConfig) *Session
func (s *Session) Connect() error
func (s *Session) Send(msg *Message) error
func (s *Session) Receive() (*Message, error)
func (s *Session) Close() error

// Session management
func (s *Session) SendLogon() error
func (s *Session) SendLogout() error  
func (s *Session) SendHeartbeat() error
func (s *Session) HandleResendRequest(start, end int) error

// Initiator
func NewInitiator(app Application, settings *Settings) (*Initiator, error)
func (i *Initiator) Start() error
func (i *Initiator) Stop() error

// Acceptor
func NewAcceptor(app Application, settings *Settings) (*Acceptor, error)  
func (a *Acceptor) Start() error
func (a *Acceptor) Stop() error

// Application interface (para callbacks)
type Application interface {
    OnCreate(sessionID SessionID)
    OnLogon(sessionID SessionID)
    OnLogout(sessionID SessionID)
    ToAdmin(msg *Message, sessionID SessionID)
    ToApp(msg *Message, sessionID SessionID) error
    FromAdmin(msg *Message, sessionID SessionID) MessageRejectError
    FromApp(msg *Message, sessionID SessionID) MessageRejectError
}
```

---

## 📝 Notas de Diseño

1. **Mantener zero-allocation** - No añadir allocations en hot path
2. **API familiar** - Mantener consistencia con QuickFIX/Go donde sea posible
3. **Modular** - Permite usar solo componentes necesarios
4. **Thread-safe** - Sesiones concurrentes

---

## 🤝 Contribuir

¿Tienes implementaciones de estas funcionalidades? Pull requests bienvenidos.

