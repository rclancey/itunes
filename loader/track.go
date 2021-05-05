package loader

import (
	"time"
)

type Track struct {
	Album                *string
	AlbumArtist          *string
	AlbumRating          *uint8
	AlbumRatingComputed  *bool
	Artist               *string
	ArtworkCount         *int
	BPM                  *uint16
	BitRate              *uint
	Clean                *bool
	Comments             *string
	Compilation          *bool
	Composer             *string
	ContentRating        *string
	DateAdded            *time.Time
	DateModified         *time.Time
	Disabled             *bool
	DiscCount            *uint8
	DiscNumber           *uint8
	Episode              *string
	EpisodeOrder         *int
	Explicit             *bool
	FileFolderCount      *int
	FileType             *int
	Genre                *string
	Grouping             *string
	HasVideo             *bool
	Kind                 *string
	LibraryFolderCount   *int
	Location             *string
	Loved                *bool
	Master               *bool
	MovementCount        *int
	MovementName         *string
	MovementNumber       *int
	Movie                *bool
	MusicVideo           *bool
	Name                 *string
	PartOfGaplessAlbum   *bool
	PersistentID         *uint64
	PlayCount            *uint
	PlayDateGarbage      *int `plist:"Play Date"`
	PlayDate             *time.Time `plist:"Play Date UTC"`
	Podcast              *bool
	Protected            *bool
	PurchaseDate         *time.Time
	Purchased            *bool
	Rating               *uint8
	RatingComputed       *bool
	ReleaseDate          *time.Time
	SampleRate           *uint
	Season               *int
	Series               *string
	Size                 *uint64
	SkipCount            *uint
	SkipDate             *time.Time
	SortAlbum            *string
	SortAlbumArtist      *string
	SortArtist           *string
	SortComposer         *string
	SortName             *string
	SortSeries           *string
	StopTime             *int
	TVShow               *bool
	TotalTime            *uint
	TrackCount           *uint8
	TrackID              *int
	TrackNumber          *uint8
	TrackType            *string
	Unplayed             *bool
	VolumeAdjustment     *uint8
	Work                 *string
	Year                 *int
}

func (tr *Track) Clone() *Track {
	xtr := *tr
	return &xtr
}

func (tr *Track) GetAlbum() string {
	if tr.Album == nil {
		return ""
	}
	return *tr.Album
}

func (tr *Track) GetAlbumArtist() string {
	if tr.AlbumArtist == nil {
		return ""
	}
	return *tr.AlbumArtist
}

func (tr *Track) GetAlbumRating() uint8 {
	if tr.AlbumRating == nil {
		return 0
	}
	return *tr.AlbumRating
}

func (tr *Track) GetAlbumRatingComputed() bool {
	if tr.AlbumRatingComputed == nil {
		return false
	}
	return *tr.AlbumRatingComputed
}

func (tr *Track) GetArtist() string {
	if tr.Artist == nil {
		return ""
	}
	return *tr.Artist
}

func (tr *Track) GetArtworkCount() int {
	if tr.ArtworkCount == nil {
		return 0
	}
	return *tr.ArtworkCount
}

func (tr *Track) GetBPM() uint16 {
	if tr.BPM == nil {
		return 0
	}
	return *tr.BPM
}

func (tr *Track) GetBitRate() uint {
	if tr.BitRate == nil {
		return 0
	}
	return *tr.BitRate
}

func (tr *Track) GetClean() bool {
	if tr.Clean == nil {
		return false
	}
	return *tr.Clean
}

func (tr *Track) GetComments() string {
	if tr.Comments == nil {
		return ""
	}
	return *tr.Comments
}

func (tr *Track) GetCompilation() bool {
	if tr.Compilation == nil {
		return false
	}
	return *tr.Compilation
}

func (tr *Track) GetComposer() string {
	if tr.Composer == nil {
		return ""
	}
	return *tr.Composer
}

func (tr *Track) GetContentRating() string {
	if tr.ContentRating == nil {
		return ""
	}
	return *tr.ContentRating
}

func (tr *Track) GetDateAdded() time.Time {
	if tr.DateAdded == nil {
		return time.Time{}
	}
	return *tr.DateAdded
}

