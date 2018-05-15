package uri

import (
	"fmt"
)

type dessert int

const (
	pie dessert = iota
	icecream
	cake
	brownie
)

var dessertArr = [...]string{"pie", "icecream", "cake", "brownie"}

func (d dessert) String() string {
	if int(d) < len(dessertArr) {
		return dessertArr[d]
	}
	return ""
}

func (d *dessert) UnmarshalText(b []byte) error {
	for i, v := range dessertArr {
		if v == string(b) {
			*d = dessert(i)
			return nil
		}
	}
	return fmt.Errorf("Unknown dessert %s", string(b))
}
