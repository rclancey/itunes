package itunes

import (
	"encoding/json"
	"fmt"
)


type Conjunction int

var ConjunctionNames = map[Conjunction]string{
	Conjunction(0x0): "AND",
	Conjunction(0x1): "OR",
}
var ConjunctionValues = map[string]Conjunction{
	"AND": Conjunction(0x0),
	"OR": Conjunction(0x1),
}
const (
	Conjunction_OR = Conjunction(0x1)
	Conjunction_AND = Conjunction(0x0)
)

func (e Conjunction) String() string {
	s, ok := ConjunctionNames[e]
	if ok {
		return s
	}
	return fmt.Sprintf("Conjunction_0x%X", int(e))
}

func (e Conjunction) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *Conjunction) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, ok := ConjunctionValues[s]
	if !ok {
		return fmt.Errorf("unknown Conjunction %s", s)
	}
	*e = v
	return nil
}


type SelectionMethod int

var SelectionMethodNames = map[SelectionMethod]string{
	SelectionMethod(0x1A): "play_date_utc",
	SelectionMethod(0x19): "play_count",
	SelectionMethod(0x1C): "rating",
	SelectionMethod(0x1): "lowest_rating",
	SelectionMethod(0x15): "date_added",
	SelectionMethod(0x2): "<random>",
	SelectionMethod(0x5): "name",
	SelectionMethod(0x6): "album",
	SelectionMethod(0x7): "artist",
	SelectionMethod(0x9): "genre",
}
var SelectionMethodValues = map[string]SelectionMethod{
	"<random>": SelectionMethod(0x2),
	"name": SelectionMethod(0x5),
	"album": SelectionMethod(0x6),
	"artist": SelectionMethod(0x7),
	"genre": SelectionMethod(0x9),
	"rating": SelectionMethod(0x1C),
	"lowest_rating": SelectionMethod(0x1),
	"date_added": SelectionMethod(0x15),
	"play_date_utc": SelectionMethod(0x1A),
	"play_count": SelectionMethod(0x19),
}
const (
	SelectionMethod_PLAY_DATE_UTC = SelectionMethod(0x1A)
	SelectionMethod_PLAY_COUNT = SelectionMethod(0x19)
	SelectionMethod_RANDOM = SelectionMethod(0x2)
	SelectionMethod_NAME = SelectionMethod(0x5)
	SelectionMethod_ALBUM = SelectionMethod(0x6)
	SelectionMethod_ARTIST = SelectionMethod(0x7)
	SelectionMethod_GENRE = SelectionMethod(0x9)
	SelectionMethod_RATING = SelectionMethod(0x1C)
	SelectionMethod_LOWEST_RATING = SelectionMethod(0x1)
	SelectionMethod_DATE_ADDED = SelectionMethod(0x15)
)

func (e SelectionMethod) String() string {
	s, ok := SelectionMethodNames[e]
	if ok {
		return s
	}
	return fmt.Sprintf("SelectionMethod_0x%X", int(e))
}

func (e SelectionMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *SelectionMethod) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, ok := SelectionMethodValues[s]
	if !ok {
		return fmt.Errorf("unknown SelectionMethod %s", s)
	}
	*e = v
	return nil
}


type Field int

