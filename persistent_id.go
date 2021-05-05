package itunes

import (
	"math/rand"
	"strconv"
	"strings"
)

type PersistentID uint64

func NewPersistentID() PersistentID {
	return PersistentID(rand.Uint64())
}

func (pid PersistentID) String() string {
	return pid.EncodeToString()
}

func (pid PersistentID) EncodeToString() string {
	v := strings.ToUpper(strconv.FormatUint(uint64(pid), 16))
	return strings.Repeat("0", 16 - len(v)) + v
}

func (pid *PersistentID) DecodeString(s string) error {
	v, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return err
	}
	*pid = PersistentID(v)
	return nil
}

func (pid PersistentID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + pid.EncodeToString() + `"`), nil
}

func (pid *PersistentID) UnmarshalJSON(data []byte) error {
	n := len(data)
	if data[0] == '"' && data[n-1] == '"' {
		return pid.DecodeString(string(data[1:n-1]))
	}
	return pid.DecodeString(string(data))
}
