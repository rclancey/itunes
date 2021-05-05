package itunes

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/dhowden/tag"
)

type Track struct {
	PersistentID         PersistentID `json:"persistent_id"`
	Album                string       `json:"album,omitempty"`
	AlbumArtist          string       `json:"album_artist,omitempty"`
	AlbumRating          uint8        `json:"album_rating,omitempty"`
	Artist               string       `json:"artist,omitempty"`
	Comments             string       `json:"comments,omitempty"`
	Compilation          bool         `json:"compilation,omitempty"`
	Composer             string       `json:"composer,omitempty"`
	DateAdded            *Time        `json:"date_added,omitempty"`
	DateModified         *Time        `json:"date_modified,omitempty"`
	DiscCount            uint8        `json:"disc_count,omitempty"`
	DiscNumber           uint8        `json:"disc_number,omitempty"`
	Genre                string       `json:"genre,omitempty"`
	Grouping             string       `json:"grouping,omitempty"`
	Kind                 string       `json:"kind,omitempty"`
	Location             string       `json:"location"`
	Loved                *bool        `json:"loved"`
	Name                 string       `json:"name,omitempty"`
	PartOfGaplessAlbum   bool         `json:"part_of_gapless_album,omitempty"`
	PlayCount            uint         `json:"play_count,omitempty"`
	PlayDate             *Time        `json:"play_date,omitempty"`
	Purchased            bool         `json:"purchased,omitempty"`
	PurchaseDate         *Time        `json:"purchase_date,omitempty"`
	Rating               uint8        `json:"rating,omitempty"`
	ReleaseDate          *Time        `json:"release_date,omitempty"`
	Size                 uint64       `json:"size,omitempty"`
	SkipCount            uint         `json:"skip_count,omitempty"`
	SkipDate             *Time        `json:"skip_date,omitempty"`
	SortAlbum            string       `json:"sort_album,omitempty"`
	SortAlbumArtist      string       `json:"sort_album_artist,omitempty"`
	SortArtist           string       `json:"sort_artist,omitempty"`
	SortComposer         string       `json:"sort_composer,omitempty"`
	SortName             string       `json:"sort_name,omitempty"`
	TotalTime            uint         `json:"total_time,omitempty"`
	TrackCount           uint8        `json:"track_count,omitempty"`
	TrackNumber          uint8        `json:"track_number,omitempty"`
	Unplayed             bool         `json:"unplayed,omitempty"`
	VolumeAdjustment     uint8        `json:"volume_adjustment,omitempty"`
	Work                 string       `json:"work,omitempty"`
}

func (t *Track) String() string {
	s := ""
	delim := ""
	if t.AlbumArtist != "" {
		s += delim + t.AlbumArtist
		delim = " / "
	} else if t.Artist != "" {
		s += delim + t.Artist
		delim = " / "
	}
	if t.Album != "" {
		s += delim + t.Album
		delim = " / "
	}
	if t.AlbumArtist != "" && t.Artist != "" && t.AlbumArtist != t.Artist {
		s += delim + t.Artist
		delim = " / "
	}
	if t.Name != "" {
		s += delim + t.Name
	}
	return s
}

func (t *Track) MediaKind() MediaKind {
	return MediaKind_MUSIC
}

type stimes []time.Time
func (s stimes) Len() int { return len(s) }
func (s stimes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s stimes) Less(i, j int) bool { return s[i].Before(s[j]) }

func (t *Track) ModDate() time.Time {
	times := []time.Time{}
	if t.DateModified != nil {
		times = append(times, t.DateModified.Get())
	}
	if t.DateAdded != nil {
		times = append(times, t.DateAdded.Get())
	}
	if t.PlayDate != nil {
		times = append(times, t.PlayDate.Get())
	}
	if t.SkipDate != nil {
		times = append(times, t.SkipDate.Get())
	}
	if len(times) == 0 {
		return time.Date(2999, time.December, 31, 23, 59, 59, 999999999, time.UTC)
	}
	sort.Sort(stimes(times))
	return times[len(times) - 1]
}

func (t *Track) Path() string {
	if t.Location == "" {
		return ""
	}
	finder := GetGlobalFinder()
	if finder != nil {
		fn, err := finder.FindFile(t.Location)
		if err == nil {
			return fn
		}
	}
	return t.Location
}

func (t *Track) getTag() (tag.Metadata, error) {
	fn := t.Path()
	f, err := os.Open(fn)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		return nil, err
	}
	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (t *Track) GetPurchaseDate() (*Time, error) {
	if t.PurchaseDate != nil {
		return t.PurchaseDate, nil
	}
	if !t.Purchased {
		return nil, nil
	}
	m, err := t.getTag()
	if err != nil {
		return nil, err
	}
	d := m.Raw()
	v, ok := d["purd"]
	if !ok {
		return nil, nil
	}
	tm, err := time.Parse("2006-01-02 15:04:05", v.(string))
	if err != nil {
		return nil, err
	}
	t.PurchaseDate = &Time{tm}
	return t.PurchaseDate, nil
}

func (t *Track) GetName() (string, error) {
	if t.Name != "" {
		return t.Name, nil
	}
	m, err := t.getTag()
	if err != nil {
		return "", err
	}
	v := m.Title()
	if v != "" {
		t.Name = v
		return v, nil
	}
	fn := t.Path()
	_, name := filepath.Split(fn)
	ext := filepath.Ext(fn)
	name = strings.TrimSuffix(name, ext)
	name = strings.Replace(name, "_", " ", -1)
	reg := regexp.MustCompile(`^\d+(\.\d+)[ \.\-]\s*`)
	name = reg.ReplaceAllString(name, "")
	t.Name = name
	return name, nil
}

