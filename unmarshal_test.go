package uri

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/jbsmith7741/trial"
)

func ExampleUnmarshal() {
	v := struct {
		Scheme string `uri:"scheme"`
		Host   string `uri:"host"`
		Path   string `uri:"path"`
		Name   string `uri:"name"`
		Count  int    `uri:"num"`
	}{}
	s := "https://localhost/root/index.html?name=Hello+World&num=11"
	Unmarshal(s, &v)
	fmt.Printf("scheme:%s\nhost:%s\npath:%s\nname:%s\ncount:%d", v.Scheme, v.Host, v.Path, v.Name, v.Count)

	// Output: scheme:https
	// host:localhost
	// path:/root/index.html
	// name:Hello World
	// count:11
}

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
	Uint     uint
	Uint32   uint32
	Uint64   uint64
	Float32  float32
	Float32P *float32
	Float64  float64
	Float64P *float64
	Rune     rune  `format:"rune"`
	RuneP    *rune `format:"rune"`

	// slice
	Strings   []string
	Ints      []int
	IntsP     []*int
	Ints32    []int32
	Ints64    []int64
	Floats32  []float32
	Floats64  []float64
	TimeSlice []time.Time `format:"2006-01-02"`

	// maps
	MString map[string]string
	MInt    map[int]int
	MStrInt map[string]int
	MIntStr map[int]string
	MFloats map[string][]float64
	MTime   map[string]time.Time `format:"2006-01-02"`

	// struct
	Time       time.Time
	TimeP      *time.Time
	TimeF      time.Time  `format:"2006-01-02"`
	TimePF     *time.Time `format:"2006-01-02"`
	Unmarshal  unmarshalStruct
	UnmarshalP *unmarshalStruct
	Struct     bStruct
	StructP    *bStruct
	private    bStruct

	// alias
	Dessert dessert

	// special case
	Dura time.Duration
	Skip int `uri:"-"`
}

type unmarshalStruct struct {
	Data string
}

type bStruct struct {
	Name  string
	Value int
}

func (s *unmarshalStruct) UnmarshalText(text []byte) error {
	s.Data = string(text)
	return nil
}

func (s unmarshalStruct) MarshalText() ([]byte, error) {
	return []byte(s.Data), nil
}

