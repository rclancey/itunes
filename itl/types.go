package itl

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/htmlindex"
)

type Sized interface {
	GetSize() int
}

type ReadableObject interface {
	Read() error
	GetType() string
	GetOffset() int
	GetSize() int
	GetData() []byte
	Buffer() *bytes.Buffer
	Parse(interface{}) error
}

type ObjectSignature [4]byte

func (v ObjectSignature) String() string {
	return string(v[:])
}

func (v ObjectSignature) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v ObjectSignature) GetEndianness() (string, binary.ByteOrder) {
	k := v.String()
	if strings.HasSuffix(k, "h") {
		out := []byte{v[3], v[2], v[1], v[0]}
		return string(out), binary.LittleEndian
	}
	return k, binary.BigEndian
}

var nullStr = string([]byte{0})
func nullTerminatedString(data []byte) string {
	return string(bytes.TrimRight(data, nullStr))
}

type ApplicationVersion [31]byte

func (o ApplicationVersion) String() string {
	return nullTerminatedString(o[:])
}

func (o ApplicationVersion) Major() (int, error) {
	return strconv.Atoi(strings.Split(o.String(), ".")[0])
}

func (o ApplicationVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

type Time uint32

func (t Time) Time() time.Time {
	return time.Unix(int64(t) + int64(macEpoch), 0)
}

func (t Time) String() string {
	return t.Time().Format(time.RFC3339)
}

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time())
}

type PersistentID uint64

func (id PersistentID) String() string {
	v := strings.ToUpper(strconv.FormatUint(uint64(id), 16))
	return strings.Repeat("0", 16 - len(v)) + v
}

