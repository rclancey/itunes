package itunes

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	//"io"
	"log"
	"reflect"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

func debugStruct(name string, obj interface{}) {
	data, _ := json.Marshal(obj)
	log.Println(name + ": " + string(data))
}

type FileKind struct {
	Name string `json:"name"`
	Extension string `json:"ext"`
}

var FileKinds = []FileKind{
	FileKind{"Protected AAC audio file", ".m4p"},
	FileKind{"MPEG audio file", ".mp3"},
	FileKind{"AIFF audio file", ".aiff"},
	FileKind{"WAV audio file", ".wav"},
	FileKind{"QuickTime movie file", ".mov"},
	FileKind{"MPEG-4 video file", ".mp4"},
	FileKind{"AAC audio file", ".m4a"},
}

const DateStartFromUnix = int64(-2082844800)

type FieldType int
const (
	RulesetField = iota + 1
	StringField
	IntField
	BooleanField
	DateField
	MediaKindField
	PlaylistField
	LoveField
	CloudField
	LocationField
)

var fieldTypes = map[Field]FieldType{
	Field(0x00): RulesetField,

	Field(0x03): StringField,
	Field(0x47): StringField,
	Field(0x04): StringField,
	Field(0x37): StringField,
	Field(0x0e): StringField,
	Field(0x12): StringField,
	Field(0x36): StringField,
	Field(0x08): StringField,
	Field(0x27): StringField,
	Field(0x09): StringField,
	Field(0x02): StringField,
	Field(0x3e): StringField,
	Field(0x4f): StringField,
	Field(0x51): StringField,
	Field(0x52): StringField,
	Field(0x4e): StringField,
	Field(0x53): StringField,
	Field(0x59): StringField,

	Field(0x23): IntField,
	Field(0x05): IntField,
	Field(0x18): IntField,
	Field(0x16): IntField,
	Field(0x19): IntField,
	Field(0x39): IntField,
	Field(0x06): IntField,
	Field(0x3f): IntField,
	Field(0x0c): IntField,
	Field(0x44): IntField,
	Field(0x0d): IntField,
	Field(0x0b): IntField,
	Field(0x07): IntField,
	Field(0x5a): IntField,

	Field(0x1f): BooleanField,
	Field(0x25): BooleanField,
	Field(0x29): BooleanField,
	Field(0x1d): BooleanField,

	Field(0x10): DateField,
	Field(0x0a): DateField,
	Field(0x17): DateField,
	Field(0x45): DateField,

	Field(0x3c): MediaKindField,

	Field(0x28): PlaylistField,

	Field(0x9a): LoveField,

	Field(0x86): CloudField,

	Field(0x85): LocationField,
}

func (f Field) Type() FieldType {
	t, ok := fieldTypes[f]
	if ok {
		return t
	}
	return FieldType(-1)
}

type SmartPlaylistInfo struct {
	CheckedOnly bool `json:"checked_only"`
	Descending bool `json:"descending"`
	HasLimit bool `json:"has_limit"`
	LiveUpdating bool `json:"live_updating"`
	LimitUnit *LimitMethod `json:"limit_method"`
	LimitSize *int `json:"limit_size"`
	SortField *SelectionMethod `json:"sort_field"`
}

type smartInfo struct {
	LiveUpdating uint8
	Match uint8
	HasLimit uint8
	LimitUnit uint8
	SortField uint32
	LimitSize uint32
	CheckedOnly uint8
	Ascending uint8
	Padding [98]byte
}

func (inf *SmartPlaylistInfo) Parse(raw []byte) error {
	buf := bytes.NewReader(raw)
	info := &smartInfo{}
	binary.Read(buf, binary.BigEndian, info)
	inf.CheckedOnly = (info.CheckedOnly > 0)
	inf.Descending = (info.Ascending == 0)
	inf.HasLimit = (info.HasLimit > 0)
	inf.LiveUpdating = (info.LiveUpdating > 0)
	if inf.HasLimit {
		lm := LimitMethod(info.LimitUnit)
		lf := SelectionMethod(info.SortField)
		sz := int(info.LimitSize)
		inf.LimitUnit = &lm
		inf.LimitSize = &sz
		inf.SortField = &lf
	} else {
		inf.LimitUnit = nil
		inf.LimitSize = nil
		inf.SortField = nil
	}
	return nil
}

