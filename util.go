package itunes

import (
	//"fmt"
	"reflect"
	"strings"
	"strconv"
	"time"
)

var pidType = reflect.TypeOf(PersistentID(0))

func SetField(s interface{}, key[]byte, kind string, val []byte) bool {
	k := strings.Replace(string(key), " ", "", -1)
	v := string(val)
	rs := reflect.ValueOf(s).Elem()
	f := rs.FieldByName(k)
	if !f.IsValid() {
		return false
	}
	switch f.Kind() {
	case reflect.Ptr:
		pval := reflect.New(f.Type().Elem())
		switch pval.Elem().Kind() {
		case reflect.Bool:
			if kind == "true" {
				pval.Elem().SetBool(true)
			} else {
				pval.Elem().SetBool(false)
			}
		case reflect.Uint64:
			var base int
			if f.Type().Elem() == pidType {
				base = 16
			} else {
				base = 10
			}
			uv, err := strconv.ParseUint(v, base, 64)
			if err != nil {
				return false
			}
			pval.Elem().SetUint(uv)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
			if kind == "integer" {
				iv, err := strconv.ParseUint(v, 10, 64)
				if err == nil {
					pval.Elem().SetUint(iv)
				}
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if kind == "integer" {
				iv, err := strconv.ParseInt(v, 10, 64)
				if err == nil {
					pval.Elem().SetInt(iv)
				}
			}
		case reflect.String:
			if kind == "string" {
				pval.Elem().SetString(v)
			}
		default:
			vi := f.Interface()
			//fmt.Printf("default for %s (%s) %T\n", string(key), kind, vi)
			switch vi.(type) {
			case *Time:
				//fmt.Println("field is time")
				t, err := time.Parse("2006-01-02T15:04:05Z", v)
				if err != nil {
					//fmt.Printf("can't parse '%s' as a time: %s\n", v, err)
					return false
				}
				pval.Elem().Set(reflect.ValueOf(Time{t}))
			default:
				//fmt.Println("field is not a time")
				return false
			}
		}
		/*
		case time.Time:
			if kind == "date" {
				it, err := time.Parse("2006-01-02T15:04:05Z", v)
				if err == nil {
					pval.SetPointer(unsafe.Pointer(&it))
				//tr := reflect.TypeOf(it).Elem()
				//f.Set(reflect.Indirect(reflect.New(tr)).Interface().(*time.Time))
			}
		default:
			return false
		}
		*/
		f.Set(pval)
		return true
	case reflect.Bool:
		if kind == "true" {
			f.SetBool(true)
		} else {
			f.SetBool(false)
		}
	case reflect.Uint64:
		var base int
		if f.Type() == pidType {
			base = 16
		} else {
			base = 10
		}
		uv, err := strconv.ParseUint(v, base, 64)
		if err != nil {
			return false
		}
		f.SetUint(uv)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		if kind == "integer" {
			iv, err := strconv.ParseUint(v, 10, 64)
			if err == nil {
				f.SetUint(iv)
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if kind == "integer" {
			iv, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				f.SetInt(iv)
			}
		}
	case reflect.String:
		if kind == "string" {
			f.SetString(v)
		}
	case reflect.Slice:
		if kind == "data" {
			bval, err := decodeb64(val)
			if err != nil {
				f.SetBytes(val)
			} else {
				f.SetBytes(bval)
			}
		}
		return true
	}
	return false
}