func (t *Track) GetAlbum() (string, error) {
	if t.Album != "" {
		return t.Album, nil
	}
	m, err := t.getTag()
	if err != nil {
		return "", err
	}
	v := m.Album()
	if v != "" {
		t.Album = v
		return v, nil
	}
	fn := t.Path()
	dir, _ := filepath.Split(fn)
	_, name := filepath.Split(dir)
	name = strings.Replace(name, "_", " ", -1)
	t.Album = name
	return name, nil
}

func (t *Track) GetArtist() (string, error) {
	if t.Artist != "" {
		return t.Artist, nil
	}
	m, err := t.getTag()
	if err != nil {
		return "", err
	}
	v := m.Artist()
	if v != "" {
		t.Artist = v
		return v, nil
	}
	fn := t.Path()
	dir, _ := filepath.Split(fn)
	dir, _ = filepath.Split(dir)
	_, name := filepath.Split(dir)
	name = strings.Replace(name, "_", " ", -1)
	t.Artist = name
	return name, nil
}

var kindExt = map[string]string{
	"Purchased AAC audio file": ".m4a",
	"Protected AAC audio file": ".m4p",
	"MPEG audio file": ".mp3",
	"WAV audio file": ".wav",
	"MPEG-4 video file": ".m4v",
	"Protected MPEG-4 video file": ".m4v",
	"Purchased MPEG-4 video file": ".m4v",
	"QuickTime movie file": ".mov",
}

func (t *Track) GetExt() string {
	if t.Location != "" {
		return path.Ext(t.Location)
	}
	if t.Kind != "" {
		ext, ok := kindExt[t.Kind]
		if ok {
			return ext
		}
	}
	return ""
}

func (t *Track) Update(orig, cur *Track) {
	mod := false
	if cur.Album != orig.Album {
		t.Album = cur.Album
		mod = true
	}
	if cur.AlbumArtist != orig.AlbumArtist {
		t.AlbumArtist = cur.AlbumArtist
		mod = true
	}
	if cur.Artist != orig.Artist {
		t.Artist = cur.Artist
		mod = true
	}
	if cur.Comments != orig.Comments {
		t.Comments = cur.Comments
		mod = true
	}
	if cur.Compilation != orig.Compilation {
		t.Compilation = cur.Compilation
		mod = true
	}
	if cur.Composer != orig.Composer {
		t.Composer = cur.Composer
		mod = true
	}
	if cur.DiscCount != orig.DiscCount {
		t.DiscCount = cur.DiscCount
		mod = true
	}
	if cur.DiscNumber != orig.DiscNumber {
		t.DiscNumber = cur.DiscNumber
		mod = true
	}
	if cur.Genre != orig.Genre {
		t.Genre = cur.Genre
		mod = true
	}
	if cur.Grouping != orig.Grouping {
		t.Grouping = cur.Grouping
		mod = true
	}
	if cur.Loved != orig.Loved {
		t.Loved = cur.Loved
		mod = true
	}
	if cur.Name != orig.Name {
		t.Name = cur.Name
		mod = true
	}
	if cur.PartOfGaplessAlbum != orig.PartOfGaplessAlbum {
		t.PartOfGaplessAlbum = cur.PartOfGaplessAlbum
		mod = true
	}
	if cur.Rating != orig.Rating {
		t.Rating = cur.Rating
		mod = true
	}
	if cur.ReleaseDate == nil {
		if orig.ReleaseDate != nil {
			t.ReleaseDate = nil
			mod = true
		}
	} else {
		if orig.ReleaseDate == nil {
			t.ReleaseDate = cur.ReleaseDate
			mod = true
		} else if !cur.ReleaseDate.Equal(orig.ReleaseDate.Get()) {
			t.ReleaseDate = cur.ReleaseDate
			mod = true
		}
	}
	if cur.SortAlbum != orig.SortAlbum {
		t.SortAlbum = cur.SortAlbum
		mod = true
	}
	if cur.SortAlbumArtist != orig.SortAlbumArtist {
		t.SortAlbumArtist = cur.SortAlbumArtist
		mod = true
	}
	if cur.SortArtist != orig.SortArtist {
		t.SortArtist = cur.SortArtist
		mod = true
	}
	if cur.SortComposer != orig.SortComposer {
		t.SortComposer = cur.SortComposer
		mod = true
	}
	if cur.SortName != orig.SortName {
		t.SortName = cur.SortName
		mod = true
	}
	if cur.TrackCount != orig.TrackCount {
		t.TrackCount = cur.TrackCount
		mod = true
	}
	if cur.TrackNumber != orig.TrackNumber {
		t.TrackNumber = cur.TrackNumber
		mod = true
	}
	if cur.VolumeAdjustment != orig.VolumeAdjustment {
		t.VolumeAdjustment = cur.VolumeAdjustment
		mod = true
	}
	if cur.Work != orig.Work {
		t.Work = cur.Work
		mod = true
	}
	if cur.PlayDate != nil {
		if t.PlayDate == nil || cur.PlayDate.After(t.PlayDate.Get()) {
			t.PlayDate = cur.PlayDate
			mod = true
		}
	}
	if cur.SkipDate != nil {
		if t.SkipDate == nil || cur.SkipDate.After(t.SkipDate.Get()) {
			t.SkipDate = cur.SkipDate
			mod = true
		}
	}
	if cur.PlayCount > orig.PlayCount {
		t.PlayCount += (cur.PlayCount - orig.PlayCount)
		mod = true
	}
	if cur.SkipCount > orig.SkipCount {
		t.SkipCount += (cur.SkipCount - orig.SkipCount)
		mod = true
	}
	if !cur.Unplayed && t.Unplayed {
		t.Unplayed = false
		mod = true
	}
	if mod {
		t.DateModified = &Time{time.Now().In(time.UTC)}
	}
}