var FieldNames = map[Field]string{
	Field(0x5A): "album_rating",
	Field(0xD): "total_time",
	Field(0x25): "has_artwork",
	Field(0x45): "skip_date",
	Field(0x9A): "love",
	Field(0x6): "sample_rate",
	Field(0x29): "purchased",
	Field(0x28): "playlist_persistent_id",
	Field(0x47): "album_artist",
	Field(0x2): "name",
	Field(0x53): "sort_series",
	Field(0x16): "play_count",
	Field(0x17): "play_date_utc",
	Field(0x37): "category",
	Field(0x36): "description",
	Field(0x4F): "sort_album",
	Field(0x23): "bpm",
	Field(0xB): "track_number",
	Field(0x19): "rating",
	Field(0x39): "podcast",
	Field(0xA): "date_modified",
	Field(0x3): "album",
	Field(0x12): "composer",
	Field(0x9): "kind",
	Field(0x3E): "series",
	Field(0x1F): "compilation",
	Field(0x3C): "media_kind",
	Field(0x27): "grouping",
	Field(0x52): "sort_composer",
	Field(0x44): "skip_count",
	Field(0x1D): "disabled",
	Field(0x85): "location",
	Field(0x8): "genre",
	Field(0x59): "video_rating",
	Field(0x5): "bit_rate",
	Field(0xC): "size",
	Field(0x86): "icloud_status",
	Field(0x3F): "season",
	Field(0x7): "year",
	Field(0x10): "date_added",
	Field(0x4): "artist",
	Field(0xE): "comments",
	Field(0x51): "sort_album_artist",
	Field(0x4E): "sort_name",
	Field(0x18): "disk_number",
}
var FieldValues = map[string]Field{
	"album": Field(0x3),
	"composer": Field(0x12),
	"kind": Field(0x9),
	"series": Field(0x3E),
	"compilation": Field(0x1F),
	"rating": Field(0x19),
	"podcast": Field(0x39),
	"date_modified": Field(0xA),
	"media_kind": Field(0x3C),
	"grouping": Field(0x27),
	"sort_composer": Field(0x52),
	"skip_count": Field(0x44),
	"disabled": Field(0x1D),
	"location": Field(0x85),
	"genre": Field(0x8),
	"video_rating": Field(0x59),
	"bit_rate": Field(0x5),
	"size": Field(0xC),
	"icloud_status": Field(0x86),
	"artist": Field(0x4),
	"comments": Field(0xE),
	"sort_album_artist": Field(0x51),
	"sort_name": Field(0x4E),
	"disk_number": Field(0x18),
	"season": Field(0x3F),
	"year": Field(0x7),
	"date_added": Field(0x10),
	"album_rating": Field(0x5A),
	"total_time": Field(0xD),
	"has_artwork": Field(0x25),
	"skip_date": Field(0x45),
	"love": Field(0x9A),
	"sample_rate": Field(0x6),
	"purchased": Field(0x29),
	"album_artist": Field(0x47),
	"name": Field(0x2),
	"sort_series": Field(0x53),
	"play_count": Field(0x16),
	"play_date_utc": Field(0x17),
	"playlist_persistent_id": Field(0x28),
	"category": Field(0x37),
	"description": Field(0x36),
	"sort_album": Field(0x4F),
	"bpm": Field(0x23),
	"track_number": Field(0xB),
}
const (
	Field_ALBUM = Field(0x3)
	Field_COMPOSER = Field(0x12)
	Field_KIND = Field(0x9)
	Field_SERIES = Field(0x3E)
	Field_COMPILATION = Field(0x1F)
	Field_RATING = Field(0x19)
	Field_PODCAST = Field(0x39)
	Field_DATE_MODIFIED = Field(0xA)
	Field_MEDIA_KIND = Field(0x3C)
	Field_GROUPING = Field(0x27)
	Field_SORT_COMPOSER = Field(0x52)
	Field_SKIP_COUNT = Field(0x44)
	Field_DISABLED = Field(0x1D)
	Field_LOCATION = Field(0x85)
	Field_GENRE = Field(0x8)
	Field_VIDEO_RATING = Field(0x59)
	Field_BIT_RATE = Field(0x5)
	Field_SIZE = Field(0xC)
	Field_ICLOUD_STATUS = Field(0x86)
	Field_ARTIST = Field(0x4)
	Field_COMMENTS = Field(0xE)
	Field_SORT_ALBUM_ARTIST = Field(0x51)
	Field_SORT_NAME = Field(0x4E)
	Field_DISK_NUMBER = Field(0x18)
	Field_SEASON = Field(0x3F)
	Field_YEAR = Field(0x7)
	Field_DATE_ADDED = Field(0x10)
	Field_ALBUM_RATING = Field(0x5A)
	Field_TOTAL_TIME = Field(0xD)
	Field_HAS_ARTWORK = Field(0x25)
	Field_SKIP_DATE = Field(0x45)
	Field_LOVE = Field(0x9A)
	Field_SAMPLE_RATE = Field(0x6)
	Field_PURCHASED = Field(0x29)
	Field_ALBUM_ARTIST = Field(0x47)
	Field_NAME = Field(0x2)
	Field_SORT_SERIES = Field(0x53)
	Field_PLAY_COUNT = Field(0x16)
	Field_PLAY_DATE_UTC = Field(0x17)
	Field_PLAYLIST_PERSISTENT_ID = Field(0x28)
	Field_CATEGORY = Field(0x37)
	Field_DESCRIPTION = Field(0x36)
	Field_SORT_ALBUM = Field(0x4F)
	Field_BPM = Field(0x23)
	Field_TRACK_NUMBER = Field(0xB)
)