func (inf *SmartPlaylistInfo) Encode() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	info := &smartInfo{}
	if inf.LiveUpdating {
		info.LiveUpdating = 1
	}
	if inf.HasLimit {
		info.HasLimit = 1
	}
	if !inf.Descending {
		info.Ascending = 1
	}
	if inf.CheckedOnly {
		info.CheckedOnly = 1
	}
	if inf.HasLimit {
		if inf.LimitUnit != nil {
			info.LimitUnit = uint8(*inf.LimitUnit)
		}
		if inf.LimitSize != nil {
			info.LimitSize = uint32(*inf.LimitSize)
		}
		if inf.SortField != nil {
			info.SortField = uint32(*inf.SortField)
		}
	}
	err := binary.Write(buf, binary.BigEndian, info)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type SmartRule interface {
	EncodeHeader([]byte) ([]byte, error)
	Encode() ([]byte, error)
	Match(track *Track, lib *Library) bool
}

type SmartPlaylistCriteria struct {
	Conjunction Conjunction `json:"conjunction"`
	Rules []SmartRule `json:"rules"`
}

func (r *SmartPlaylistCriteria) Match(track *Track, lib *Library) bool {
	if r.Conjunction == Conjunction_OR {
		for _, rule := range r.Rules {
			if rule != nil && rule.Match(track, lib) {
				return true
			}
		}
		return false
	}
	for _, rule := range r.Rules {
		if rule != nil && !rule.Match(track, lib) {
			return false
		}
	}
	return true
}

type RuleSetHeader struct {
	Junk1 [8]byte
	RuleCount uint32
	ConjunctionId uint32
	Junk2 [120]byte
}

type RuleHeader struct {
	FieldId uint32
	LogicSignId uint8
	Junk1 byte
	LogicRuleId uint16
	Junk2 [44]byte
	Length uint32
}

func (rh RuleHeader) Field() Field {
	return Field(rh.FieldId)
}

func (rh RuleHeader) LogicSign() LogicSign {
	return LogicSign(rh.LogicSignId)
}

func (rh RuleHeader) LogicRule() LogicRule {
	return LogicRule(rh.LogicRuleId)
}

type IntRuleData struct {
	Junk1 [4]byte
	IntA uint32
	RelA int64
	Junk2 [4]byte
	BoolB uint32
	Junk3 [4]byte
	IntB uint32
	Junk4 [12]byte
	BoolC uint32
	Junk5 [4]byte
	IntC uint32
	Junk6 [12]byte
}

func (ird IntRuleData) Ints() []int64 {
	ints := []int64{int64(ird.IntA)}
	if ird.BoolB > 0 {
		ints = append(ints, int64(ird.IntB))
		if ird.BoolC > 0 {
			ints = append(ints, int64(ird.IntC))
		}
	}
	return ints
}

func (ird IntRuleData) Times() []*Time {
	ints := ird.Ints()
	if ints[len(ints)-1] == 0 {
		ints = ints[:len(ints)-1]
	}
	times := make([]*Time, len(ints))
	for i, v := range ints {
		t := time.Unix(v+DateStartFromUnix, 0)
		times[i] = &Time{t}
	}
	return times
}

func (ird *IntRuleData) Encode() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, binary.BigEndian, ird)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ird *IntRuleData) Decode(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.BigEndian, ird)
}

