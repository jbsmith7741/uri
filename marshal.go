package uri

import (
	"fmt"
	"log"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	// sliceDelim used for slices
	sliceDelim = ","
	mapDelim   = "|"

	// supported struct tags
	uriTag      = "uri"
	jsonTag     = "json"
	defaultTag  = "default"
	requiredTag = "required"

	// supported tag values
	scheme    = "scheme"
	host      = "host"
	path      = "path"
	userinfo  = "userinfo"
	password  = "password"
	username  = "username"
	filename  = "filename"
	authority = "authority" // userinfo@host
	origin    = "origin"    // scheme://host/path - see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Origin
	fragment  = "fragment"  // anything after hash #
)

// Marshal a struct into a string representation of a uri
// Note: Marshal panics if a struct or pointer to a struct is not provided
func Marshal(v interface{}) (s string) {
	u := &url.URL{}
	uVal := &url.Values{}
	vStruct := reflect.ValueOf(v)
	if vStruct.Kind() == reflect.Ptr {
		if vStruct.IsNil() {
			vStruct = reflect.New(vStruct.Type().Elem())
		}
		vStruct = vStruct.Elem()
	}

	parseStruct(u, uVal, vStruct)

	// Note: url values are sorted by string value as they are encoded
	u.RawQuery = uVal.Encode()
	return u.String()
}

// MarshalUnescaped is the same as marshal but without url encoding the values
func MarshalUnescaped(v interface{}) string {
	m := Marshal(v)
	s, err := url.QueryUnescape(m)
	if err != nil {
		log.Println(err)
		return m
	}
	return s

}

func parseStruct(u *url.URL, uVal *url.Values, vStruct reflect.Value) {
	for i := 0; i < vStruct.NumField(); i++ {
		field := vStruct.Field(i)
		if !field.CanInterface() {
			continue // skip unexported variables
		}
		// check for embedded struct and handle recursively
		if field.Kind() == reflect.Struct {
			ptr := reflect.New(field.Type())
			if !implementsMarshaler(ptr) {
				parseStruct(u, uVal, field)
				continue
			}
		} else if field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct {
			if !implementsMarshaler(field) {
				parseStruct(u, uVal, field.Elem())
				continue
			}
		}
		var name string
		structTag := vStruct.Type().Field(i).Tag
		tag := parseURITag(structTag)

		fs := GetFieldString(field, structTag)

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
		case fragment:
			u.Fragment = fs
			continue
		case userinfo:
			u.User = url.User(fs)
			continue
		case username:
			p, set := u.User.Password()
			if set {
				u.User = url.UserPassword(fs, p)
			} else {
				u.User = url.User(fs)
			}
			continue
		case password:
			u.User = url.UserPassword(u.User.Username(), fs)
			continue
		case origin: // scheme://host/path
			l, err := url.Parse(fs)
			if err == nil {
				u.Host = l.Host
				u.Scheme = l.Scheme
				u.Path = l.Path
			}
			continue
		case authority: //userinfo@host
			l, err := url.Parse(fs)
			if err == nil {
				u.User = l.User
				u.Host = l.Host
			}
			continue
		case "-": // skip disabled fields
			continue
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
			for _, v := range strings.Split(fs, sliceDelim) {
				uVal.Add(name, v)
			}
		} else {
			uVal.Add(name, fs)
		}
	}
}

// GetFieldString returns a string representation of a Value
// booleans become true/false
// nil pointers return "nil"
// slices combine elements with a comma. []int{1,2,3} -> "1,2,3"
func GetFieldString(value reflect.Value, sTag reflect.StructTag) string {

	format := sTag.Get("format")
	if format != "" {
		if value.Type() == reflect.TypeOf(time.Time{}) {
			return value.Interface().(time.Time).Format(format)
		}
		if value.Type() == reflect.TypeOf(&time.Time{}) {
			return value.Interface().(*time.Time).Format(format)
		}
	}

	if format == "rune" && value.Kind() == reflect.Int32 {
		return string(value.Interface().(rune))
	}

	switch value.Kind() {
	case reflect.String:
		return value.Interface().(string)
	case reflect.Bool:
		if value.Interface().(bool) == true {
			return "true"
		}
		return "false"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%v", value.Interface())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%v", value.Interface())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", value.Interface())
	case reflect.Ptr:
		if value.IsNil() {
			return "nil"
		}
		return GetFieldString(value.Elem(), sTag)
	case reflect.Slice:
		var s string
		for i := 0; i < value.Len(); i++ {
			s += GetFieldString(value.Index(i), sTag) + sliceDelim
		}
		return strings.TrimRight(s, sliceDelim)
	case reflect.Struct:
		s, _ := tryMarshal(value)
		return s
	case reflect.Map:
		iter := value.MapRange()
		s := make([]string, 0)
		for iter.Next() {
			k := GetFieldString(iter.Key(), sTag)
			v := GetFieldString(iter.Value(), sTag)
			s = append(s, k+":"+v)
		}
		sort.Sort(sort.StringSlice(s)) // sorted for consistency
		return strings.Join(s, mapDelim)
	default:
		return ""
	}
}
