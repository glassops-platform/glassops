package facts

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// roundTrip writes d and reads it back, so values take the shapes JSON decoding gives them.
func roundTrip(t *testing.T, d *Document) *Document {
	t.Helper()
	path := filepath.Join(t.TempDir(), "nested", "facts.json")
	if err := Write(path, d); err != nil {
		t.Fatalf("Write: %v", err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	return got
}

func TestNewSetGet(t *testing.T) {
	d := New()
	if d.Schema != Schema {
		t.Fatalf("schema = %q", d.Schema)
	}
	if _, ok := d.Get(TargetName); ok {
		t.Fatal("empty document reports a value")
	}
	d.Set(TargetName, "prod")
	if v, ok := d.Get(TargetName); !ok || v != "prod" {
		t.Fatalf("Get = %v, %v", v, ok)
	}
	var zero Document
	zero.Set("a", 1)
	if v, _ := zero.Get("a"); v != 1 {
		t.Fatal("Set on a zero document did not store the value")
	}
}

func TestString(t *testing.T) {
	d := New()
	d.Set(CIActor, "octocat")
	d.Set("n", 3)
	for _, doc := range []*Document{d, roundTrip(t, d)} {
		if v, ok := doc.String(CIActor); !ok || v != "octocat" {
			t.Fatalf("String = %q, %v", v, ok)
		}
		if _, ok := doc.String("n"); ok {
			t.Fatal("String accepted a number")
		}
		if _, ok := doc.String("missing"); ok {
			t.Fatal("String accepted a missing key")
		}
	}
}

func TestInt(t *testing.T) {
	cases := []struct {
		name string
		v    any
		want int
		ok   bool
	}{
		{"int", 7, 7, true},
		{"int64", int64(-3), -3, true},
		{"whole float", float64(12), 12, true},
		{"fraction", 1.5, 0, false},
		{"too large float", 1e30, 0, false},
		{"string", "7", 0, false},
		{"bool", true, 0, false},
		{"missing", nil, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := New()
			if tc.v != nil {
				d.Set("k", tc.v)
			}
			got, ok := d.Int("k")
			if got != tc.want || ok != tc.ok {
				t.Fatalf("Int = %d, %v, want %d, %v", got, ok, tc.want, tc.ok)
			}
		})
	}
	d := New()
	d.Set(ChangeDestructiveCount, 2)
	if got, ok := roundTrip(t, d).Int(ChangeDestructiveCount); !ok || got != 2 {
		t.Fatalf("Int after JSON = %d, %v", got, ok)
	}
}

func TestBool(t *testing.T) {
	d := New()
	d.Set(ChangeEmpty, true)
	d.Set("s", "true")
	for _, doc := range []*Document{d, roundTrip(t, d)} {
		if v, ok := doc.Bool(ChangeEmpty); !ok || !v {
			t.Fatalf("Bool = %v, %v", v, ok)
		}
		if _, ok := doc.Bool("s"); ok {
			t.Fatal("Bool accepted a string")
		}
	}
}

func TestStrings(t *testing.T) {
	d := New()
	d.Set(CILabels, []string{"a", "b"})
	d.Set("mixed", []any{"a", 1})
	d.Set("empty", []string{})
	for _, doc := range []*Document{d, roundTrip(t, d)} {
		if v, ok := doc.Strings(CILabels); !ok || !reflect.DeepEqual(v, []string{"a", "b"}) {
			t.Fatalf("Strings = %v, %v", v, ok)
		}
		if _, ok := doc.Strings("mixed"); ok {
			t.Fatal("Strings accepted a list with a number")
		}
		if v, ok := doc.Strings("empty"); !ok || len(v) != 0 {
			t.Fatalf("empty Strings = %v, %v", v, ok)
		}
	}
}

func TestApprovals(t *testing.T) {
	at := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	want := []Approval{
		{Mechanism: "environment", Approver: "octocat", State: "approved", Environment: "glassops-deploy", At: at, Comment: "ok"},
		{Mechanism: "label", Approver: "hubot", Label: "approved-destructive", At: at},
	}
	d := New()
	if got := d.Approvals(); got != nil {
		t.Fatalf("missing approvals = %v", got)
	}
	d.Set(CIApprovals, want)
	if got := d.Approvals(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Approvals = %+v", got)
	}
	if got := roundTrip(t, d).Approvals(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Approvals after JSON = %+v", got)
	}
	d.Set(CIApprovals, []any{"not an approval"})
	if got := d.Approvals(); got != nil {
		t.Fatalf("malformed approvals = %v", got)
	}
}

func TestMerge(t *testing.T) {
	src := map[string]any{"salesforce.org_type": "scratch", "salesforce.api": "62.0", "validate.passed": true}
	d := New()
	d.Merge("salesforce.", src)
	if len(d.Facts) != 2 {
		t.Fatalf("prefix merge = %v", d.Facts)
	}
	d.Merge("", src)
	if len(d.Facts) != 3 {
		t.Fatalf("full merge = %v", d.Facts)
	}
}

func TestReadChecksSchema(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte(`{"schema":"other/v1","facts":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(bad); err == nil || !strings.Contains(err.Error(), Schema) {
		t.Fatalf("wrong schema: %v", err)
	}
	if _, err := Read(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("missing file read without error")
	}
	noFacts := filepath.Join(dir, "nofacts.json")
	if err := os.WriteFile(noFacts, []byte(`{"schema":"`+Schema+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := Read(noFacts)
	if err != nil || d.Facts == nil {
		t.Fatalf("no facts: %v, %v", d, err)
	}
}
