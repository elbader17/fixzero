package fixzero

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultHeartbeatInterval = 30
	DefaultReconnectInterval = 30
	DefaultConnectTimeout    = 10 * time.Second
	DefaultReadTimeout       = 30 * time.Second
	DefaultWriteTimeout      = 30 * time.Second
)

var (
	ErrNotConnected     = errors.New("not connected")
	ErrConnectionClosed = errors.New("connection closed")
	ErrSessionNotFound  = errors.New("session not found")
)

type MessageRejectError error

type Settings struct {
	GlobalSessions map[SessionID]*SessionSettings
}

type SessionSettings struct {
	SessionID SessionID

	Host              string
	Port              int
	ReconnectEnabled  bool
	ReconnectInterval int

	HeartbeatInterval int

	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration

	TLSConfig *tls.Config
}

func (s *SessionSettings) Addr() string {
	return net.JoinHostPort(s.Host, formatUintNoAlloc(uint64(s.Port)))
}

func (s *SessionSettings) Heartbeat() int {
	if s.HeartbeatInterval > 0 {
		return s.HeartbeatInterval
	}
	return DefaultHeartbeatInterval
}

func (s *SessionSettings) Reconnect() int {
	if s.ReconnectInterval > 0 {
		return s.ReconnectInterval
	}
	return DefaultReconnectInterval
}

func (s *SessionSettings) ConnectDuration() time.Duration {
	if s.ConnectTimeout > 0 {
		return s.ConnectTimeout
	}
	return DefaultConnectTimeout
}

func (s *SessionSettings) ReadDuration() time.Duration {
	if s.ReadTimeout > 0 {
		return s.ReadTimeout
	}
	return DefaultReadTimeout
}

func (s *SessionSettings) WriteDuration() time.Duration {
	if s.WriteTimeout > 0 {
		return s.WriteTimeout
	}
	return DefaultWriteTimeout
}

type Application interface {
	OnCreate(sessionID SessionID)
	OnLogon(sessionID SessionID)
	OnLogout(sessionID SessionID)
	ToAdmin(msg *Message, sessionID SessionID)
	ToApp(msg *Message, sessionID SessionID) error
	FromAdmin(msg *Message, sessionID SessionID) MessageRejectError
	FromApp(msg *Message, sessionID SessionID) MessageRejectError
}

type ApplicationFuncs struct {
	OnCreateFunc  func(sessionID SessionID)
	OnLogonFunc   func(sessionID SessionID)
	OnLogoutFunc  func(sessionID SessionID)
	ToAdminFunc   func(msg *Message, sessionID SessionID)
	ToAppFunc     func(msg *Message, sessionID SessionID) error
	FromAdminFunc func(msg *Message, sessionID SessionID) MessageRejectError
	FromAppFunc   func(msg *Message, sessionID SessionID) MessageRejectError
}

func (a *ApplicationFuncs) OnCreate(sessionID SessionID) {
	if a.OnCreateFunc != nil {
		a.OnCreateFunc(sessionID)
	}
}

func (a *ApplicationFuncs) OnLogon(sessionID SessionID) {
	if a.OnLogonFunc != nil {
		a.OnLogonFunc(sessionID)
	}
}

func (a *ApplicationFuncs) OnLogout(sessionID SessionID) {
	if a.OnLogoutFunc != nil {
		a.OnLogoutFunc(sessionID)
	}
}

func (a *ApplicationFuncs) ToAdmin(msg *Message, sessionID SessionID) {
	if a.ToAdminFunc != nil {
		a.ToAdminFunc(msg, sessionID)
	}
}

func (a *ApplicationFuncs) ToApp(msg *Message, sessionID SessionID) error {
	if a.ToAppFunc != nil {
		return a.ToAppFunc(msg, sessionID)
	}
	return nil
}

func (a *ApplicationFuncs) FromAdmin(msg *Message, sessionID SessionID) MessageRejectError {
	if a.FromAdminFunc != nil {
		return a.FromAdminFunc(msg, sessionID)
	}
	return nil
}

func (a *ApplicationFuncs) FromApp(msg *Message, sessionID SessionID) MessageRejectError {
	if a.FromAppFunc != nil {
		return a.FromAppFunc(msg, sessionID)
	}
	return nil
}

type Initiator struct {
	app      Application
	settings *Settings

	sessions    map[SessionID]*InitiatorSession
	sessionsMut sync.RWMutex

	dialer *net.Dialer

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	started atomic.Bool
	stopped atomic.Bool
}

