package itunes

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type Time struct {
	time.Time
}

func (t *Time) Get() time.Time {
	return t.Time
}

func (t *Time) Set(tm time.Time) {
	t.Time = tm
}

func (t *Time) SetEpochMS(ms int64) {
	s := ms / 1000
	ns := (ms % 1000) * 1e6
	t.Time = time.Unix(s, ns)
}

func (t *Time) EpochMS() int64 {
	s := t.Unix()
	ns := int64(t.Nanosecond())
	return (s * 1000) + (ns / 1e6)
}

func (t *Time) MarshalJSON() ([]byte, error) {
	if t == nil || t.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(strconv.FormatInt(t.EpochMS(), 10)), nil
}

func (t *Time) UnmarshalJSON(data []byte) error {
	v, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	t.SetEpochMS(v)
	return nil
}

func (t *Time) UnmarshalPlist(obj interface{}) error {
	if tm, isa := obj.(time.Time); isa {
		t.Set(tm)
		return nil
	}
	rv := reflect.ValueOf(obj)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		t.SetEpochMS(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		t.SetEpochMS(int64(rv.Uint()))
	case reflect.Float32, reflect.Float64:
		t.SetEpochMS(int64(rv.Float()))
	case reflect.String:
		tm, err := time.Parse("2006-01-02T15:04:05Z", rv.String())
		if err != nil {
			return err
		}
		t.Set(tm)
	default:
		return fmt.Errorf("don't know how to use %T as time", obj)
	}
	return nil
}

/*
func (t *Time) MarshalPlist(e *plist.Encoder) error {
	if t == nil {
		err := e.StartTag("null")
		if err != nil {
			return err
		}
		return e.EndTag("null")
	}
	return e.Encode(t.Time)
}
*/
