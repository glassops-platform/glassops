package telemetry

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/pipeline"
	"github.com/glassops-platform/glassops/internal/core/run"
)

// syntheticRun is a run record built here, not captured from a real run. The shape follows
// what internal/cli/run.go writes. [unverified against a real run.json]
func syntheticRun() *run.Result {
	t0 := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	rec := run.New("run", func() time.Time { return t0 })
	rec.Ended = t0.Add(9 * time.Second)
	rec.Provider = "salesforce"
	zero, one := 0, 1
	rec.Facts["run.pipeline"] = "pr"
	rec.Facts["run.exit_code"] = 1
	rec.Facts["run.empty_plan"] = false
	rec.Facts["run.cancelled"] = false
	// Values that must never be exported.
	rec.Facts["salesforce.username"] = "someone@example.com"
	rec.Facts["org.id"] = "00D000000000001AAA"
	rec.Errors = []string{"secret failure text"}
	rec.Facts["run.steps"] = []pipeline.Record{
		{Index: 1, Step: "plan", When: "success", Argv: []string{"--target", "prod", "plan"}, ExitCode: &zero, Started: t0, DurationMS: 2000},
		{Index: 2, Step: "gate", When: "success", Argv: []string{"gate", "--target=prod", "--login", "someone@example.com"}, ExitCode: &one, Started: t0.Add(2 * time.Second), DurationMS: 1500},
		{Index: 3, Step: "deploy", When: "success", Skipped: true, SkipReason: "an earlier step failed"},
	}
	return rec
}

func testConfig(endpoint string) *config.Config {
	return &config.Config{
		Targets: map[string]config.Target{"prod": {Tier: "prod"}},
		Telemetry: config.Telemetry{OTLP: config.OTLP{
			Endpoint: endpoint,
			Headers:  map[string]string{"Authorization": "$OTEL_TOKEN"},
		}},
	}
}

type capture struct {
	mu     sync.Mutex
	hits   int
	path   string
	header http.Header
	body   []byte
}

func collector(t *testing.T, code int) (*httptest.Server, *capture) {
	t.Helper()
	c := &capture{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		c.mu.Lock()
		c.hits++
		c.path, c.header, c.body = r.URL.Path, r.Header.Clone(), b
		c.mu.Unlock()
		w.WriteHeader(code)
	}))
	t.Cleanup(srv.Close)
	return srv, c
}

func fakeEnv(t *testing.T, env map[string]string) {
	t.Helper()
	old := getenv
	getenv = func(k string) string { return env[k] }
	t.Cleanup(func() { getenv = old })
}

func TestExportIsOptIn(t *testing.T) {
	rec := run.New("run", nil)
	if err := Export(context.Background(), nil, rec); err != nil {
		t.Fatalf("nil config: %v", err)
	}
	if err := Export(context.Background(), &config.Config{}, rec); err != nil {
		t.Fatalf("no endpoint: %v", err)
	}
}

type decoded struct {
	ResourceSpans []struct {
		Resource struct {
			Attributes []map[string]any `json:"attributes"`
		} `json:"resource"`
		ScopeSpans []struct {
			Spans []map[string]any `json:"spans"`
		} `json:"scopeSpans"`
	} `json:"resourceSpans"`
}

func attrs(sp map[string]any) map[string]any {
	out := map[string]any{}
	for _, a := range sp["attributes"].([]any) {
		kv := a.(map[string]any)
		for _, v := range kv["value"].(map[string]any) {
			out[kv["key"].(string)] = v
		}
	}
	return out
}

