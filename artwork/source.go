package artwork

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/rclancey/itunes/persistentId"
)

type ArtworkSource interface {
	GetJPEG(id pid.PersistentID) ([]byte, error)
	Close() error
}

func NewArtworkSource(homedir string, libid pid.PersistentID) (ArtworkSource, error) {
	var src ArtworkSource
	var err error
	src, err = NewArtworkDB(homedir, libid)
	if err == nil {
		log.Println("got artworkdb")
		return src, nil
	}
	log.Println("no artwork db:", err)
	return NewItunesSource(homedir, libid)
}

type ItunesSource struct {
	root string
	libid pid.PersistentID
}

func NewItunesSource(homedir string, libid pid.PersistentID) (*ItunesSource, error) {
	root := filepath.Join(
		homedir,
		"Music",
		"iTunes",
		"Album Artwork",
		"Cache",
		libid.String(),
	)
	_, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	return &ItunesSource{root: root, libid: libid}, nil
}

func (src *ItunesSource) GetJPEG(id pid.PersistentID) ([]byte, error) {
	idbytes := []byte(id.String())
	n := len(idbytes) - 1
	fn := filepath.Join(
		src.root,
		fmt.Sprintf("%02d", idbytes[n]),
		fmt.Sprintf("%02d", idbytes[n-1]),
		fmt.Sprintf("%02d", idbytes[n-2]),
		fmt.Sprintf("%s-%s.itc", src.libid, id),
	)
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	obj := NewITC(f)
	var biggest *Item
	maxSize := 0
	for {
		x, err := obj.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		item, isa := x.(*Item)
		if !isa {
			continue
		}
		size := item.Width * item.Height
		if biggest == nil || size > maxSize {
			biggest = item
			maxSize = size
		}
	}
	if biggest == nil {
		return nil, os.ErrNotExist
	}
	buf := &bytes.Buffer{}
	err = biggest.ExportJPEG(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (src *ItunesSource) Close() error {
	return nil
}