func (c *SmartPlaylistCriteria) Parse(raw []byte) error {
	buf := bytes.NewReader(raw)
	rulesetHeader := &RuleSetHeader{}
	err := binary.Read(buf, binary.BigEndian, rulesetHeader)
	if err != nil {
		//fmt.Println("error reading ruleset header:", err)
		return err
	}
	//debugStruct("rulesetHeader", rulesetHeader)
	c.Conjunction = Conjunction(rulesetHeader.ConjunctionId)
	c.Rules = make([]SmartRule, int(rulesetHeader.RuleCount))
	//fmt.Printf("parsing smart criteria (%d rules)\n", rulesetHeader.RuleCount)
	for i := uint32(0); i < rulesetHeader.RuleCount; i++ {
		ruleHeader := &RuleHeader{}
		binary.Read(buf, binary.BigEndian, ruleHeader)
		//debugStruct("ruleHeader", ruleHeader)
		data := make([]byte, int(ruleHeader.Length))
		buf.Read(data)
		var rule SmartRule
		switch ruleHeader.Field().Type() {
		case RulesetField:
			// TODO
			sub := &SmartPlaylistCriteria{}
			err := sub.Parse(data)
			if err != nil {
				return err
			}
			rule = sub
		case StringField:
			rule = NewSmartPlaylistStringRule(ruleHeader, data)
		case IntField:
			rule = NewSmartPlaylistIntegerRule(ruleHeader, data)
		case BooleanField:
			rule = NewSmartPlaylistBooleanRule(ruleHeader, data)
		case DateField:
			rule = NewSmartPlaylistDateRule(ruleHeader, data)
		case MediaKindField:
			rule = NewSmartPlaylistMediaKindRule(ruleHeader, data)
		case PlaylistField:
			rule = NewSmartPlaylistPlaylistRule(ruleHeader, data)
		case LoveField:
			rule = NewSmartPlaylistLoveRule(ruleHeader, data)
		case CloudField:
			rule = NewSmartPlaylistCloudRule(ruleHeader, data)
		case LocationField:
			rule = NewSmartPlaylistLocationRule(ruleHeader, data)
		default:
			return fmt.Errorf("unknown rule type: %d / %s / %s", ruleHeader.FieldId, ruleHeader.Field(), ruleHeader.Field().Type())
		}
		//debugStruct("rule", rule)
		//rhd, _ := json.Marshal(ruleHeader)
		//rd, _ := json.Marshal(rule)
		//fmt.Println(string(rhd), string(rd))
		c.Rules[int(i)] = rule
	}
	return nil
}

func (c *SmartPlaylistCriteria) EncodeHeader(value []byte) ([]byte, error) {
	rh := &RuleHeader{}
	rh.FieldId = 0
	rh.LogicSignId = 0
	rh.LogicRuleId = 1
	rh.Junk2[0] = 1
	rh.Length = uint32(len(value))
	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, binary.BigEndian, rh)
	if err != nil {
		return nil, err
	}
	buf.Write(value)
	return buf.Bytes(), nil
}

