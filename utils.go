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
	return strings.Contains(v.Type().String(), ".")
}

func implementsUnmarshaler(v reflect.Value) bool {
	return v.Type().Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem())
}

func implementsMarshaler(v reflect.Value) bool {
	return v.Type().Implements(reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem())
}

func tryMarshal(v reflect.Value) (string, error) {
	// does it implement TextMarshaler?
	if implementsMarshaler(v) {
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
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}

// parseURITag gets structTag field from uriTag or jsonTag.
// If the jsonTag value is found the only the value before the comma is returned.
func parseURITag(v reflect.StructTag) string {
	if tag := v.Get(uriTag); tag != "" {
		return tag
	}
	return strings.Split(v.Get(jsonTag), ",")[0]
}
