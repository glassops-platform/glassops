package config

import (
	"strings"
	"testing"
)

// A hostile glassops.yaml must be a config error, never a crash of the process that loads it.
func TestParseRefusesAnAliasThatRefersToItself(t *testing.T) {
	for _, in := range []string{"&a [*a]\n", "a: &x {b: *x}\n", "&a [&b [*a]]\n"} {
		if _, err := Parse([]byte(in)); err == nil {
			t.Errorf("%q: a cyclic config must be refused", in)
		}
	}
}

func TestParseRefusesAliasesThatExpandWithoutLimit(t *testing.T) {
	var b strings.Builder
	b.WriteString("a0: &a0 [x, x, x, x, x, x, x, x, x, x]\n")
	for i := 1; i <= 9; i++ {
		b.WriteString("a" + string(rune('0'+i)) + ": &a" + string(rune('0'+i)) + " [")
		for j := 0; j < 10; j++ {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteString("*a" + string(rune('0'+i-1)))
		}
		b.WriteString("]\n")
	}
	if _, err := Parse([]byte(b.String())); err == nil {
		t.Fatal("ten levels of tenfold aliases must be refused")
	}
}

func TestParseRefusesDeepNesting(t *testing.T) {
	in := strings.Repeat("[", 200) + strings.Repeat("]", 200) + "\n"
	if _, err := Parse([]byte(in)); err == nil {
		t.Fatal("200 levels of nesting must be refused")
	}
}
