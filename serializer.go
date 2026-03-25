package fixzero

import (
	"strconv"
)

// Serializer ultra-optimizado
type Serializer struct{}

var globalSerializer = &Serializer{}

func Serialize(msg *Message) []byte {
	return globalSerializer.SerializeMessage(msg)
}

func SerializeTo(msg *Message, buf []byte) []byte {
	buf = buf[:0]
	return globalSerializer.serializeToBuffer(msg, buf)
}

// SerializeMessage con buffer pre-sized
func (s *Serializer) SerializeMessage(msg *Message) []byte {
	size := s.calculateSize(msg)
	buf := make([]byte, size)
	n := s.serializeToBufferFixed(msg, buf)
	return buf[:n]
}

// serializeToBufferFixed escribe directamente al buffer pre-sized
func (s *Serializer) serializeToBufferFixed(msg *Message, buf []byte) int {
	pos := 0
	// fieldsPos // Posición donde empiezan los campos después de header
	// bodyLenStart // Donde empieza el body length
	
	// BeginString: 8=FIX.X.X|SOH
	if msg.BeginString != "" {
		// fieldsPos = pos
		pos = writeFieldFixed(buf, pos, TagBeginString, msg.BeginString)
	}
	
	// MsgType: 35=X|SOH
	if msg.MsgType != "" {
		// fieldsPos = pos
		pos = writeFieldFixed(buf, pos, TagMsgType, msg.MsgType)
	}
	
	// Campos
	for i := 0; i < msg.numFields; i++ {
		f := &msg.fields[i]
		if f.Tag == TagBeginString || f.Tag == TagMsgType || f.Tag == TagCheckSum {
			continue
		}
		pos = writeFieldFixed(buf, pos, f.Tag, f.Value)
	}
	
	// Checksum: 10=XXX|SOH
	checksum := s.calculateChecksum(buf[:pos])
	pos = writeFieldFixed(buf, pos, TagCheckSum, checksum)
	
	// Ahora calcular body length y re-estructurar
	// Body = todo después de 35=X|
	bodyStart := findAfterTag(buf[:pos], TagMsgType)
	if bodyStart > 0 {
		// Encontrar posición de 10=
		checksumStart := -1
		for i := bodyStart; i < len(buf[:pos])-3; i++ {
			if buf[i] == '1' && buf[i+1] == '0' && buf[i+2] == '=' {
				checksumStart = i
				break
			}
		}
		
		if checksumStart > 0 {
			bodyLen := checksumStart - bodyStart
			bodyLenStr := formatUint32(uint32(bodyLen))
			bodyLenLen := len(bodyLenStr)
			
			// Shift del cuerpo para hacer espacio
			remaining := checksumStart - bodyStart
			copy(buf[bodyStart+2+bodyLenLen+1:], buf[bodyStart:bodyStart+remaining])
			
			// Escribir "9=bodylen|"
			copy(buf[bodyStart:], "9=")
			copy(buf[bodyStart+2:], bodyLenStr)
			buf[bodyStart+2+bodyLenLen] = SOH
			
			pos += 2 + bodyLenLen + 1
		}
	}
	
	return pos
}

func writeFieldFixed(buf []byte, pos int, tag Tag, value string) int {
	pos += fastFormatUint(buf[pos:], uint64(tag))
	buf[pos] = '='
	pos++
	if len(value) > 0 {
		copy(buf[pos:], value)
		pos += len(value)
	}
	buf[pos] = SOH
	pos++
	return pos
}

func fastFormatUint(buf []byte, v uint64) int {
	if v == 0 {
		buf[0] = '0'
		return 1
	}
	var tmp [20]byte
	i := len(tmp)
	for v > 0 {
		i--
		tmp[i] = byte(v%10) + '0'
		v /= 10
	}
	copy(buf, tmp[i:])
	return len(tmp) - i
}

func formatUint32(v uint32) string {
	if v == 0 {
		return "0"
	}
	var buf [10]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte(v%10) + '0'
		v /= 10
	}
	return string(buf[i:])
}

func (s *Serializer) calculateSize(msg *Message) int {
	size := 0
	if msg.BeginString != "" {
		size += 1 + 1 + len(msg.BeginString) + 1
	}
	if msg.MsgType != "" {
		size += 2 + 1 + len(msg.MsgType) + 1
	}
	for i := 0; i < msg.numFields; i++ {
		f := &msg.fields[i]
		if f.Tag == TagBeginString || f.Tag == TagMsgType || f.Tag == TagCheckSum {
			continue
		}
		tagLen := 1
		if f.Tag >= 1000 { tagLen = 4 } else if f.Tag >= 100 { tagLen = 3 } else if f.Tag >= 10 { tagLen = 2 }
		size += tagLen + 1 + len(f.Value) + 1
	}
	// Checksum 10=XXX|SOH + body length 9=XXXXX|SOH + padding
	size += 4 + 3 + 1 + 3 + 6 + 1 + 32
	return size
}

func (s *Serializer) calculateChecksum(data []byte) string {
	var cs uint8
	for i := 0; i < len(data); i++ {
		cs ^= data[i]
	}
	return string([]byte{
		hexDigit(cs >> 4),
		hexDigit(cs & 0x0F),
		'0',
	})
}

func hexDigit(n uint8) byte {
	if n < 10 {
		return '0' + n
	}
	return 'A' + n - 10
}

func findAfterTag(buf []byte, tag Tag) int {
	tagStr := strconv.FormatUint(uint64(tag), 10)
	tagLen := len(tagStr)
	
	for i := 0; i < len(buf)-tagLen-2; i++ {
		found := true
		for j := 0; j < tagLen; j++ {
			if buf[i+j] != tagStr[j] {
				found = false
				break
			}
		}
		if found && buf[i+tagLen] == '=' {
			for j := i + tagLen + 1; j < len(buf); j++ {
				if buf[j] == SOH {
					return j + 1
				}
			}
		}
	}
	return -1
}

// Legacy - para compatibilidad
func (s *Serializer) serializeToBuffer(msg *Message, buf []byte) []byte {
	size := s.calculateSize(msg)
	buf2 := make([]byte, size)
	n := s.serializeToBufferFixed(msg, buf2)
	return buf2[:n]
}

func AppendField(buf []byte, tag Tag, value string) []byte {
	_ = fastFormatUint
	_ = formatUint32
	buf = strconv.AppendUint(buf, uint64(tag), 10)
	buf = append(buf, '=')
	buf = append(buf, value...)
	buf = append(buf, SOH)
	return buf
}
