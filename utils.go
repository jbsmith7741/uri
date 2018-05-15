package uri

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
)

func isAlias(v reflect.Value) bool {
	if v.Kind() == reflect.Struct || v.Kind() == reflect.Ptr {
		return false
	}
	s := fmt.Sprint(v.Type())
	return strings.Contains(s, ".")
}

func implementsUnmarshaler(v reflect.Value) bool {
	return v.Type().Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem())
}

func tryMarshal(v reflect.Value) (string, error) {
	// does it implement TextMarshaler?
	if v.Type().Implements(reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()) {
		b, err := v.Interface().(encoding.TextMarshaler).MarshalText()
		return string(b), err
	} else if v.Type().Implements(reflect.TypeOf((*fmt.Stringer)(nil)).Elem()) {
		return v.Interface().(fmt.Stringer).String(), nil
	}
	return "", nil
}

func isZero(v reflect.Value) bool {
	if !v.CanInterface() {
		return false
	}
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}