func (e Field) String() string {
	s, ok := FieldNames[e]
	if ok {
		return s
	}
	return fmt.Sprintf("Field_0x%X", int(e))
}

func (e Field) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *Field) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, ok := FieldValues[s]
	if !ok {
		return fmt.Errorf("unknown Field %s", s)
	}
	*e = v
	return nil
}


type LogicSign int

var LogicSignNames = map[LogicSign]string{
	LogicSign(0x3): "str_neg",
	LogicSign(0x0): "int_pos",
	LogicSign(0x1): "str_pos",
	LogicSign(0x2): "int_neg",
}
var LogicSignValues = map[string]LogicSign{
	"int_pos": LogicSign(0x0),
	"str_pos": LogicSign(0x1),
	"int_neg": LogicSign(0x2),
	"str_neg": LogicSign(0x3),
}
const (
	LogicSign_STR_POS = LogicSign(0x1)
	LogicSign_INT_NEG = LogicSign(0x2)
	LogicSign_STR_NEG = LogicSign(0x3)
	LogicSign_INT_POS = LogicSign(0x0)
)

func (e LogicSign) String() string {
	s, ok := LogicSignNames[e]
	if ok {
		return s
	}
	return fmt.Sprintf("LogicSign_0x%X", int(e))
}

func (e LogicSign) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *LogicSign) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, ok := LogicSignValues[s]
	if !ok {
		return fmt.Errorf("unknown LogicSign %s", s)
	}
	*e = v
	return nil
}


type LogicRule int

var LogicRuleNames = map[LogicRule]string{
	LogicRule(0x0): "other",
	LogicRule(0x2): "contains",
	LogicRule(0x8): "endswith",
	LogicRule(0x10): "greaterthan",
	LogicRule(0x100): "between",
	LogicRule(0x1): "is",
	LogicRule(0x4): "startswith",
	LogicRule(0x40): "lessthan",
	LogicRule(0x200): "within",
}
var LogicRuleValues = map[string]LogicRule{
	"within": LogicRule(0x200),
	"is": LogicRule(0x1),
	"startswith": LogicRule(0x4),
	"lessthan": LogicRule(0x40),
	"greaterthan": LogicRule(0x10),
	"between": LogicRule(0x100),
	"other": LogicRule(0x0),
	"contains": LogicRule(0x2),
	"endswith": LogicRule(0x8),
}
const (
	LogicRule_BETWEEN = LogicRule(0x100)
	LogicRule_OTHER = LogicRule(0x0)
	LogicRule_CONTAINS = LogicRule(0x2)
	LogicRule_ENDSWITH = LogicRule(0x8)
	LogicRule_GREATERTHAN = LogicRule(0x10)
	LogicRule_IS = LogicRule(0x1)
	LogicRule_STARTSWITH = LogicRule(0x4)
	LogicRule_LESSTHAN = LogicRule(0x40)
	LogicRule_WITHIN = LogicRule(0x200)
)

func (e LogicRule) String() string {
	s, ok := LogicRuleNames[e]
	if ok {
		return s
	}
	return fmt.Sprintf("LogicRule_0x%X", int(e))
}