func (c *SmartPlaylistCriteria) Encode() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	rsh := &RuleSetHeader{}
	rsh.Junk1 = [8]byte{83, 76, 115, 116, 0, 1, 0, 1}
	rsh.RuleCount = uint32(len(c.Rules))
	rsh.ConjunctionId = uint32(c.Conjunction)
	err := binary.Write(buf, binary.BigEndian, rsh)
	if err != nil {
		return nil, err
	}
	for _, r := range c.Rules {
		if r == nil {
			return nil, errors.New("nil rule")
		}
		data, err := r.Encode()
		if err != nil {
			return nil, err
		}
		data, err = r.EncodeHeader(data)
		if err != nil {
			return nil, err
		}
		_, err = buf.Write(data)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

type SmartPlaylistCommonRule struct {
	Field Field `json:"field"`
	Sign LogicSign `json:"sign"`
	Operator LogicRule `json:"operator"`
	idx []int
}

func NewSmartPlaylistCommonRule(ruleHeader *RuleHeader, value []byte) *SmartPlaylistCommonRule {
	return &SmartPlaylistCommonRule{
		Field: ruleHeader.Field(),
		Sign: ruleHeader.LogicSign(),
		Operator: ruleHeader.LogicRule(),
		idx: nil,
	}
}

func (r *SmartPlaylistCommonRule) EncodeHeader(value []byte) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	rh := &RuleHeader{}
	rh.FieldId = uint32(r.Field)
	rh.LogicSignId = uint8(r.Sign)
	rh.LogicRuleId = uint16(r.Operator)
	rh.Length = uint32(len(value))
	err := binary.Write(buf, binary.BigEndian, rh)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var BadFieldError = errors.New("bad rule field")

func (r *SmartPlaylistCommonRule) GetField(track *Track, kind reflect.Kind, typ reflect.Type) (reflect.Value, error) {
	if r.idx == nil {
		r.idx = []int{-1}
		fn := r.Field.String()
		rt := reflect.TypeOf(Track{})
		n := rt.NumField()
		for i := 0; i < n; i++ {
			f := rt.Field(i)
			ft := f.Type
			if f.Name == fn || strings.Split(f.Tag.Get("json"), ",")[0] == fn {
				if typ != nil {
					if typ.Kind() != reflect.Ptr && ft.Kind() == reflect.Ptr {
						ft = ft.Elem()
					}
					if ft == typ {
						r.idx = f.Index
					} else {
						err := fmt.Errorf("field %s (%s) is not of type %s", fn, f.Name, typ.Name())
						log.Println(err)
						return reflect.Value{}, err
					}
				} else if kind == reflect.Invalid {
					r.idx = f.Index
				} else {
					if ft.Kind() == reflect.Ptr {
						ft = ft.Elem()
					}
					if ft.Kind() == kind {
						r.idx = f.Index
					} else {
						err := fmt.Errorf("field %s (%s) is not of kind %s", fn, f.Name, kind)
						log.Println(err)
						return reflect.Value{}, err
					}
				}
				break
			}
		}
		if len(r.idx) == 0 || r.idx[0] == -1 {
			err := fmt.Errorf("field %s not found", r.Field)
			log.Println(err)
			return reflect.Value{}, err
		}
	}
	if len(r.idx) == 0 || r.idx[0] == -1 {
		return reflect.Value{}, BadFieldError
	}
	rv := reflect.ValueOf(*track).FieldByIndex(r.idx)
	if typ != nil {
		if rv.Kind() == reflect.Ptr {
			if typ.Kind() != reflect.Ptr {
				if rv.IsNil() {
					rv = reflect.Zero(rv.Type().Elem())
				} else {
					rv = rv.Elem()
				}
			}
		}
		if rv.Type() != typ {
			return reflect.Value{}, fmt.Errorf("field %s is not of type %s", r.Field, typ.Name())
		}
		return rv, nil
	}
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv = reflect.Zero(rv.Type().Elem())
		} else {
			rv = rv.Elem()
		}
	}
	if kind == reflect.Invalid {
		return rv, nil
	}
	if rv.Kind() != kind {
		return reflect.Value{}, fmt.Errorf("field %s is not of kind %s", r.Field, kind)
	}
	return rv, nil
}

type SmartPlaylistStringRule struct {
	*SmartPlaylistCommonRule
	RuleType string `json:"type"`
	Value string `json:"value"`
}

// UTF16BytesToString converts UTF-16 encoded bytes, in big or little endian byte order,
// to a UTF-8 encoded string.
func UTF16BytesToString(b []byte, o binary.ByteOrder) string {
    utf := make([]uint16, (len(b)+(2-1))/2)
    for i := 0; i+(2-1) < len(b); i += 2 {
        utf[i/2] = o.Uint16(b[i:])
    }
    if len(b)/2 < len(utf) {
        utf[len(utf)-1] = utf8.RuneError
    }
    return string(utf16.Decode(utf))
}

func StringToUTF16Bytes(s string, o binary.ByteOrder) []byte {
	utf := utf16.Encode([]rune(s))
	b := make([]byte, len(utf) * 2)
	for i, u := range utf {
		o.PutUint16(b[i*2:], u)
	}
	return b
}

func NewSmartPlaylistStringRule(ruleHeader *RuleHeader, value []byte) *SmartPlaylistStringRule {
	return &SmartPlaylistStringRule{
		SmartPlaylistCommonRule: NewSmartPlaylistCommonRule(ruleHeader, value),
		RuleType: "string",
		Value: UTF16BytesToString(value, binary.BigEndian),
	}
}

func (r *SmartPlaylistStringRule) Encode() ([]byte, error) {
	return StringToUTF16Bytes(r.Value, binary.BigEndian), nil
}

func (r *SmartPlaylistStringRule) Match(track *Track, lib *Library) bool {
	rv, err := r.GetField(track, reflect.String, nil)
	if err != nil {
		return true
	}
	switch r.Sign {
	case LogicSign_INT_POS, LogicSign_STR_POS:
		return r.basicMatch(rv.String())
	case LogicSign_INT_NEG, LogicSign_STR_NEG:
		return !r.basicMatch(rv.String())
	}
	return false
}