func (id PersistentID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

type DataType uint32

var compactDataTypes = map[DataType]bool{
	1: true,
	19: true,
	21: true,
	36: true,
	48: true,
	50: true,
	54: true,
	56: true,
	66: true,
	101: true,
	102: true,
	103: true,
	105: true,
	107: true,
	109: true,
	402: true,
	514: true,
	700: true,
	800: true,
}

func (dt DataType) String() string {
	s, ok := DATA_TYPE[uint32(dt)]
	if !ok {
		return fmt.Sprintf("Z%d", uint32(dt))
	}
	return s
}

func (dt DataType) HasExtendedData() bool {
	_, ok := compactDataTypes[dt]
	if ok {
		return false
	}
	return true
}

type DataSetType uint32

func (dst DataSetType) String() string {
	s, ok := DATASET_TYPE[uint32(dst)]
	if !ok {
		return fmt.Sprintf("Z%d", uint32(dst))
	}
	return s
}

type SortField uint16

func (sf SortField) String() string {
	s, ok := DATA_TYPE[uint32(sf)]
	if !ok {
		return fmt.Sprintf("Z%d", uint32(sf))
	}
	return s
}

type PlaylistKind uint32

func (pk PlaylistKind) String() string {
	s, ok := PLAYLIST_KIND_TYPE[uint32(pk)]
	if !ok {
		return fmt.Sprintf("Z%d", uint32(pk))
	}
	return s
}

func utf16ToUtf8(data []byte, typ DataType) (string, error) {
	n := len(data)
	if n % 2 != 0 {
		log.Printf("utf16 value %s for %s has odd number of bytes", string(data), typ)
		return string(data), nil
	}
	buf := &bytes.Buffer{}
	u16buf := make([]uint16, 1)
	u8buf := make([]byte, 4)
	for i := 0; i < n; i += 2 {
		u16buf[0] = (uint16(data[i]) << 8) + uint16(data[i+1])
		r := utf16.Decode(u16buf)
		m := utf8.EncodeRune(u8buf, r[0])
		buf.Write(u8buf[:m])
	}
	return buf.String(), nil
}

type StandardObject struct {
	Type string
	Offset int
	Preface interface{}
	Size int
	ByteOrder binary.ByteOrder
	Data []byte
}

func ReadObject(r io.Reader, offset int) (int, interface{}, error) {
	n, std, err := NewStandardObject(r, offset)
	if err != nil {
		return n, std, err
	}
	obj := std.CreateSubObject()
	err = obj.Read()
	return n, obj, err
}

func NewStandardObject(r io.Reader, offset int) (int, *StandardObject, error) {
	bytesRead := 0
	var sig ObjectSignature
	err := binary.Read(r, binary.BigEndian, &sig)
	bytesRead += 4
	if err != nil {
		return bytesRead, nil, errors.WithStack(err)
	}
	if sig[0] != 'h' && sig[3] != 'h' {
		return bytesRead, nil, io.EOF
	}
	o := &StandardObject{
		Type: sig.String(),
		Offset: offset,
		Size: 0,
		ByteOrder: binary.BigEndian,
		Data: []byte{},
	}
	if sig[3] == 'h' {
		o.ByteOrder = binary.LittleEndian
		o.Type = string([]byte{sig[3], sig[2], sig[1], sig[0]})
	}
	if o.Type == "hlrm" {
		return bytesRead, nil, io.EOF
	}
	buf := &bytes.Buffer{}
	buf.Write(sig[:])
	var size uint32
	err = binary.Read(r, o.ByteOrder, &size)
	bytesRead += 4
	if err != nil {
		return bytesRead, nil, errors.WithStack(err)
	}
	binary.Write(buf, o.ByteOrder, size)
	if o.Type == "hohm" {
		o.Preface = int(size)
		err = binary.Read(r, o.ByteOrder, &size)
		bytesRead += 4
		if err != nil {
			return bytesRead, nil, errors.WithStack(err)
		}
		binary.Write(buf, o.ByteOrder, size)
	}
	o.Size = int(size)
	o.Data = make([]byte, o.Size)
	if o.Size == 0 {
		log.Printf("%s has size zero", sig)
		return bytesRead, o, io.EOF
	}
	copy(o.Data[:buf.Len()], buf.Bytes())
	n, err := io.ReadFull(r, o.Data[bytesRead:])
	bytesRead += n
	if err != nil && errors.Is(err, io.ErrUnexpectedEOF) {
		o.Data = o.Data[:bytesRead]
		return bytesRead, o, nil
	}
	return bytesRead, o, errors.WithStack(err)
}

func (o *StandardObject) CreateSubObject() ReadableObject {
	switch o.Type {
	case "hdfm":
		return &Database{StandardObject: o}
	case "hdsm":
		return &DataSet{StandardObject: o}
	case "hghm":
		return &HGHM{StandardObject: o}
	case "hohm":
		return &DataObject{StandardObject: o}
	case "halm":
		return &AlbumList{StandardObject: o}
	case "haim":
		return &Album{StandardObject: o}
	case "hilm":
		return &ArtistList{StandardObject: o}
	case "hiim":
		return &Artist{StandardObject: o}
	case "htlm":
		return &TrackList{StandardObject: o}
	case "htim":
		return &Track{StandardObject: o}
	case "hplm":
		return &PlaylistList{StandardObject: o}
	case "hpim":
		return &Playlist{StandardObject: o}
	case "hptm":
		return &PlaylistItem{StandardObject: o}
	case "hqlm":
		return &QueryList{StandardObject: o}
	case "hqim":
		return &QueryItem{StandardObject: o}
	}
	return &Unhandled{StandardObject: o}
}

func (o *StandardObject) Buffer() *bytes.Buffer {
	return bytes.NewBuffer(o.Data)
}

func (o *StandardObject) Parse(x interface{}) error {
	buf := o.Buffer()
	err := binary.Read(buf, o.ByteOrder, x)
	return errors.WithStack(err)
}

func (o *StandardObject) GetType() string {
	return o.Type
}

func (o *StandardObject) GetOffset() int {
	return o.Offset
}

func (o *StandardObject) GetSize() int {
	return o.Size
}

func (o *StandardObject) GetData() []byte {
	return o.Data
}

// hdfm
// IIB31sIQIBBBBIIIHHIIIIi
type Database struct {
	*StandardObject
	Parsed *DatabaseInner
	PersistentID PersistentID
	ApplicationVersion ApplicationVersion
	MajorVersion int
	MinorVersion int
	MaxCryptSize int
	TZOffset int
}

type DatabaseInner struct {
	Type ObjectSignature // 0
	Size uint32 // 4
	A [4]byte // 8
	B [4]byte // 12
	StrLen uint8 // 16
	ApplicationVersion ApplicationVersion `json:",string"` // 17
	E [4]byte // 48
	PersistentID PersistentID `json:",string"` // 52
	G [4]byte  // 60
	H uint8 // 64
	MajorVersion uint8 // 65
	J uint8 // 66
	MinorVersion uint8 // 67
	L uint32 // 68
	M [4]byte // 72
	N [4]byte // 76
	O uint16 // 80
	P uint16 // 82
	Q [4]byte // 84
	//DatabaseDate Time `json:",string"`
	R uint32 // 88
	MaxCryptSize uint32 // 92
	T uint32 // 96
	TZOffset int32 // 100
}

func (o *Database) Read() error {
	o.Parsed = &DatabaseInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.PersistentID = o.Parsed.PersistentID
	o.ApplicationVersion = o.Parsed.ApplicationVersion
	o.MajorVersion = int(o.Parsed.MajorVersion)
	o.MinorVersion = int(o.Parsed.MinorVersion)
	o.MaxCryptSize = int(o.Parsed.MaxCryptSize)
	o.TZOffset = int(o.Parsed.TZOffset)
	return nil
}

//hdsm
type DataSet struct {
	*StandardObject
	Parsed *DataSetInner
	RecordBytes int
}

func (o *DataSet) Read() error {
	o.Parsed = &DataSetInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.RecordBytes = int(o.Parsed.RecordBytes)
	return nil
}

type DataSetInner struct {
	Type ObjectSignature
	Size uint32
	RecordBytes uint32
}

//hghm
type HGHM struct {
	*StandardObject
	Parsed *HGHMInner
	RecordCount int
	LibraryDate time.Time
}

func (o *HGHM) Read() error {
	o.Parsed = &HGHMInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.RecordCount = int(o.Parsed.RecordCount)
	o.LibraryDate = o.Parsed.LibraryDate.Time()
	return nil
}

type HGHMInner struct {
	Type ObjectSignature
	Size uint32
	RecordCount uint32
	Unknown1 uint32
	LibraryDate Time
}

// hohm
type DataObject struct {
	*StandardObject
	Parsed *DataObjectInner
	TypeID DataType
	Raw []byte
	Str string
}

type DataObjectInner struct {
	Type ObjectSignature
	Unknown1 uint32
	Size uint32
	TypeID DataType
	Unknown2 uint32
	Unknown3 uint32
	Unknown4 uint32
	DataSize uint32
	Unknown5 uint32
	Unknown6 uint32
}

func (o *DataObject) Read() error {
	r := o.Buffer()
	inner := &DataObjectInner{}
	err := binary.Read(r, o.ByteOrder, inner)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return errors.WithStack(err)
	}
	o.Parsed = inner
	o.TypeID = inner.TypeID

	//xsize := inner.Size - (inner.Unknown1 + 8)
	if o.Parsed.Unknown4 & 256 != 0 || inner.DataSize > inner.Size {
		return nil
	}
	/*
	if inner.DataSize > inner.Size - uint32(binary.Size(inner)) {
	//if xsize <= 12 || !inner.TypeID.HasExtendedData() {
		return nil
	}
	*/

	/*
	if inner.DataSize > 2048 {
		log.Println("big data object", *inner)
		//return errors.New("big data object")
	}
	*/
	/*
	if inner.DataSize > inner.Size {
		log.Printf("data object overflow: %d > %d", inner.DataSize, inner.Size)
		log.Println("object", *inner)
		//return errors.WithStack(ErrTooBig)
	}
	*/
	data := make([]byte, int(inner.DataSize))
	n, err := io.ReadFull(r, data)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			dump, _ := json.Marshal(o)
			log.Printf("trying to read %d bytes, got %d bytes and EOF (%s)", len(data), n, string(dump))
			return nil
		}
		return errors.WithStack(err)
	}
	o.Raw = data
	switch o.TypeID {
	case 1, 19, 101, 102, 104, 105:
	case 500, 504, 505, 506, 508:
		o.Str, err = utf16ToUtf8(data, o.TypeID)
		if err != nil {
			return err
		}
	case 11:
		o.Str = string(o.Raw)
	case 303:
		out := make([]byte, len(o.Raw))
		copy(out, o.Raw)
		for i := 0; i + 1 < len(o.Raw); i += 2 {
			out[i], out[i+1] = o.Raw[i+1], o.Raw[i]
		}
		o.Str = string(out)
	default:
		if inner.Unknown4 & 2 == 2 {
			enc, err := htmlindex.Get("ISO-8859-1")
			if err != nil {
				return errors.WithStack(err)
			}
			txt, err := enc.NewDecoder().Bytes(data)
			if err != nil {
				return errors.WithStack(err)
			}
			o.Str = string(txt)
		} else if inner.Unknown4 & 1 == 1 {
			o.Str, err = utf16ToUtf8(data, o.TypeID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// halm
type AlbumList struct {
	*StandardObject
	Parsed *AlbumListInner
	RecordCount int
}

func (o *AlbumList) Read() error {
	o.Parsed = &AlbumListInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.RecordCount = int(o.Parsed.RecordCount)
	return nil
}

type AlbumListInner struct {
	Type ObjectSignature
	Size uint32
	RecordCount uint32
}

// haim
type Album struct {
	*StandardObject
	Parsed *AlbumInner
	RecordCount int
	Sequence int
	AlbumID PersistentID
	ArtistID PersistentID
	AlbumRating int
}

func (o *Album) Read() error {
	o.Parsed = &AlbumInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.RecordCount = int(o.Parsed.RecordCount)
	o.Sequence = int(o.Parsed.Sequence)
	o.AlbumID = o.Parsed.AlbumID
	o.ArtistID = o.Parsed.ArtistID
	o.AlbumRating = int(o.Parsed.AlbumRating)
	return nil
}

type AlbumInner struct {
	Type ObjectSignature
	Size uint32
	SubBytes uint32
	RecordCount uint32
	Sequence uint32
	AlbumID PersistentID `json:",string"`
	Unknown1 [4]byte
	ArtistID PersistentID `json:",string"`
	Unknown2 uint16
	AlbumRating int8
}

// hilm
type ArtistList struct {
	*StandardObject
	Parsed *ArtistListInner
	RecordCount int
}

func (o *ArtistList) Read() error {
	o.Parsed = &ArtistListInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.RecordCount = int(o.Parsed.RecordCount)
	return nil
}

type ArtistListInner struct {
	Type ObjectSignature
	Size uint32
	RecordCount uint32
}

// hiim
type Artist struct {
	*StandardObject
	Parsed *ArtistInner
	RecordCount int
	Sequence int
	ArtistID PersistentID
}

func (o *Artist) Read() error {
	o.Parsed = &ArtistInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.RecordCount = int(o.Parsed.RecordCount)
	o.Sequence = int(o.Parsed.Sequence)
	o.ArtistID = o.Parsed.ArtistID
	return nil
}

type ArtistInner struct {
	Size uint32
	SubBytes uint32
	RecordCount uint32
	Sequence uint32
	ArtistID PersistentID `json:",string"`
}

// htlm
type TrackList struct {
	*StandardObject
	Parsed *TrackListInner
	RecordCount int
}

func (o *TrackList) Read() error {
	o.Parsed = &TrackListInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.RecordCount = int(o.Parsed.RecordCount)
	return nil
}

type TrackListInner struct {
	Type ObjectSignature
	Size uint32
	RecordCount uint32
}

// htim
type Track struct {
	*StandardObject
	Parsed *TrackInner
	RecordCount int
	TrackID int
	BlockType int
	FileType string
	DateModified time.Time
	FileSize int
	TotalTime int
	TrackNumber int
	TrackCount int
	Year int
	BitRate int
	SampleRate int
	VolumeAdjustment int
	StartTime int
	StopTime int
	PlayCount int
	Compilation bool
	PlayDate time.Time
	DiscNumber int
	DiscCount int
	Rating int
	BPM int
	DateAdded time.Time
	Disabled bool
	PersistentID PersistentID
	PurchaseDate time.Time
	ReleaseDate time.Time
	AlbumSequence int
	BackupDate time.Time
	SampleCount int
	SkipCount int
	SkipDate time.Time
}

func (o *Track) Read() error {
	o.Parsed = &TrackInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		expSize := binary.Size(o.Parsed)
		log.Printf("expected %d bytes for track, but only %d available", expSize, o.StandardObject.Size)
		return err
	}
	o.RecordCount = int(o.Parsed.RecordCount)
	o.TrackID = int(o.Parsed.TrackID)
	o.BlockType = int(o.Parsed.BlockType)
	o.FileType = string(o.Parsed.FileType[:])
	o.DateModified = o.Parsed.DateModified.Time()
	o.FileSize = int(o.Parsed.FileSize)
	o.TotalTime = int(o.Parsed.TotalTime)
	o.TrackNumber = int(o.Parsed.TrackNumber)
	o.TrackCount = int(o.Parsed.TrackCount)
	o.Year = int(o.Parsed.Year)
	o.BitRate = int(o.Parsed.BitRate)
	o.SampleRate = int(o.Parsed.SampleRate)
	o.VolumeAdjustment = int(o.Parsed.VolumeAdjustment)
	o.StartTime = int(o.Parsed.StartTime)
	o.StopTime = int(o.Parsed.StopTime)
	o.PlayCount = int(o.Parsed.PlayCount)
	o.Compilation = o.Parsed.Compilation != 0
	o.PlayDate = o.Parsed.PlayDate.Time()
	o.DiscNumber = int(o.Parsed.DiscNumber)
	o.DiscCount = int(o.Parsed.DiscCount)
	o.Rating = int(o.Parsed.Rating)
	o.BPM = int(o.Parsed.BPM)
	o.DateAdded = o.Parsed.DateAdded.Time()
	o.Disabled = o.Parsed.Disabled != 0
	o.PersistentID = o.Parsed.PersistentID
	o.PurchaseDate = o.Parsed.PurchaseDate.Time()
	o.ReleaseDate = o.Parsed.ReleaseDate.Time()
	o.AlbumSequence = int(o.Parsed.AlbumSequence)
	o.BackupDate = o.Parsed.BackupDate.Time()
	o.SampleCount = int(o.Parsed.SampleCount)
	if len(o.StandardObject.Data) >= 288 {
		buf := bytes.NewBuffer(o.StandardObject.Data[280:])
		var count uint32
		var t Time
		err := binary.Read(buf, o.StandardObject.ByteOrder, &count)
		if err == nil {
			o.SkipCount = int(count)
		}
		err = binary.Read(buf, o.StandardObject.ByteOrder, &t)
		if err == nil {
			o.SkipDate = t.Time()
		}
	}
	return nil
}


type TrackInner struct {
	Type ObjectSignature      // 0
	Size uint32               // 4
	SubBytes uint32           // 8
	RecordCount uint32        // 12
	TrackID uint32            // 16
	BlockType uint32          // 20
	Unknown1 uint32           // 24
	FileType [4]byte          // 28
	DateModified Time         // 32
	FileSize uint32           // 36
	TotalTime uint32          // 40
	TrackNumber uint32        // 44
	TrackCount uint32         // 48
	Unknown2 uint16           // 52
	Year uint16               // 54
	Unknown3 uint16           // 56
	BitRate uint16            // 58
	SampleRate uint16         // 60
	Unknown4 uint16           // 62
	VolumeAdjustment int32    // 64
	StartTime uint32          // 68
	StopTime uint32           // 72
	PlayCount uint32          // 76
	Unknown5 uint16           // 80
	Compilation uint16        // 82
	Unknown6 [12]byte         // 84
	Unknown7 uint32           // 96
	PlayDate Time             // 100
	DiscNumber uint16         // 104
	DiscCount uint16          // 106
	Rating uint8              // 108
	BPM uint8                 // 109
	Unknown8 [10]byte         // 110
	DateAdded Time            // 120
	Disabled uint32           // 124
	PersistentID PersistentID // 128
	Unknown9 uint32           // 136
	FileType2 [4]byte         // 140
	Unknown10 [3]uint32       // 144
	PurchaseDate Time         // 156
	ReleaseDate Time // 39    // 160
	Unknown11 [13]uint32      // 164
	Protected uint32          // 216
	AlbumSequence uint32      // 220
	BackupDate Time           // 224
	Unknown12 [5]uint32       // 228
	SampleCount uint32        // 248
	//Unknown13 [7]uint32       // 252
	//SkipCount uint32          // 280
	//SkipDate Time             // 284
	//Remainder [420]byte
}

// hplm
type PlaylistList struct {
	*StandardObject
	Parsed *PlaylistListInner
	RecordCount int
}

func (o *PlaylistList) Read() error {
	o.Parsed = &PlaylistListInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.RecordCount = int(o.Parsed.RecordCount)
	return nil
}

type PlaylistListInner struct {
	Type ObjectSignature
	Size uint32
	RecordCount uint32
}

// hpim
type Playlist struct {
	*StandardObject
	Parsed *PlaylistInner
	RecordCount int
	TrackCount int
	SortOrderFieldType SortField
	DateAdded time.Time
	DateModified time.Time
	PersistentID PersistentID
	ParentPersistentID PersistentID
	Folder bool
	GeniusTrackID int

	DistinguishedKind PlaylistKind `json:",string"`
	Master bool
	Audiobooks bool
	Movies bool
	Music bool
	Podcasts bool
	PurchasedMusic bool
	TVShows bool
}

func (o *Playlist) Read() error {
	o.Parsed = &PlaylistInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		expSize := binary.Size(o.Parsed)
		log.Printf("expected %d bytes for playlist, but only %d available", expSize, o.StandardObject.Size)
		/*
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil
		}
		*/
		return err
	}
	o.RecordCount = int(o.Parsed.RecordCount)
	o.TrackCount = int(o.Parsed.TrackCount)
	o.SortOrderFieldType = o.Parsed.SortOrderFieldType
	o.DateAdded = o.Parsed.DateAdded.Time()
	if len(o.StandardObject.Data) >= 632 {
		buf := bytes.NewBuffer(o.StandardObject.Data[628:632])
		var t Time
		err := binary.Read(buf, o.StandardObject.ByteOrder, &t)
		if err == nil {
			o.DateModified = t.Time()
		}
	}
	o.PersistentID = o.Parsed.PersistentID
	o.ParentPersistentID = o.Parsed.ParentPersistentID
	data := o.StandardObject.Data[8:]
	o.Folder = data[514] != 0
	o.DistinguishedKind = PlaylistKind(int(data[561]))
	switch o.DistinguishedKind.String() {
	case "Books":
		o.Audiobooks = true
	case "Movies":
		o.Movies = true
	case "TV Shows":
		o.TVShows = true
	case "Music":
		o.Music = true
	case "Podcasts":
		o.Podcasts = true
	case "Purchased":
		o.PurchasedMusic = true
	}
	return nil
}

type PlaylistInner struct {
	Type ObjectSignature                     // 0
	Size uint32                              // 4
	Unknown1 uint32 `json:",omitempty"`      // 8
	RecordCount uint32                       // 12
	TrackCount uint32                        // 16
	Unknown2 uint32 `json:",omitempty"`      // 20
	SortOrderFieldType SortField             // 24
	Unknown3 uint16 `json:",omitempty"`      // 26
	DateAdded Time                           // 28
	Unknown4 [36]uint32 `json:",omitempty"`  // 32
	Unknown5 [65]uint32 `json:",omitempty"`  // 176
	Unknown6 uint16 `json:",omitempty"`      // 436
	Unknown7 uint16 `json:",omitempty"`      // 438
	PersistentID PersistentID                // 440
	Unknown8 [18]uint32 `json:",omitempty"`  // 448
	Unknown9 uint16 `json:",omitempty"`      // 520
	Folder uint16                            // 522
	Unknown10 uint32 `json:",omitempty"`     // 524
	ParentPersistentID PersistentID          // 528
	Unknown11 [7]uint32 `json:",omitempty"`  // 536
	PersistentID2 PersistentID               // 564
	//Unknown12 [14]uint32 `json:",omitempty"` // 572
	//DateModified Time                        // 628
}

// hptm
type PlaylistItem struct {
	*StandardObject
	Parsed *PlaylistItemInner
	Sequence int
	TrackID int
	Position int
}

func (o *PlaylistItem) Read() error {
	o.Parsed = &PlaylistItemInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.Sequence = int(o.Parsed.Sequence)
	o.TrackID = int(o.Parsed.TrackID)
	o.Position = int(o.Parsed.Position)
	return nil
}

type PlaylistItemInner struct {
	Type ObjectSignature
	Size uint32
	Unknown1 uint32
	Unknown2 uint32
	Sequence uint32
	Unknown4 uint32
	TrackID uint32
	Unknown5 uint32
	Position uint32
}

// hqlm
type QueryList struct {
	*StandardObject
	Parsed *QueryListInner
}

func (o *QueryList) Read() error {
	o.Parsed = &QueryListInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	return nil
}

type QueryListInner struct {
	Type ObjectSignature
	Size uint32
}

// hqim
type QueryItem struct {
	*StandardObject
	Parsed *QueryItemInner
}

func (o *QueryItem) Read() error {
	o.Parsed = &QueryItemInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	return nil
}

type QueryItemInner struct {
	Type ObjectSignature
	Size uint32
}

// hrlm, hrpm
type Unhandled struct {
	*StandardObject
	Parsed *UnhandledInner
}

func (o *Unhandled) Read() error {
	o.Parsed = &UnhandledInner{}
	return o.StandardObject.Parse(o.Parsed)
}

type UnhandledInner struct {
	Type ObjectSignature
	Size uint32
}
