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
	s := GetFieldString(reflect.ValueOf(name), "")
	fmt.Println(s)

	var i *int
	s = GetFieldString(reflect.ValueOf(i), "")
	fmt.Println(s)

	v := []int{2, 1, 3}
	s = GetFieldString(reflect.ValueOf(v), "")
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
		return MarshalUnescaped(args[0]), nil
	}
	cases := trial.Cases{
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
				R1 rune `uri:"r1" format:"rune"`
				R2 rune `uri:"r2" format:"rune"`
				R3 rune `uri:"r3" format:"rune"`
			}{
				R1: '\t',
				R2: 8984,
				R3: 'я',
			},
			Expected: "?r1=\t&r2=⌘&r3=я",
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
				Uint   uint
				String string
			}{Int: 10, Uint: 20, String: "hello"},
			Expected: "?Int=10&String=hello&Uint=20",
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
			Expected: "?struct=Input&time=2018-04-04T00:00:00Z",
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
			Expected: "?Int=10&String=brown fox",
		},
		"embedded *struct (ptr)": {
			Input: struct {
				*Embedded
			}{Embedded: &Embedded{Int: 10, String: "brown fox"}},
			Expected: "?Int=10&String=brown fox",
		},
		"fragment": {
			Input: struct {
				Int   int
				Value string `uri:"fragment"`
			}{Int: 11, Value: "this is a fragment"},
			Expected: "?Int=11#this is a fragment",
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
				Authority: "http://user:pass@localhost:8080/path/to/file.txt",
			},
			Expected: "//user:pass@localhost:8080",
		},
		"username_password": {
			Input: struct {
				Password string `uri:"password"`
				Username string `uri:"username"`
			}{
				Username: "user",
				Password: "pass",
			},
			Expected: "//user:pass@",
		},
		"userinfo": {
			Input: struct {
				UserInfo string `uri:"userinfo"`
			}{
				UserInfo: "userinfo:asd",
			},
			Expected: "//userinfo:asd@",
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
		"skip fields": {
			Input: struct {
				Skip   int `uri:"-"`
				String string
			}{Skip: 10, String: "apple"},
			Expected: "?String=apple",
		},
		"map[string]string": {
			Input: struct {
				Map map[string]string `uri:"map"`
			}{Map: map[string]string{"fruit": "apple"}},
			Expected: "?map=fruit:apple", // fruit:apple
		},
		"map[int]string": {
			Input: struct {
				Map map[int]string `uri:"map"`
			}{Map: map[int]string{1: "apple", 2: "banana"}},
			Expected: "?map=1:apple|2:banana",
		},
		"map[int][]string]": {
			Input: struct {
				Map map[int][]string `uri:"map"`
			}{Map: map[int][]string{0: {"apple", "banana"}}},
			Expected: "?map=0:apple,banana",
		},
		"map[string]time.Time": {
			Input: struct {
				Map map[string]time.Time `uri:"map" format:"2006-01-02"`
			}{Map: map[string]time.Time{"a": trial.TimeDay("2020-01-01")}},
			Expected: "?map=a:2020-01-01",
		},
		"map(nil)": {
			Input: struct {
				Map map[string]string `uri:"map"`
			}{Map: nil},
			Expected: "",
		},
		"json simple": {
			Input: struct {
				String string `json:"apple"`
			}{String: "Fuji"},
			Expected: "?apple=Fuji",
		},
		"json with uri": {
			Input: struct {
				String string `json:"apple" uri:"fruit"`
			}{String: "Fuji"},
			Expected: "?fruit=Fuji",
		},
		"json ignore": {
			Input: struct {
				String string `json:"-" uri:"fruit"`
			}{String: "Fuji"},
			Expected: "?fruit=Fuji",
		},
		"json uri ignore": {
			Input: struct {
				String string `json:"apple" uri:"-"`
			}{String: "Fuji"},
			Expected: "",
		},
		"json omitempty": {
			Input: struct {
				String string `json:"apple,omitempty"`
			}{String: "Fuji"},
			Expected: "?apple=Fuji",
		},
		"json blank": {
			Input: struct {
				String string `json:",omitempty"`
			}{String: "Fuji"},
			Expected: "?String=Fuji",
		},
		"private": {
			Input: struct {
				Int  int
				name string
			}{Int: 7, name: "hello"},
			Expected: "?Int=7",
		},
	}
	trial.New(fn, cases).SubTest(t)
}
