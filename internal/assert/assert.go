package assert

import (
	"reflect"
	"testing"
)

// ErrorExists asserts that the error exists or not
func ErrorExists(t *testing.T, act error, exp bool) {
	if act != nil && !exp {
		t.Errorf("got %v, expected nil", act)
	}

	if act == nil && exp {
		t.Error("got nil, expected an error")
	}
}

// DeepEqual asserts that the two arguments are deep equal
func DeepEqual(t *testing.T, act, exp interface{}) {
	if !reflect.DeepEqual(act, exp) {
		t.Errorf("got %v, expected %v", act, exp)
	}
}
