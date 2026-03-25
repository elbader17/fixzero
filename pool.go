package fixzero

import (
	"sync"
	"unsafe"
)

// BufferPool proporciona gestión de buffers sin asignaciones.
var BufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 2048)
		return &buf
	},
}

// GetBuffer returns a pooled buffer with sufficient capacity.
// The buffer is extended to the required size.
func GetBuffer(minSize int) []byte {
	bp := BufferPool.Get().(*[]byte)
	if cap(*bp) < minSize {
		*bp = make([]byte, minSize)
	} else {
		// Extend to minSize for direct indexing
		*bp = (*bp)[:minSize]
	}
	// Reset to empty but keep capacity
	*bp = (*bp)[:0]
	return *bp
}

// GetBufferExact returns buffer with exact size for direct indexing
func GetBufferExact(minSize int) []byte {
	bp := BufferPool.Get().(*[]byte)
	if cap(*bp) < minSize {
		*bp = make([]byte, minSize)
	} else {
		*bp = (*bp)[:minSize]
	}
	return *bp
}

// PutBuffer returns buffer to pool.
func PutBuffer(buf []byte) {
	if cap(buf) > 0 && cap(buf) <= 8192 {
		BufferPool.Put(&buf)
	}
}

// StringToBytes convierte string a []byte sin asignación.
func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString convierte []byte a string sin asignación.
func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// Field representa un campo FIX con acceso zero-copy.
type Field struct {
	Tag   Tag
	Value string
}

// Message representa un mensaje FIX parseado.
// Usa array index para acceso O(1) sin map.
type Message struct {
	rawData []byte

	BeginString string
	MsgType     string

	fields [64]Field
	numFields int

	tagIndex [10000]int16

	Checksum string
}

// Reset limpia el mensaje para reutilizarlo.
func (m *Message) Reset() {
	m.rawData = m.rawData[:0]
	m.BeginString = ""
	m.MsgType = ""
	m.numFields = 0
	for i := 0; i < m.numFields; i++ {
		m.tagIndex[m.fields[i].Tag] = -1
	}
	m.Checksum = ""
}

// NumFields retorna el número de campos.
func (m *Message) NumFields() int {
	return m.numFields
}

// GetFieldIndex retorna índice del campo para un tag, -1 si no existe.
func (m *Message) GetFieldIndex(tag Tag) int {
	if tag < Tag(len(m.tagIndex)) {
		idx := m.tagIndex[tag]
		if idx >= 0 {
			return int(idx)
		}
	}
	return -1
}

// AddField agrega un campo al mensaje.
func (m *Message) AddField(tag Tag, value string) {
	if m.numFields < len(m.fields) {
		m.fields[m.numFields] = Field{Tag: tag, Value: value}
		if tag < Tag(len(m.tagIndex)) {
			m.tagIndex[tag] = int16(m.numFields)
		}
		m.numFields++
	}
}

var msgAllocated int64
