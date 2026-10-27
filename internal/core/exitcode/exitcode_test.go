package exitcode

import (
	"errors"
	"testing"
)

func TestWorst(t *testing.T) {
	cases := []struct {
		codes []int
		want  int
	}{
		{nil, OK},
		{[]int{OK, OK}, OK},
		{[]int{OK, Fail}, Fail},
		{[]int{Error, Fail, OK}, Error},
	}
	for _, tc := range cases {
		if got := Worst(tc.codes...); got != tc.want {
			t.Errorf("Worst(%v) = %d, want %d", tc.codes, got, tc.want)
		}
	}
}

func TestExitErrorUnwraps(t *testing.T) {
	inner := errors.New("inner")
	err := &ExitError{Code: Fail, Err: inner}
	if !errors.Is(err, inner) {
		t.Fatal("ExitError does not unwrap")
	}
	if err.Error() != "inner" {
		t.Fatalf("Error() = %q", err.Error())
	}
}
