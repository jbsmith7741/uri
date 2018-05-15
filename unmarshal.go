package uri

import (
	"encoding"
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/jbsmith7741/go-tools/appenderr"
)

// Unmarshal copies a standard parsable uri to a predefined struct
// [scheme:][//[userinfo@]host][/]path[?query][#fragment]
// scheme:opaque[?query][#fragment]
func Unmarshal(uri string, v interface{}) error {
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}
	values := u.Query()

	//verify that v is a pointer
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("struct must be a pointer")
	}
	vStruct := reflect.ValueOf(v).Elem()
	errs := appenderr.New()
	for i := 0; i < vStruct.NumField(); i++ {
		field := vStruct.Field(i)

		if !field.CanSet() { // skip private variables
			continue
		}

		name := vStruct.Type().Field(i).Name
		tag := vStruct.Type().Field(i).Tag.Get(uriTag)
		if tag != "" {
			name = tag
			tag = strings.ToLower(tag)
		}

		// check default values
		def := vStruct.Type().Field(i).Tag.Get(defaultTag)
		if def != "" {
			if err := SetField(field, def); err != nil {
				errs.Add(fmt.Errorf("default value %s can not be set to %s (%s)", def, name, field.Type()))
			}
		}

		required := vStruct.Type().Field(i).Tag.Get(requiredTag)
		data := values.Get(name)
		switch tag {
		case scheme:
			data = u.Scheme
		case host:
			data = u.Host
		case path:
			data = u.Path
		case filename:
			_, data = filepath.Split(u.Path)
		case origin:
			data = fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
			if u.Scheme == "" && u.Host == "" {
				data = u.Path
			}
		case authority:
			data = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		case fragment:
			data = u.Fragment
		default:
			if len(values[name]) == 0 && !(required == "true" && def == "") {
				continue
			}
		}

		if required == "true" && data == "" && def == "" {
			errs.Add(fmt.Errorf("%s is required", name))
			continue
		}

		if field.Kind() == reflect.Slice {
			data = strings.Join(values[name], Separator)
		}

		if err := SetField(field, data); err != nil {
			errs.Add(fmt.Errorf("%s can not be set to %s (%s)", data, name, field.Type()))
		}
	}

	return errs.ErrOrNil()
}

// SetField converts the string s to the type of value and sets the value if possible.
// Pointers and slices are recursively dealt with by deferencing the pointer
// or creating a generic slice of type value.
// All structs and alias' that implement encoding.TextUnmarshaler are suppported
func SetField(value reflect.Value, s string) error {
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
	}
	switch value.Kind() {
	case reflect.String:
		value.SetString(s)
	case reflect.Bool:
		b := strings.ToLower(s) == "true" || s == ""
		value.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
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
		SetField(z.Elem(), s)

		value.Set(z)

	case reflect.Slice:
		// create a generate slice and recursively assign the elements
		baseType := reflect.TypeOf(value.Interface()).Elem()
		data := strings.Split(s, Separator)
		slice := reflect.MakeSlice(value.Type(), 0, len(data))
		for _, v := range data {
			baseValue := reflect.New(baseType).Elem()
			SetField(baseValue, v)
			slice = reflect.Append(slice, baseValue)
		}
		value.Set(slice)

	case reflect.Struct:
		v := reflect.New(value.Type())
		if implementsUnmarshaler(v) {
			err := v.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s))
			if err != nil {
				return err
			}
		}
		value.Set(v.Elem())

	default:
		return fmt.Errorf("Unsupported type %v", value.Kind())
	}
	return nil
}
