package fixzero

import (
	"strconv"
	"unsafe"
)

// Field accessor methods for Message.
// Optimized for zero-allocation using array indexing.

// Has returns true if the field exists.
//go:inline
func (m *Message) Has(tag Tag) bool {
	return m.GetFieldIndex(tag) >= 0
}

// GetString returns the field value as string.
//go:inline
func (m *Message) GetString(tag Tag) string {
	idx := m.GetFieldIndex(tag)
	if idx >= 0 {
		return m.fields[idx].Value
	}
	return ""
}

// GetBytes returns the field value as []byte.
// This creates a copy (not zero-copy).
func (m *Message) GetBytes(tag Tag) []byte {
	idx := m.GetFieldIndex(tag)
	if idx >= 0 {
		return []byte(m.fields[idx].Value)
	}
	return nil
}

// GetBytesZeroCopy returns zero-copy view.
//go:inline
func (m *Message) GetBytesZeroCopy(tag Tag) []byte {
	idx := m.GetFieldIndex(tag)
	if idx >= 0 {
		val := m.fields[idx].Value
		return unsafe.Slice(unsafe.StringData(val), len(val))
	}
	return nil
}

// GetInt returns the field value as int64.
//go:inline
func (m *Message) GetInt(tag Tag) (int64, bool) {
	idx := m.GetFieldIndex(tag)
	if idx >= 0 {
		val := m.fields[idx].Value
		if len(val) == 0 {
			return 0, false
		}
		n, err := parseFixedInteger(val)
		return n, err == nil
	}
	return 0, false
}

// GetUint returns the field value as uint64.
//go:inline
func (m *Message) GetUint(tag Tag) (uint64, bool) {
	idx := m.GetFieldIndex(tag)
	if idx >= 0 {
		val := m.fields[idx].Value
		if len(val) == 0 {
			return 0, false
		}
		n, err := parseFixedUnsigned(val)
		return n, err == nil
	}
	return 0, false
}

// GetFloat returns the field value as float64.
//go:inline
func (m *Message) GetFloat(tag Tag) (float64, bool) {
	idx := m.GetFieldIndex(tag)
	if idx >= 0 {
		val := m.fields[idx].Value
		if len(val) == 0 {
			return 0, false
		}
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	}
	return 0, false
}

// GetBool returns the field value as bool.
//go:inline
func (m *Message) GetBool(tag Tag) (bool, bool) {
	idx := m.GetFieldIndex(tag)
	if idx >= 0 {
		val := m.fields[idx].Value
		if len(val) == 0 {
			return false, false
		}
		return val[0] == 'Y', true
	}
	return false, false
}

// GetTag returns the tag number for a field.
//go:inline
func (f *Field) GetTag() Tag {
	return f.Tag
}

// GetValue returns the field value.
//go:inline
func (f *Field) GetValue() string {
	return f.Value
}

// Iterate calls fn for each field in the message.
func (m *Message) Iterate(fn func(tag Tag, value string) bool) {
	for i := 0; i < m.numFields; i++ {
		f := &m.fields[i]
		if !fn(f.Tag, f.Value) {
			break
		}
	}
}

// IterateFields iterates over all fields.
func (m *Message) IterateFields(fn func(Field) bool) {
	for i := 0; i < m.numFields; i++ {
		if !fn(m.fields[i]) {
			break
		}
	}
}

// parseFixedInteger parses a FIX integer (may have leading zeros).
//go:inline
func parseFixedInteger(s string) (int64, error) {
	if len(s) == 0 {
		return 0, strconv.ErrSyntax
	}
	var neg bool
	start := 0
	if s[0] == '-' {
		neg = true
		start = 1
	} else if s[0] == '+' {
		start = 1
	}
	var n int64
	for i := start; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, strconv.ErrSyntax
		}
		n = n*10 + int64(c-'0')
	}
	if neg {
		return -n, nil
	}
	return n, nil
}

// parseFixedUnsigned parses a FIX unsigned integer.
//go:inline
func parseFixedUnsigned(s string) (uint64, error) {
	if len(s) == 0 {
		return 0, strconv.ErrSyntax
	}
	var n uint64
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, strconv.ErrSyntax
		}
		n = n*10 + uint64(c-'0')
	}
	return n, nil
}

// RawData returns the raw message data.
//go:inline
func (m *Message) RawData() []byte {
	return m.rawData
}

// Fields returns all fields in the message.
func (m *Message) Fields() []Field {
	return m.fields[:m.numFields]
}

// Len returns the number of fields.
//go:inline
func (m *Message) Len() int {
	return m.numFields
}

// Quick accessors for common fields.
//go:inline

func (m *Message) GetSenderCompID() string {
	return m.GetString(TagSenderCompID)
}

func (m *Message) GetTargetCompID() string {
	return m.GetString(TagTargetCompID)
}

func (m *Message) GetMsgSeqNum() (int64, bool) {
	return m.GetInt(TagMsgSeqNum)
}

func (m *Message) GetSendingTime() string {
	return m.GetString(TagSendingTime)
}

func (m *Message) GetSymbol() string {
	return m.GetString(TagSymbol)
}

func (m *Message) GetSide() string {
	return m.GetString(TagSide)
}

func (m *Message) GetOrdType() string {
	return m.GetString(TagOrdType)
}

func (m *Message) GetPrice() (float64, bool) {
	return m.GetFloat(TagPrice)
}

func (m *Message) GetQuantity() (int64, bool) {
	return m.GetInt(TagQuantity)
}

func (m *Message) GetClOrdID() string {
	return m.GetString(TagClOrdID)
}

func (m *Message) GetOrdID() string {
	return m.GetString(TagOrdID)
}

func (m *Message) GetExecType() string {
	return m.GetString(TagExecType)
}

func (m *Message) GetOrdStatus() string {
	return m.GetString(TagOrdStatus)
}

func (m *Message) GetLeavesQty() (int64, bool) {
	return m.GetInt(TagLeavesQty)
}

func (m *Message) GetCumQty() (int64, bool) {
	return m.GetInt(TagCumQty)
}

func (m *Message) GetAvgPx() (float64, bool) {
	return m.GetFloat(TagAvgPx)
}
