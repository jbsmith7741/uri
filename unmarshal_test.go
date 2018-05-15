package uri

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jbsmith7741/go-tools/trial"
)

type testStruct struct {
	// basic types
	String   string
	Bool     bool
	Int      int
	IntP     *int
	Int32    int32
	Int32P   *int32
	Int64    int64
	Int64P   *int64
	Float32  float32
	Float32P *float32
	Float64  float64
	Float64P *float64

	// slice
	Strings  []string
	Ints     []int
	IntsP    []*int
	Ints32   []int32
	Ints64   []int64
	Floats32 []float32
	Floats64 []float64

	// struct
	Time       time.Time
	TimeP      *time.Time
	Unmarshal  unmarshalStruct
	UnmarshalP *unmarshalStruct

	// alias
	Dessert dessert
}

type unmarshalStruct struct {
	Data string
}

func (s *unmarshalStruct) UnmarshalText(text []byte) error {
	s.Data = string(text)
	return nil
}

func (s unmarshalStruct) MarshalText() ([]byte, error) {
	return []byte(s.Data), nil
}

func TestUnmarshal(t *testing.T) {
	tm, _ := time.Parse(time.RFC3339, "2017-10-10T12:12:12Z")
	cases := map[string]struct {
		uri       string
		shouldErr bool
		expected  testStruct
	}{
		"string": {
			uri:      "?String=hello",
			expected: testStruct{String: "hello"},
		},
		"integer: int, int32, int64": {
			uri:      "?Int=10&Int32=32&Int64=64",
			expected: testStruct{Int: 10, Int32: 32, Int64: 64},
		},
		"pointer: *int, *int32, *int64": {
			uri:      "?IntP=77&Int32P=11&Int64P=222",
			expected: testStruct{IntP: trial.IntP(77), Int32P: trial.Int32P(11), Int64P: trial.Int64P(222)},
		},
		"invalid integer": {
			uri:       "?Int=abc",
			shouldErr: true,
		},
		"float32, float64": {
			uri:      "?Float32=12.2&Float64=33.3",
			expected: testStruct{Float32: 12.2, Float64: 33.3},
		},
		"pointer: *float32, *float64": {
			uri:      "?Float32P=12.2&Float64P=33.3",
			expected: testStruct{Float32P: trial.Float32P(12.2), Float64P: trial.Float64P(33.3)},
		},
		"invalid float": {
			uri:       "?Float32=abc",
			shouldErr: true,
		},
		"time.Time": {
			uri:      "?Time=2017-10-10T12:12:12Z",
			expected: testStruct{Time: tm},
		},
		"*time.Time": {
			uri:      "?TimeP=2017-10-10T12:12:12Z",
			expected: testStruct{TimeP: &tm},
		},
		"invalid time": {
			uri:       "?Time=2017-10-",
			shouldErr: true,
		},
		"struct with UnMarshalText": {
			uri: "?Unmarshal=abc&UnmarshalP=def",
			expected: testStruct{
				Unmarshal:  unmarshalStruct{Data: "abc"},
				UnmarshalP: &unmarshalStruct{Data: "def"},
			},
		},
		"bool": {
			uri: "?Bool=true",
			expected: testStruct{
				Bool: true,
			},
		},
		"bool implicit true": {
			uri: "?Bool&Test",
			expected: testStruct{
				Bool: true,
			},
		},
		"slice of string": {
			uri: "?Strings=a&Strings=b&Strings=c",
			expected: testStruct{
				Strings: []string{"a", "b", "c"},
			},
		},
		"slice: int, int32, int64": {
			uri: "?Ints=1&Ints=2&Ints=3&Ints32=4,5,6&Ints64=7,8,9",
			expected: testStruct{
				Ints:   []int{1, 2, 3},
				Ints32: []int32{4, 5, 6},
				Ints64: []int64{7, 8, 9},
			},
		},
		"slice: float32, float64": {
			uri: "?Floats32=1.1&Floats32=2.2&Floats32=3.3&Floats64=4.4,5.5,6.6",
			expected: testStruct{
				Floats32: []float32{1.1, 2.2, 3.3},
				Floats64: []float64{4.4, 5.5, 6.6},
			},
		},
		"slice of *int": {
			uri: "?IntsP=1,2,3",
			expected: testStruct{
				IntsP: []*int{trial.IntP(1), trial.IntP(2), trial.IntP(3)},
			},
		},
		"slice of *int with nil": {
			uri: "?IntsP=1,2,nil,3",
			expected: testStruct{
				IntsP: []*int{trial.IntP(1), trial.IntP(2), nil, trial.IntP(3)},
			},
		},
		"alias type (dessert)": {
			uri:      "?Dessert=brownie",
			expected: testStruct{Dessert: brownie},
		},
		"invalid alias type": {
			uri:       "?Dessert=cat",
			shouldErr: true,
		},
	}
	for msg, test := range cases {
		var d testStruct
		err := Unmarshal(test.uri, &d)
		if err != nil != test.shouldErr {
			t.Errorf("FAIL: %v error mismatch %v", msg, err)
		} else if !test.shouldErr && !cmp.Equal(d, test.expected) {
			t.Errorf("FAIL: %v values did not match %s", msg, cmp.Diff(d, test.expected))
		} else {
			t.Logf("PASS: %v", msg)
		}
	}
}

