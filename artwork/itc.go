package artwork

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	//"log"
)

const (
	MethodLocal = "locl"
	MethodDownload = "down"
	FormatPNG = "PNGf"
	FormatJPEG = "JPEG"
	FormatARGB = "ARGb"
	Itunes9 = 208
	ItunesOld = 216
)

type ITC struct {
	r io.Reader
}

func NewITC(r io.Reader) *ITC {
	return &ITC{r}
}

type FrameHeader struct {
	Size uint32
	Kind [4]byte
}

func (itc *ITC) Read() (interface{}, error) {
	header := &FrameHeader{}
	err := binary.Read(itc.r, binary.BigEndian, header)
	if err != nil {
		return nil, err
	}
	//log.Printf("read header: %#v", *header)
	switch string(header.Kind[:]) {
	case "itch":
		return NewITCH(itc, header)
	case "artw":
		return NewARTW(itc, header)
	case "item":
		return NewItem(itc, header)
	}
	return nil, errors.New("unknown section kind: " + string(header.Kind[:]))
}

type ITCH struct {
	Header *FrameHeader
	Unknown1 []byte
	Subframe interface{}
}

func NewITCH(itc *ITC, header *FrameHeader) (*ITCH, error) {
	data := make([]byte, 12)
	_, err := io.ReadFull(itc.r, data)
	if err != nil {
		return nil, err
	}
	subframe, err := itc.Read()
	if err != nil {
		return nil, err
	}
	return &ITCH{header, data, subframe}, nil
}

type ARTW struct {
	Header *FrameHeader
	Unknown1 []byte
}

func NewARTW(itc *ITC, header *FrameHeader) (*ARTW, error) {
	data := make([]byte, 256)
	_, err := io.ReadFull(itc.r, data)
	if err != nil {
		return nil, err
	}
	return &ARTW{header, data}, nil
}

type Item struct {
	Header *FrameHeader
	Offset uint32
	Preamble []byte
	LibraryID uint64
	TrackID uint64
	Method string
	Format string
	Width int
	Height int
	Data []byte
}

func NewItem(itc *ITC, header *FrameHeader) (*Item, error) {
	bytesRead := 0
	item := &Item{Header: header}
	var offset uint32
	err := binary.Read(itc.r, binary.BigEndian, &offset)
	if err != nil {
		return nil, err
	}
	bytesRead += 4
	item.Offset = offset
	var preamble []byte
	if offset == Itunes9 {
		preamble = make([]byte, 16)
	} else if offset == ItunesOld {
		preamble = make([]byte, 20)
	} else {
		preamble = make([]byte, 16)
	}
	//log.Printf("offset = %d, preamble = %d bytes", offset, len(preamble))
	_, err = io.ReadFull(itc.r, preamble)
	if err != nil {
		return nil, err
	}
	bytesRead += len(preamble)
	item.Preamble = preamble
	var info struct {
		LibraryID uint64
		TrackID uint64
		Method [4]byte
		Format [4]byte
	}
	err = binary.Read(itc.r, binary.BigEndian, &info)
	if err != nil {
		return nil, err
	}
	//log.Println("item info = %#v", info)
	bytesRead += binary.Size(&info)
	item.LibraryID = info.LibraryID
	item.TrackID = info.TrackID
	item.Method = string(info.Method[:])
	if info.Format[0] == 0 {
		switch info.Format[3] {
		case 0x0e:
			item.Format = FormatPNG
		case 0x0d:
			item.Format = FormatJPEG
		default:
			item.Format = ""
		}
	} else {
		item.Format = string(info.Format[:])
	}
	padding := make([]byte, 4)
	_, err = io.ReadFull(itc.r, padding)
	if err != nil {
		return nil, err
	}
	bytesRead += len(padding)
	var dims struct {
		Width uint32
		Height uint32
	}
	err = binary.Read(itc.r, binary.BigEndian, &dims)
	if err != nil {
		return nil, err
	}
	bytesRead += binary.Size(&dims)
	item.Width = int(dims.Width)
	item.Height = int(dims.Height)
	padding = make([]byte, int(item.Offset) - bytesRead - 8)
	_, err = io.ReadFull(itc.r, padding)
	if err != nil {
		return nil, err
	}
	dataSize := int(header.Size - item.Offset)
	data := make([]byte, dataSize)
	_, err = io.ReadFull(itc.r, data)
	if err != nil {
		return nil, err
	}
	item.Data = data
	return item, nil
}

func (item *Item) MakeImage() (image.Image, error) {
	switch item.Format {
	case FormatPNG:
		return item.ParsePNG()
	case FormatJPEG:
		return item.ParseJPEG()
	case FormatARGB:
		return item.ParseARGB()
	}
	return nil, errors.New("unknown image format " + item.Format)
}

func (item *Item) ParsePNG() (image.Image, error) {
	buf := bytes.NewBuffer(item.Data)
	return png.Decode(buf)
}

func (item *Item) ParseJPEG() (image.Image, error) {
	buf := bytes.NewBuffer(item.Data)
	return png.Decode(buf)
}

func (item *Item) ParseARGB() (image.Image, error) {
	img := image.NewRGBA(image.Rect(0, 0, item.Width, item.Height))
	n := len(item.Data)
	if item.Width * item.Height * 4 != n {
		return nil, errors.New("invalid ARGB image dimensions")
	}
	var x, y int
	px := item.Data
	for i := 0; i < n; i += 4 {
		img.SetRGBA(x, y, color.RGBA{px[i+1], px[i+2], px[i+3], px[i]})
		x += 1
		if x >= item.Width {
			x = 0
			y += 1
		}
	}
	return img, nil
}

func (item *Item) ExportJPEG(w io.Writer) error {
	if item.Format == FormatJPEG {
		_, err := w.Write(item.Data)
		return err
	}
	img, err := item.MakeImage()
	if err != nil {
		return err
	}
	return jpeg.Encode(w, img, &jpeg.Options{Quality: 75})
}

func (item *Item) ExportPNG(w io.Writer) error {
	if item.Format == FormatPNG {
		_, err := w.Write(item.Data)
		return err
	}
	img, err := item.MakeImage()
	if err != nil {
		return err
	}
	return png.Encode(w, img)
}