func TestExportSendsOneTracePerRunWithAllowlistedAttributes(t *testing.T) {
	srv, c := collector(t, 200)
	fakeEnv(t, map[string]string{"OTEL_TOKEN": "tok-123"})
	cfg := testConfig(srv.URL)
	cfg.Telemetry.OTLP.ServiceName = "my-service"

	if err := Export(context.Background(), cfg, syntheticRun()); err != nil {
		t.Fatalf("export: %v", err)
	}
	if c.hits != 1 || c.path != "/v1/traces" {
		t.Fatalf("hits %d path %q", c.hits, c.path)
	}
	if got := c.header.Get("Authorization"); got != "tok-123" {
		t.Fatalf("header from $VAR = %q", got)
	}
	if ct := c.header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content type %q", ct)
	}
	var d decoded
	if err := json.Unmarshal(c.body, &d); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	rs := d.ResourceSpans[0]
	if rs.Resource.Attributes[0]["key"] != "service.name" {
		t.Fatalf("resource attrs %v", rs.Resource.Attributes)
	}
	spans := rs.ScopeSpans[0].Spans
	if len(spans) != 4 {
		t.Fatalf("want a root plus 3 step spans, got %d", len(spans))
	}
	trace := spans[0]["traceId"]
	if len(trace.(string)) != 32 {
		t.Fatalf("trace id %v", trace)
	}
	for _, sp := range spans[1:] {
		if sp["traceId"] != trace || sp["parentSpanID"] != nil || sp["parentSpanId"] != spans[0]["spanId"] {
			t.Fatalf("step span not a child of the root in one trace: %v", sp)
		}
	}
	gate := attrs(spans[2])
	if gate["glassops.step"] != "gate" || gate["glassops.target.name"] != "prod" ||
		gate["glassops.target.tier"] != "prod" || gate["glassops.step.outcome"] != "fail" || gate["glassops.exit_code"] != "1" {
		t.Fatalf("gate attrs %v", gate)
	}
	if spans[2]["status"].(map[string]any)["code"] != float64(2) {
		t.Fatalf("a failed step is not an error status: %v", spans[2]["status"])
	}
	if got := attrs(spans[1])["glassops.target.name"]; got != "prod" {
		t.Fatalf("--target X form: %v", got)
	}
	deploy := attrs(spans[3])
	if deploy["glassops.step.outcome"] != "skipped" || deploy["glassops.step.skipped"] != true {
		t.Fatalf("skipped attrs %v", deploy)
	}
	if spans[1]["startTimeUnixNano"] != "1793448000000000000" || spans[1]["endTimeUnixNano"] != "1793448002000000000" {
		t.Fatalf("times %v %v", spans[1]["startTimeUnixNano"], spans[1]["endTimeUnixNano"])
	}
	// Nothing outside the allowlist, and none of the planted values, left the process.
	body := string(c.body)
	for _, bad := range []string{"someone@example.com", "00D000000000001AAA", "secret failure text", "tok-123", "an earlier step failed", "--login"} {
		if strings.Contains(body, bad) {
			t.Errorf("body contains %q", bad)
		}
	}
	allowed := map[string]bool{}
	for _, k := range []string{attrPipeline, attrProvider, attrExit, attrEmptyPlan, attrCancelled, attrStep, attrIndex,
		attrWhen, attrOutcome, attrSkipped, attrDuration, attrTargetName, attrTargetTier} {
		allowed[k] = true
	}
	for _, sp := range spans {
		for k := range attrs(sp) {
			if !allowed[k] {
				t.Errorf("attribute %q is not on the allowlist", k)
			}
		}
	}
	if rs.Resource.Attributes[0]["value"].(map[string]any)["stringValue"] != "my-service" {
		t.Fatalf("service name not used")
	}
}

func TestUnknownTargetIsNotExported(t *testing.T) {
	srv, c := collector(t, 200)
	fakeEnv(t, map[string]string{"OTEL_TOKEN": "x"})
	rec := syntheticRun()
	rec.Facts["run.steps"] = []pipeline.Record{{Step: "gate", Argv: []string{"--target", "someone@example.com"}}}
	if err := Export(context.Background(), testConfig(srv.URL), rec); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(c.body), "someone@example.com") || strings.Contains(string(c.body), "glassops.target") {
		t.Fatalf("an unconfigured target leaked: %s", c.body)
	}
}