func (r *SmartPlaylistStringRule) basicMatch(s string) bool {
	switch r.Operator {
	case LogicRule_IS:
		return strings.ToLower(r.Value) == strings.ToLower(s)
	case LogicRule_CONTAINS:
		return strings.Contains(strings.ToLower(s), strings.ToLower(r.Value))
	case LogicRule_STARTSWITH:
		return strings.HasPrefix(strings.ToLower(s), strings.ToLower(r.Value))
	case LogicRule_ENDSWITH:
		return strings.HasSuffix(strings.ToLower(s), strings.ToLower(r.Value))
	}
	return false
}

type SmartPlaylistIntegerRule struct {
	*SmartPlaylistCommonRule
	RuleType string `json:"type"`
	Values []int64 `json:"values"`
}

func NewSmartPlaylistIntegerRule(ruleHeader *RuleHeader, value []byte) *SmartPlaylistIntegerRule {
	buf := bytes.NewReader(value)
	ird := &IntRuleData{}
	binary.Read(buf, binary.BigEndian, ird)
	return &SmartPlaylistIntegerRule{
		SmartPlaylistCommonRule: NewSmartPlaylistCommonRule(ruleHeader, value),
		RuleType: "int",
		Values: ird.Ints(),
	}
}

func (r *SmartPlaylistIntegerRule) Encode() ([]byte, error) {
	ird := &IntRuleData{}
	ird.IntA = uint32(r.Values[0])
	ird.BoolB = 1
	ird.BoolC = 1
	if len(r.Values) > 1 {
		ird.IntB = uint32(r.Values[1])
		if len(r.Values) > 2 {
			ird.IntC = uint32(r.Values[2])
		}
	} else {
		ird.IntB = ird.IntA
	}
	return ird.Encode()
}

func (r *SmartPlaylistIntegerRule) Match(track *Track, lib *Library) bool {
	if len(r.Values) == 0 {
		return false
	}
	rv, err := r.GetField(track, reflect.Invalid, nil)
	if err != nil {
		return true
	}
	var iv int64
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		iv = rv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		iv = int64(rv.Uint())
	default:
		return false
	}
	switch r.Sign {
	case LogicSign_INT_POS, LogicSign_STR_POS:
		return r.basicMatch(iv)
	case LogicSign_INT_NEG, LogicSign_STR_NEG:
		return !r.basicMatch(iv)
	}
	return false
}

func (r *SmartPlaylistIntegerRule) basicMatch(v int64) bool {
	switch r.Operator {
	case LogicRule_IS:
		return v == r.Values[0]
	case LogicRule_GREATERTHAN:
		return v > r.Values[0]
	case LogicRule_LESSTHAN:
		return v < r.Values[0]
	case LogicRule_BETWEEN:
		if len(r.Values) < 2 {
			return false
		}
		return v >= r.Values[0] && v <= r.Values[1]
	}
	return false
}

type SmartPlaylistBooleanRule struct {
	*SmartPlaylistCommonRule
	RuleType string `json:"type"`
	Value bool `json:"value"`
}

func NewSmartPlaylistBooleanRule(ruleHeader *RuleHeader, value []byte) *SmartPlaylistBooleanRule {
	buf := bytes.NewReader(value)
	ird := &IntRuleData{}
	binary.Read(buf, binary.BigEndian, ird)
	return &SmartPlaylistBooleanRule{
		SmartPlaylistCommonRule: NewSmartPlaylistCommonRule(ruleHeader, value),
		RuleType: "bool",
		Value: ird.IntA == 1,
	}
}

func (r *SmartPlaylistBooleanRule) Encode() ([]byte, error) {
	ird := &IntRuleData{}
	if r.Value {
		ird.IntA = 1
		ird.IntB = 1
	}
	ird.BoolB = 1
	ird.BoolC = 1
	return ird.Encode()
}

func (r *SmartPlaylistBooleanRule) Match(track *Track, lib *Library) bool {
	rv, err := r.GetField(track, reflect.Bool, nil)
	if err != nil {
		return true
	}
	switch r.Sign {
	case LogicSign_INT_POS, LogicSign_STR_POS:
		return r.basicMatch(rv.Bool())
	case LogicSign_INT_NEG, LogicSign_STR_NEG:
		return !r.basicMatch(rv.Bool())
	}
	return false
}