func (e LogicRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *LogicRule) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, ok := LogicRuleValues[s]
	if !ok {
		return fmt.Errorf("unknown LogicRule %s", s)
	}
	*e = v
	return nil
}


type MediaKind int

var MediaKindNames = map[MediaKind]string{
	MediaKind(0x400): "HomeVideo",
	MediaKind(0x10000): "ITunesExtras",
	MediaKind(0x20A004): "UndesiredOther",
	MediaKind(0x4): "Podcast",
	MediaKind(0xC00008): "BookOrAudiobook",
	MediaKind(0x8): "Audiobook",
	MediaKind(0x40): "TVShow",
	MediaKind(0x1021B1): "OtherMusic",
	MediaKind(0x208004): "UndesiredMusic",
	MediaKind(0x1): "Music",
	MediaKind(0x2): "Movie",
	MediaKind(0x20): "MusicVideo",
	MediaKind(0x100000): "VoiceMemo",
	MediaKind(0x200000): "ITunesU",
	MediaKind(0xC00000): "Book",
}
var MediaKindValues = map[string]MediaKind{
	"BookOrAudiobook": MediaKind(0xC00008),
	"Podcast": MediaKind(0x4),
	"TVShow": MediaKind(0x40),
	"Audiobook": MediaKind(0x8),
	"Movie": MediaKind(0x2),
	"MusicVideo": MediaKind(0x20),
	"VoiceMemo": MediaKind(0x100000),
	"ITunesU": MediaKind(0x200000),
	"Book": MediaKind(0xC00000),
	"OtherMusic": MediaKind(0x1021B1),
	"UndesiredMusic": MediaKind(0x208004),
	"Music": MediaKind(0x1),
	"ITunesExtras": MediaKind(0x10000),
	"UndesiredOther": MediaKind(0x20A004),
	"HomeVideo": MediaKind(0x400),
}
const (
	MediaKind_UNDESIREDOTHER = MediaKind(0x20A004)
	MediaKind_HOMEVIDEO = MediaKind(0x400)
	MediaKind_ITUNESEXTRAS = MediaKind(0x10000)
	MediaKind_PODCAST = MediaKind(0x4)
	MediaKind_BOOKORAUDIOBOOK = MediaKind(0xC00008)
	MediaKind_AUDIOBOOK = MediaKind(0x8)
	MediaKind_TVSHOW = MediaKind(0x40)
	MediaKind_MUSICVIDEO = MediaKind(0x20)
	MediaKind_VOICEMEMO = MediaKind(0x100000)
	MediaKind_ITUNESU = MediaKind(0x200000)
	MediaKind_BOOK = MediaKind(0xC00000)
	MediaKind_OTHERMUSIC = MediaKind(0x1021B1)
	MediaKind_UNDESIREDMUSIC = MediaKind(0x208004)
	MediaKind_MUSIC = MediaKind(0x1)
	MediaKind_MOVIE = MediaKind(0x2)
)

func (e MediaKind) String() string {
	s, ok := MediaKindNames[e]
	if ok {
		return s
	}
	return fmt.Sprintf("MediaKind_0x%X", int(e))
}

func (e MediaKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *MediaKind) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, ok := MediaKindValues[s]
	if !ok {
		return fmt.Errorf("unknown MediaKind %s", s)
	}
	*e = v
	return nil
}


type ICloudStatus int

var ICloudStatusNames = map[ICloudStatus]string{
	ICloudStatus(0x2): "Matched",
	ICloudStatus(0x3): "Uploaded",
	ICloudStatus(0x4): "Ineligible",
	ICloudStatus(0x5): "LocalOnly",
	ICloudStatus(0x6): "Duplicate",
	ICloudStatus(0x1): "Purchased",
}
var ICloudStatusValues = map[string]ICloudStatus{
	"Purchased": ICloudStatus(0x1),
	"Matched": ICloudStatus(0x2),
	"Uploaded": ICloudStatus(0x3),
	"Ineligible": ICloudStatus(0x4),
	"LocalOnly": ICloudStatus(0x5),
	"Duplicate": ICloudStatus(0x6),
}
const (
	ICloudStatus_LOCALONLY = ICloudStatus(0x5)
	ICloudStatus_DUPLICATE = ICloudStatus(0x6)
	ICloudStatus_PURCHASED = ICloudStatus(0x1)
	ICloudStatus_MATCHED = ICloudStatus(0x2)
	ICloudStatus_UPLOADED = ICloudStatus(0x3)
	ICloudStatus_INELIGIBLE = ICloudStatus(0x4)
)