func TestDefaultServiceName(t *testing.T) {
	srv, c := collector(t, 200)
	fakeEnv(t, map[string]string{"OTEL_TOKEN": "x"})
	if err := Export(context.Background(), testConfig(srv.URL), syntheticRun()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(c.body), `"stringValue":"glassops"`) {
		t.Fatalf("default service name missing: %s", c.body)
	}
}

func TestEndpointPathKept(t *testing.T) {
	srv, c := collector(t, 200)
	fakeEnv(t, map[string]string{"OTEL_TOKEN": "x"})
	if err := Export(context.Background(), testConfig(srv.URL+"/custom/traces"), syntheticRun()); err != nil {
		t.Fatal(err)
	}
	if c.path != "/custom/traces" {
		t.Fatalf("path %q", c.path)
	}
}

func TestSendFailuresReturnErrorsWithoutSecrets(t *testing.T) {
	fakeEnv(t, map[string]string{"OTEL_TOKEN": "tok-secret"})
	srv, _ := collector(t, 500)
	err := Export(context.Background(), testConfig(srv.URL), syntheticRun())
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("a 500 must be an error: %v", err)
	}
	// A dead collector.
	dead, _ := collector(t, 200)
	url := dead.URL
	dead.Close()
	err = Export(context.Background(), testConfig(url), syntheticRun())
	if err == nil {
		t.Fatal("a refused connection must be an error")
	}
	if strings.Contains(err.Error(), "tok-secret") || strings.Contains(err.Error(), url) {
		t.Fatalf("error leaks the token or URL: %v", err)
	}
}

func TestHeaderProblems(t *testing.T) {
	srv, c := collector(t, 200)
	cfg := testConfig(srv.URL)

	fakeEnv(t, map[string]string{})
	err := Export(context.Background(), cfg, syntheticRun())
	if err == nil || !strings.Contains(err.Error(), "OTEL_TOKEN") || c.hits != 0 {
		t.Fatalf("unset variable: %v hits %d", err, c.hits)
	}

	cfg.Telemetry.OTLP.Headers = map[string]string{"Authorization": "Bearer literal"}
	err = Export(context.Background(), cfg, syntheticRun())
	if err == nil || strings.Contains(err.Error(), "literal") || c.hits != 0 {
		t.Fatalf("literal header: %v hits %d", err, c.hits)
	}
}

func TestEndpointRules(t *testing.T) {
	fakeEnv(t, map[string]string{"OTEL_TOKEN": "x"})
	cases := []struct {
		name, endpoint string
		headers        bool
		wantErr        bool
	}{
		{"https with headers", "https://otel.example.com", true, false},
		{"plain http loopback with headers", "http://127.0.0.1:4318", true, false},
		{"plain http remote with headers", "http://otel.example.com", true, true},
		{"plain http remote no headers", "http://otel.example.com", false, false},
		{"credentials in the URL", "https://user:pw@otel.example.com", false, true},
		{"not http", "grpc://otel.example.com", false, true},
		{"no host", "https://", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveEndpoint(tc.endpoint, tc.headers)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil && strings.Contains(err.Error(), "pw") {
				t.Fatalf("error leaks the password: %v", err)
			}
			if err == nil && !strings.HasSuffix(got, "/v1/traces") {
				t.Fatalf("path not added: %s", got)
			}
		})
	}
}

func TestContextCancelStopsSend(t *testing.T) {
	fakeEnv(t, map[string]string{"OTEL_TOKEN": "x"})
	srv, _ := collector(t, 200)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Export(ctx, testConfig(srv.URL), syntheticRun()); err == nil {
		t.Fatal("a cancelled context must fail the send")
	}
}

func TestJSONRecordsAndTypedRecordsAgree(t *testing.T) {
	typed := syntheticRun()
	b, _ := json.Marshal(typed.Facts)
	var generic map[string]any
	if err := json.Unmarshal(b, &generic); err != nil {
		t.Fatal(err)
	}
	if a, g := len(stepsOf(typed.Facts)), len(stepsOf(generic)); a != 3 || g != 3 {
		t.Fatalf("typed %d decoded %d", a, g)
	}
}
