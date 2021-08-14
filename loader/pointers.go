package loader

import (
	"time"
)

func Stringp(s string) *string {
	return &s
}

func Boolp(b bool) *bool {
	return &b
}

func Uintp(i uint) *uint {
	return &i
}

func Uint8p(i uint8) *uint8 {
	return &i
}

func Uint16p(i uint16) *uint16 {
	return &i
}

func Uint32p(i uint32) *uint32 {
	return &i
}

func Uint64p(i uint64) *uint64 {
	return &i
}

func Intp(i int) *int {
	return &i
}

func Int8p(i int8) *int8 {
	return &i
}

func Int16p(i int16) *int16 {
	return &i
}

func Int32p(i int32) *int32 {
	return &i
}

func Int64p(i int64) *int64 {
	return &i
}

func Timep(t time.Time) *time.Time {
	return &t
}
