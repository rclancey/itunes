package itunes

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var serializationRoot string

func SetSerializationRoot(dn string) {
	serializationRoot = dn
}

type Serializable interface {
	Serialize(int) ([]byte, error)
	SerializationPath() []string
}

func EnsureDir(fn string) error {
	dn := filepath.Dir(fn)
	st, err := os.Stat(dn)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		err = os.MkdirAll(dn, os.FileMode(0755))
		if err != nil {
			return err
		}
		return nil
	}
	if !st.IsDir() {
		return fmt.Errorf("%s exists and is not a directory", dn)
	}
	return nil
}

func Serialize(obj Serializable) error {
	data, err := obj.Serialize(2)
	if err != nil {
		return err
	}
	parts := obj.SerializationPath()
	if parts == nil || len(parts) == 0 {
		return errors.New("no serialization path")
	}
	fn := filepath.Join(serializationRoot, filepath.Join(parts...) + ".pb")
	err = EnsureDir(fn)
	if err != nil {
		return err
	}
	f, err := os.Create(fn)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}