type InitiatorSession struct {
	settings     *SessionSettings
	conn         net.Conn
	connected    atomic.Bool
	connectedMut sync.Mutex

	sessionState *SessionState

	sendCh chan []byte
	recvCh chan []byte
	doneCh chan struct{}

	reconnectAttempts int
	reconnectMut      sync.Mutex
}

func NewInitiator(app Application, settings *Settings) *Initiator {
	ctx, cancel := context.WithCancel(context.Background())
	return &Initiator{
		app:      app,
		settings: settings,
		sessions: make(map[SessionID]*InitiatorSession),
		dialer: &net.Dialer{
			Timeout: DefaultConnectTimeout,
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

func (i *Initiator) Start() error {
	if i.started.Swap(true) {
		return errors.New("initiator already started")
	}
	i.stopped.Store(false)

	if i.settings == nil || i.settings.GlobalSessions == nil {
		return errors.New("no sessions configured")
	}

	for sessionID, sessionSettings := range i.settings.GlobalSessions {
		i.app.OnCreate(sessionID)

		is := &InitiatorSession{
			settings: sessionSettings,
			sendCh:   make(chan []byte, 256),
			recvCh:   make(chan []byte, 256),
			doneCh:   make(chan struct{}),
		}
		is.sessionState = GetSessionState()
		is.sessionState.SessionID = sessionID
		is.sessionState.HeartBtInt = sessionSettings.Heartbeat()

		i.sessionsMut.Lock()
		i.sessions[sessionID] = is
		i.sessionsMut.Unlock()

		i.wg.Add(1)
		go i.runSession(is, sessionID)
	}

	return nil
}

func (i *Initiator) Stop() error {
	if !i.started.Load() {
		return errors.New("initiator not started")
	}
	if i.stopped.Swap(true) {
		return errors.New("initiator already stopped")
	}

	i.cancel()
	i.wg.Wait()

	i.sessionsMut.Lock()
	for _, is := range i.sessions {
		if is.conn != nil {
			is.conn.Close()
		}
		close(is.sendCh)
		close(is.recvCh)
		close(is.doneCh)
		PutSessionState(is.sessionState)
	}
	i.sessions = nil
	i.sessionsMut.Unlock()

	return nil
}

func (i *Initiator) Connect(sessionID SessionID) error {
	i.sessionsMut.RLock()
	is, ok := i.sessions[sessionID]
	i.sessionsMut.RUnlock()

	if !ok {
		return ErrSessionNotFound
	}

	return i.connectSession(is)
}

func (i *Initiator) connectSession(is *InitiatorSession) error {
	is.connectedMut.Lock()
	defer is.connectedMut.Unlock()

	if is.conn != nil {
		return nil
	}

	conn, err := i.dialWithTLS(is.settings)
	if err != nil {
		return err
	}

	is.conn = conn
	is.connected.Store(true)

	i.wg.Add(1)
	go i.readLoop(is)
	i.wg.Add(1)
	go i.writeLoop(is)

	return nil
}

func (i *Initiator) dialWithTLS(settings *SessionSettings) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(i.ctx, settings.ConnectDuration())
	defer cancel()

	if settings.TLSConfig != nil {
		dialer := &net.Dialer{}
		return dialer.DialContext(ctx, "tcp", settings.Addr())
	}

	return i.dialer.DialContext(ctx, "tcp", settings.Addr())
}

func (i *Initiator) runSession(is *InitiatorSession, sessionID SessionID) {
	defer i.wg.Done()

	for {
		select {
		case <-i.ctx.Done():
			return
		case <-is.doneCh:
			return
		default:
		}

		if err := i.connectSession(is); err != nil {
			i.scheduleReconnect(is)
			continue
		}

		select {
		case <-i.ctx.Done():
			return
		case <-is.doneCh:
			return
		}
	}
}

func (i *Initiator) scheduleReconnect(is *InitiatorSession) {
	is.reconnectMut.Lock()
	interval := time.Duration(is.settings.Reconnect()) * time.Second
	attempts := is.reconnectAttempts
	is.reconnectAttempts++
	is.reconnectMut.Unlock()

	backoff := interval
	if attempts > 0 {
		backoff = interval + time.Duration(attempts*5)*time.Second
		if backoff > 5*interval {
			backoff = 5 * interval
		}
	}

	select {
	case <-i.ctx.Done():
		return
	case <-time.After(backoff):
	}
}

func (i *Initiator) readLoop(is *InitiatorSession) {
	defer i.wg.Done()

	for {
		select {
		case <-i.ctx.Done():
			return
		case <-is.doneCh:
			return
		default:
		}

		is.conn.SetReadDeadline(time.Now().Add(is.settings.ReadDuration()))

		data, err := readFrame(is.conn)
		if err != nil {
			i.handleDisconnect(is)
			return
		}

		select {
		case is.recvCh <- data:
		case <-i.ctx.Done():
			return
		case <-is.doneCh:
			return
		}
	}
}

func (i *Initiator) writeLoop(is *InitiatorSession) {
	defer i.wg.Done()

	for {
		select {
		case <-i.ctx.Done():
			return
		case <-is.doneCh:
			return
		case data := <-is.sendCh:
			is.conn.SetWriteDeadline(time.Now().Add(is.settings.WriteDuration()))
			if _, err := is.conn.Write(data); err != nil {
				i.handleDisconnect(is)
				return
			}
		}
	}
}

func (i *Initiator) handleDisconnect(is *InitiatorSession) {
	is.connectedMut.Lock()
	if is.conn != nil {
		is.conn.Close()
		is.conn = nil
	}
	is.connected.Store(false)
	is.connectedMut.Unlock()

	is.sessionState.IsLogon = false
	close(is.doneCh)
	is.doneCh = make(chan struct{})
}

func (i *Initiator) Send(sessionID SessionID, msg *Message) error {
	i.sessionsMut.RLock()
	is, ok := i.sessions[sessionID]
	i.sessionsMut.RUnlock()

	if !ok {
		return ErrSessionNotFound
	}

	is.connectedMut.Lock()
	connected := is.connected.Load()
	is.connectedMut.Unlock()

	if !connected {
		return ErrNotConnected
	}

	data := Serialize(msg)

	select {
	case is.sendCh <- data:
		return nil
	default:
		return errors.New("send channel full")
	}
}

func (i *Initiator) Receive(sessionID SessionID) ([]byte, error) {
	i.sessionsMut.RLock()
	is, ok := i.sessions[sessionID]
	i.sessionsMut.RUnlock()

	if !ok {
		return nil, ErrSessionNotFound
	}

	select {
	case data := <-is.recvCh:
		return data, nil
	case <-i.ctx.Done():
		return nil, i.ctx.Err()
	case <-is.doneCh:
		return nil, ErrConnectionClosed
	}
}

type Acceptor struct {
	app      Application
	settings *Settings

	listener net.Listener

	sessions    map[SessionID]*AcceptorSession
	sessionsMut sync.RWMutex

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	started atomic.Bool
	stopped atomic.Bool
}

type AcceptorSession struct {
	conn net.Conn
	addr string

	settings *SessionSettings

	sendCh chan []byte
	recvCh chan []byte
	doneCh chan struct{}
}

func NewAcceptor(app Application, settings *Settings) *Acceptor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Acceptor{
		app:      app,
		settings: settings,
		sessions: make(map[SessionID]*AcceptorSession),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (a *Acceptor) Start() error {
	if a.settings == nil || len(a.settings.GlobalSessions) == 0 {
		return errors.New("no sessions configured")
	}

	for sessionID := range a.settings.GlobalSessions {
		a.app.OnCreate(sessionID)
	}

	if a.settings.GlobalSessions != nil {
		for _, ss := range a.settings.GlobalSessions {
			if ss.Port == 0 {
				continue
			}

			var listener net.Listener
			var err error

			if ss.TLSConfig != nil {
				listener, err = tls.Listen("tcp", ss.Addr(), ss.TLSConfig)
			} else {
				listener, err = net.Listen("tcp", ss.Addr())
			}

			if err != nil {
				return err
			}

			a.listener = listener
			break
		}
	}

	if a.listener == nil {
		return errors.New("no valid listener configured")
	}

	if a.started.Swap(true) {
		return errors.New("acceptor already started")
	}
	a.stopped.Store(false)

	a.wg.Add(1)
	go a.acceptLoop()

	return nil
}

func (a *Acceptor) Stop() error {
	if !a.started.Load() {
		return errors.New("acceptor not started")
	}
	if a.stopped.Swap(true) {
		return errors.New("acceptor already stopped")
	}

	a.cancel()

	if a.listener != nil {
		a.listener.Close()
	}

	a.wg.Wait()

	a.sessionsMut.Lock()
	for _, as := range a.sessions {
		if as.conn != nil {
			as.conn.Close()
		}
		close(as.sendCh)
		close(as.recvCh)
		close(as.doneCh)
	}
	a.sessions = nil
	a.sessionsMut.Unlock()

	return nil
}

func (a *Acceptor) acceptLoop() {
	defer a.wg.Done()

	for {
		conn, err := a.listener.Accept()
		if err != nil {
			select {
			case <-a.ctx.Done():
				return
			default:
				continue
			}
		}

		as := &AcceptorSession{
			conn:   conn,
			addr:   conn.RemoteAddr().String(),
			sendCh: make(chan []byte, 256),
			recvCh: make(chan []byte, 256),
			doneCh: make(chan struct{}),
		}

		a.sessionsMut.Lock()
		for sessionID, settings := range a.settings.GlobalSessions {
			as.settings = settings
			a.sessions[sessionID] = as
			break
		}
		a.sessionsMut.Unlock()

		a.wg.Add(1)
		go a.handleConnection(as)
	}
}

func (a *Acceptor) handleConnection(as *AcceptorSession) {
	defer a.wg.Done()
	defer as.conn.Close()

	a.wg.Add(1)
	go a.acceptReadLoop(as)
	a.wg.Add(1)
	go a.acceptWriteLoop(as)

	<-as.doneCh
}

func (a *Acceptor) acceptReadLoop(as *AcceptorSession) {
	defer a.wg.Done()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-as.doneCh:
			return
		default:
		}

		as.conn.SetReadDeadline(time.Now().Add(as.settings.ReadDuration()))

		data, err := readFrame(as.conn)
		if err != nil {
			close(as.doneCh)
			return
		}

		select {
		case as.recvCh <- data:
		case <-a.ctx.Done():
			return
		case <-as.doneCh:
			return
		}
	}
}

func (a *Acceptor) acceptWriteLoop(as *AcceptorSession) {
	defer a.wg.Done()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-as.doneCh:
			return
		case data := <-as.sendCh:
			as.conn.SetWriteDeadline(time.Now().Add(as.settings.WriteDuration()))
			if _, err := as.conn.Write(data); err != nil {
				close(as.doneCh)
				return
			}
		}
	}
}

