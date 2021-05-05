package loader

import (
	"time"
)

type Library struct {
	FileName *string
	MajorVersion *int
	MinorVersion *int
	ApplicationVersion *string
	Date *time.Time
	Features *int
	ShowContentRatings *bool
	PersistentID *uint64 `plist:"Library Persistent ID"`
	MusicFolder *string
	Tracks *int
	Playlists *int
}

func (lib *Library) GetFileName() string {
	if lib.FileName == nil {
		return ""
	}
	return *lib.FileName
}

func (lib *Library) GetMajorVersion() int {
	if lib.MajorVersion == nil {
		return 0
	}
	return *lib.MajorVersion
}

func (lib *Library) GetMinorVersion() int {
	if lib.MinorVersion == nil {
		return 0
	}
	return *lib.MinorVersion
}

func (lib *Library) GetApplicationVersion() string {
	if lib.ApplicationVersion == nil {
		return ""
	}
	return *lib.ApplicationVersion
}

func (lib *Library) GetDate() time.Time {
	if lib.Date == nil {
		return time.Time{}
	}
	return *lib.Date
}

func (lib *Library) GetFeatures() int {
	if lib.Features == nil {
		return 0
	}
	return *lib.Features
}

func (lib *Library) GetShowContentRatings() bool {
	if lib.ShowContentRatings == nil {
		return false
	}
	return *lib.ShowContentRatings
}

func (lib *Library) GetPersistentID() uint64 {
	if lib.PersistentID == nil {
		return 0
	}
	return *lib.PersistentID
}

func (lib *Library) GetMusicFolder() string {
	if lib.MusicFolder == nil {
		return ""
	}
	return *lib.MusicFolder
}

func (lib *Library) GetTracks() int {
	if lib.Tracks == nil {
		return 0
	}
	return *lib.Tracks
}

func (lib *Library) GetPlaylists() int {
	if lib.Playlists == nil {
		return 0
	}
	return *lib.Playlists
}

