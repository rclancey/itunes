package pid

import (
	"database/sql/driver"
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type PersistentID uint64

func NewPersistentID() PersistentID {
	return PersistentID(rand.Uint64())
}

func (pid PersistentID) Pointer() *PersistentID {
	return &pid
}

func (pid PersistentID) String() string {
	v := strings.ToUpper(strconv.FormatUint(uint64(pid), 16))
	return strings.Repeat("0", 16 - len(v)) + v
}

func (pid *PersistentID) Decode(s string) error {
	v, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return errors.Wrap(err, "can't parse persistent id " + s)
	}
	*pid = PersistentID(v)
	return nil
}

func (pid PersistentID) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(pid.String())
	return data, errors.Wrapf(err, "can't json marshal perstent id %d", pid)
}

func (pid *PersistentID) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return errors.Wrap(err, "can't json unmarshal persistent id " + string(data))
	}
	return pid.Decode(s)
}

func (pid PersistentID) Int64() int64 {
	iv := int64(pid & 0x7fffffffffffffff)
	if pid > 0x7fffffffffffffff {
		iv *= -1
	}
	return iv
}

func (pid PersistentID) Value() (driver.Value, error) {
	return pid.Int64(), nil
}

func (pid *PersistentID) Scan(value interface{}) error {
	if value == nil {
		*pid = PersistentID(0)
		return nil
	}
	switch v := value.(type) {
	case int64:
		if v < 0 {
			*pid = PersistentID(uint64(-1 * v) | 0x8000000000000000)
			return nil
		}
		*pid = PersistentID(uint64(v))
		return nil
	case string:
		return pid.Decode(v)
	}
	return errors.Errorf("don't know how to convert %T into persistent id", value)
}

