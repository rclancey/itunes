package mdb

import (
	"encoding/json"
	"fmt"

)

const macEpoch = -2082844800

type bomaType struct {
	ID uint32
	Kind int
	Name string
}

const (
	BomaTypeNumeric = iota
	BomaTypeWideChar
	BomaTypeShortHeader
	BomaTypeLongHeader
	BomaTypeBook
	BomaTypePlaylistItem
	BomaTypeVideoInfo
	BomaTypeUnknown
	BomaTypeTimestamps
	BomaTypeGeniusInfo
)

var bomaTypes = map[uint32]bomaType{
	0x01: bomaType{1, BomaTypeNumeric, "Numeric"},

	0x02: bomaType{0x02, BomaTypeWideChar, "Title"},
	0x03: bomaType{0x03, BomaTypeWideChar, "Album"},
	0x04: bomaType{0x04, BomaTypeWideChar, "Artist"},
	0x05: bomaType{0x05, BomaTypeWideChar, "Genre"},
	0x06: bomaType{0x06, BomaTypeWideChar, "Kind"},
	0x07: bomaType{0x07, BomaTypeWideChar, "Unmapped0x7"},
	0x08: bomaType{0x08, BomaTypeWideChar, "Comment"},
	0x0b: bomaType{0x0b, BomaTypeWideChar, "Location"},
	0x0c: bomaType{0x0c, BomaTypeWideChar, "Composer"},
	0x0e: bomaType{0x0e, BomaTypeWideChar, "Grouping"},
	0x12: bomaType{0x12, BomaTypeWideChar, "YearSequence"},
	0x1b: bomaType{0x1b, BomaTypeWideChar, "AlbumArtist"},
	0x1e: bomaType{0x1e, BomaTypeWideChar, "SortTitle"},
	0x1f: bomaType{0x1f, BomaTypeWideChar, "SortAlbum"},
	0x20: bomaType{0x20, BomaTypeWideChar, "SortArtist"},
	0x21: bomaType{0x21, BomaTypeWideChar, "SortAlbumArtist"},
	0x22: bomaType{0x22, BomaTypeWideChar, "SortComposer"},
	0x2b: bomaType{0x2b, BomaTypeWideChar, "CopyrightHolder"},
	0x2e: bomaType{0x2e, BomaTypeWideChar, "CopyrightInfo"},
	0x34: bomaType{0x34, BomaTypeWideChar, "Flavor"},
	0x3b: bomaType{0x3b, BomaTypeWideChar, "PurchaserEmail"},
	0x3c: bomaType{0x3c, BomaTypeWideChar, "PurchaserName"},
	0x3f: bomaType{0x3f, BomaTypeWideChar, "Work"},
	0x40: bomaType{0x40, BomaTypeWideChar, "Movement"},
	0xc8: bomaType{0xc8, BomaTypeWideChar, "PlaylistName"},
	0x12c: bomaType{0x12c, BomaTypeWideChar, "iamaAlbum"},
	0x12d: bomaType{0x12d, BomaTypeWideChar, "iamaAlbumArtist"},
	0x12e: bomaType{0x12e, BomaTypeWideChar, "iamaArtist"},
	0x190: bomaType{0x190, BomaTypeWideChar, "iAmaName"},
	0x191: bomaType{0x191, BomaTypeWideChar, "iAmaSortName"},
	0x1f3: bomaType{0x1f8, BomaTypeWideChar, "Unmapped0x1f4"},
	0x1f8: bomaType{0x1f8, BomaTypeWideChar, "MediaFolder"},
	0x2be: bomaType{0x2be, BomaTypeWideChar, "ApplicationTitle"},
	0x2bf: bomaType{0x2bf, BomaTypeWideChar, "ApplicationArtist"},

	0x17: bomaType{0x17, BomaTypeTimestamps, "Timestamps"},
	0xcb: bomaType{0xcb, BomaTypeGeniusInfo, "GeniusInfo"},

	0x36: bomaType{0x36, BomaTypeShortHeader, "Unmapped0x36"},
	0x38: bomaType{0x38, BomaTypeShortHeader, "Unmapped0x38"},
	0x192: bomaType{0x192, BomaTypeShortHeader, "ArtworkURL"},

	0x1d: bomaType{0x1d, BomaTypeLongHeader, "FlavorPlist0x1d"},
	0xcd: bomaType{0xcd, BomaTypeLongHeader, "Unmapped0xcd"},
	0x2bc: bomaType{0x2bc, BomaTypeLongHeader, "Unmapped0x2bc"},
	0x3cc: bomaType{0x3cc, BomaTypeLongHeader, "Unmapped0x3cc"},

	0x42: bomaType{0x42, BomaTypeBook, "LocationComponentsBook0x42"},
	0x1fc: bomaType{0x1fc, BomaTypeBook, "Unmapped0x1fc"},
	0x1fd: bomaType{0x1fd, BomaTypeBook, "Unmapped0x1fd"},
	0x200: bomaType{0x200, BomaTypeBook, "Unmapped0x200"},

	0xce: bomaType{0xce, BomaTypePlaylistItem, "PlaylistItem"},

	0x24: bomaType{0x24, BomaTypeVideoInfo, "VideoInfo"},

	0xca: bomaType{0xca, BomaTypeUnknown, "SmartInfo"},
	0xc9: bomaType{0xc9, BomaTypeUnknown, "SmartCriteria"},
	0x1f6: bomaType{0x1f6, BomaTypeUnknown, "Unmapped0x1f6"},

}

