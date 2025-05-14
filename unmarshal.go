package uri

import (
	"encoding"
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jbsmith7741/go-tools/appenderr"
)

var regMapSplit = regexp.MustCompile("^(.*?)[:](.*)")

// Unmarshal copies a standard parsable uri to a predefined struct
// [scheme:][//[userinfo@]host][/]path[?query][#fragment]
// scheme:opaque[?query][#fragment]
func Unmarshal(uri string, v interface{}) error {
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}
	if strings.Contains(u.RawQuery, ";") {
		u.RawQuery = strings.Replace(u.RawQuery, ";", "%3B", -1)
	}
	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return err
	}

	//verify that v is a pointer
	if value := reflect.ValueOf(v); value.Kind() != reflect.Ptr || value.IsNil() {
		return fmt.Errorf("%v must be a non nil pointer", reflect.TypeOf(v))
	}

	vStruct := reflect.ValueOf(v).Elem()
	errs := appenderr.New()
	for i := 0; i < vStruct.NumField(); i++ {
		field := vStruct.Field(i)

		if !field.CanSet() { // skip private variables
			continue
		}

		name := vStruct.Type().Field(i).Name
		tag := parseURITag(vStruct.Type().Field(i).Tag)
		if tag == "-" {
			continue
		}
		if tag != "" {
			name = tag
			tag = strings.ToLower(tag)
		}

		// check default values
		def := vStruct.Type().Field(i).Tag.Get(defaultTag)
		if def != "" {
			if err := SetField(field, def, vStruct.Type().Field(i)); err != nil {
				errs.Add(fmt.Errorf("default value %s can not be set to %s (%s)", def, name, field.Type()))
			}
		}

		skip, err := handleEmbeddeStruct(uri, field)
		errs.Add(err)
		if skip {
			continue
		}

		required := vStruct.Type().Field(i).Tag.Get(requiredTag)
		data := values.Get(name)
		if field.Kind() == reflect.Slice {
			data = strings.Join(values[name], sliceDelim)
		}
		if field.Kind() == reflect.Map {
			data = strings.Join(values[name], mapDelim)
		}
		switch tag {
		case scheme:
			data = u.Scheme
		case host:
			data = u.Host
		case path:
			data = u.Path
		case userinfo:
			data = u.User.String()
		case username:
			data = u.User.Username()
		case password:
			data, _ = u.User.Password()
		case filename:
			_, data = filepath.Split(u.Path)
		case origin:
			data = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
			if u.Scheme == "" && u.Host == "" {
				data = u.Path
			}
		case authority:
			data = u.Host
			if userinfo := u.User.String(); userinfo != "" {
				data = fmt.Sprintf("%s@%s", userinfo, u.Host)
			}
		case fragment:
			data = u.Fragment
		default:
			if len(values[name]) == 0 && !(required == "true" && def == "") {
				continue
			}
		}

		if required == "true" && data == "" && def == "" {
			errs.Addf("%s is required", name)
			continue
		}

		if err := SetField(field, data, vStruct.Type().Field(i)); err != nil {
			errs.Wrapf(err, "%q can not be set to %s (%s)", data, name, field.Type())
		}
	}

	return errs.ErrOrNil()
}

// UnmarshalQuery is a comparable to the url.ParseQuery()
func UnmarshalQuery(query string, v interface{}) error {
	u := url.URL{}
	u.RawQuery = query
	return Unmarshal(u.String(), v)
}

func handleEmbeddeStruct(uri string, value reflect.Value) (bool, error) {
	// do we have an embedded struct
	switch value.Kind() {
	case reflect.Struct:
		v := value.Addr()
		// if the struct implements the unmarshaler let SetField handle the parsing
		if implementsUnmarshaler(v) {
			return false, nil
		}

		err := Unmarshal(uri, v.Interface())
		return true, err
	case reflect.Ptr:
		v := reflect.New(value.Type().Elem())
		if v.Elem().Kind() != reflect.Struct {
			return false, nil
		}
		// if the struct implements the unmarshaler let SetField handle the parsing
		if implementsUnmarshaler(value) {
			return false, nil
		}
		if value.IsNil() {
			err := Unmarshal(uri, v.Interface())
			v2 := reflect.New(value.Type().Elem())
			// only set the pointer if values changed, otherwise keep it as nil
			if !reflect.DeepEqual(v.Interface(), v2.Interface()) {
				value.Set(v)
			}
			return true, err
		}
		err := Unmarshal(uri, value.Interface())
		return true, err
	}

	return false, nil
}

