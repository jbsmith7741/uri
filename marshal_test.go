package uri

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jbsmith7741/go-tools/trial"
)

func TestMarshal(t *testing.T) {
	cases := map[string]struct {
		data     interface{}
		expected string
	}{
		"default values": {
			data: struct {
				Int    int
				Amount float64 `uri:"float" default:"1.1"`
				Slice  []int   `default:"1,2,3"`
			}{
				Amount: 1.1,
				Slice:  []int{1, 2, 3},
			},
		},
		"slices": {
			data: struct {
				Ints    []int
				Nil     []int
				Strings []string `uri:"strings"`
			}{
				Ints:    []int{1, 2, 3},
				Strings: []string{"hello", "world"},
			},
			expected: "?Ints=1&Ints=2&Ints=3&strings=hello&strings=world",
		},
		/* todo how to handle this case?
		"empty slice with default value": {
			data: struct {
				Floats []float64 `uri:"float" default:"3.14,2.7,7.7"`
			}{},
			expected: "?float=[]",
		},*/
		"pointers": {
			data: struct {
				Int     *int
				Nil     *int
				Default *int `default:"1"`
			}{
				Int: trial.IntP(10),
			},
			expected: "?Default=nil&Int=10",
		},
		"structs": {
			data: struct {
				Time   time.Time       `uri:"time"`
				Struct unmarshalStruct `uri:"struct"`
			}{
				Time:   trial.Time(time.RFC3339, "2018-04-04T00:00:00Z"),
				Struct: unmarshalStruct{Data: "data"},
			},
			expected: "?struct=data&time=2018-04-04T00%3A00%3A00Z",
		},
		"bools": {
			data: struct {
				BoolT bool
				BoolF bool `default:"true"`
			}{true, false},
			expected: "?BoolF=false&BoolT=true",
		},
	}
	for msg, test := range cases {
		s := Marshal(test.data)
		if !cmp.Equal(s, test.expected) {
			t.Errorf("FAIL: %q %s", msg, cmp.Diff(s, test.expected))
		} else {
			t.Logf("PASS: %q", msg)
		}
	}
}
