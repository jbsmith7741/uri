package uri

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/jbsmith7741/go-tools/trial"
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
	}).EqualFn(trial.ContainsFn).Test(t)
}