func (e ICloudStatus) String() string {
	s, ok := ICloudStatusNames[e]
	if ok {
		return s
	}
	return fmt.Sprintf("ICloudStatus_0x%X", int(e))
}

func (e ICloudStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *ICloudStatus) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, ok := ICloudStatusValues[s]
	if !ok {
		return fmt.Errorf("unknown ICloudStatus %s", s)
	}
	*e = v
	return nil
}


type LimitMethod int

var LimitMethodNames = map[LimitMethod]string{
	LimitMethod(0x5): "GB",
	LimitMethod(0x1): "minutes",
	LimitMethod(0x2): "MB",
	LimitMethod(0x3): "items",
	LimitMethod(0x4): "hours",
}
var LimitMethodValues = map[string]LimitMethod{
	"hours": LimitMethod(0x4),
	"GB": LimitMethod(0x5),
	"minutes": LimitMethod(0x1),
	"MB": LimitMethod(0x2),
	"items": LimitMethod(0x3),
}
const (
	LimitMethod_MINUTES = LimitMethod(0x1)
	LimitMethod_MB = LimitMethod(0x2)
	LimitMethod_ITEMS = LimitMethod(0x3)
	LimitMethod_HOURS = LimitMethod(0x4)
	LimitMethod_GB = LimitMethod(0x5)
)

func (e LimitMethod) String() string {
	s, ok := LimitMethodNames[e]
	if ok {
		return s
	}
	return fmt.Sprintf("LimitMethod_0x%X", int(e))
}

func (e LimitMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *LimitMethod) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, ok := LimitMethodValues[s]
	if !ok {
		return fmt.Errorf("unknown LimitMethod %s", s)
	}
	*e = v
	return nil
}


type LoveStatus int

var LoveStatusNames = map[LoveStatus]string{
	LoveStatus(0x0): "None",
	LoveStatus(0x2): "Loved",
	LoveStatus(0x3): "Disliked",
}
var LoveStatusValues = map[string]LoveStatus{
	"None": LoveStatus(0x0),
	"Loved": LoveStatus(0x2),
	"Disliked": LoveStatus(0x3),
}
const (
	LoveStatus_NONE = LoveStatus(0x0)
	LoveStatus_LOVED = LoveStatus(0x2)
	LoveStatus_DISLIKED = LoveStatus(0x3)
)

func (e LoveStatus) String() string {
	s, ok := LoveStatusNames[e]
	if ok {
		return s
	}
	return fmt.Sprintf("LoveStatus_0x%X", int(e))
}

func (e LoveStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *LoveStatus) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, ok := LoveStatusValues[s]
	if !ok {
		return fmt.Errorf("unknown LoveStatus %s", s)
	}
	*e = v
	return nil
}


type LocationStatus int

var LocationStatusNames = map[LocationStatus]string{
	LocationStatus(0x1): "Computer",
	LocationStatus(0x10): "iCloud",
}
var LocationStatusValues = map[string]LocationStatus{
	"Computer": LocationStatus(0x1),
	"iCloud": LocationStatus(0x10),
}
const (
	LocationStatus_COMPUTER = LocationStatus(0x1)
	LocationStatus_ICLOUD = LocationStatus(0x10)
)

func (e LocationStatus) String() string {
	s, ok := LocationStatusNames[e]
	if ok {
		return s
	}
	return fmt.Sprintf("LocationStatus_0x%X", int(e))
}

func (e LocationStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *LocationStatus) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, ok := LocationStatusValues[s]
	if !ok {
		return fmt.Errorf("unknown LocationStatus %s", s)
	}
	*e = v
	return nil
}


