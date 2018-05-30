package conv

import (
	"reflect"
)

func IntPtrTo64(ptr interface{}) (value int64) {
	if v := reflect.ValueOf(ptr); v.Kind() == reflect.Ptr {
		p := v.Elem()
		switch p.Kind() {
		case reflect.Int:
			value = int64(*ptr.(*int))
		case reflect.Int8:
			value = int64(*ptr.(*int8))
		case reflect.Int16:
			value = int64(*ptr.(*int16))
		case reflect.Int32:
			value = int64(*ptr.(*int32))
		case reflect.Int64:
			value = *ptr.(*int64)
		}
	}
	return
}

func UintPtrTo64(ptr interface{}) (value uint64) {
	if v := reflect.ValueOf(ptr); v.Kind() == reflect.Ptr {
		p := v.Elem()
		switch p.Kind() {
		case reflect.Uint:
			value = uint64(*ptr.(*uint))
		case reflect.Uint8:
			value = uint64(*ptr.(*uint8))
		case reflect.Uint16:
			value = uint64(*ptr.(*uint16))
		case reflect.Uint32:
			value = uint64(*ptr.(*uint32))
		case reflect.Uint64:
			value = *ptr.(*uint64)
		}
	}
	return
}