func TestUnmarshal(t *testing.T) {
	fn := func(args ...interface{}) (interface{}, error) {
		var d interface{}
		if len(args) > 1 {
			d = args[1]
		} else {
			d = &testStruct{}
		}
		err := Unmarshal(args[0].(string), d)
		return d, err
	}
	cases := trial.Cases{
		"string": {
			Input:    "?String=hello",
			Expected: &testStruct{String: "hello"},
		},
		"integer: int, int32, int64": {
			Input:    "?Int=10&Int32=32&Int64=64",
			Expected: &testStruct{Int: 10, Int32: 32, Int64: 64},
		},
		"uint, uint32, uint64": {
			Input:    "?Uint=10&Uint32=32&Uint64=64",
			Expected: &testStruct{Uint: 10, Uint32: 32, Uint64: 64},
		},
		"runes": {
			Input:    "?Rune=%09&RuneP=%D1%8F",
			Expected: &testStruct{Rune: '\t', RuneP: trial.Int32P('я')},
		},
		"pointer: *int, *int32, *int64": {
			Input:    "?IntP=77&Int32P=11&Int64P=222",
			Expected: &testStruct{IntP: trial.IntP(77), Int32P: trial.Int32P(11), Int64P: trial.Int64P(222)},
		},
		"invalid integer": {
			Input:     "?Int=abc",
			ShouldErr: true,
		},
		"invalid uint": {
			Input:     "?Uint=abc",
			ShouldErr: true,
		},
		"duration as string": {
			Input:    "?Dura=10m",
			Expected: &testStruct{Dura: 10 * time.Minute},
		},
		"duration as integer": {
			Input:    "?Dura=1000000",
			Expected: &testStruct{Dura: time.Millisecond},
		},
		"float32, float64": {
			Input:    "?Float32=12.2&Float64=33.3",
			Expected: &testStruct{Float32: 12.2, Float64: 33.3},
		},
		"pointer: *float32, *float64": {
			Input:    "?Float32P=12.2&Float64P=33.3",
			Expected: &testStruct{Float32P: trial.Float32P(12.2), Float64P: trial.Float64P(33.3)},
		},
		"invalid float": {
			Input:     "?Float32=abc",
			ShouldErr: true,
		},
		"time.Time": {
			Input:    "?Time=2017-10-10T12:12:12Z",
			Expected: &testStruct{Time: trial.Time(time.RFC3339, "2017-10-10T12:12:12Z")},
		},
		"*time.Time": {
			Input:    "?TimeP=2017-10-10T12:12:12Z",
			Expected: &testStruct{TimeP: trial.TimeP(time.RFC3339, "2017-10-10T12:12:12Z")},
		},
		"(custom) time.Time": {
			Input:    "?TimeF=2017-10-10",
			Expected: &testStruct{TimeF: trial.TimeDay("2017-10-10")},
		},
		"(custom) time parse error": {
			Input:     "?TimeF=abcde",
			ShouldErr: true,
		},
		"(custom) *time.Time": {
			Input:    "?TimePF=2017-10-10",
			Expected: &testStruct{TimePF: trial.TimeP("2006-01-02", "2017-10-10")},
		},
		"(custom) []time.Time": {
			Input:    "?TimeSlice=2017-10-10,2018-11-11",
			Expected: &testStruct{TimeSlice: []time.Time{trial.TimeDay("2017-10-10"), trial.TimeDay("2018-11-11")}},
		},
		"invalid time": {
			Input:     "?Time=2017-10-",
			ShouldErr: true,
		},
		"struct with UnMarshalText": {
			Input: "?Unmarshal=abc&UnmarshalP=def",
			Expected: &testStruct{
				Unmarshal:  unmarshalStruct{Data: "abc"},
				UnmarshalP: &unmarshalStruct{Data: "def"},
			},
		},
		"default struct without MarshalText": {
			Input:    trial.Args("?", &testStruct{Struct: bStruct{Name: "hello", Value: 10}}),
			Expected: &testStruct{Struct: bStruct{Name: "hello", Value: 10}},
		},
		"default *struct without MarshalText": {
			Input:    trial.Args("?", &testStruct{StructP: &bStruct{Name: "hello", Value: 10}}),
			Expected: &testStruct{StructP: &bStruct{Name: "hello", Value: 10}},
		},
		"default private struct": {
			Input:    trial.Args("?", &testStruct{private: bStruct{Name: "hello", Value: 10}}),
			Expected: &testStruct{private: bStruct{Name: "hello", Value: 10}},
		},
		"bool": {
			Input: "?Bool=true",
			Expected: &testStruct{
				Bool: true,
			},
		},
		"bool implicit true": {
			Input: "?Bool&Test",
			Expected: &testStruct{
				Bool: true,
			},
		},
		"slice of string": {
			Input: "?Strings=a&Strings=b&Strings=c",
			Expected: &testStruct{
				Strings: []string{"a", "b", "c"},
			},
		},
		"string slice with ,": {
			Input: "?Strings=a,b,c",
			Expected: &testStruct{
				Strings: []string{"a", "b", "c"},
			},
		},
		"slice: int, int32, int64": {
			Input: "?Ints=1&Ints=2&Ints=3&Ints32=4,5,6&Ints64=7,8,9",
			Expected: &testStruct{
				Ints:   []int{1, 2, 3},
				Ints32: []int32{4, 5, 6},
				Ints64: []int64{7, 8, 9},
			},
		},
		"slice: float32, float64": {
			Input: "?Floats32=1.1&Floats32=2.2&Floats32=3.3&Floats64=4.4,5.5,6.6",
			Expected: &testStruct{
				Floats32: []float32{1.1, 2.2, 3.3},
				Floats64: []float64{4.4, 5.5, 6.6},
			},
		},
		"slice of *int": {
			Input: "?IntsP=1,2,3",
			Expected: &testStruct{
				IntsP: []*int{trial.IntP(1), trial.IntP(2), trial.IntP(3)},
			},
		},
		"slice of *int with nil": {
			Input: "?IntsP=1,2,nil,3",
			Expected: &testStruct{
				IntsP: []*int{trial.IntP(1), trial.IntP(2), nil, trial.IntP(3)},
			},
		},
		"alias type (dessert)": {
			Input:    "?Dessert=brownie",
			Expected: &testStruct{Dessert: brownie},
		},
		"invalid alias type": {
			Input:     "?Dessert=cat",
			ShouldErr: true,
		},
		"skip": {
			Input:    "?-=10",
			Expected: &testStruct{},
		},
		"maps": {
			Input: "?MString=fruit:apple|mammal:dog&MInt=1:2|3:4&MStrInt=fruit:1|dog:2&MIntStr=1:fruit&MIntStr=2:dog",
			Expected: &testStruct{
				MString: map[string]string{"fruit": "apple", "mammal": "dog"},
				MInt:    map[int]int{1: 2, 3: 4},
				MStrInt: map[string]int{"fruit": 1, "dog": 2},
				MIntStr: map[int]string{1: "fruit", 2: "dog"},
			},
		},
		"map_slice": {
			Input: "?MFloats=cat:1.2,2.3,3.4|dog:4.4,5.5|bat:0.0",
			Expected: &testStruct{
				MFloats: map[string][]float64{"cat": {1.2, 2.3, 3.4}, "dog": {4.4, 5.5}, "bat": {0.0}},
			},
		},
		"map_time": {
			Input: "?MTime=a:2020-01-02|b:2020-02-02",
			Expected: &testStruct{
				MTime: map[string]time.Time{
					"a": trial.TimeDay("2020-01-02"),
					"b": trial.TimeDay("2020-02-02"),
				},
			},
		},
		"map_blank": {
			Input:     "?MString",
			ShouldErr: true,
		},
		"map_invalid": {
			Input:     "?MInt=a:b&MStrInt=a:b",
			ShouldErr: true,
		},
	}
	trial.New(fn, cases).SubTest(t)
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
	type Embedded struct {
		Int    int
		String string
	}
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
		"host as a []string": {
			uri: "https://host1,host2",
			expected: &struct {
				Hosts []string `uri:"host"`
			}{Hosts: []string{"host1", "host2"}},
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
		"embedded struct": {
			uri: "?Int=10&String=hello",
			expected: &struct {
				Embedded
			}{
				Embedded: Embedded{Int: 10, String: "hello"},
			},
		},
		"embedded *struct": {
			uri: "?Int=10&String=hello",
			expected: &struct {
				*Embedded
			}{
				Embedded: &Embedded{Int: 10, String: "hello"},
			},
		},
	}
	for name, test := range cases {
		v := reflect.New(reflect.TypeOf(test.expected).Elem()).Interface()
		Unmarshal(test.uri, v)
		if equal, msg := trial.Equal(v, test.expected); !equal {
			t.Errorf("FAIL: %v values did not match %s", name, msg)
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
		"nil pointer": {
			data:      (*sliceDefault)(nil),
			shouldErr: true,
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
