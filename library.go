package itunes

import (
	"math/rand"
	"path"
	"sort"
	"strings"

	"github.com/rclancey/itunes/loader"
	"github.com/rclancey/itunes/persistentId"
)

type Library struct {
	FileName string
	MajorVersion int
	MinorVersion int
	ApplicationVersion string
	Date Time
	Features int
	ShowContentRatings bool
	PersistentID pid.PersistentID
	MusicFolder string
	Tracks []*Track
	Playlists map[pid.PersistentID]*Playlist
	PlaylistTree []*Playlist
}

func NewLibrary() *Library {
	lib := &Library{}
	lib.Tracks = make([]*Track, 0)
	lib.Playlists = map[pid.PersistentID]*Playlist{}
	lib.PlaylistTree = []*Playlist{}
	return lib
}

func (lib *Library) Load(fn string) error {
	l := NewLoader(fn)
	go l.LoadFile(fn)
	for {
		ch := l.GetChan()
		update, ok := <-ch
		if !ok {
			return nil
		}
		switch tupdate := update.(type) {
		case *loader.Library:
			lib.FileName = tupdate.GetFileName()
			lib.MajorVersion = tupdate.GetMajorVersion()
			lib.MinorVersion = tupdate.GetMinorVersion()
			lib.ApplicationVersion = tupdate.GetApplicationVersion()
			lib.Date = Time{tupdate.GetDate()}
			lib.Features = tupdate.GetFeatures()
			lib.ShowContentRatings = tupdate.GetShowContentRatings()
			lib.PersistentID = pid.PersistentID(tupdate.GetPersistentID())
			lib.MusicFolder = tupdate.GetMusicFolder()
		case *loader.Track:
			if tupdate.GetMusicVideo() {
				continue
			}
			if tupdate.GetPodcast() {
				continue
			}
			if tupdate.GetMovie() {
				continue
			}
			if tupdate.GetTVShow() {
				continue
			}
			if tupdate.GetHasVideo() {
				continue
			}
			if path.Ext(tupdate.GetLocation()) == ".m4b" {
				// audiobook
				continue
			}
			tr := &Track{
				PersistentID:       pid.PersistentID(tupdate.GetPersistentID()),
				Album:              tupdate.GetAlbum(),
				AlbumArtist:        tupdate.GetAlbumArtist(),
				AlbumRating:        tupdate.GetAlbumRating(),
				Artist:             tupdate.GetArtist(),
				Comments:           tupdate.GetComments(),
				Compilation:        tupdate.GetCompilation(),
				Composer:           tupdate.GetComposer(),
				DateAdded:          &Time{tupdate.GetDateAdded()},
				DateModified:       &Time{tupdate.GetDateModified()},
				DiscCount:          tupdate.GetDiscCount(),
				DiscNumber:         tupdate.GetDiscNumber(),
				Genre:              tupdate.GetGenre(),
				Grouping:           tupdate.GetGrouping(),
				Kind:               tupdate.GetKind(),
				Location:           tupdate.GetLocation(),
				Loved:              tupdate.Loved,
				Name:               tupdate.GetName(),
				PartOfGaplessAlbum: tupdate.GetPartOfGaplessAlbum(),
				PlayCount:          tupdate.GetPlayCount(),
				Purchased:          tupdate.GetPurchased(),
				Rating:             tupdate.GetRating(),
				Size:               tupdate.GetSize(),
				SkipCount:          tupdate.GetSkipCount(),
				SortAlbum:          tupdate.GetSortAlbum(),
				SortAlbumArtist:    tupdate.GetSortAlbumArtist(),
				SortArtist:         tupdate.GetSortArtist(),
				SortComposer:       tupdate.GetSortComposer(),
				SortName:           tupdate.GetSortName(),
				TotalTime:          tupdate.GetTotalTime(),
				TrackCount:         tupdate.GetTrackCount(),
				TrackNumber:        tupdate.GetTrackNumber(),
				Unplayed:           tupdate.GetUnplayed(),
				VolumeAdjustment:   tupdate.GetVolumeAdjustment(),
				Work:               tupdate.GetWork(),
			}
			if tupdate.PlayDate != nil {
				tr.PlayDate = &Time{*tupdate.PlayDate}
			}
			if tupdate.PurchaseDate != nil {
				tr.PurchaseDate = &Time{*tupdate.PurchaseDate}
			}
			if tupdate.ReleaseDate != nil {
				tr.ReleaseDate = &Time{*tupdate.ReleaseDate}
			}
			if tupdate.SkipDate != nil {
				tr.SkipDate = &Time{*tupdate.SkipDate}
			}
			lib.AddTrack(tr)
		case *loader.Playlist:
			if tupdate.GetMaster() {
				continue
			}
			if tupdate.GetMusic() {
				continue
			}
			if tupdate.GetMovies() {
				continue
			}
			if tupdate.GetPodcasts() {
				continue
			}
			if tupdate.GetPurchasedMusic() {
				continue
			}
			if tupdate.GetAudiobooks() {
				continue
			}
			if !tupdate.GetVisible() {
				continue
			}
			pl := &Playlist{
				PersistentID: pid.PersistentID(tupdate.GetPersistentID()),
				Folder: tupdate.GetFolder(),
				Name: tupdate.GetName(),
			}
			if tupdate.ParentPersistentID != nil {
				pid := pid.PersistentID(*tupdate.ParentPersistentID)
				pl.ParentPersistentID = &pid
			}
			if tupdate.GeniusTrackID != nil {
				pid := pid.PersistentID(*tupdate.GeniusTrackID)
				pl.GeniusTrackID = &pid
			}
			if tupdate.IsSmart() {
				pl.Smart, _ = ParseSmartPlaylist(tupdate.SmartInfo, tupdate.SmartCriteria)
			}
			if !pl.Folder && pl.Smart == nil {
				pl.TrackIDs = make([]pid.PersistentID, len(tupdate.TrackIDs))
				for i, id := range tupdate.TrackIDs {
					pl.TrackIDs[i] = pid.PersistentID(id)
				}
			}
			if pl.Folder {
				pl.Children = []*Playlist{}
			}
			lib.Playlists[pl.PersistentID] = pl
		case error:
			return tupdate
		}
	}
	return nil
}

