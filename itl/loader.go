package itl

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rclancey/itunes/binary"
	"github.com/rclancey/itunes/loader"
)

type Loader struct {
	*loader.BaseLoader
	header *Database
	offset int
	trackIdMap map[int]uint64
}

var zeroTime = time.Unix(0, 0)

func NewLoader() *Loader {
	return &Loader{loader.NewBaseLoader(), nil, 0, nil}
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
	db, isa := obj.(*Database)
	if !isa {
		return nil, errors.WithStack(ErrInvalidHeader)
	}
	l.header = db
	/*
	major, err := db.ApplicationVersion.Major()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	*/
	var cryptSize int
	if db.MajorVersion == 1 && db.MinorVersion == 0 {
		cryptSize = -2
	} else if db.MajorVersion == 1 && db.MinorVersion == 1 {
		cryptSize = -1
	} else if db.MajorVersion == 2 {
		cryptSize = 102400
	} else {
		return nil, errors.Wrapf(ErrUnknownVersion, "%d.%d", db.MajorVersion, db.MinorVersion)
	}
	if db.MaxCryptSize != 0 {
		cryptSize = db.MaxCryptSize
	}
	payload, err := binary.NewPayloadReader(f, cryptSize)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (l *Loader) Parse(f io.Reader) {
	ch := l.GetChan()
	if ch != nil {
		ch <- l.header
	}
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
		l.Shutdown(errors.WithStack(err))
		return
	}
	l.trackIdMap = map[int]uint64{}
	for {
		obj, err := l.getNext(payload)
		ch := l.GetChan()
		if ch == nil {
			return
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				l.Shutdown(nil)
				return
			}
			l.Shutdown(err)
			return
		}
		if obj != nil {
			t, isa := obj.(*loader.Track)
			if isa && t.TrackID != nil && t.PersistentID != nil {
				l.trackIdMap[*t.TrackID] = *t.PersistentID
			}
			ch <- obj
		}
	}
}

func (l *Loader) getNext(payload io.Reader) (interface{}, error) {
	n, obj, err := ReadObject(payload, l.offset)
	l.offset += n
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, io.EOF
	}
	switch xobj := obj.(type) {
	case *Database:
		lib := &loader.Library{
			MajorVersion: loader.Intp(int(xobj.MajorVersion)),
			MinorVersion: loader.Intp(int(xobj.MinorVersion)),
			ApplicationVersion: loader.Stringp(xobj.ApplicationVersion.String()),
			Date: nil,
			Features: nil,
			ShowContentRatings: nil,
			PersistentID: loader.Uint64p(uint64(xobj.PersistentID)),
			MusicFolder: nil,
		}
		return lib, nil
	case *Track:
		return l.getTrack(xobj, payload)
	case *Playlist:
		return l.getPlaylist(xobj, payload)
	}
	return nil, nil
}