type BomaSubType uint32

func (b BomaSubType) String() string {
	t, ok := bomaTypes[uint32(b)]
	if !ok {
		return fmt.Sprintf("Unknown0x%x", uint32(b))
	}
	return t.Name
}

func (b BomaSubType) Kind() int {
	t, ok := bomaTypes[uint32(b)]
	if !ok {
		return BomaTypeUnknown
	}
	return t.Kind
}

func (b BomaSubType) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.String())
}

//var UTF16_TYPES = map[uint32]bool{}
//var DATA_TYPE = map[uint32]string{}
//var DATASET_TYPE = map[uint32]string{}
var SORT_FIELD_TYPE = map[uint32]string{}
//var COLUMN_TYPE = map[uint32]string{}
var PLAYLIST_KIND_TYPE = map[uint32]string{}
//var SPECIAL_SMART_TYPE = map[uint32]string{}

func init() {
	// byte 561 = special kind:
	// 0=user,
	// 2=Movies,
	// 3=TV Shows,
	// 4=Music,
	// 5=Books,
	// 6=Tones,
	// 7=Rentals,
	// 10=Podcasts,
	// 19=Purchased,
	// 22=iTunes DJ,
	// 26=Genius,
	// 31=iTunes U,
	// 32=Genius Mix,
	// 35=Genius Mixes
	PLAYLIST_KIND_TYPE[2]  = "Movies"
	PLAYLIST_KIND_TYPE[3]  = "TV Shows"
	PLAYLIST_KIND_TYPE[4]  = "Music"
	PLAYLIST_KIND_TYPE[5]  = "Books"
	PLAYLIST_KIND_TYPE[6]  = "Tones"
	PLAYLIST_KIND_TYPE[7]  = "Rentals"
	PLAYLIST_KIND_TYPE[10] = "Podcasts"
	PLAYLIST_KIND_TYPE[19] = "Purchased"
	PLAYLIST_KIND_TYPE[22] = "iTunes DJ"
	PLAYLIST_KIND_TYPE[26] = "Genius"
	PLAYLIST_KIND_TYPE[31] = "iTunes U"
	PLAYLIST_KIND_TYPE[32] = "Genius Mix"
	PLAYLIST_KIND_TYPE[35] = "Genius Mixes"
	PLAYLIST_KIND_TYPE[47] = "Music Videos"
	PLAYLIST_KIND_TYPE[48] = "Home Videos"
	PLAYLIST_KIND_TYPE[65] = "Downloaded"

	SORT_FIELD_TYPE[1]   = "Playlist Order"
	SORT_FIELD_TYPE[5]   = "Name"
	SORT_FIELD_TYPE[6]   = "Album"
	SORT_FIELD_TYPE[7]   = "Artist"
	SORT_FIELD_TYPE[8]   = "Bit Rate"
	SORT_FIELD_TYPE[9]   = "Genre"
	SORT_FIELD_TYPE[10]  = "Kind"
	SORT_FIELD_TYPE[11]  = "Date Modified"
	SORT_FIELD_TYPE[12]  = "Track Number"
	SORT_FIELD_TYPE[13]  = "Release Date"
	SORT_FIELD_TYPE[13]  = "Size"
	SORT_FIELD_TYPE[14]  = "Total Time"
	SORT_FIELD_TYPE[15]  = "Year"
	SORT_FIELD_TYPE[16]  = "Sample Rate"
	SORT_FIELD_TYPE[17]  = "Category"
	SORT_FIELD_TYPE[20]  = "Comments"
	SORT_FIELD_TYPE[21]  = "Date Added"
	SORT_FIELD_TYPE[23]  = "Composer"
	SORT_FIELD_TYPE[25]  = "Play Count"
	SORT_FIELD_TYPE[26]  = "Play Date"
	SORT_FIELD_TYPE[27]  = "Disc Number"
	SORT_FIELD_TYPE[28]  = "Rating"
	SORT_FIELD_TYPE[32]  = "Movement Number"
	SORT_FIELD_TYPE[32]  = "Media Kind"
	SORT_FIELD_TYPE[33]  = "BPM"
	SORT_FIELD_TYPE[35]  = "Grouping"
	SORT_FIELD_TYPE[36]  = "Description"
	SORT_FIELD_TYPE[38]  = "Series"
	SORT_FIELD_TYPE[39]  = "Season"
	SORT_FIELD_TYPE[42]  = "Skip Count"
	SORT_FIELD_TYPE[43]  = "Skip Date"
	SORT_FIELD_TYPE[44]  = "Album Artist"
	SORT_FIELD_TYPE[47]  = "Album Rating"
	SORT_FIELD_TYPE[48]  = "Rental"
	SORT_FIELD_TYPE[116] = "Love"
	SORT_FIELD_TYPE[118] = "Work"
	SORT_FIELD_TYPE[119] = "Purchase Date"
	SORT_FIELD_TYPE[120] = "Movement Name"

}

type PlaylistKind uint32

func (k PlaylistKind) String() string {
	s, ok := PLAYLIST_KIND_TYPE[uint32(k)]
	if !ok {
		return fmt.Sprintf("Kind0x%x", uint32(k))
	}
	return s
}

func (k PlaylistKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}