func (lib *Library) FindPlaylists(name string) []*Playlist {
	playlists := make([]*Playlist, 0)
	for _, p := range lib.Playlists {
		if p.Name == name {
			playlists = append(playlists, p)
		}
	}
	return playlists
}

func (lib *Library) GetPlaylistByPath(path string) *Playlist {
	parts := strings.Split(path, "/")
	for _, p := range lib.PlaylistTree {
		if p.Name == path {
			return p
		}
		if p.Name == parts[0] {
			m := p.GetByPath(strings.Join(parts[1:], "/"))
			if m != nil {
				return m
			}
		}
	}
	return nil
}

func (lib *Library) CreatePlaylist(name string, parentId *pid.PersistentID) *Playlist {
	p := &Playlist{
		PersistentID: pid.PersistentID(rand.Uint64()),
		ParentPersistentID: parentId,
		Name: name,
		TrackIDs: []pid.PersistentID{},
	}
	lib.Playlists[p.PersistentID] = p
	p.Nest(lib)
	return nil
}

func (lib *Library) TrackList() *TrackList {
	tl := TrackList(lib.Tracks)
	return &tl
}

type sts []*Track
func (s sts) Len() int { return len(s) }
func (s sts) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sts) Less(i, j int) bool { return s[i].PersistentID < s[j].PersistentID }

