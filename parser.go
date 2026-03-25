package fixzero

import (
	"errors"
	"unsafe"
)

var (
	ErrInvalidFormat       = errors.New("fixzero: invalid message format")
	ErrMissingBeginString = errors.New("fixzero: missing BeginString (tag 8)")
	ErrMissingMsgType     = errors.New("fixzero: missing MsgType (tag 35)")
)

const SOH byte = 0x01

type Parser struct{}

var globalParser = &Parser{}

func Parse(data []byte) (*Message, error) {
	return globalParser.Parse(data)
}

func ParseInto(data []byte, msg *Message) error {
	return globalParser.ParseInto(data, msg)
}

func (p *Parser) Parse(data []byte) (*Message, error) {
	msg := GetMessage()
	if err := p.parseMessage(data, msg); err != nil {
		msg.numFields = 0
		PutMessage(msg)
		return nil, err
	}
	return msg, nil
}

func (p *Parser) ParseInto(data []byte, msg *Message) error {
	msg.numFields = 0
	msg.BeginString = ""
	msg.MsgType = ""
	msg.Checksum = ""
	return p.parseMessage(data, msg)
}

// Parser optimizado
func (p *Parser) parseMessage(data []byte, msg *Message) error {
	if len(data) < 10 {
		return ErrInvalidFormat
	}

	msg.rawData = data
	n := len(data)
	
	var tag Tag
	var pos, numFields int
	
	for pos < n {
		eqPos := pos
		for eqPos < n && data[eqPos] != '=' {
			eqPos++
		}
		if eqPos >= n {
			break
		}
		
		tag = 0
		for i := pos; i < eqPos; i++ {
			tag = tag*10 + Tag(data[i]-'0')
		}
		
		sohPos := eqPos + 1
		for sohPos < n && data[sohPos] != SOH {
			sohPos++
		}
		if sohPos >= n {
			return ErrInvalidFormat
		}
		
		// Zero-copy: crear string de slice sin copiar
		valueSlice := data[eqPos+1:sohPos]
		value := *(*string)(unsafe.Pointer(&valueSlice))
		
		msg.fields[numFields] = Field{Tag: tag, Value: value}
		if tag < 10000 {
			msg.tagIndex[tag] = int16(numFields)
		}
		numFields++
		
		if tag == 8 {
			msg.BeginString = value
		} else if tag == 35 {
			msg.MsgType = value
		} else if tag == 10 {
			msg.Checksum = value
		}
		
		pos = sohPos + 1
	}
	
	msg.numFields = numFields
	
	if msg.BeginString == "" {
		return ErrMissingBeginString
	}
	if msg.MsgType == "" {
		return ErrMissingMsgType
	}
	
	return nil
}

func ParseTagString(s string) (Tag, error) {
	if len(s) == 0 {
		return 0, ErrInvalidFormat
	}
	var val Tag
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, ErrInvalidFormat
		}
		val = val*10 + Tag(c-'0')
	}
	return val, nil
}
