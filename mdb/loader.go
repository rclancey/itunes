package mdb

import (
	"encoding/base64"
	xbin "encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/rclancey/itunes/binary"
	"github.com/rclancey/itunes/loader"
	"github.com/rclancey/itunes/persistentId"
)

type Loader struct {
	*loader.BaseLoader
	offset int
	albumRatings map[pid.PersistentID]uint8
}

func NewLoader() *Loader {
	return &Loader{loader.NewBaseLoader(), 0, map[pid.PersistentID]uint8{}}
}

func (l *Loader) LoadFile(fn string) {
	f, err := os.Open(fn)
	if err != nil {
		l.Shutdown(errors.Wrap(err, "can't open library file " +fn))
		return
	}
	l.Load(f)
}

func (l *Loader) Decrypt(f io.ReadCloser) (io.ReadCloser, error) {
	_, obj, err := ReadObject(f, 0)
	if err != nil {
		return nil, err
	}
	env, isa := obj.(*Envelope)
	if !isa {
		return nil, errors.WithStack(ErrInvalidHeader)
	}
	cryptSize := int64(env.MaxCryptSize)
	payload, err := binary.NewPayloadReader(f, int(cryptSize))
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (l *Loader) Parse(f io.Reader) {
	for {
		n, obj, err := ReadObject(f, l.offset)
		l.offset += n
		if err != nil {
			l.Shutdown(err)
			return
		}
		ch := l.GetChan()
		if ch != nil {
			ch <- obj
		} else {
			break
		}
	}
}

func (l *Loader) Load(f io.ReadCloser) {
	defer f.Close()
	payload, err := l.Decrypt(f)
	if err != nil {
		l.Shutdown(err)
		return
	}
	defer payload.Close()
	ch := l.GetChan()
	for {
		obj, err := l.getNext(payload)
		/*
		switch {
		case <-l.QuitCh:
			l.Shutdown(errors.WithStack(loader.AbortError))
			return
		default:
		*/
			if err != nil {
				if errors.Is(err, io.EOF) {
					l.Shutdown(nil)
					return
				}
				l.Shutdown(err)
				return
			}
			if obj != nil {
				ch <- obj
			}
		//}
	}
}

func (l *Loader) getNext(payload io.Reader) (interface{}, error) {
	n, obj, err := ReadObject(payload, l.offset)
	l.offset += n
	if err != nil {
		return nil, err
	}
	switch xobj := obj.(type) {
	case *Envelope:
		lib := &loader.Library{
			MajorVersion: loader.Intp(int(xobj.MajorVersion)),
			MinorVersion: loader.Intp(int(xobj.MinorVersion)),
			ApplicationVersion: loader.Stringp(xobj.ApplicationVersion.String()),
			Date: nil,
			Features: nil,
			ShowContentRatings: nil,
			PersistentID: xobj.PersistentID.Pointer(),
			MusicFolder: nil,
			Tracks: loader.Intp(int(xobj.ItemCount)),
			Playlists: loader.Intp(int(xobj.PlaylistCount)),
		}
		return lib, nil
	case *LibraryMaster:
		lib := &loader.Library{}
		for i := 0; i < int(xobj.DataObjectCount); i += 1 {
			n, child, err := ReadObject(payload, l.offset)
			l.offset += n
			if err != nil {
				return lib, err
			}
			dobj, isa := child.(*DataObject)
			if !isa {
				return lib, errors.Wrap(ErrUnexpectedObject, fmt.Sprintf("expected *DataObject, got %T", child))
			}
			if dobj.Key == "MediaFolder" {
				lib.MusicFolder = loader.Stringp(dobj.WideCharData.StrData)
			}
		}
		return lib, nil
	case *Album:
		l.albumRatings[xobj.PersistentID] = xobj.AlbumRating
	case *Track:
		return l.getTrack(xobj, payload)
	case *Playlist:
		return l.getPlaylist(xobj, payload)
	}
	return nil, nil
}

func (l *Loader) getTrack(t *Track, payload io.Reader) (*loader.Track, error) {
	track := &loader.Track{
		PersistentID: t.PersistentID.Pointer(),
		Rating: loader.Uint8p(uint8(t.Stars)),
	}
	if t.Love {
		track.Loved = loader.Boolp(true)
	}
	if t.MovementCount != 0 {
		track.MovementCount = loader.Intp(int(t.MovementCount))
	}
	if t.MovementNumber != 0 {
		track.MovementNumber = loader.Intp(int(t.MovementNumber))
	}
	if t.TrackCount != 0 {
		track.TrackCount = loader.Uint8p(uint8(t.TrackCount))
	}
	if t.TrackNumber != 0 {
		track.TrackNumber = loader.Uint8p(uint8(t.TrackNumber))
	}
	if t.DiscCount != 0 {
		track.DiscCount = loader.Uint8p(uint8(t.DiscCount))
	}
	if t.DiscNumber != 0 {
		track.DiscNumber = loader.Uint8p(uint8(t.DiscNumber))
	}
	if t.Year != 0 {
		track.Year = loader.Intp(int(t.Year))
	}
	if t.Disabled {
		track.Disabled = loader.Boolp(true)
	}
	for i := 0; i < int(t.DataObjectCount); i += 1 {
		n, child, err := ReadObject(payload, l.offset)
		l.offset += n
		if err != nil {
			return track, err
		}
		dobj, isa := child.(*DataObject)
		if !isa {
			return track, errors.Wrap(ErrUnexpectedObject, fmt.Sprintf("expected *DataObject, got %T", child))
		}
		switch dobj.Key {
		case "Title":
			track.Name = loader.Stringp(dobj.WideCharData.StrData)
		case "Artist":
			track.Artist = loader.Stringp(dobj.WideCharData.StrData)
		case "AlbumArtist":
			track.AlbumArtist = loader.Stringp(dobj.WideCharData.StrData)
		case "Composer":
			track.Composer = loader.Stringp(dobj.WideCharData.StrData)
		case "Album":
			track.Album = loader.Stringp(dobj.WideCharData.StrData)
		case "Genre":
			track.Genre = loader.Stringp(dobj.WideCharData.StrData)
		case "Grouping":
			track.Grouping = loader.Stringp(dobj.WideCharData.StrData)
		case "Kind":
			track.Kind = loader.Stringp(dobj.WideCharData.StrData)
		case "CopyrightInfo":
			// noop
		case "SortTitle":
			track.SortName = loader.Stringp(dobj.WideCharData.StrData)
		case "SortAlbum":
			track.SortAlbum = loader.Stringp(dobj.WideCharData.StrData)
		case "SortArtist":
			track.SortArtist = loader.Stringp(dobj.WideCharData.StrData)
		case "SortAlbumArtist":
			track.SortAlbumArtist = loader.Stringp(dobj.WideCharData.StrData)
		case "SortComposer":
			track.SortComposer = loader.Stringp(dobj.WideCharData.StrData)
		case "Work":
			track.Work = loader.Stringp(dobj.WideCharData.StrData)
		case "Numeric":
			v := dobj.NumericData
			track.FileType = loader.Intp(int(v.FileType))
			track.FileFolderCount = loader.Intp(int(v.FileFolderCount))
			track.LibraryFolderCount = loader.Intp(int(v.LibraryFolderCount))
			track.BitRate = loader.Uintp(uint(v.BitRate))
			track.SampleRate = loader.Uintp(uint(v.SampleRate))
			track.DateAdded = loader.Timep(v.DateAdded.Time())
			if v.DateModified != 0 {
				track.DateModified = loader.Timep(v.DateModified.Time())
			}
			if v.DatePurchased != 0 {
				track.PurchaseDate = loader.Timep(v.DatePurchased.Time())
			}
			if v.ReleaseDate != 0 {
				track.ReleaseDate = loader.Timep(v.ReleaseDate.Time())
			}
			track.TotalTime = loader.Uintp(uint(v.Duration))
			track.Size = loader.Uint64p(uint64(v.FileSize))
		case "CopyrightHolder":
			// noop
		case "PurchaserEmail":
			track.Purchased = loader.Boolp(true)
		case "PurchaserName":
			// noop
		case "Location":
			track.Location = loader.Stringp(dobj.WideCharData.StrData)
			if strings.HasPrefix(dobj.WideCharData.StrData, "file://") {
				track.TrackType = loader.Stringp("File")
			} else {
				track.TrackType = loader.Stringp("URL")
			}
		case "Timestamps":
			v := dobj.TimestampsData
			if v.PlayDate != 0 {
				track.PlayDate = loader.Timep(v.PlayDate.Time())
				track.PlayDateGarbage = loader.Intp(int(v.PlayDate))
			}
			if v.PlayCount != 0 {
				track.PlayCount = loader.Uintp(uint(v.PlayCount))
			}
			if v.SkipDate != 0 {
				track.SkipDate = loader.Timep(v.SkipDate.Time())
			}
			if v.SkipCount != 0 {
				track.SkipCount = loader.Uintp(uint(v.SkipCount))
			}
		}
	}
	albumRating := l.albumRatings[t.AlbumID]
	if albumRating != 0 {
		track.AlbumRating = loader.Uint8p(albumRating)
		track.AlbumRatingComputed = loader.Boolp(true)
	}
	return track, nil
}

func (l *Loader) getPlaylist(p *Playlist, payload io.Reader) (*loader.Playlist, error) {
	pl := &loader.Playlist{
		PersistentID: p.PersistentID.Pointer(),
		DateAdded: loader.Timep(p.DateAdded),
		DateModified: loader.Timep(p.DateModified),
		TrackIDs: []pid.PersistentID{},
		AllItems: loader.Boolp(true),
		Visible: loader.Boolp(true),
	}
	if p.ParentPersistentID != 0 {
		pl.ParentPersistentID = p.ParentPersistentID.Pointer()
	}
	if p.Folder {
		pl.Folder = loader.Boolp(true)
	}
	if p.DistinguishedKind != 0 {
		pl.DistinguishedKind = loader.Intp(int(p.DistinguishedKind))
	}
	for i := 0; i < int(p.DataObjectCount); i += 1 {
		n, child, err := ReadObject(payload, l.offset)
		l.offset += n
		if err != nil {
			return pl, err
		}
		dobj, isa := child.(*DataObject)
		if !isa {
			return pl, errors.Wrap(ErrUnexpectedObject, fmt.Sprintf("expected *DataObject, got %T", child))
		}
		switch dobj.Key {
		case "PlaylistName":
			pl.Name = loader.Stringp(dobj.WideCharData.StrData)
		case "SmartCriteria":
			pl.SmartCriteria = dobj.Raw[4:]
		case "SmartInfo":
			pl.SmartInfo = dobj.Raw[4:]
		case "GeniusInfo":
			pl.GeniusTrackID = dobj.GeniusInfoData.GeniusTrackID.Pointer()
		case "PlaylistItem":
			pl.TrackIDs = append(pl.TrackIDs, dobj.PlaylistItemData.TrackID)
		}
	}
	if pl.Folder != nil && *pl.Folder {
		pl.SmartCriteria = nil
		pl.SmartInfo = nil
		pl.TrackIDs = []pid.PersistentID{}
	}
	if pl.SmartCriteria != nil && pl.SmartInfo != nil {
		info := base64.StdEncoding.EncodeToString(pl.SmartInfo)
		crit := base64.StdEncoding.EncodeToString(pl.SmartCriteria)
		var err error
		pl.Smart, err = loader.ParseSmartPlaylist([]byte(info), []byte(crit), xbin.BigEndian)
		if err != nil {
			log.Printf("%s %+v", *pl.Name, err)
		}
		if pl.GeniusTrackID == nil && pl.Smart != nil && len(pl.Smart.Criteria.Rules) == 0 {
			pl.GeniusTrackID = pl.TrackIDs[0].Pointer()
			pl.Smart = nil
			pl.SmartCriteria = nil
			pl.SmartInfo = nil
		}
	}
	return pl, nil
}