// SetField converts the string s to the type of value and sets the value if possible.
// Pointers and slices are recursively dealt with by deferencing the pointer
// or creating a generic slice of type value.
// All structs and alias' that implement encoding.TextUnmarshaler are suppported
func SetField(value reflect.Value, s string, sField reflect.StructField) error {
	if isAlias(value) {
		v := reflect.New(value.Type())
		if implementsUnmarshaler(v) {
			err := v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s))
			if err != nil {
				return err
			}
			value.Set(v.Elem())
			return nil
		}
		if value.Type() == reflect.TypeOf(time.Second) {
			if d, err := time.ParseDuration(s); err == nil {
				value.Set(reflect.ValueOf(d))
				return nil
			}
		}
	}
	switch value.Kind() {
	case reflect.String:
		value.SetString(s)
	case reflect.Bool:
		b := strings.ToLower(s) == "true" || s == ""
		value.SetBool(b)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(s, 10, 0)
		if err != nil {
			return err
		}
		value.SetUint(i)
	case reflect.Int32:
		if sField.Tag.Get("format") == "rune" {
			r, _ := utf8.DecodeRuneInString(s)
			value.Set(reflect.ValueOf(r))
			return nil
		}
		fallthrough
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			return err
		}
		value.SetInt(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 0)
		if err != nil {
			return err
		}
		value.SetFloat(f)
	case reflect.Ptr:
		// create non pointer type and recursively assign
		z := reflect.New(value.Type().Elem())
		if s == "nil" {
			return nil
		}
		SetField(z.Elem(), s, sField)
		value.Set(z)
	case reflect.Slice:
		// create a generate slice and recursively assign the elements
		baseType := reflect.TypeOf(value.Interface()).Elem()
		if s == "" { // ignore empty slices
			return nil
		}
		data := strings.Split(s, sliceDelim)
		slice := reflect.MakeSlice(value.Type(), 0, len(data))
		for _, v := range data {
			baseValue := reflect.New(baseType).Elem()
			SetField(baseValue, v, sField)
			slice = reflect.Append(slice, baseValue)
		}
		value.Set(slice)
	case reflect.Struct:
		v := reflect.New(value.Type())
		if value.Type() == reflect.TypeOf(time.Time{}) {
			format := sField.Tag.Get("format")
			if format == "" {
				format = time.RFC3339
			}
			t, err := time.Parse(format, s)
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(t))
			return nil

		}
		if implementsUnmarshaler(v) {
			err := v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s))
			if err != nil {
				return err
			}
			value.Set(v.Elem())
		}
	case reflect.Map:
		// set map if nil
		if value.IsNil() {
			m := reflect.MakeMap(value.Type())
			value.Set(m)
		}

		// Type for map key anv value
		kType := value.Type().Key()
		vType := value.Type().Elem()

		// Split string into fields and key,value pairs
		for _, row := range strings.Split(s, mapDelim) {
			d := regMapSplit.FindStringSubmatch(row)
			if len(d) != 3 {
				return fmt.Errorf("invalid map format expected key:value got %v", row)
			}
			k, v := d[1], d[2] // d[0] is match string
			// set key value
			kValue := reflect.New(kType).Elem()
			if err := SetField(kValue, k, sField); err != nil {
				return err
			}
			// set value value
			vValue := reflect.New(vType).Elem()
			if err := SetField(vValue, v, sField); err != nil {
				return err
			}
			// add key/value pair to map
			value.SetMapIndex(kValue, vValue)
		}
	default:
		return fmt.Errorf("Unsupported type %v", value.Kind())
	}
	return nil
}