func (tr *Track) GetDateModified() time.Time {
	if tr.DateModified == nil {
		return time.Time{}
	}
	return *tr.DateModified
}

func (tr *Track) GetDisabled() bool {
	if tr.Disabled == nil {
		return false
	}
	return *tr.Disabled
}

func (tr *Track) GetDiscCount() uint8 {
	if tr.DiscCount == nil {
		return 0
	}
	return *tr.DiscCount
}

func (tr *Track) GetDiscNumber() uint8 {
	if tr.DiscNumber == nil {
		return 0
	}
	return *tr.DiscNumber
}

func (tr *Track) GetEpisode() string {
	if tr.Episode == nil {
		return ""
	}
	return *tr.Episode
}

func (tr *Track) GetEpisodeOrder() int {
	if tr.EpisodeOrder == nil {
		return 0
	}
	return *tr.EpisodeOrder
}

func (tr *Track) GetExplicit() bool {
	if tr.Explicit == nil {
		return false
	}
	return *tr.Explicit
}

func (tr *Track) GetFileFolderCount() int {
	if tr.FileFolderCount == nil {
		return 0
	}
	return *tr.FileFolderCount
}

func (tr *Track) GetFileType() int {
	if tr.FileType == nil {
		return 0
	}
	return *tr.FileType
}

func (tr *Track) GetGenre() string {
	if tr.Genre == nil {
		return ""
	}
	return *tr.Genre
}

func (tr *Track) GetGrouping() string {
	if tr.Grouping == nil {
		return ""
	}
	return *tr.Grouping
}

func (tr *Track) GetHasVideo() bool {
	if tr.HasVideo == nil {
		return false
	}
	return *tr.HasVideo
}

func (tr *Track) GetKind() string {
	if tr.Kind == nil {
		return ""
	}
	return *tr.Kind
}

func (tr *Track) GetLibraryFolderCount() int {
	if tr.LibraryFolderCount == nil {
		return 0
	}
	return *tr.LibraryFolderCount
}

func (tr *Track) GetLocation() string {
	if tr.Location == nil {
		return ""
	}
	return *tr.Location
}

func (tr *Track) GetLoved() bool {
	if tr.Loved == nil {
		return false
	}
	return *tr.Loved
}

func (tr *Track) GetMaster() bool {
	if tr.Master == nil {
		return false
	}
	return *tr.Master
}

func (tr *Track) GetMovementCount() int {
	if tr.MovementCount == nil {
		return 0
	}
	return *tr.MovementCount
}

func (tr *Track) GetMovementName() string {
	if tr.MovementName == nil {
		return ""
	}
	return *tr.MovementName
}

func (tr *Track) GetMovementNumber() int {
	if tr.MovementNumber == nil {
		return 0
	}
	return *tr.MovementNumber
}

func (tr *Track) GetMovie() bool {
	if tr.Movie == nil {
		return false
	}
	return *tr.Movie
}

func (tr *Track) GetMusicVideo() bool {
	if tr.MusicVideo == nil {
		return false
	}
	return *tr.MusicVideo
}

func (tr *Track) GetName() string {
	if tr.Name == nil {
		return ""
	}
	return *tr.Name
}

func (tr *Track) GetPartOfGaplessAlbum() bool {
	if tr.PartOfGaplessAlbum == nil {
		return false
	}
	return *tr.PartOfGaplessAlbum
}

func (tr *Track) GetPersistentID() uint64 {
	if tr.PersistentID == nil {
		return 0
	}
	return *tr.PersistentID
}

func (tr *Track) GetPlayCount() uint {
	if tr.PlayCount == nil {
		return 0
	}
	return *tr.PlayCount
}

func (tr *Track) GetPlayDateGarbage() int {
	if tr.PlayDateGarbage == nil {
		return 0
	}
	return *tr.PlayDateGarbage
}

func (tr *Track) GetPlayDate() time.Time {
	if tr.PlayDate == nil {
		return time.Time{}
	}
	return *tr.PlayDate
}

func (tr *Track) GetPodcast() bool {
	if tr.Podcast == nil {
		return false
	}
	return *tr.Podcast
}

