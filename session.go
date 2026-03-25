package fixzero

import (
	"sync"
	"time"
)

type SessionID struct {
	BeginString  string
	SenderCompID string
	TargetCompID string
}

func (s SessionID) String() string {
	return s.BeginString + ":" + s.SenderCompID + "->" + s.TargetCompID
}

type SessionState struct {
	SessionID    SessionID
	SenderCompID string
	TargetCompID string
	BeginString  string

	OutMsgSeqNum int
	InMsgSeqNum  int

	LastSentTime time.Time
	LastRecvTime time.Time
	HeartBtInt   int

	LogonTime time.Time
	IsLogon   bool
}

func (s *SessionState) Reset() {
	s.SessionID = SessionID{}
	s.SenderCompID = ""
	s.TargetCompID = ""
	s.BeginString = ""
	s.OutMsgSeqNum = 0
	s.InMsgSeqNum = 0
	s.LastSentTime = time.Time{}
	s.LastRecvTime = time.Time{}
	s.HeartBtInt = 0
	s.LogonTime = time.Time{}
	s.IsLogon = false
}

func (s *SessionState) IncrOutMsgSeqNum() {
	s.OutMsgSeqNum++
}

func (s *SessionState) IncrInMsgSeqNum() {
	s.InMsgSeqNum++
}

func (s *SessionState) SetOutMsgSeqNum(n int) {
	s.OutMsgSeqNum = n
}

func (s *SessionState) SetInMsgSeqNum(n int) {
	s.InMsgSeqNum = n
}

func (s *SessionState) UpdateLastSent() {
	s.LastSentTime = time.Now()
}

func (s *SessionState) UpdateLastRecv() {
	s.LastRecvTime = time.Now()
}

func (s *SessionState) NeedHeartbeat() bool {
	if !s.IsLogon || s.HeartBtInt <= 0 {
		return false
	}
	elapsed := time.Since(s.LastSentTime)
	return elapsed >= time.Duration(s.HeartBtInt)*time.Second
}

func (s *SessionState) CheckHeartbeat() error {
	if !s.IsLogon || s.HeartBtInt <= 0 {
		return nil
	}
	elapsed := time.Since(s.LastRecvTime)
	timeout := time.Duration(s.HeartBtInt)*time.Second*2 + time.Second*10
	if elapsed > timeout {
		return &HeartbeatTimeoutError{Elapsed: elapsed, Timeout: timeout}
	}
	return nil
}

type HeartbeatTimeoutError struct {
	Elapsed time.Duration
	Timeout time.Duration
}

func (e *HeartbeatTimeoutError) Error() string {
	return "heartbeat timeout"
}

func (s *SessionState) CheckGap(newSeqNum int) (bool, int) {
	if newSeqNum > s.InMsgSeqNum+1 {
		return true, newSeqNum - s.InMsgSeqNum - 1
	}
	return false, 0
}

func (s *SessionState) CheckSeqNum(newSeqNum int) (bool, error) {
	if newSeqNum < s.InMsgSeqNum {
		return false, &DuplicateError{Sequence: newSeqNum, Expected: s.InMsgSeqNum + 1}
	}
	if newSeqNum > s.InMsgSeqNum+1 {
		return false, &GapError{Sequence: newSeqNum, Expected: s.InMsgSeqNum + 1}
	}
	return true, nil
}

type DuplicateError struct {
	Sequence int
	Expected int
}

func (e *DuplicateError) Error() string {
	return "duplicate message"
}

type GapError struct {
	Sequence int
	Expected int
}

func (e *GapError) Error() string {
	return "sequence gap detected"
}

func (s *SessionState) HandleResendRequest(start, end int, getMessage func(int) []byte) [][]byte {
	if end == 0 {
		end = s.OutMsgSeqNum
	}
	var results [][]byte
	for i := start; i <= end && i <= s.OutMsgSeqNum; i++ {
		msg := getMessage(i)
		if msg != nil {
			results = append(results, msg)
		}
	}
	return results
}

func (s *SessionState) HandleSequenceReset(newSeqNum int, gapFill bool) {
	if gapFill {
		s.InMsgSeqNum = newSeqNum
	} else {
		s.OutMsgSeqNum = newSeqNum
	}
}

var sessionPool = sync.Pool{
	New: func() interface{} {
		return &SessionState{}
	},
}

func GetSessionState() *SessionState {
	ss := sessionPool.Get().(*SessionState)
	ss.Reset()
	return ss
}

func PutSessionState(ss *SessionState) {
	if ss != nil {
		ss.Reset()
		sessionPool.Put(ss)
	}
}

type ResendRequest struct {
	BeginSeqNo int
	EndSeqNo   int
}

func ParseResendRequest(msg *Message) *ResendRequest {
	rr := &ResendRequest{}
	for i := 0; i < msg.numFields; i++ {
		field := &msg.fields[i]
		switch field.Tag {
		case TagMsgSeqNum:
			if v, ok := parseInt(field.Value); ok {
				rr.BeginSeqNo = v
			}
		case 433:
			if v, ok := parseInt(field.Value); ok {
				rr.EndSeqNo = v
			}
		}
	}
	return rr
}

func parseInt(s string) (int, bool) {
	var n int
	var neg bool
	for i, c := range s {
		if c == '-' && i == 0 {
			neg = true
			continue
		}
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		n = -n
	}
	return n, true
}

func BuildResendRequest(beginSeqNo, endSeqNo int) *Message {
	msg := GetMessage()
	msg.BeginString = "FIX.4.4"
	msg.MsgType = MsgTypeResendRequest
	msg.AddField(TagMsgSeqNum, formatIntNoAlloc(beginSeqNo))
	if endSeqNo > 0 {
		msg.AddField(Tag(433), formatIntNoAlloc(endSeqNo))
	}
	return msg
}

func BuildSequenceReset(newSeqNum int, gapFill bool) *Message {
	msg := GetMessage()
	msg.BeginString = "FIX.4.4"
	msg.MsgType = MsgTypeSequenceReset
	msg.AddField(TagMsgSeqNum, formatIntNoAlloc(newSeqNum))
	if gapFill {
		msg.AddField(Tag(141), "Y")
	}
	return msg
}

func BuildLogon(heartbeat int, resetSeq bool) *Message {
	msg := GetMessage()
	msg.BeginString = "FIX.4.4"
	msg.MsgType = "A"
	msg.AddField(TagHeartBtInt, formatIntNoAlloc(heartbeat))
	if resetSeq {
		msg.AddField(TagResetSeqNumFlag, "Y")
	}
	return msg
}

func BuildLogout(reason string) *Message {
	msg := GetMessage()
	msg.BeginString = "FIX.4.4"
	msg.MsgType = MsgTypeLogout
	if reason != "" {
		msg.AddField(TagText, reason)
	}
	return msg
}

func BuildHeartbeat() *Message {
	msg := GetMessage()
	msg.BeginString = "FIX.4.4"
	msg.MsgType = MsgTypeHeartbeat
	return msg
}

func BuildTestRequest(testReqID string) *Message {
	msg := GetMessage()
	msg.BeginString = "FIX.4.4"
	msg.MsgType = MsgTypeTestRequest
	msg.AddField(Tag(112), testReqID)
	return msg
}

func formatIntNoAlloc(v int) string {
	if v == 0 {
		return "0"
	}
	if v < 0 {
		return "-" + formatUintNoAlloc(uint64(-v))
	}
	return formatUintNoAlloc(uint64(v))
}

func formatUintNoAlloc(v uint64) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte(v%10) + '0'
		v /= 10
	}
	return string(buf[i:])
}