func (r *SmartPlaylistBooleanRule) basicMatch(v bool) bool {
	switch r.Operator {
	case LogicRule_IS:
		return v == r.Value
	}
	return false
}

type SmartPlaylistMediaKindRule struct {
	*SmartPlaylistCommonRule
	RuleType string `json:"type"`
	Value MediaKind `json:"value"`
}

func NewSmartPlaylistMediaKindRule(ruleHeader *RuleHeader, value []byte) *SmartPlaylistMediaKindRule {
	buf := bytes.NewReader(value)
	ird := &IntRuleData{}
	binary.Read(buf, binary.BigEndian, ird)
	return &SmartPlaylistMediaKindRule{
		SmartPlaylistCommonRule: NewSmartPlaylistCommonRule(ruleHeader, value),
		RuleType: "media",
		Value: MediaKind(ird.Ints()[0]),
	}
}

func (r *SmartPlaylistMediaKindRule) Encode() ([]byte, error) {
	ird := &IntRuleData{}
	ird.IntA = uint32(r.Value)
	ird.IntB = ird.IntA
	ird.BoolB = 1
	ird.BoolC = 1
	return ird.Encode()
}

func (r *SmartPlaylistMediaKindRule) Match(track *Track, lib *Library) bool {
	mk := track.MediaKind()
	switch r.Sign {
	case LogicSign_INT_POS, LogicSign_STR_POS:
		return mk == r.Value
	case LogicSign_INT_NEG, LogicSign_STR_NEG:
		return mk != r.Value
	}
	return false
}

type SmartPlaylistDateRule struct {
	*SmartPlaylistCommonRule
	RuleType string `json:"type"`
	Values []*Time `json:"values"`
	Relative int64 `json:"relative"`
}

func NewSmartPlaylistDateRule(ruleHeader *RuleHeader, value []byte) *SmartPlaylistDateRule {
	buf := bytes.NewReader(value)
	ird := &IntRuleData{}
	binary.Read(buf, binary.BigEndian, ird)
	return &SmartPlaylistDateRule{
		SmartPlaylistCommonRule: NewSmartPlaylistCommonRule(ruleHeader, value),
		RuleType: "date",
		Values: ird.Times(),
		Relative: ird.RelA * int64(ird.BoolB) * 1000,
	}
}

func (r *SmartPlaylistDateRule) Encode() ([]byte, error) {
	ird := &IntRuleData{}
	if r.Relative != 0 {
		rel := r.Relative / 1000
		if rel % (365 * 86400) == 0 {
			ird.BoolB = 365 * 86400
			ird.RelA = rel / (365 * 86400)
		} else if rel % (30 * 86400) == 0 {
			ird.BoolB = 30 * 86400
			ird.RelA = rel / (30 * 86400)
		} else if rel % (7 * 86400) == 0 {
			ird.BoolB = 7 * 86400
			ird.RelA = rel / (7 * 86400)
		} else if rel % 86400 == 0 {
			ird.BoolB = 86400
			ird.RelA = rel / 86400
		} else if rel % 3600 == 0 {
			ird.BoolB = 3600
			ird.RelA = rel / 3600
		} else if rel % 60 == 0 {
			ird.BoolB = 60
			ird.RelA = rel / 60
		} else {
			ird.BoolB = 1
			ird.RelA = rel
		}
		ird.Junk1 = [4]byte{45, 174, 45, 174}
		ird.IntA = 766389678
		ird.Junk3 = [4]byte{45, 174, 45, 174}
		ird.IntB = 766389678
	} else {
		ird.IntA = timeToRuleInt(r.Values[0])
		ird.BoolB = 1
		ird.BoolC = 1
		if len(r.Values) > 1 {
			ird.IntB = timeToRuleInt(r.Values[1])
		} else {
			ird.IntB = timeToRuleInt(r.Values[0])
		}
	}
	return ird.Encode()
}

func timeToRuleInt(t *Time) uint32 {
	return uint32(t.Unix() - DateStartFromUnix)
}

var timeType = reflect.TypeOf(Time{})

