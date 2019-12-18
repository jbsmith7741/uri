package uri

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/jbsmith7741/trial"
)

func ExampleMarshal() {
	v := struct {
		Scheme string `uri:"scheme"`
		Host   string `uri:"host"`
		Path   string `uri:"path"`
		Name   string `uri:"name"`
		Count  int    `uri:"num"`
	}{
		Scheme: "https",
		Host:   "localhost",
		Path:   "root/index.html",
		Name:   "Hello World",
		Count:  11,
	}
	s := Marshal(v)
	fmt.Println(s)

	// Output: https://localhost/root/index.html?name=Hello+World&num=11
}

func ExampleGetFieldString() {
	name := "hello world"
	s := GetFieldString(reflect.ValueOf(name))
	fmt.Println(s)

	var i *int
	s = GetFieldString(reflect.ValueOf(i))
	fmt.Println(s)

	v := []int{2, 1, 3}
	s = GetFieldString(reflect.ValueOf(v))
	fmt.Println(s)

	// Output: hello world
	// nil
	// 2,1,3
}

func TestMarshal(t *testing.T) {
	type Embedded struct {
		Int    int
		String string
	}
	fn := func(args ...interface{}) (interface{}, error) {
		return Marshal(args[0]), nil
	}
	trial.New(fn, trial.Cases{
		"default values": {
			Input: struct {
				Int    int
				Amount float64 `uri:"float" default:"1.1"`
				Slice  []int   `default:"1,2,3"`
			}{
				Amount: 1.1,
				Slice:  []int{1, 2, 3},
			},
			Expected: "",
		},
		"runes": {
			Input: struct {
				R1 rune `uri:"r1"`
				R2 rune `uri:"r2"`
				R3 rune `uri:"r3"`
			}{
				R1: '\t',
				R2: 8984,
				R3: 'я',
			},
			Expected: "?r1=%09&r2=%E2%8C%98&r3=%D1%8F",
		},
		"slices": {
			Input: struct {
				Ints    []int
				Nil     []int
				Strings []string `uri:"strings"`
			}{
				Ints:    []int{1, 2, 3},
				Strings: []string{"hello", "world"},
			},
			Expected: "?Ints=1&Ints=2&Ints=3&strings=hello&strings=world",
		},
		"*struct with values": {
			Input: &struct {
				Int    int
				String string
			}{Int: 10, String: "hello"},
			Expected: "?Int=10&String=hello",
		},
		"nil *struct": {
			Input:    (*Embedded)(nil),
			Expected: "",
		},
		"time.Duration": {
			Input: struct {
				Dura time.Duration
			}{
				Dura: 10 * time.Minute,
			},
			Expected: "?Dura=10m0s",
		},
		"nil *struct with defaults": {
			Input: (*struct {
				Value string `default:"apple"`
			})(nil),
			Expected: "?Value=", // because there is a default uri sets the value to blank indicating it should be blank
		},
		/* todo how to handle this case?
		"empty slice with default value": {
			Input: struct {
				Floats []float64 `uri:"float" default:"3.14,2.7,7.7"`
			}{},
			Expected: "?float=[]",
		},*/
		"pointers": {
			Input: struct {
				Int     *int
				Nil     *int
				Default *int `default:"1"`
			}{
				Int: trial.IntP(10),
			},
			Expected: "?Default=nil&Int=10",
		},
		"structs": {
			Input: struct {
				Time   time.Time       `uri:"time"`
				Struct unmarshalStruct `uri:"struct"`
			}{
				Time:   trial.Time(time.RFC3339, "2018-04-04T00:00:00Z"),
				Struct: unmarshalStruct{Data: "Input"},
			},
			Expected: "?struct=Input&time=2018-04-04T00%3A00%3A00Z",
		},
		"custom time.Time": {
			Input: struct {
				Day time.Time `uri:"day" format:"2006-01-02"`
			}{Day: trial.TimeDay("2019-10-11")},
			Expected: "?day=2019-10-11",
		},
		"custom *time.Time": {
			Input: struct {
				Hour *time.Time `uri:"hour" format:"2006-01-02T15"`
			}{Hour: trial.TimeP("2006-01-02T15", "2019-10-11T12")},
			Expected: "?hour=2019-10-11T12",
		},
		"bools": {
			Input: struct {
				BoolT bool
				BoolF bool `default:"true"`
			}{true, false},
			Expected: "?BoolF=false&BoolT=true",
		},
		"embedded struct": {
			Input: struct {
				Embedded
			}{Embedded: Embedded{Int: 10, String: "brown fox"}},
			Expected: "?Int=10&String=brown+fox",
		},
		"embedded *struct (ptr)": {
			Input: struct {
				*Embedded
			}{Embedded: &Embedded{Int: 10, String: "brown fox"}},
			Expected: "?Int=10&String=brown+fox",
		},
		"fragment": {
			Input: struct {
				Int   int
				Value string `uri:"fragment"`
			}{Int: 11, Value: "this is a fragment"},
			Expected: "?Int=11#this%20is%20a%20fragment",
		},
		"scheme": {
			Input: struct {
				Scheme string `uri:"scheme"`
			}{Scheme: "http"},
			Expected: "http:",
		},
		"scheme + host + path": {
			Input: struct {
				Scheme string `uri:"scheme"`
				Host   string `uri:"host"`
				Path   string `uri:"path"`
			}{
				Scheme: "http",
				Host:   "localhost:8080",
				Path:   "path/to/file.txt",
			},
			Expected: "http://localhost:8080/path/to/file.txt",
		},
		"origin": {
			Input: struct {
				Origin string `uri:"origin"`
			}{
				Origin: "http://localhost:8080/path/to/file.txt",
			},
			Expected: "http://localhost:8080/path/to/file.txt",
		},
		"authority": {
			Input: struct {
				Authority string `uri:"authority"`
			}{
				Authority: "http://localhost:8080/path/to/file.txt",
			},
			Expected: "http://localhost:8080",
		},
		/* how should this case be handled?
		"origin override": {
			Input: struct {
				Origin string `uri:"origin"`
				Scheme string `uri:"scheme"`
				Host   string `uri:"host"`
				Path   string `uri:"path"`
			}{
				Origin: "http://localhost:8080/path/to/file.txt",
				Host:   "127.0.0.1",
				Path:   "other.txt",
			},
			Expected: "http://localhost:8080/path/to/file.txt",
		}, */
		"ignore default time": {
			Input: struct {
				Int  int `uri:"z"`
				Time time.Time
			}{Int: 10, Time: time.Time{}},
			Expected: "?z=10",
		},
		"alias type with stringer": {
			Input: struct {
				Dessert dessert
			}{Dessert: cake},
			Expected: "?Dessert=cake",
		},
	}).Test(t)
}
