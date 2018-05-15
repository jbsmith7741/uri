package uri

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

var (
	// Separator used for slices
	Separator = ","

	// supported struct tags
	uriTag      = "uri"
	defaultTag  = "default"
	requiredTag = "required"

	// supported tag values
	scheme    = "scheme"
	host      = "host"
	path      = "path"
	filename  = "filename"
	authority = "authority" // scheme://host
	origin    = "origin"    // scheme://host/path
	fragment  = "fragment"  // anything after hash #
)

func Marshal(v interface{}) (s string) {
	var u url.URL
	uVal := url.Values{}
	var vStruct reflect.Value
	if reflect.TypeOf(vStruct).Kind() == reflect.Ptr {
		vStruct = reflect.ValueOf(v).Elem()
	} else {
		vStruct = reflect.ValueOf(v)
	}

	for i := 0; i < vStruct.NumField(); i++ {
		field := vStruct.Field(i)

		var name string
		tag := vStruct.Type().Field(i).Tag.Get(uriTag)

		fs := GetFieldString(field)
		switch tag {
		case scheme:
			u.Scheme = fs
			continue
		case host:
			u.Host = fs
			continue
		case path:
			u.Path = fs
			continue
		case origin:
		case authority:
		case "":
			name = vStruct.Type().Field(i).Name
		default:
			name = tag
		}
		def := vStruct.Type().Field(i).Tag.Get(defaultTag)
		// skip default fields
		if def == "" && isZero(field) {
			continue
		} else if fs == def {
			continue
		}

		if field.Kind() == reflect.Slice {
			for _, v := range strings.Split(fs, ",") {
				uVal.Add(name, v)
			}
		} else {
			uVal.Add(name, fs)
		}
	}

	// Note: url values are sorted by string value as they are encoded
	u.RawQuery = uVal.Encode()

	return u.String()
}

func GetFieldString(value reflect.Value) string {
	switch value.Kind() {
	case reflect.String:
		return value.Interface().(string)
	case reflect.Bool:
		if value.Interface().(bool) == true {
			return "true"
		} else {
			return "false"
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%v", value.Interface())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", value.Interface())
	case reflect.Ptr:
		if value.IsNil() {
			return "nil"
		}
		return GetFieldString(value.Elem())
	case reflect.Slice:
		var s string
		for i := 0; i < value.Len(); i++ {
			s += GetFieldString(value.Index(i)) + ","
		}
		return strings.TrimRight(s, ",")
	case reflect.Struct:
		s, _ := tryMarshal(value)
		return s
	default:
		return ""
	}
}
