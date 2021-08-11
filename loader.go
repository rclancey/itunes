package itunes

import (
	"strings"

	"github.com/rclancey/itunes/itl"
	"github.com/rclancey/itunes/loader"
	"github.com/rclancey/itunes/mdb"
	"github.com/rclancey/itunes/plist"
)

func NewLoader(fn string) loader.Loader {
	if strings.HasSuffix(fn, ".xml") {
		return plist.NewLoader()
	}
	if strings.HasSuffix(fn, ".itl") {
		return itl.NewLoader()
	}
	if strings.HasSuffix(fn, ".musicdb") {
		return mdb.NewLoader()
	}
	return itl.NewLoader()
}
