package fixzero

type GroupDef struct {
	tag      int
	numField int
	fields   []int
}

func NewGroupDef(tag, numField int, fields []int) *GroupDef {
	return &GroupDef{
		tag:      tag,
		numField: numField,
		fields:   fields,
	}
}

func (gd *GroupDef) Tag() int {
	return gd.tag
}

func (gd *GroupDef) NumField() int {
	return gd.numField
}

func (gd *GroupDef) Fields() []int {
	return gd.fields
}

func CountGroupsByTag(msg *Message, groupTag int) int {
	if msg.numFields == 0 {
		return 0
	}
	count := 0
	for i := 0; i < msg.numFields; i++ {
		if int(msg.fields[i].Tag) == groupTag {
			count++
		}
	}
	return count
}

func FindGroupStart(msg *Message, groupTag, firstFieldTag int) int {
	found := false
	for i := 0; i < msg.numFields; i++ {
		tag := int(msg.fields[i].Tag)
		if tag == groupTag {
			found = true
		}
		if found && tag == firstFieldTag {
			return i
		}
	}
	return -1
}

func IsGroupField(tag int, groupDef *GroupDef) bool {
	if groupDef == nil {
		return false
	}
	for _, f := range groupDef.fields {
		if f == tag {
			return true
		}
	}
	return false
}

func IterateGroups(msg *Message, groupTag int, fn func(groupIndex int, fields []Field) bool) {
	if msg.numFields == 0 {
		return
	}
	currentGroup := -1
	startIdx := -1

	for i := 0; i < msg.numFields; i++ {
		tag := int(msg.fields[i].Tag)

		if tag == groupTag {
			if currentGroup >= 0 && startIdx >= 0 {
				groupFields := msg.fields[startIdx:i]
				if !fn(currentGroup, groupFields) {
					return
				}
			}
			currentGroup++
			startIdx = i + 1
			continue
		}

		if currentGroup >= 0 && startIdx == -1 {
			startIdx = i
		}
	}

	if currentGroup >= 0 && startIdx >= 0 && startIdx < msg.numFields {
		groupFields := msg.fields[startIdx:msg.numFields]
		fn(currentGroup, groupFields)
	}
}

func GetGroupByTag(msg *Message, groupTag, index int) []Field {
	if msg.numFields == 0 || index < 0 {
		return nil
	}

	currentGroup := -1
	startIdx := -1

	for i := 0; i < msg.numFields; i++ {
		tag := int(msg.fields[i].Tag)

		if tag == groupTag {
			if currentGroup == index && startIdx >= 0 {
				return msg.fields[startIdx:i]
			}
			currentGroup++
			startIdx = -1
			continue
		}

		if currentGroup == index && startIdx < 0 {
			startIdx = i
		}
	}

	if currentGroup == index && startIdx >= 0 {
		return msg.fields[startIdx:]
	}

	return nil
}

func GetGroup(msg *Message, groupTag string, index int) []Field {
	tag := tagFromString(groupTag)
	if tag < 0 {
		return nil
	}
	return GetGroupByTag(msg, tag, index)
}

func GetGroupByField(msg *Message, firstFieldTag, index int) []Field {
	if msg.numFields == 0 || index < 0 {
		return nil
	}

	currentGroup := -1
	startIdx := -1

	for i := 0; i < msg.numFields; i++ {
		tag := int(msg.fields[i].Tag)

		if tag == firstFieldTag {
			if currentGroup == index && startIdx >= 0 {
				return msg.fields[startIdx:i]
			}
			currentGroup++
			startIdx = i
			continue
		}
	}

	if currentGroup == index && startIdx >= 0 {
		return msg.fields[startIdx:]
	}

	return nil
}

func ValidateGroup(msg *Message, groupDef *GroupDef) []ValidationError {
	var errs []ValidationError

	if groupDef == nil {
		return errs
	}

	counterField := groupDef.numField
	var expectedCount int64 = 0

	if counterField > 0 {
		idx := msg.GetFieldIndex(Tag(counterField))
		if idx >= 0 {
			val := msg.fields[idx].Value
			if len(val) > 0 {
				n, err := parseFixedInteger(val)
				if err == nil {
					expectedCount = n
				}
			}
		}
	}

	if len(groupDef.fields) == 0 {
		return errs
	}

	firstFieldTag := groupDef.fields[0]
	actualCount := CountGroupsByTag(msg, firstFieldTag)

	if expectedCount > 0 && int(expectedCount) != actualCount {
		errs = append(errs, ValidationError{
			Tag:       groupDef.tag,
			FieldName: "",
			Message:   "group count mismatch",
		})
	}

	for i := 0; i < actualCount; i++ {
		groupFields := GetGroupByField(msg, firstFieldTag, i)
		if len(groupFields) == 0 {
			continue
		}

		for _, reqTag := range groupDef.fields {
			found := false
			for _, f := range groupFields {
				if int(f.Tag) == reqTag {
					found = true
					break
				}
			}
			if !found {
				errs = append(errs, ValidationError{
					Tag:       reqTag,
					FieldName: "",
					Message:   "required field missing in group",
				})
			}
		}
	}

	return errs
}

func tagFromString(s string) int {
	if len(s) == 0 {
		return -1
	}
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return -1
		}
		n = n*10 + int(c-'0')
	}
	return n
}