type (
	primitiveDefault struct {
		// basic types
		String  string  `default:"hello"`
		Bool    bool    `default:"true"`
		Int     int     `default:"42"`
		Float32 float32 `default:"12.34"`
	}
	sliceDefault struct {
		Strings []string `default:"hello,world"`
		Ints    []int    `default:"11"`
	}
	aliasDefault struct {
		Dessert dessert `default:"cake"`
	}
)

func TestTags(t *testing.T) {
	cases := map[string]struct {
		uri      string
		expected interface{}
	}{
		"Scheme uri tag": {
			uri: "https://localhost:8080/usr/bin",
			expected: &struct {
				Schema string `uri:"scheme"`
			}{Schema: "https"},
		},
		"Host uri tag": {
			uri: "https://localhost:8080/usr/bin",
			expected: &struct {
				Host string `uri:"host"`
			}{Host: "localhost:8080"},
		},
		"Path uri tag": {
			uri: "https://localhost:8080/usr/bin/file.txt",
			expected: &struct {
				Path string `uri:"path"`
				File string `uri:"filename"`
			}{
				Path: "/usr/bin/file.txt",
				File: "file.txt",
			},
		},
		"Authority uri tag": {
			uri: "https://localhost:8080/usr/bin",
			expected: &struct {
				Authority string `uri:"authority"`
			}{Authority: "https://localhost:8080"},
		},
		"Origin uri tag": {
			uri: "https://localhost:8080/usr/bin",
			expected: &struct {
				Origin string `uri:"Origin"`
			}{Origin: "https://localhost:8080/usr/bin"},
		},
		"Origin uri tag without authority": {
			uri: "/usr/bin",
			expected: &struct {
				Origin string `uri:"Origin"`
			}{Origin: "/usr/bin"},
		},
		"Custom int name": {
			uri: "?NewInt=10",
			expected: &struct {
				OldInt int `uri:"NewInt"`
			}{OldInt: 10},
		},
		"Var named Host without tag": {
			uri: "https://local/usr/bin?Host=hello",
			expected: &struct {
				Host string
			}{Host: "hello"},
		},
		"default tag for primitive types": {
			expected: &primitiveDefault{String: "hello", Bool: true, Int: 42, Float32: 12.34},
		},
		"override default tag for primitive types": {
			uri:      "?String=world&Bool=false&Int=0&Float32=0.1",
			expected: &primitiveDefault{String: "world", Bool: false, Int: 0, Float32: 0.1},
		},
		"default tag for slices": {
			expected: &sliceDefault{Strings: []string{"hello", "world"}, Ints: []int{11}},
		},
		"override default tag for slices": {
			uri:      "?Strings=test&Ints=1&Ints=2&Ints=3",
			expected: &sliceDefault{Strings: []string{"test"}, Ints: []int{1, 2, 3}},
		},
		"default tag unmarshalText struct": {
			expected: &struct {
				Time time.Time `default:"2018-01-01T00:00:00Z"`
			}{Time: trial.Time(time.RFC3339, "2018-01-01T00:00:00Z")},
		},
		"override default tag unmarshalText struct": {
			uri: "?Time=2017-04-24T12:00:00Z",
			expected: &struct {
				Time time.Time `default:"2018-01-01T00:00:00Z"`
			}{Time: trial.Time(time.RFC3339, "2017-04-24T12:00:00Z")},
		},
		"uri with fragment tag": {
			uri: "?Int=10#hello world",
			expected: &struct {
				Int     int
				Message string `uri:"fragment"`
			}{Int: 10, Message: "hello world"},
		},
	}
	for name, test := range cases {
		v := reflect.New(reflect.TypeOf(test.expected).Elem()).Interface()
		Unmarshal(test.uri, v)
		if !cmp.Equal(v, test.expected) {
			t.Errorf("FAIL: %v values did not match %s", name, cmp.Diff(v, test.expected))
		} else {
			t.Logf("PASS: %v", name)
		}
	}
}

func TestValidate(t *testing.T) {
	cases := map[string]struct {
		uri       string
		data      interface{}
		shouldErr bool
	}{
		"cannot write to struct": {
			data:      struct{}{},
			shouldErr: true,
		},
		"invalid uri": {
			uri:       "://",
			data:      &struct{}{},
			shouldErr: true,
		},
		"invalid default tag": {
			data: &struct {
				Value int `default:"abc"`
			}{},
			shouldErr: true,
		},
		"private variables": {
			uri: "?string=hello&int=1",
			data: &struct {
				int    int    `uri:"int"`
				String string `uri:"string"`
			}{},
		},
		"private variables with same name": {
			uri: "int=1",
			data: &struct {
				int int
				Int int `uri:"int"`
			}{Int: 1},
		},
		"required fields error is not provided": {
			uri: "",
			data: &struct {
				Int int `uri:"int" required:"true"`
			}{},
			shouldErr: true,
		},
		"required field ok when provided": {
			uri: "?int=10",
			data: &struct {
				Int int `uri:"int" required:"true"`
			}{Int: 10},
		},
		"required field ignored on default": {
			uri: "",
			data: &struct {
				Int int `uri:"int" required:"true" default:"10"`
			}{Int: 10},
		},
		"required fragment valid when provided": {
			uri: "?Int=10#name=hello",
			data: &struct {
				Int  int
				Name string `uri:"fragment" required:"true"`
			}{Int: 10, Name: "hello"},
		},
	}
	for name, test := range cases {
		err := Unmarshal(test.uri, test.data)
		if err != nil != test.shouldErr {
			t.Errorf("FAIL: %q", name)
		} else {
			t.Logf("PASS: %q data: %v", name, test.data)
		}
	}
}
