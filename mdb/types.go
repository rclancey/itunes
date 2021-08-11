package mdb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Sized interface {
	GetSize() int
}

type ReadableObject interface {
	Read() error
}

type ObjectSignature [4]byte

func (v ObjectSignature) String() string {
	return string(v[:])
}

func (v ObjectSignature) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

var nullStr = string([]byte{0})
func nullTerminatedString(data []byte) string {
	return string(bytes.TrimRight(data, nullStr))
}

type ApplicationVersion [32]byte

func (v ApplicationVersion) String() string {
	return nullTerminatedString(v[:])
}

func (v ApplicationVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

type Time uint32

func (t Time) Time() time.Time {
	return time.Unix(int64(t) + int64(macEpoch), 0)
}

func (t Time) String() string {
	return t.Time().Format(time.RFC3339)
}

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

type PersistentID uint64

func (id PersistentID) String() string {
	v := strings.ToUpper(strconv.FormatUint(uint64(id), 16))
	return strings.Repeat("0", 16 - len(v)) + v
}

func (id PersistentID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

type StandardObject struct {
	Type string
	Offset int
	Preface interface{}
	Size int
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
	err := binary.Read(r, binary.LittleEndian, &sig)
	bytesRead += 4
	if err != nil {
		return bytesRead, nil, errors.WithStack(err)
	}
	o := &StandardObject{
		Type: sig.String(),
		Offset: offset,
		Size: 0,
		Data: []byte{},
	}
	buf := &bytes.Buffer{}
	buf.Write(sig[:])
	var size uint32
	err = binary.Read(r, binary.LittleEndian, &size)
	bytesRead += 4
	if err != nil {
		return bytesRead, nil, errors.WithStack(err)
	}
	binary.Write(buf, binary.LittleEndian, size)
	if o.Type == "boma" {
		o.Preface = int(size)
		err = binary.Read(r, binary.LittleEndian, &size)
		bytesRead += 4
		if err != nil {
			return bytesRead, nil, errors.WithStack(err)
		}
		binary.Write(buf, binary.LittleEndian, size)
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
	if err != nil  && errors.Is(err, io.ErrUnexpectedEOF) {
		o.Data = o.Data[:bytesRead]
		return bytesRead, o, nil
	}
	return bytesRead, o, errors.WithStack(err)
}

func (o *StandardObject) CreateSubObject() ReadableObject {
	switch o.Type {
	case "hfma":
		return &Envelope{StandardObject: o}
	case "hsma":
		return &SectionBoundary{StandardObject: o}
	case "plma":
		return &LibraryMaster{StandardObject: o}
	case "lama":
		return &AlbumList{StandardObject: o}
	case "iama":
		return &Album{StandardObject: o}
	case "lAma":
		return &ArtistList{StandardObject: o}
	case "iAma":
		return &Artist{StandardObject: o}
	case "ltma":
		return &TrackList{StandardObject: o}
	case "itma":
		return &Track{StandardObject: o}
	case "lPma":
		return &PlaylistList{StandardObject: o}
	case "lpma":
		return &Playlist{StandardObject: o}
	case "boma":
		return &DataObject{StandardObject: o}
	}
	return &Unhandled{StandardObject: o}
}

func (o *StandardObject) Buffer() *bytes.Buffer {
	return bytes.NewBuffer(o.Data)
}

func (o *StandardObject) Parse(x interface{}) error {
	buf := o.Buffer()
	err := binary.Read(buf, binary.LittleEndian, x)
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

type Envelope struct {
	*StandardObject
	Parsed *EnvelopeInner
	PersistentID PersistentID
	ApplicationVersion ApplicationVersion
	MajorVersion int
	MinorVersion int
	MaxCryptSize int
	ItemCount int
	PlaylistCount int
	TZOffset int
}
// hfma
func (o *Envelope) Read() error {
	o.Parsed = &EnvelopeInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.PersistentID = o.Parsed.PersistentID
	o.ApplicationVersion = o.Parsed.ApplicationVersion
	o.MajorVersion = int(o.Parsed.MajorVersion)
	o.MinorVersion = int(o.Parsed.MinorVersion)
	o.MaxCryptSize = int(o.Parsed.MaxCryptSize)
	o.ItemCount = int(o.Parsed.ItemCount)
	o.PlaylistCount = int(o.Parsed.PlaylistCount)
	o.TZOffset = int(o.Parsed.TZOffset)
	return nil
}

type EnvelopeInner struct {
	Type ObjectSignature
	Size uint32
	FileLength uint32
	MajorVersion uint16
	MinorVersion uint16
	ApplicationVersion ApplicationVersion
	PersistentID PersistentID
	FileTypeID uint32
	Unknown1 [2]uint32
	ItemCount uint32
	PlaylistCount uint32
	CollectionCount uint32
	ArtistCount uint32
	MaxCryptSize uint32
	TZOffset int32
	AppleStoreID PersistentID
	LibraryDate Time
	Unknown3 uint32
	ITunesLibraryPersistentID PersistentID
}

type SectionBoundary struct {
	*StandardObject
	Parsed *SectionBoundaryInner
	SubType int
	SectionsLength int
}

func (o *SectionBoundary) Read() error {
	o.Parsed = &SectionBoundaryInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.SubType = int(o.Parsed.SubType)
	o.SectionsLength = int(o.Parsed.SectionsLength)
	return nil
}

type SectionBoundaryInner struct {
	Type ObjectSignature
	Size uint32
	SectionsLength uint32
	SubType uint32
}

type LibraryMaster struct {
	*StandardObject
	Parsed *LibraryMasterInner
	PersistentID PersistentID
	DataObjectCount int
}

func (o *LibraryMaster) Read() error {
	o.Parsed = &LibraryMasterInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.PersistentID = o.Parsed.LibraryPersistentID
	o.DataObjectCount = int(o.Parsed.DataObjectCount)
	return nil
}

type LibraryMasterInner struct {
	Type ObjectSignature
	Size uint32
	DataObjectCount uint32
	Unknown1 [11]uint32
	Unknown2 uint16
	LibraryPersistentID PersistentID
	Unknown3 [5]uint32
	LibraryPersistentID2 PersistentID
}

type AlbumList struct {
	*StandardObject
	Parsed *AlbumListInner
	AlbumCount int
}

func (o *AlbumList) Read() error {
	o.Parsed = &AlbumListInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.AlbumCount = int(o.Parsed.AlbumCount)
	return nil
}

type AlbumListInner struct {
	Type ObjectSignature
	Size uint32
	AlbumCount uint32
	Unknown [4]byte
}

type Album struct {
	*StandardObject
	Parsed *AlbumInner
	PersistentID PersistentID
	DataObjectCount int
}

func (o *Album) Read() error {
	o.Parsed = &AlbumInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.PersistentID = o.Parsed.PersistentID
	o.DataObjectCount = int(o.Parsed.DataObjectCount)
	return nil
}

type AlbumInner struct {
	Type ObjectSignature
	Size uint32
	SectionsLength uint32
	DataObjectCount uint32
	PersistentID PersistentID
}

type ArtistList struct {
	*StandardObject
	Parsed *ArtistListInner
	ArtistCount int
}

func (o *ArtistList) Read() error {
	o.Parsed = &ArtistListInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.ArtistCount = int(o.Parsed.ArtistCount)
	return nil
}

type ArtistListInner struct {
	Type ObjectSignature
	Size uint32
	ArtistCount uint32
}

type Artist struct {
	*StandardObject
	Parsed *ArtistInner
	PersistentID PersistentID
	DataObjectCount int
}

func (o *Artist) Read() error {
	o.Parsed = &ArtistInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.PersistentID = o.Parsed.PersistentID
	o.DataObjectCount = int(o.Parsed.DataObjectCount)
	return nil
}

type ArtistInner struct {
	Type ObjectSignature
	Size uint32
	SectionsLength uint32
	DataObjectCount uint32
	PersistentID PersistentID
}

type TrackList struct {
	*StandardObject
	Parsed *TrackListInner
	TrackCount int
}

func (o *TrackList) Read() error {
	o.Parsed = &TrackListInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.TrackCount = int(o.Parsed.TrackCount)
	return nil
}

type TrackListInner struct {
	Type ObjectSignature
	Size uint32
	TrackCount uint32
}

type Track struct {
	*StandardObject
	Parsed *TrackInner
	PersistentID PersistentID
	DataObjectCount int
	Disabled bool
	Love bool
	Stars int
	MovementCount int
	MovementNumber int
	TrackCount int
	TrackNumber int
	DiscCount int
	DiscNumber int
	Year int
	AlbumID PersistentID
	ArtistID PersistentID
}

func (o *Track) Read() error {
	o.Parsed = &TrackInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.DataObjectCount = int(o.Parsed.DataObjectCount)
	o.PersistentID = o.Parsed.PersistentID
	o.Disabled = o.Parsed.Disabled != 0
	o.Love = o.Parsed.Love != 0
	o.Stars = int(o.Parsed.Stars)
	o.MovementCount = int(o.Parsed.MovementCount)
	o.MovementNumber = int(o.Parsed.MovementNumber)
	o.TrackCount = int(o.Parsed.TrackCount)
	o.TrackNumber = int(o.Parsed.TrackNumber)
	o.DiscCount = int(o.Parsed.DiscCount)
	o.DiscNumber = int(o.Parsed.DiscNumber)
	o.Year = int(o.Parsed.Year)
	o.AlbumID = o.Parsed.AlbumID
	o.ArtistID = o.Parsed.ArtistID
	return nil
}

type TrackInner struct {
	Type ObjectSignature        // 0
	Size uint32                 // 4
	Unknown1 uint32             // 8
	DataObjectCount uint32      // 12
	PersistentID PersistentID   // 16
	Unknown2 [4]uint32          // 24
	Unknown3 uint16             // 40
	Disabled uint16             // 42
	Unknown4 [4]uint32          // 44
	Unknown5 uint16             // 60
	Love uint16                 // 62
	Unknown6 uint8              // 64
	Stars uint8                 // 65
	Unknown7 [4]uint32          // 66
	Unknown8 uint16             // 82
	DiscNumber uint16           // 84
	MovementCount uint16        // 86
	MovementNumber uint16       // 88
	DiscCount uint16            // 90
	Unknown9 uint16             // 92
	Unknown10 [5]uint32         // 94
	Unknown11 uint16            // 114
	TrackCount uint16           // 116
	Unknown12 [10]uint32        // 118
	Unknown13 uint16            // 158
	TrackNumber uint16          // 162
	Unknown14 uint16            // 164
	Unknown15 uint32            // 166
	Year uint16                 // 170
	Unknown16 uint16            // 172
	AlbumID PersistentID        // 174
	ArtistID PersistentID       // 182
	Unknown17 [21]uint32        // 190
	PersistentID2 PersistentID  // 274
}

type PlaylistList struct {
	*StandardObject
	Parsed *PlaylistListInner
	PlaylistCount int
}

func (o *PlaylistList) Read() error {
	o.Parsed = &PlaylistListInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.PlaylistCount = int(o.Parsed.PlaylistCount)
	return nil
}

type PlaylistListInner struct {
	Type ObjectSignature
	Size uint32
	PlaylistCount uint32
}

type Playlist struct {
	*StandardObject
	Parsed *PlaylistInner
	PersistentID PersistentID
	ParentPersistentID PersistentID
	DataObjectCount int
	TrackCount int
	DateAdded time.Time
	DateModified time.Time
	Folder bool
	DistinguishedKind PlaylistKind
}

func (o *Playlist) Read() error {
	o.Parsed = &PlaylistInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.PersistentID = o.Parsed.PersistentID
	o.ParentPersistentID = o.Parsed.ParentPersistentID
	o.DataObjectCount = int(o.Parsed.DataObjectCount)
	o.TrackCount = int(o.Parsed.TrackCount)
	o.DateAdded = o.Parsed.DateAdded.Time()
	o.DateModified = o.Parsed.DateModified.Time()
	o.Folder = o.Parsed.Folder != 0
	o.DistinguishedKind = PlaylistKind(o.StandardObject.Data[79])
	return nil
}

type PlaylistInner struct {
	Type ObjectSignature             // 0
	Size uint32                      // 4
	SectionsLength uint32            // 8
	DataObjectCount uint32           // 12
	TrackCount uint32                // 16
	Unknown1 uint16                  // 20
	DateAdded Time                   // 22
	Unknown2 uint16                  // 26
	Unknown3 uint16                  // 28
	PersistentID PersistentID        // 30
	Unknown4 [2]uint32               // 38
	Unknown5 [3]byte                 // 46
	Folder uint8                     // 49
	ParentPersistentID PersistentID  // 50
	Unknown6 [5]uint32               // 58
	Unknown7 byte                    // 78
	PlaylistKind uint8               // 79
	Unknown8 [2]byte                 // 80
	Unknown9 [14]uint32              // 82
	DateModified Time                // 138
}

type DataObject struct {
	*StandardObject
	Parsed *DataObjectInner
	Raw []byte
	Nums []uint32
	Key string
	NumericData *NumericDataObject `json:"NumericData,omitempty"`
	WideCharData *WideCharDataObject `json:"WideCharData,omitempty"`
	ShortHeaderData *ShortHeaderDataObject `json:"ShortHeaderData,omitempty"`
	LongHeaderData *LongHeaderDataObject `json:"LongHeaderData,omitempty"`
	PlaylistItemData *PlaylistItemDataObject `json:"PlaylistItemData,omitempty"`
	VideoInfoData *VideoInfoDataObject `json:"VideoInfoData,omitempty"`
	BookData *BookDataObject `json:"BookData,omitempty"`
	TimestampsData *TimestampsDataObject `json:"TimestampsData,omitempty"`
	GeniusInfoData *GeniusInfoDataObject `json:"GeniusInfoData,omitempty"`
}

func (o *DataObject) Read() error {
	o.Parsed = &DataObjectInner{}
	err := o.StandardObject.Parse(o.Parsed)
	if err != nil {
		return err
	}
	o.Raw = o.StandardObject.Data[binary.Size(o.Parsed):]
	buf := o.StandardObject.Buffer()
	data := make([]byte, binary.Size(o.Parsed))
	_, err = io.ReadFull(buf, data)
	if err != nil {
		return errors.WithStack(err)
	}
	o.Key = o.Parsed.Subtype.String()
	kind := o.Parsed.Subtype.Kind()
	switch kind {
	case BomaTypeNumeric:
		numData := &NumericDataObject{}
		err = binary.Read(buf, binary.LittleEndian, numData)
		if err != nil {
			return errors.WithStack(err)
		}
		o.NumericData = numData
	case BomaTypeWideChar:
		strData := &WideCharDataObject{}
		err = strData.Read(buf)
		if err != nil {
			return errors.WithStack(err)
		}
		o.WideCharData = strData
	case BomaTypeTimestamps:
		timestamps := &TimestampsDataObject{}
		err = timestamps.Read(buf)
		if err != nil {
			return errors.WithStack(err)
		}
		o.TimestampsData = timestamps
	case BomaTypeGeniusInfo:
		info := &GeniusInfoDataObject{}
		err = info.Read(buf)
		if err != nil {
			return errors.WithStack(err)
		}
		o.GeniusInfoData = info
	case BomaTypeShortHeader:
		hData := &ShortHeaderDataObject{}
		err = hData.Read(buf)
		if err != nil {
			return errors.WithStack(err)
		}
		o.ShortHeaderData = hData
	case BomaTypeLongHeader:
		hData := &LongHeaderDataObject{}
		err = hData.Read(buf)
		if err != nil {
			return errors.WithStack(err)
		}
		o.LongHeaderData = hData
	case BomaTypePlaylistItem:
		item := &PlaylistItemDataObject{}
		err = item.Read(buf)
		if err != nil {
			return errors.WithStack(err)
		}
		o.PlaylistItemData = item
	case BomaTypeVideoInfo:
		info := &VideoInfoDataObject{}
		err = info.Read(buf)
		if err != nil {
			return errors.WithStack(err)
		}
		o.VideoInfoData = info
	case BomaTypeBook:
		book := &BookDataObject{}
		err = book.Read(buf)
		if err != nil {
			return errors.WithStack(err)
		}
		o.BookData = book
	default:
		n := (o.StandardObject.Size - binary.Size(o.Parsed)) / 4
		nums := make([]uint32, n)
		for i := range nums {
			var num uint32
			binary.Read(buf, binary.LittleEndian, &num)
			nums[i] = num
		}
		o.Nums = nums
	}
	return nil
}

type DataObjectInner struct {
	Type ObjectSignature
	Unknown uint32
	Size uint32
	Subtype BomaSubType
}

type Unhandled struct {
	*StandardObject
}

func (o *Unhandled) Read() error {
	return nil
}
