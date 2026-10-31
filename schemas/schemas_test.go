package schemas

import (
	"bytes"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func compile(t *testing.T, url string, data []byte) *jsonschema.Schema {
	t.Helper()
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("parse %s: %v", url, err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(url, doc); err != nil {
		t.Fatal(err)
	}
	s, err := c.Compile(url)
	if err != nil {
		t.Fatalf("compile %s: %v", url, err)
	}
	return s
}

func TestEngineResultSchema(t *testing.T) {
	s := compile(t, "https://github.com/glassops-platform/glassops/schemas/engine-result.json", EngineResult)
	cases := []struct {
		name string
		doc  string
		ok   bool
	}{
		{"minimal", `{"schema":"glassops.engine-result/v1","step":"deploy","passed":true}`, true},
		{"full", `{"schema":"glassops.engine-result/v1","step":"deploy","passed":false,"id":"d-1","components_deployed":3,"problems":["one"],"facts":{"exec.tool":"x"}}`, true},
		{"no passed", `{"schema":"glassops.engine-result/v1","step":"deploy"}`, false},
		{"unknown step", `{"schema":"glassops.engine-result/v1","step":"test","passed":true}`, false},
		{"fact outside exec", `{"schema":"glassops.engine-result/v1","step":"deploy","passed":true,"facts":{"target.id":"x"}}`, false},
		{"unknown key", `{"schema":"glassops.engine-result/v1","step":"deploy","passed":true,"extra":1}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := jsonschema.UnmarshalJSON(strings.NewReader(tc.doc))
			if err != nil {
				t.Fatal(err)
			}
			err = s.Validate(v)
			if (err == nil) != tc.ok {
				t.Fatalf("valid = %v, want %v: %v", err == nil, tc.ok, err)
			}
		})
	}
}
