package junit

import (
	"encoding/xml"
	"strings"
	"testing"
)

// The inputs below are synthetic.
func sample() Suites {
	var s Suites
	s.Name = "apex"
	s.Add(Suite{Name: "GreeterTest", Cases: []Case{
		{Name: "greetsByName", ClassName: "GreeterTest", Time: 0.033},
		{Name: "fails", ClassName: "GreeterTest", Time: 0.02, Failure: &Failure{
			Message: `Assertion <failed> & "quoted"`, Type: "Fail", Text: "Class.GreeterTest.fails: line 1",
		}},
		{Name: "later", ClassName: "GreeterTest", Skipped: &Skipped{Message: "skipped"}},
	}})
	return s
}

func TestRoundTrip(t *testing.T) {
	data, err := sample().Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "<?xml") {
		t.Fatalf("no declaration: %s", data)
	}
	var back Suites
	if err := xml.Unmarshal(data, &back); err != nil {
		t.Fatalf("does not parse: %v\n%s", err, data)
	}
	if back.Tests != 3 || back.Failures != 1 || back.Skipped != 1 || len(back.Suites) != 1 {
		t.Fatalf("totals: %+v", back)
	}
	cases := back.Suites[0].Cases
	if len(cases) != 3 || cases[1].Failure == nil || cases[1].Failure.Message != `Assertion <failed> & "quoted"` {
		t.Fatalf("cases: %+v", cases)
	}
	if cases[1].Failure.Text != "Class.GreeterTest.fails: line 1" || cases[0].Failure != nil || cases[2].Skipped == nil {
		t.Fatalf("cases: %+v", cases)
	}
	if back.Suites[0].Name != "GreeterTest" || back.Time < 0.05 || back.Time > 0.06 {
		t.Fatalf("suite: %+v", back)
	}
}

func TestEmptySuitesStillParse(t *testing.T) {
	data, err := (Suites{}).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	var back Suites
	if err := xml.Unmarshal(data, &back); err != nil || back.Tests != 0 {
		t.Fatalf("%v %+v", err, back)
	}
}

func TestAddRecomputesCounts(t *testing.T) {
	var s Suites
	s.Add(Suite{Name: "x", Tests: 99, Failures: 99, Cases: []Case{{Name: "a"}}})
	if s.Tests != 1 || s.Failures != 0 || s.Suites[0].Tests != 1 {
		t.Fatalf("%+v", s)
	}
}