func (lib *Library) AddTrack(tr *Track) {
	id := tr.PersistentID
	f := func(i int) bool {
		return lib.Tracks[i].PersistentID >= id
	}
	idx := sort.Search(len(lib.Tracks), f)
	var tracks []*Track
	//alloced := false
	if len(lib.Tracks) == cap(lib.Tracks) {
		//alloced = true
		tracks = make([]*Track, len(lib.Tracks) + 1, len(lib.Tracks) + 1000)
		for i, x := range lib.Tracks[:idx] {
			tracks[i] = x
		}
		tracks[idx] = tr
		for i, x := range lib.Tracks[idx:] {
			tracks[i+idx+1] = x
		}
	} else {
		tracks = append(lib.Tracks, nil)
		for i := len(lib.Tracks) - 1; i >= idx; i-- {
			tracks[i+1] = tracks[i]
		}
		tracks[idx] = tr
	}
	lib.Tracks = tracks
	/*
	if alloced {
		runtime.GC()
	}
	*/
}

func (lib *Library) RemoveTrack(id pid.PersistentID) {
	f := func(i int) bool {
		return lib.Tracks[i].PersistentID >= id
	}
	idx := sort.Search(len(lib.Tracks), f)
	if idx < 0 || idx >= len(lib.Tracks) || lib.Tracks[idx].PersistentID != id {
		// track not in library, ignore
		return
	}
	tracks := append(lib.Tracks[:idx], lib.Tracks[idx+1:]...)
	lib.Tracks = tracks
	for _, pl := range lib.Playlists {
		if pl.Folder || pl.Smart != nil {
			continue
		}
		found := 0
		for _, tid := range pl.TrackIDs {
			if tid == id {
				found += 1
			}
		}
		if found > 0 {
			ids := make([]pid.PersistentID, len(pl.TrackIDs) - found)
			i := 0
			for _, tid := range pl.TrackIDs {
				if tid != id {
					ids[i] = tid
					i += 1
				}
			}
			pl.TrackIDs = ids
		}
	}
}

func (lib *Library) GetTrack(id pid.PersistentID) *Track {
	f := func(i int) bool {
		return lib.Tracks[i].PersistentID >= id
	}
	idx := sort.Search(len(lib.Tracks), f)
	if idx < 0 || idx >= len(lib.Tracks) {
		return nil
	}
	tr := lib.Tracks[idx]
	if tr.PersistentID == id {
		return tr
	}
	return nil
}

type spls []*Playlist
func (s spls) Len() int { return len(s) }
func (s spls) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s spls) Less(i, j int) bool {
	ap := s[i].Priority()
	bp := s[j].Priority()
	if ap < bp {
		return true
	}
	if ap > bp {
		return false
	}
	return s[i].Name < s[j].Name
}

func (l *Library) RenestPlaylists() {
	l.PlaylistTree = []*Playlist{}
	for _, pl := range l.Playlists {
		if pl.Folder {
			pl.Children = []*Playlist{}
		} else {
			pl.Children = nil
		}
	}
	for _, pl := range l.Playlists {
		if pl.ParentPersistentID == nil {
			l.PlaylistTree = append(l.PlaylistTree, pl)
		} else {
			parent, ok := l.Playlists[*pl.ParentPersistentID]
			if !ok || !parent.Folder {
				l.PlaylistTree = append(l.PlaylistTree, pl)
			} else {
				parent.Children = append(parent.Children, pl)
			}
		}
	}
	for _, pl := range l.Playlists {
		if pl.Folder {
			sort.Sort(spls(pl.Children))
		}
	}
	sort.Sort(spls(l.PlaylistTree))
}

func (l *Library) MovePlaylist(p *Playlist, parentId *pid.PersistentID) error {
	if p.ParentPersistentID == nil && parentId == nil {
		return nil
	}
	if p.ParentPersistentID != nil && parentId != nil && *p.ParentPersistentID == *parentId {
		return nil
	}
	p.Unnest(l)
	p.ParentPersistentID = parentId
	p.Nest(l)
	return nil
}
