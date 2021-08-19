package mdb

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	//"log"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/pkg/errors"

	"github.com/rclancey/itunes/persistentId"
)

type NumericDataObject struct {
	Unknown1 [16]uint32
	SampleRate float32
	Unknown2 uint32
	FileType uint32
	FileFolderCount int16
	LibraryFolderCount int16
	Unknown3 [3]uint32
	BitRate uint32
	DateAdded Time
	Unknown4 [8]uint32
	DateModified Time
	Normalization uint32
	DatePurchased Time
	ReleaseDate Time
	Unknown5 [3]uint32
	Duration uint32
	Unknown6 [34]uint32
	FileSize uint32
}

type TimestampsDataObject struct {
	Unknown1 uint32
	Unknown2 uint32
	Unknown3 uint32
	PlayDate Time
	PlayCount uint32
	Unknown4 uint32
	Unknown5 uint32
	Unknown6 uint32
	SkipDate Time
	SkipCount uint32
}

func (o *TimestampsDataObject) Read(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, o)
}

type GeniusInfoDataObject struct {
	Unknown1 uint32
	GeniusTrackID pid.PersistentID
}

func (o *GeniusInfoDataObject) Read(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, o)
}

type WideCharDataObject struct {
	Unknown1 uint32
	CharType uint32
	StringByteLength uint32
	Unknown2 [2]uint32
	Raw []byte `json:",omitemtpy"`
	StrData string `json:",omitemtpy"`
}

func utf16ToUtf8(data []byte) (string, error) {
	n := len(data)
	buf := &bytes.Buffer{}
	u16buf := make([]uint16, 1)
	u8buf := make([]byte, 4)
	for i := 0; i < n; i += 2 {
		u16buf[0] = uint16(data[i]) + (uint16(data[i+1]) << 8)
		r := utf16.Decode(u16buf)
		m := utf8.EncodeRune(u8buf, r[0])
		buf.Write(u8buf[:m])
	}
	return buf.String(), nil
}

func (o *WideCharDataObject) Read(r io.Reader) error {
	var unk1, unk2 [2]uint32
	var bytelen uint32
	err := binary.Read(r, binary.LittleEndian, &unk1)
	if err != nil {
		return errors.WithStack(err)
	}
	o.Unknown1 = unk1[0]
	o.CharType = unk1[1]
	err = binary.Read(r, binary.LittleEndian, &bytelen)
	if err != nil {
		return errors.WithStack(err)
	}
	o.StringByteLength = bytelen
	err = binary.Read(r, binary.LittleEndian, &unk2)
	if err != nil {
		return errors.WithStack(err)
	}
	o.Unknown2 = unk2
	data := make([]byte, int(bytelen))
	_, err = io.ReadFull(r, data)
	if err != nil {
		return errors.WithStack(err)
	}
	if o.CharType == 1 {
		o.StrData, err = utf16ToUtf8(data)
	} else if o.CharType == 2 {
		o.StrData = string(data)
	} else {
		o.Raw = data
	}
	return errors.WithStack(err)
}

type ShortHeaderDataObject struct {
	Unknown1 uint32
	XML string
}

func (o *ShortHeaderDataObject) Read(r io.Reader) error {
	var unk uint32
	err := binary.Read(r, binary.LittleEndian, &unk)
	if err != nil {
		return errors.WithStack(err)
	}
	o.Unknown1 = unk
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.WithStack(err)
	}
	o.XML = string(data)
	return nil
}

type LongHeaderDataObject struct {
	//Unknown1 [2]uint32
	ByteLength uint32
	//Unknown2 [2]uint32
	XML string
}

func (o *LongHeaderDataObject) Read(r io.Reader) error {
	//var unk1, unk2 [2]uint32
	var bytelen uint32
	/*
	err := binary.Read(r, binary.LittleEndian, &unk1)
	if err != nil {
		return errors.WithStack(err)
	}
	o.Unknown1 = unk1
	*/
	err := binary.Read(r, binary.LittleEndian, &bytelen)
	if err != nil {
		return errors.WithStack(err)
	}
	o.ByteLength = bytelen
	/*
	err = binary.Read(r, binary.LittleEndian, &unk2)
	if err != nil {
		return errors.WithStack(err)
	}
	o.Unknown2 = unk2
	*/
	/*
	data := make([]byte, int(bytelen))
	n, err := io.ReadFull(r, data)
	if err != nil {
		log.Printf("ignoring %s; read %d, expected %d", err, n, bytelen)
		return nil
	}
	*/
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.WithStack(err)
	}
	o.XML = string(data)
	return nil
}

type BookDataObject struct {
	Unknown1 uint32
	Signature string
	Unknown2 [12]uint32
	Strings []string
}

func (o *BookDataObject) Read(r io.Reader) error {
	o.Strings = []string{}
	var unk uint32
	err := binary.Read(r, binary.LittleEndian, &unk)
	if err != nil {
		return errors.WithStack(err)
	}
	o.Unknown1 = unk
	sig := make([]byte, 4)
	_, err = io.ReadFull(r, sig)
	if err != nil {
		return errors.WithStack(err)
	}
	o.Signature = string(sig)
	var unk2 [12]uint32
	err = binary.Read(r, binary.LittleEndian, &unk2)
	if err != nil {
		return errors.WithStack(err)
	}
	o.Unknown2 = unk2
	var size, strFlag uint32
	for {
		size = strFlag
		err = binary.Read(r, binary.LittleEndian, &strFlag)
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			break
		}
		if strFlag == 0x101 || strFlag == 0x901 || strFlag == 0x201 {
			data := make([]byte, int(size))
			_, err = io.ReadFull(r, data)
			if err != nil {
				if errors.Is(err, io.ErrUnexpectedEOF) {
					break
				} else {
					return errors.WithStack(err)
				}
			}
			o.Strings = append(o.Strings, string(data))
			if size % 4 != 0 {
				padSize := 4 - (size % 4)
				padding := make([]byte, padSize)
				_, err = io.ReadFull(r, padding)
				if err != nil {
					if errors.Is(err, io.ErrUnexpectedEOF) {
						break
					} else {
						return errors.WithStack(err)
					}
				}
			}
		}
	}
	return nil
}

type PlaylistItemDataObject struct {
	Unknown1 uint32
	SubSectionStart uint32
	SectionLength uint32
	Unknown2 uint32
	IpfaID pid.PersistentID
	TrackID pid.PersistentID
	Unknown3 [4]uint32
	IpfaID2 pid.PersistentID
}

func (o *PlaylistItemDataObject) Read(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, o)
}

type VideoInfoDataObject struct {
	Unknown1 uint32
	Height uint32
	Width uint32
	Unknown [10]uint32
	FrameRate uint32
}

func (o *VideoInfoDataObject) Read(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, o)
}