func (r *SmartPlaylistDateRule) Match(track *Track, lib *Library) bool {
	rv, err := r.GetField(track, reflect.Struct, timeType)
	if err != nil {
		return true
	}
	t, ok := rv.Interface().(Time)
	if !ok {
		return false
	}
	switch r.Sign {
	case LogicSign_INT_POS, LogicSign_STR_POS:
		return r.basicMatch(t.Get())
	case LogicSign_INT_NEG, LogicSign_STR_NEG:
		return !r.basicMatch(t.Get())
	}
	return false
}

func (r *SmartPlaylistDateRule) basicMatch(v time.Time) bool {
	switch r.Operator {
	case LogicRule_IS:
		return r.Values[0].Equal(v)
	case LogicRule_GREATERTHAN:
		return r.Values[0].Before(v)
	case LogicRule_LESSTHAN:
		return r.Values[0].After(v)
	case LogicRule_BETWEEN:
		if len(r.Values) < 2 {
			return false
		}
		return (r.Values[0].Equal(v) || r.Values[0].Before(v)) && (r.Values[1].Equal(v) || r.Values[0].After(v))
	case LogicRule_WITHIN:
		/*
		limit := time.Now().Add(time.Duration(r.Relative) * 1e6)
		if limit.Before(v) {
			log.Printf("track time %v after cutoff time %v", v, limit)
			return true
		}
		return false
		*/
		return time.Now().Add(time.Duration(r.Relative) * 1e6).Before(v)
	}
	return false
}

type SmartPlaylistPlaylistRule struct {
	*SmartPlaylistCommonRule
	RuleType string `json:"type"`
	Value PersistentID `json:"value"`
	tracks map[PersistentID]bool
}

func NewSmartPlaylistPlaylistRule(ruleHeader *RuleHeader, value []byte) *SmartPlaylistPlaylistRule {
	buf := bytes.NewReader(value)
	var id uint64
	binary.Read(buf, binary.BigEndian, &id)
	return &SmartPlaylistPlaylistRule{
		SmartPlaylistCommonRule: NewSmartPlaylistCommonRule(ruleHeader, value),
		RuleType: "playlist",
		Value: PersistentID(id),
		tracks: nil,
	}
}

func (r *SmartPlaylistPlaylistRule) Encode() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, binary.BigEndian, uint64(r.Value))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *SmartPlaylistPlaylistRule) Match(track *Track, lib *Library) bool {
	if r.tracks == nil {
		pl := lib.Playlists[r.Value]
		r.tracks = map[PersistentID]bool{}
		if pl != nil {
			for _, tr := range pl.Populate(lib).PlaylistItems {
				r.tracks[tr.PersistentID] = true
			}
		}
	}
	switch r.Sign {
	case LogicSign_INT_POS, LogicSign_STR_POS:
		return r.tracks[track.PersistentID]
	case LogicSign_INT_NEG, LogicSign_STR_NEG:
		return !r.tracks[track.PersistentID]
	}
	return false
}

type SmartPlaylistLoveRule struct {
	*SmartPlaylistCommonRule
	RuleType string `json:"type"`
	Value LoveStatus `json:"value"`
}

func NewSmartPlaylistLoveRule(ruleHeader *RuleHeader, value []byte) *SmartPlaylistLoveRule {
	buf := bytes.NewReader(value)
	ird := &IntRuleData{}
	binary.Read(buf, binary.BigEndian, ird)
	return &SmartPlaylistLoveRule{
		SmartPlaylistCommonRule: NewSmartPlaylistCommonRule(ruleHeader, value),
		RuleType: "love",
		Value: LoveStatus(ird.IntA),
	}
}

func (r *SmartPlaylistLoveRule) Encode() ([]byte, error) {
	ird := &IntRuleData{}
	ird.IntA = uint32(r.Value)
	ird.IntB = ird.IntA
	ird.BoolB = 1
	ird.BoolC = 1
	return ird.Encode()
}

func (r *SmartPlaylistLoveRule) Match(track *Track, lib *Library) bool {
	var ls LoveStatus
	if track.Loved == nil {
		ls = LoveStatus_NONE
	} else if *track.Loved {
		ls = LoveStatus_LOVED
	} else {
		ls = LoveStatus_DISLIKED
	}
	switch r.Sign {
	case LogicSign_INT_POS, LogicSign_STR_POS:
		return ls == r.Value
	case LogicSign_INT_NEG, LogicSign_STR_NEG:
		return ls != r.Value
	}
	return false
}

