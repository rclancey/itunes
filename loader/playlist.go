package loader

import (
	"time"

	"github.com/rclancey/itunes/persistentId"
)

type Playlist struct {
	AllItems             *bool
	Audiobooks           *bool
	DistinguishedKind    *int
	Folder               *bool
	GeniusTrackID        *pid.PersistentID
	Master               *bool
	Movies               *bool
	Music                *bool
	Name                 *string
	ParentPersistentID   *pid.PersistentID
	PlaylistID           *int
	PersistentID         *pid.PersistentID `plist:"Playlist Persistent ID"`
	Podcasts             *bool
	PurchasedMusic       *bool
	SmartCriteria        []byte
	SmartInfo            []byte
	TVShows              *bool
	Visible              *bool
	DateAdded            *time.Time
	DateModified         *time.Time
	Smart                *SmartPlaylist
	SortField            *string
	TrackIDs             []pid.PersistentID
}

func NewPlaylist() *Playlist {
	p := &Playlist{}
	p.TrackIDs = make([]pid.PersistentID, 0)
	return p
}

func (p *Playlist) IsSmart() bool {
	if p.GetFolder() {
		return false
	}
	if p.GeniusTrackID != nil {
		return false
	}
	if p.SmartInfo == nil || p.SmartCriteria == nil {
		return false
	}
	return len(p.SmartInfo) > 0 && len(p.SmartCriteria) > 0
}

func (pl *Playlist) GetAllItems() bool {
	if pl.AllItems == nil {
		return false
	}
	return *pl.AllItems
}

func (pl *Playlist) GetAudiobooks() bool {
	if pl.Audiobooks == nil {
		return false
	}
	return *pl.Audiobooks
}

func (pl *Playlist) GetDistinguishedKind() int {
	if pl.DistinguishedKind == nil {
		return 0
	}
	return *pl.DistinguishedKind
}

func (pl *Playlist) GetFolder() bool {
	if pl.Folder == nil {
		return false
	}
	return *pl.Folder
}

func (pl *Playlist) GetGeniusTrackID() pid.PersistentID {
	if pl.GeniusTrackID == nil {
		return 0
	}
	return *pl.GeniusTrackID
}

func (pl *Playlist) GetMaster() bool {
	if pl.Master == nil {
		return false
	}
	return *pl.Master
}

func (pl *Playlist) GetMovies() bool {
	if pl.Movies == nil {
		return false
	}
	return *pl.Movies
}

func (pl *Playlist) GetMusic() bool {
	if pl.Music == nil {
		return false
	}
	return *pl.Music
}

func (pl *Playlist) GetName() string {
	if pl.Name == nil {
		return ""
	}
	return *pl.Name
}

func (pl *Playlist) GetParentPersistentID() pid.PersistentID {
	if pl.ParentPersistentID == nil {
		return 0
	}
	return *pl.ParentPersistentID
}

func (pl *Playlist) GetPlaylistID() int {
	if pl.PlaylistID == nil {
		return 0
	}
	return *pl.PlaylistID
}

func (pl *Playlist) GetPersistentID() pid.PersistentID {
	if pl.PersistentID == nil {
		return 0
	}
	return *pl.PersistentID
}

func (pl *Playlist) GetPodcasts() bool {
	if pl.Podcasts == nil {
		return false
	}
	return *pl.Podcasts
}

func (pl *Playlist) GetPurchasedMusic() bool {
	if pl.PurchasedMusic == nil {
		return false
	}
	return *pl.PurchasedMusic
}

func (pl *Playlist) GetTVShows() bool {
	if pl.TVShows == nil {
		return false
	}
	return *pl.TVShows
}

func (pl *Playlist) GetVisible() bool {
	if pl.Visible == nil {
		return true
	}
	return *pl.Visible
}