func (l *Loader) getTrack(t *Track, payload io.Reader) (*loader.Track, error) {
	track := &loader.Track{
		TrackID: loader.Intp(int(t.TrackID)),
		Size: loader.Uint64p(uint64(t.FileSize)),
		DateModified: loader.Timep(t.DateModified),
		TotalTime: loader.Uintp(uint(t.TotalTime)),
		BitRate: loader.Uintp(uint(t.BitRate)),
		SampleRate: loader.Uintp(uint(t.SampleRate)),
		DateAdded: loader.Timep(t.DateAdded),
		PersistentID: loader.Uint64p(uint64(t.PersistentID)),
	}
	if t.TrackNumber != 0 {
		track.TrackNumber = loader.Uint8p(uint8(t.TrackNumber))
	}
	if t.TrackCount != 0 {
		track.TrackCount = loader.Uint8p(uint8(t.TrackCount))
	}
	if t.Year != 0 {
		track.Year = loader.Intp(int(t.Year))
	}
	if t.VolumeAdjustment != 0 {
		track.VolumeAdjustment = loader.Uint8p(uint8(t.VolumeAdjustment))
	}
	if t.StartTime != 0 {
		track.StartTime = loader.Intp(int(t.StartTime))
	}
	if t.StopTime != 0 {
		track.StopTime = loader.Intp(int(t.StopTime))
	}
	if t.PlayCount != 0 {
		track.PlayCount = loader.Uintp(uint(t.PlayCount))
	}
	if t.Compilation {
		track.Compilation = loader.Boolp(true)
	}
	if !t.PlayDate.Equal(zeroTime){
		track.PlayDate = loader.Timep(t.PlayDate)
	}
	if t.DiscNumber != 0 {
		track.DiscNumber = loader.Uint8p(uint8(t.DiscNumber))
	}
	if t.DiscCount != 0 {
		track.DiscCount = loader.Uint8p(uint8(t.DiscCount))
	}
	if t.Rating != 0 {
		track.Rating = loader.Uint8p(uint8(t.Rating))
	}
	if t.BPM != 0 {
		track.BPM = loader.Uint16p(uint16(t.BPM))
	}
	if t.Disabled {
		track.Disabled = loader.Boolp(true)
	}
	if !t.PurchaseDate.Equal(zeroTime) {
		track.PurchaseDate = loader.Timep(t.PurchaseDate)
	}
	if !t.ReleaseDate.Equal(zeroTime) {
		track.ReleaseDate = loader.Timep(t.ReleaseDate)
	}
	if t.SkipCount != 0 {
		track.SkipCount = loader.Uintp(uint(t.SkipCount))
	}
	if !t.SkipDate.Equal(zeroTime) {
		track.SkipDate = loader.Timep(t.SkipDate)
	}

	/*
	if t.Love != 0 {
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
	if t.TrackYear != 0 {
		track.Year = loader.Intp(int(t.TrackYear))
	}
	if t.Unchecked != 0 {
		track.Disabled = loader.Boolp(true)
	}
	*/
	for i := 0; i < int(t.RecordCount); i += 1 {
		n, child, err := ReadObject(payload, l.offset)
		l.offset += n
		if err != nil {
			return track, err
		}
		dobj, isa := child.(*DataObject)
		if !isa {
			return track, errors.Wrap(ErrUnexpectedObject, fmt.Sprintf("expected *DataObject, got %T", child))
		}
		switch dobj.TypeID.String() {
		case "Title":
			track.Name = loader.Stringp(dobj.Str)
		case "Artist":
			track.Artist = loader.Stringp(dobj.Str)
		case "AlbumArtist":
			track.AlbumArtist = loader.Stringp(dobj.Str)
		case "Composer":
			track.Composer = loader.Stringp(dobj.Str)
		case "Album":
			track.Album = loader.Stringp(dobj.Str)
		case "Genre":
			track.Genre = loader.Stringp(dobj.Str)
		case "Grouping":
			track.Grouping = loader.Stringp(dobj.Str)
		case "Kind":
			track.Kind = loader.Stringp(dobj.Str)
		case "CopyrightInfo":
			// noop
		case "SortTitle":
			track.SortName = loader.Stringp(dobj.Str)
		case "SortAlbum":
			track.SortAlbum = loader.Stringp(dobj.Str)
		case "SortArtist":
			track.SortArtist = loader.Stringp(dobj.Str)
		case "SortAlbumArtist":
			track.SortAlbumArtist = loader.Stringp(dobj.Str)
		case "SortComposer":
			track.SortComposer = loader.Stringp(dobj.Str)
		case "Work":
			track.Work = loader.Stringp(dobj.Str)
		case "CopyrightHolder":
			// noop
		case "PurchaserEmail":
			track.Purchased = loader.Boolp(true)
		case "PurchaserName":
			// noop
		case "Location":
			track.Location = loader.Stringp(dobj.Str)
		}
	}
	return track, nil
}

func (l *Loader) getPlaylist(p *Playlist, payload io.Reader) (*loader.Playlist, error) {
	pl := &loader.Playlist{
		PersistentID: loader.Uint64p(uint64(p.PersistentID)),
		Name: loader.Stringp(""),
		DateAdded: loader.Timep(p.DateAdded),
		DateModified: loader.Timep(p.DateModified),
		TrackIDs: []uint64{},
		AllItems: loader.Boolp(true),
		Visible: loader.Boolp(true),
	}
	if p.ParentPersistentID != 0 {
		pl.ParentPersistentID = loader.Uint64p(uint64(p.ParentPersistentID))
	}
	if p.Folder {
		pl.Folder = loader.Boolp(true)
	}
	if p.DistinguishedKind != 0 {
		pl.DistinguishedKind = loader.Intp(int(p.DistinguishedKind))
	}
	for i := 0; i < int(p.RecordCount); i += 1 {
		n, child, err := ReadObject(payload, l.offset)
		l.offset += n
		if err != nil {
			return pl, err
		}
		dobj, isa := child.(*DataObject)
		if !isa {
			return pl, errors.Wrap(ErrUnexpectedObject, fmt.Sprintf("expected *DataObject, got %T", child))
		}
		switch dobj.TypeID.String() {
		case "Playlist Name":
			pl.Name = loader.Stringp(dobj.Str)
		case "Smart Criteria":
			pl.SmartCriteria = dobj.Data
		case "Smart Info":
			pl.SmartInfo = dobj.Data
		//case "GeniusInfo":
		//	pl.GeniusTrackID = loader.Uint64p(uint64(dobj.GeniusInfoData.GeniusTrackID))
		}
	}
	i := 0
	for i < int(p.TrackCount) {
		n, child, err := ReadObject(payload, l.offset)
		l.offset += n
		if err != nil {
			return pl, err
		}
		pt, isa := child.(*PlaylistItem)
		if !isa {
			continue
		}
		pid, ok := l.trackIdMap[int(pt.TrackID)]
		if ok {
			pl.TrackIDs = append(pl.TrackIDs, pid)
		}
		i += 1
	}
	return pl, nil
}