func (a *Acceptor) Receive(sessionID SessionID) ([]byte, error) {
	a.sessionsMut.RLock()
	as, ok := a.sessions[sessionID]
	a.sessionsMut.RUnlock()

	if !ok {
		return nil, ErrSessionNotFound
	}

	select {
	case data := <-as.recvCh:
		return data, nil
	case <-a.ctx.Done():
		return nil, a.ctx.Err()
	case <-as.doneCh:
		return nil, ErrConnectionClosed
	}
}

func (a *Acceptor) Send(sessionID SessionID, msg *Message) error {
	a.sessionsMut.RLock()
	as, ok := a.sessions[sessionID]
	a.sessionsMut.RUnlock()

	if !ok {
		return ErrSessionNotFound
	}

	if as.conn == nil {
		return ErrNotConnected
	}

	data := Serialize(msg)

	select {
	case as.sendCh <- data:
		return nil
	default:
		return errors.New("send channel full")
	}
}

var frameBufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 2+1024)
		return &buf
	},
}

func readFrame(conn net.Conn) ([]byte, error) {
	bufp := frameBufferPool.Get().(*[]byte)
	defer frameBufferPool.Put(bufp)
	return readFrameInto(conn, *bufp)
}

func readFrameInto(conn net.Conn, buf []byte) ([]byte, error) {
	if len(buf) < 2 {
		return nil, errors.New("buffer too small")
	}

	if _, err := conn.Read(buf[:2]); err != nil {
		return nil, err
	}

	if buf[0] == SOH {
		firstLine := buf[:0]

		tmp := make([]byte, 1)
		for {
			if _, err := conn.Read(tmp); err != nil {
				return nil, err
			}
			if tmp[0] == SOH {
				break
			}
			firstLine = append(firstLine, tmp[0])
		}

		return firstLine, nil
	}

	length := int(buf[0])<<8 | int(buf[1])

	if len(buf) < 2+length {
		return nil, errors.New("buffer too small")
	}

	if _, err := conn.Read(buf[2 : 2+length]); err != nil {
		return nil, err
	}

	return buf[:2+length], nil
}