func (tr *Track) GetProtected() bool {
	if tr.Protected == nil {
		return false
	}
	return *tr.Protected
}

func (tr *Track) GetPurchaseDate() time.Time {
	if tr.PurchaseDate == nil {
		return time.Time{}
	}
	return *tr.PurchaseDate
}

func (tr *Track) GetPurchased() bool {
	if tr.Purchased == nil {
		return false
	}
	return *tr.Purchased
}

func (tr *Track) GetRating() uint8 {
	if tr.Rating == nil {
		return 0
	}
	return *tr.Rating
}

func (tr *Track) GetRatingComputed() bool {
	if tr.RatingComputed == nil {
		return false
	}
	return *tr.RatingComputed
}

func (tr *Track) GetReleaseDate() time.Time {
	if tr.ReleaseDate == nil {
		return time.Time{}
	}
	return *tr.ReleaseDate
}

func (tr *Track) GetSampleRate() uint {
	if tr.SampleRate == nil {
		return 0
	}
	return *tr.SampleRate
}

func (tr *Track) GetSeason() int {
	if tr.Season == nil {
		return 0
	}
	return *tr.Season
}

func (tr *Track) GetSeries() string {
	if tr.Series == nil {
		return ""
	}
	return *tr.Series
}

func (tr *Track) GetSize() uint64 {
	if tr.Size == nil {
		return 0
	}
	return *tr.Size
}

func (tr *Track) GetSkipCount() uint {
	if tr.SkipCount == nil {
		return 0
	}
	return *tr.SkipCount
}

func (tr *Track) GetSkipDate() time.Time {
	if tr.SkipDate == nil {
		return time.Time{}
	}
	return *tr.SkipDate
}

func (tr *Track) GetSortAlbum() string {
	if tr.SortAlbum == nil {
		return ""
	}
	return *tr.SortAlbum
}

func (tr *Track) GetSortAlbumArtist() string {
	if tr.SortAlbumArtist == nil {
		return ""
	}
	return *tr.SortAlbumArtist
}

func (tr *Track) GetSortArtist() string {
	if tr.SortArtist == nil {
		return ""
	}
	return *tr.SortArtist
}

func (tr *Track) GetSortComposer() string {
	if tr.SortComposer == nil {
		return ""
	}
	return *tr.SortComposer
}

func (tr *Track) GetSortName() string {
	if tr.SortName == nil {
		return ""
	}
	return *tr.SortName
}

func (tr *Track) GetSortSeries() string {
	if tr.SortSeries == nil {
		return ""
	}
	return *tr.SortSeries
}

func (tr *Track) GetStopTime() int {
	if tr.StopTime == nil {
		return 0
	}
	return *tr.StopTime
}

func (tr *Track) GetTVShow() bool {
	if tr.TVShow == nil {
		return false
	}
	return *tr.TVShow
}

func (tr *Track) GetTotalTime() uint {
	if tr.TotalTime == nil {
		return 0
	}
	return *tr.TotalTime
}

func (tr *Track) GetTrackCount() uint8 {
	if tr.TrackCount == nil {
		return 0
	}
	return *tr.TrackCount
}

func (tr *Track) GetTrackID() int {
	if tr.TrackID == nil {
		return 0
	}
	return *tr.TrackID
}

func (tr *Track) GetTrackNumber() uint8 {
	if tr.TrackNumber == nil {
		return 0
	}
	return *tr.TrackNumber
}

func (tr *Track) GetTrackType() string {
	if tr.TrackType == nil {
		return ""
	}
	return *tr.TrackType
}

func (tr *Track) GetUnplayed() bool {
	if tr.Unplayed == nil {
		return false
	}
	return *tr.Unplayed
}

func (tr *Track) GetVolumeAdjustment() uint8 {
	if tr.VolumeAdjustment == nil {
		return 0
	}
	return *tr.VolumeAdjustment
}

func (tr *Track) GetWork() string {
	if tr.Work == nil {
		return ""
	}
	return *tr.Work
}

func (tr *Track) GetYear() int {
	if tr.Year == nil {
		return 0
	}
	return *tr.Year
}