type SmartPlaylistCloudRule struct {
	*SmartPlaylistCommonRule
	RuleType string `json:"type"`
	Value ICloudStatus `json:"value"`
}

func NewSmartPlaylistCloudRule(ruleHeader *RuleHeader, value []byte) *SmartPlaylistCloudRule {
	buf := bytes.NewReader(value)
	ird := &IntRuleData{}
	binary.Read(buf, binary.BigEndian, ird)
	return &SmartPlaylistCloudRule{
		SmartPlaylistCommonRule: NewSmartPlaylistCommonRule(ruleHeader, value),
		RuleType: "cloud",
		Value: ICloudStatus(ird.IntA),
	}
}

func (r *SmartPlaylistCloudRule) Encode() ([]byte, error) {
	ird := &IntRuleData{}
	ird.IntA = uint32(r.Value)
	ird.IntB = ird.IntA
	ird.BoolB = 1
	ird.BoolC = 1
	return ird.Encode()
}

func (r *SmartPlaylistCloudRule) Match(track *Track, lib *Library) bool {
	// TODO
	return false
}

type SmartPlaylistLocationRule struct {
	*SmartPlaylistCommonRule
	RuleType string `json:"type"`
	Value LocationStatus `json:"value"`
}

func NewSmartPlaylistLocationRule(ruleHeader *RuleHeader, value []byte) *SmartPlaylistLocationRule {
	buf := bytes.NewReader(value)
	ird := &IntRuleData{}
	binary.Read(buf, binary.BigEndian, ird)
	return &SmartPlaylistLocationRule{
		SmartPlaylistCommonRule: NewSmartPlaylistCommonRule(ruleHeader, value),
		RuleType: "playlist",
		Value: LocationStatus(ird.IntA),
	}
}

func (r *SmartPlaylistLocationRule) Encode() ([]byte, error) {
	ird := &IntRuleData{}
	ird.IntA = uint32(r.Value)
	ird.IntB = ird.IntA
	ird.BoolB = 1
	ird.BoolC = 1
	return ird.Encode()
}

func (r *SmartPlaylistLocationRule) Match(track *Track, lib *Library) bool {
	switch r.Sign {
	case LogicSign_INT_POS, LogicSign_STR_POS:
		switch r.Value {
		case LocationStatus_COMPUTER:
			return track.Location != ""
		case LocationStatus_ICLOUD:
			return track.Location == "" || track.Purchased
		default:
			return false
		}
	case LogicSign_INT_NEG, LogicSign_STR_NEG:
		switch r.Value {
		case LocationStatus_COMPUTER:
			return track.Location == ""
		case LocationStatus_ICLOUD:
			// TODO
			return track.Purchased && track.Location != ""
		}
	}
	return false
}

type SmartPlaylist struct {
	rawInfo []byte
	rawCriteria []byte
	Info *SmartPlaylistInfo `json:"info"`
	Criteria *SmartPlaylistCriteria `json:"criteria"`
}

func decodeb64(data []byte) ([]byte, error) {
	s := strings.TrimSpace(string(data))
	s = strings.Join(strings.Fields(strings.TrimSpace(s)), "")
	return base64.StdEncoding.DecodeString(s)
}

func encodeb64(data []byte) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dst, data)
	return data
}

func ParseSmartPlaylist(info, criteria []byte) (*SmartPlaylist, error) {
	dinfo, err := decodeb64(info)
	if err != nil {
		return nil, err
	}
	dcrit, err := decodeb64(criteria)
	if err != nil {
		return nil, err
	}
	p := &SmartPlaylist{
		rawInfo: dinfo,
		rawCriteria: dcrit,
		Info: &SmartPlaylistInfo{},
		Criteria: &SmartPlaylistCriteria{},
	}
	err = p.Info.Parse(dinfo)
	if err != nil {
		return nil, err
	}
	err = p.Criteria.Parse(dcrit)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *SmartPlaylist) Encode() (info []byte, criteria []byte, err error) {
	info, err = s.Info.Encode()
	if err != nil {
		return nil, nil, err
	}
	criteria, err = s.Criteria.Encode()
	if err != nil {
		return nil, nil, err
	}
	return encodeb64(info), encodeb64(criteria), nil
}
