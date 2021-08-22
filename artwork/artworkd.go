package artwork

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/rclancey/itunes/persistentId"
)

type ArtworkDB struct {
	root string
	fn string
	db *sqlx.DB
	libid pid.PersistentID
}

func NewArtworkDB(homedir string, libid pid.PersistentID) (*ArtworkDB, error) {
	root := filepath.Join(
		homedir,
		"Library",
		"Containers",
		"com.apple.AMPArtworkAgent",
		"Data",
		"Documents",
	)
	fn := filepath.Join(root, "artworkd.sqlite")
	_, err := os.Stat(fn)
	if err != nil {
		return nil, err
	}
	db, err := sqlx.Connect("sqlite3", fn)
	if err != nil {
		return nil, err
	}
	return &ArtworkDB{
		root: root,
		fn: fn,
		db: db,
		libid: libid,
	}, nil
}

type Format uint32

func (f Format) String() string {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, uint32(f))
	return string(buf.Bytes())
}

type ArtworkItem struct {
	LibraryID pid.PersistentID `db:"ZDBID"`
	PersistentID pid.PersistentID `db:"ZPERSISTENTID"`
	Hash string `db:"ZHASHSTRING"`
	Width float64 `db:"ZWIDTH"`
	Height float64 `db:"ZHEIGHT"`
	Format Format `db:"ZFORMAT"`
	Kind int `db:"ZKIND"`
	URL *string `db:"ZURL"`
}

func (db *ArtworkDB) GetArtworkItem(id pid.PersistentID) (*ArtworkItem, error) {
	qs := `
SELECT db.ZDBID,
       db.ZPERSISTENTID,
       src.ZURL,
       img.ZHASHSTRING,
       img.ZKIND,
       c.ZWIDTH,
       c.ZHEIGHT,
       c.ZFORMAT
  FROM ZDATABASEITEMINFO db
  LEFT JOIN ZSOURCEINFO src ON db.ZSOURCEINFO = src.Z_PK
  LEFT JOIN ZIMAGEINFO img ON src.ZIMAGEINFO = img.Z_PK
  LEFT JOIN ZCACHEITEM c ON src.ZIMAGEINFO = c.ZIMAGEINFO
 WHERE db.ZDBID = ?
   AND db.ZPERSISTENTID = ?`
	log.Println(qs, db.libid.Int64(), id.Int64())
	rows, err := db.db.Queryx(qs, db.libid, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item := &ArtworkItem{}
		err := rows.StructScan(item)
		if err != nil {
			return nil, err
		}
		return item, nil
	}
	return nil, sql.ErrNoRows
}

func (db *ArtworkDB) GetArtworkFile(id pid.PersistentID) (string, error) {
	item, err := db.GetArtworkItem(id)
	if err != nil {
		return "", err
	}
	var ext string
	switch item.Format.String() {
	case "JPEG":
		ext = "jpeg"
	case "PNGf":
		ext = "png"
	default:
		return "", errors.New("unknown format " + item.Format.String())
	}
	fn := filepath.Join(db.root, "artwork", fmt.Sprintf("%s_sk_%d_cid_1.%s", item.Hash, item.Kind, ext))
	_, err = os.Stat(fn)
	if err != nil {
		return "", err
	}
	return fn, nil
}

func (db *ArtworkDB) GetArtworkURL(id pid.PersistentID) (string, error) {
	item, err := db.GetArtworkItem(id)
	if err != nil {
		return "", err
	}
	if item.URL == nil {
		return "", nil
	}
	return *item.URL, nil
}

func (db *ArtworkDB) GetJPEG(id pid.PersistentID) ([]byte, error) {
	fn, err := db.GetArtworkFile(id)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(fn, ".jpeg") {
		f, err := os.Open(fn)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return ioutil.ReadAll(f)
	}
	if strings.HasSuffix(fn, ".png") {
		f, err := os.Open(fn)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		img, err := png.Decode(f)
		if err != nil {
			return nil, err
		}
		buf := &bytes.Buffer{}
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 75})
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	return nil, errors.New("unknown image format")
}
