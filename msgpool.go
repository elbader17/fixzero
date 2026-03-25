package fixzero

import (
	"strconv"
	"sync"
	"unsafe"
)

// MessagePool provides pooled Message instances.
var MessagePool = sync.Pool{
	New: func() interface{} {
		msg := &Message{
			// Pre-initialize tagIndex with -1
		}
		// Initialize tagIndex to -1 for all slots
		for i := range msg.tagIndex {
			msg.tagIndex[i] = -1
		}
		return msg
	},
}

// GetMessage returns a pooled Message.
func GetMessage() *Message {
	msg := MessagePool.Get().(*Message)
	// Reset tagIndex to -1
	for i := 0; i < msg.numFields; i++ {
		msg.tagIndex[msg.fields[i].Tag] = -1
	}
	msg.numFields = 0
	msg.BeginString = ""
	msg.MsgType = ""
	msg.Checksum = ""
	msg.rawData = msg.rawData[:0]
	return msg
}

// PutMessage returns a Message to the pool.
func PutMessage(msg *Message) {
	if msg != nil {
		// Reset tagIndex
		for i := 0; i < msg.numFields; i++ {
			msg.tagIndex[msg.fields[i].Tag] = -1
		}
		msg.numFields = 0
		MessagePool.Put(msg)
	}
}

// Builder provides a fluent API for building FIX messages.
type Builder struct {
	msg         *Message
	beginString string
	msgType     string
}

// NewBuilder creates a new message Builder.
func NewBuilder() *Builder {
	return &Builder{
		msg: GetMessage(),
	}
}

// BeginString sets the BeginString field.
func (b *Builder) BeginString(v string) *Builder {
	b.beginString = v
	return b
}

// MsgType sets the MsgType field.
func (b *Builder) MsgType(v string) *Builder {
	b.msgType = v
	return b
}

// Set sets a field to a string value.
func (b *Builder) Set(tag Tag, value string) *Builder {
	switch tag {
	case TagBeginString:
		b.beginString = value
	case TagMsgType:
		b.msgType = value
	default:
		b.msg.AddField(tag, value)
	}
	return b
}

// SetInt sets a field to an integer value.
func (b *Builder) SetInt(tag Tag, value int64) *Builder {
	if tag != TagBeginString && tag != TagMsgType {
		b.msg.AddField(tag, formatInt(value))
	}
	return b
}

// SetFloat sets a field to a float value.
func (b *Builder) SetFloat(tag Tag, value float64) *Builder {
	if tag != TagBeginString && tag != TagMsgType {
		b.msg.AddField(tag, strconv.FormatFloat(value, 'f', -1, 64))
	}
	return b
}

// Build builds and returns the Message.
func (b *Builder) Build() *Message {
	b.msg.BeginString = b.beginString
	b.msg.MsgType = b.msgType
	return b.msg
}

// Reset resets the builder for reuse.
func (b *Builder) Reset() {
	// Return current message to pool
	if b.msg != nil {
		PutMessage(b.msg)
	}
	b.msg = GetMessage()
	b.beginString = ""
	b.msgType = ""
}

// formatInt formats int64 to string.
//go:inline
func formatInt(v int64) string {
	if v == 0 {
		return "0"
	}
	if v < 0 {
		return "-" + formatUint(uint64(-v))
	}
	return formatUint(uint64(v))
}

//go:inline
func formatUint(v uint64) string {
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

// formatFloat formats float64 to string.
func formatFloat(v float64) string {
	if v == 0 {
		return "0.0"
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// UnsafeStringToBytes converts string to []byte without allocation.
//go:inline
func UnsafeStringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
