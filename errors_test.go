package gomo

import (
	"errors"
	"testing"
)

func withWrapTest(t *testing.T) {
	err := NewError("top error message", errors.New("Internal error message"))
	if err.Unwrap() == nil {
		t.Errorf("no wrapped error found")

	} else {
		t.Logf("top error is: %s", err)
		t.Logf("Internal error is: %s", err.Unwrap().Error())
	}
}

func noWrapTest(t *testing.T) {
	err := NewError("top error message", nil)
	if err.Unwrap() != nil {
		t.Errorf("internal error should have been nil")
	} else {
		t.Logf("top error: %s", err)
	}
}

func TestGomongoError(t *testing.T) {
	t.Run("wrap", withWrapTest)
	t.Run("no-wrap", noWrapTest)

}
