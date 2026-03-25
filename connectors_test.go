package fixzero

import (
	"testing"
)

type testApp struct{}

func (t *testApp) OnCreate(sessionID SessionID)                  {}
func (t *testApp) OnLogon(sessionID SessionID)                   {}
func (t *testApp) OnLogout(sessionID SessionID)                  {}
func (t *testApp) ToAdmin(msg *Message, sessionID SessionID)     {}
func (t *testApp) ToApp(msg *Message, sessionID SessionID) error { return nil }
func (t *testApp) FromAdmin(msg *Message, sessionID SessionID) MessageRejectError {
	return nil
}
func (t *testApp) FromApp(msg *Message, sessionID SessionID) MessageRejectError {
	return nil
}

func TestNewInitiator(t *testing.T) {
	app := &testApp{}
	settings := &Settings{
		GlobalSessions: map[SessionID]*SessionSettings{
			{SenderCompID: "SENDER", TargetCompID: "TARGET"}: {
				SessionID:        SessionID{SenderCompID: "SENDER", TargetCompID: "TARGET"},
				Host:             "localhost",
				Port:             9998,
				ReconnectEnabled: false,
			},
		},
	}

	initiator := NewInitiator(app, settings)
	if initiator == nil {
		t.Error("NewInitiator returned nil")
	}

	if initiator.app == nil {
		t.Error("Initiator app is nil")
	}

	if initiator.settings == nil {
		t.Error("Initiator settings is nil")
	}

	if initiator.sessions == nil {
		t.Error("Initiator sessions map is nil")
	}
}

func TestNewAcceptor(t *testing.T) {
	app := &testApp{}
	settings := &Settings{
		GlobalSessions: map[SessionID]*SessionSettings{
			{SenderCompID: "SENDER", TargetCompID: "TARGET"}: {
				SessionID: SessionID{SenderCompID: "SENDER", TargetCompID: "TARGET"},
				Host:      "localhost",
				Port:      9999,
			},
		},
	}

	acceptor := NewAcceptor(app, settings)
	if acceptor == nil {
		t.Error("NewAcceptor returned nil")
	}

	if acceptor.app == nil {
		t.Error("Acceptor app is nil")
	}

	if acceptor.settings == nil {
		t.Error("Acceptor settings is nil")
	}

	if acceptor.sessions == nil {
		t.Error("Acceptor sessions map is nil")
	}
}
