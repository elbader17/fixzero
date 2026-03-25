package fixzero

import (
	"encoding/xml"
	"errors"
	"os"
	"strconv"
	"strings"
)

type DataDictionary struct {
	messageTypes map[string]*MessageDef
	fields       map[int]*FieldDef
}

func NewDataDictionary() *DataDictionary {
	return &DataDictionary{
		messageTypes: make(map[string]*MessageDef),
		fields:       make(map[int]*FieldDef),
	}
}

func (dd *DataDictionary) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return dd.ParseXML(data)
}

func (dd *DataDictionary) ParseXML(data []byte) error {
	type XMLField struct {
		Tag  int    `xml:"name,attr"`
		Name string `xml:"name,attr"`
		Type string `xml:"type,attr"`
		Enum string `xml:"enum,attr"`
		Req  string `xml:"required,attr"`
	}

	type XMLGroup struct {
		Name   string     `xml:"name,attr"`
		NumTag int        `xml:"numtag,attr"`
		NumReq string     `xml:"numrequired,attr"`
		Fields []XMLField `xml:"field"`
	}

	type XMLMsg struct {
		Name    string     `xml:"name,attr"`
		MsgType string     `xml:"msgtype,attr"`
		Fields  []XMLField `xml:"field"`
		Groups  []XMLGroup `xml:"group"`
	}

	type XMLDict struct {
		Fields []XMLField `xml:"fields>field"`
		Msgs   []XMLMsg   `xml:"messages>msg"`
	}

	var dict XMLDict
	if err := xml.Unmarshal(data, &dict); err != nil {
		return err
	}

	dd.fields = make(map[int]*FieldDef, len(dict.Fields))
	for _, f := range dict.Fields {
		field := &FieldDef{
			tag:       f.Tag,
			name:      f.Name,
			fieldType: f.Type,
			required:  f.Req == "Y",
		}
		if f.Enum != "" {
			field.enumValues = strings.Split(f.Enum, "|")
		}
		dd.fields[f.Tag] = field
	}

	dd.messageTypes = make(map[string]*MessageDef, len(dict.Msgs))
	for _, m := range dict.Msgs {
		msgDef := &MessageDef{
			name:     m.Name,
			required: make([]int, 0, len(m.Fields)),
		}

		for _, f := range m.Fields {
			msgDef.required = append(msgDef.required, f.Tag)
		}

		if len(m.Groups) > 0 {
			msgDef.repeatingGroups = make([]*RepeatingGroup, 0, len(m.Groups))
			for _, g := range m.Groups {
				group := &RepeatingGroup{
					tag:      g.NumTag,
					numField: g.NumTag,
				}
				if g.NumReq == "Y" {
					group.numRequired = true
				}
				group.fields = make([]int, 0, len(g.Fields))
				for _, f := range g.Fields {
					group.fields = append(group.fields, f.Tag)
				}
				msgDef.repeatingGroups = append(msgDef.repeatingGroups, group)
			}
		}

		dd.messageTypes[m.MsgType] = msgDef
	}

	return nil
}

func (dd *DataDictionary) GetMessageDef(msgType string) *MessageDef {
	return dd.messageTypes[msgType]
}

func (dd *DataDictionary) GetFieldDef(tag int) *FieldDef {
	return dd.fields[tag]
}

type MessageDef struct {
	name            string
	required        []int
	repeatingGroups []*RepeatingGroup
}

func (md *MessageDef) Validate(msg *Message) []ValidationError {
	var errs []ValidationError

	for _, tag := range md.required {
		if !msg.Has(Tag(tag)) {
			errs = append(errs, ValidationError{
				Tag:       tag,
				FieldName: "",
				Message:   "required field missing",
			})
		}
	}

	for i := 0; i < msg.NumFields(); i++ {
		f := &msg.fields[i]
		fieldDef := fields[int(f.Tag)]
		if fieldDef == nil {
			continue
		}

		if err := ValidateField(int(f.Tag), f.Value, fieldDef); err != nil {
			errs = append(errs, ValidationError{
				Tag:       int(f.Tag),
				FieldName: fieldDef.name,
				Message:   err.Error(),
			})
		}
	}

	for _, group := range md.repeatingGroups {
		if groupErrs := ValidateRepeatingGroup(msg, group); len(groupErrs) > 0 {
			errs = append(errs, groupErrs...)
		}
	}

	return errs
}

type FieldDef struct {
	tag        int
	name       string
	fieldType  string
	enumValues []string
	required   bool
}

func (fd *FieldDef) Tag() int {
	return fd.tag
}

func (fd *FieldDef) Name() string {
	return fd.name
}

func (fd *FieldDef) Type() string {
	return fd.fieldType
}

func (fd *FieldDef) Required() bool {
	return fd.required
}

func (fd *FieldDef) EnumValues() []string {
	return fd.enumValues
}

var fields = make(map[int]*FieldDef)

func RegisterField(tag int, name, fieldType string, required bool, enumValues ...string) {
	f := &FieldDef{
		tag:        tag,
		name:       name,
		fieldType:  fieldType,
		required:   required,
		enumValues: enumValues,
	}
	fields[tag] = f
}

type ValidationError struct {
	Tag       int
	FieldName string
	Message   string
}

func (ve ValidationError) Error() string {
	return ve.Message
}

func ValidateField(tag int, value string, fieldDef *FieldDef) error {
	if fieldDef == nil {
		return nil
	}

	if err := ValidateType(value, fieldDef.fieldType); err != nil {
		return err
	}

	if len(fieldDef.enumValues) > 0 {
		if err := ValidateEnum(value, fieldDef.enumValues); err != nil {
			return err
		}
	}

	return nil
}

func ValidateType(value string, fieldType string) error {
	if value == "" {
		return nil
	}

	switch fieldType {
	case "STRING", "CHAR", "MULTIPLEVALUESTRING", "COUNTRY", "CURRENCY", "EXCHANGE", "LANGUAGE":
		return nil
	case "INT", "NUMINGROUP", "SEQNUM", "DAYOFMONTH":
		_, err := strconv.ParseInt(value, 10, 64)
		return err
	case "FLOAT", "AMT", "PRICE", "QTY", "PERCENTAGE", "PRICEOFFSET":
		_, err := strconv.ParseFloat(value, 64)
		return err
	case "BOOLEAN":
		if value != "Y" && value != "N" {
			return errors.New("invalid boolean value")
		}
		return nil
	case "DATE":
		if len(value) != 8 {
			return errors.New("invalid date format")
		}
		_, err := strconv.Atoi(value)
		return err
	case "TIME":
		if len(value) != 8 && len(value) != 12 {
			return errors.New("invalid time format")
		}
		return nil
	case "DATETIME":
		if len(value) != 15 && len(value) != 21 {
			return errors.New("invalid datetime format")
		}
		return nil
	case "DECIMAL":
		return nil
	default:
		return nil
	}
}

func ValidateEnum(value string, enumValues []string) error {
	for _, ev := range enumValues {
		if value == ev {
			return nil
		}
	}
	return errors.New("invalid enum value")
}

type RepeatingGroup struct {
	tag         int
	numField    int
	numRequired bool
	fields      []int
}

func (rg *RepeatingGroup) Tag() int {
	return rg.tag
}

func (rg *RepeatingGroup) NumField() int {
	return rg.numField
}

func (rg *RepeatingGroup) Fields() []int {
	return rg.fields
}

func CountGroups(msg *Message, firstFieldTag int) int {
	count := 0
	seenFirst := false

	for i := 0; i < msg.NumFields(); i++ {
		f := &msg.fields[i]
		if int(f.Tag) == firstFieldTag {
			if !seenFirst {
				seenFirst = true
			}
			count++
		}
	}

	return count
}

func ValidateRepeatingGroup(msg *Message, groupDef *RepeatingGroup) []ValidationError {
	var errs []ValidationError

	numGroups := CountGroups(msg, groupDef.fields[0])
	if numGroups == 0 && groupDef.numRequired {
		errs = append(errs, ValidationError{
			Tag:       groupDef.tag,
			FieldName: "",
			Message:   "required repeating group missing",
		})
		return errs
	}

	if numGroups == 0 {
		return errs
	}

	firstField := groupDef.fields[0]
	idx := 0
	currentGroup := 0

	for i := 0; i < msg.NumFields(); i++ {
		f := &msg.fields[i]
		tag := int(f.Tag)

		if tag == firstField {
			currentGroup++
			idx = 0
			continue
		}

		if currentGroup > 0 && idx < len(groupDef.fields) {
			if tag == groupDef.fields[idx] {
				idx++
			} else if tag > groupDef.fields[idx] {
				idx++
				if idx >= len(groupDef.fields) {
					currentGroup++
					idx = 0
				}
			}
		}
	}

	return errs
}

func (m *Message) GetMsgType() string {
	return m.MsgType
}
