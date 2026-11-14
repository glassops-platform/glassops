// Package telemetry exports a finished run as trace spans, when telemetry.otlp is configured.
// Export is opt-in, and a failed export never changes a run's exit code.
//
// The export is OTLP over HTTP with the JSON encoding, written by hand with net/http, so the
// module gains no dependency. One run is one trace: a root span for the run and one child span
// per step. Only an allowlist of attributes is ever sent (see the constants below). Facts
// values, logins, org Ids and error text are never copied into a span.
package telemetry

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/run"
)

// The attribute allowlist. Anything not named here is not exported.
const (
	attrPipeline   = "glassops.pipeline"
	attrProvider   = "glassops.provider"
	attrExit       = "glassops.exit_code"
	attrEmptyPlan  = "glassops.empty_plan"
	attrCancelled  = "glassops.cancelled"
	attrStep       = "glassops.step"
	attrIndex      = "glassops.step.index"
	attrWhen       = "glassops.step.when"
	attrOutcome    = "glassops.step.outcome"
	attrSkipped    = "glassops.step.skipped"
	attrDuration   = "glassops.duration_ms"
	attrTargetName = "glassops.target.name"
	attrTargetTier = "glassops.target.tier"
)

// tracesPath is appended to the endpoint when it has no path of its own, as the OTLP/HTTP
// specification does for a base endpoint.
const tracesPath = "/v1/traces"

// sendTimeout bounds one export so a dead collector cannot hold a finished run.
const sendTimeout = 10 * time.Second

// envRef is a header value: a whole-value $VAR reference, the same form config validates.
var envRef = regexp.MustCompile(`^\$([A-Z_][A-Z0-9_]*)$`)

// Seams replaced in tests.
var (
	getenv     = os.Getenv
	httpClient = &http.Client{}
	newTraceID = func() string { return randomHex(16) }
	newSpanID  = func() string { return randomHex(8) }
)

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// Not reached on supported platforms. A zero id is still valid JSON.
		return strings.Repeat("0", n*2)
	}
	return hex.EncodeToString(b)
}

// Export sends rec, the run step's result, to the endpoint in cfg.Telemetry as one trace.
// With no endpoint configured it does nothing and returns nil. The error never carries a
// header value or the endpoint's credentials.
func Export(ctx context.Context, cfg *config.Config, rec *run.Result) error {
	if cfg == nil || !cfg.Telemetry.Enabled() {
		return nil
	}
	if rec == nil {
		return errors.New("no run record to export")
	}
	o := cfg.Telemetry.OTLP
	endpoint, err := resolveEndpoint(o.Endpoint, len(o.Headers) > 0)
	if err != nil {
		return err
	}
	headers, err := resolveHeaders(o.Headers)
	if err != nil {
		return err
	}
	body, err := json.Marshal(buildRequest(cfg, rec))
	if err != nil {
		return fmt.Errorf("encode spans: %w", err)
	}
	return post(ctx, endpoint, headers, body)
}

// resolveEndpoint validates the endpoint and adds the traces path when it has none. Headers
// are only sent over https, except to a loopback host, so a token never crosses a network
// in clear text.
func resolveEndpoint(raw string, hasHeaders bool) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", errors.New("telemetry.otlp.endpoint is not a valid URL")
	}
	switch u.Scheme {
	case "https":
	case "http":
		if hasHeaders && !isLoopback(u.Hostname()) {
			return "", errors.New("telemetry.otlp.headers are not sent over plain http to a non-loopback host")
		}
	default:
		return "", errors.New("telemetry.otlp.endpoint must be an http or https URL")
	}
	if u.User != nil {
		return "", errors.New("telemetry.otlp.endpoint must not carry credentials, use headers")
	}
	if u.Path == "" || u.Path == "/" {
		u.Path = tracesPath
	}
	return u.String(), nil
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// resolveHeaders reads each $VAR value from the environment. A header that is not a $VAR
// reference or whose variable is unset is an error, so a missing credential is visible. The
// error names the header and the variable, never a value.
func resolveHeaders(h map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(h))
	names := make([]string, 0, len(h))
	for k := range h {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		m := envRef.FindStringSubmatch(h[k])
		if m == nil {
			return nil, fmt.Errorf("telemetry.otlp.headers.%s must be a $VAR reference", k)
		}
		v := getenv(m[1])
		if v == "" {
			return nil, fmt.Errorf("telemetry.otlp.headers.%s: environment variable %s is not set", k, m[1])
		}
		out[k] = v
	}
	return out, nil
}

func post(ctx context.Context, endpoint string, headers map[string]string, body []byte) error {
	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.New("build the export request failed")
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		// A *url.Error repeats the URL. Report only the cause.
		var ue *url.Error
		if errors.As(err, &ue) {
			err = ue.Err
		}
		return fmt.Errorf("send spans: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("collector answered HTTP %d", resp.StatusCode)
	}
	return nil
}

// The OTLP/HTTP JSON shapes, per the OpenTelemetry protocol specification. 64-bit integers
// are strings in the JSON encoding and ids are hex strings. [unverified against a collector
// until a probe run]

type otlpRequest struct {
	ResourceSpans []resourceSpans `json:"resourceSpans"`
}

type resourceSpans struct {
	Resource   resource     `json:"resource"`
	ScopeSpans []scopeSpans `json:"scopeSpans"`
}

type resource struct {
	Attributes []keyValue `json:"attributes"`
}

type scopeSpans struct {
	Scope scope  `json:"scope"`
	Spans []span `json:"spans"`
}

type scope struct {
	Name string `json:"name"`
}

type span struct {
	TraceID           string     `json:"traceId"`
	SpanID            string     `json:"spanId"`
	ParentSpanID      string     `json:"parentSpanId,omitempty"`
	Name              string     `json:"name"`
	Kind              int        `json:"kind"`
	StartTimeUnixNano string     `json:"startTimeUnixNano"`
	EndTimeUnixNano   string     `json:"endTimeUnixNano"`
	Attributes        []keyValue `json:"attributes"`
	Status            status     `json:"status"`
}

type status struct {
	Code int `json:"code"`
}

type keyValue struct {
	Key   string   `json:"key"`
	Value anyValue `json:"value"`
}

type anyValue struct {
	StringValue *string `json:"stringValue,omitempty"`
	IntValue    *string `json:"intValue,omitempty"`
	BoolValue   *bool   `json:"boolValue,omitempty"`
}

func str(k, v string) keyValue { return keyValue{k, anyValue{StringValue: &v}} }
func num(k string, v int64) keyValue {
	s := strconv.FormatInt(v, 10)
	return keyValue{k, anyValue{IntValue: &s}}
}
func flag(k string, v bool) keyValue { return keyValue{k, anyValue{BoolValue: &v}} }

const (
	kindInternal = 1
	statusUnset  = 0
	statusError  = 2
)

// stepRecord is the part of pipeline.Record the export reads. It is decoded from JSON so
// this package does not import the pipeline.
type stepRecord struct {
	Index      int       `json:"index"`
	Step       string    `json:"step"`
	When       string    `json:"when"`
	Argv       []string  `json:"argv"`
	ExitCode   *int      `json:"exit_code"`
	Started    time.Time `json:"started"`
	DurationMS int64     `json:"duration_ms"`
	Skipped    bool      `json:"skipped"`
}

// stepsOf reads the run.steps fact, whether it holds typed records or decoded JSON.
func stepsOf(facts map[string]any) []stepRecord {
	raw, ok := facts["run.steps"]
	if !ok {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var recs []stepRecord
	if json.Unmarshal(b, &recs) != nil {
		return nil
	}
	return recs
}

func buildRequest(cfg *config.Config, rec *run.Result) otlpRequest {
	traceID := newTraceID()
	rootID := newSpanID()
	steps := stepsOf(rec.Facts)

	runStart, runEnd := rec.Started, rec.Ended
	if runEnd.Before(runStart) {
		runEnd = runStart
	}
	spans := make([]span, 0, len(steps)+1)

	root := span{
		TraceID: traceID, SpanID: rootID, Name: "glassops run", Kind: kindInternal,
		StartTimeUnixNano: nano(runStart), EndTimeUnixNano: nano(runEnd),
		Attributes: []keyValue{num(attrDuration, runEnd.Sub(runStart).Milliseconds())},
		Status:     status{Code: statusUnset},
	}
	if v, ok := rec.Facts["run.pipeline"].(string); ok && v != "" {
		root.Attributes = append(root.Attributes, str(attrPipeline, v))
	}
	if rec.Provider != "" {
		root.Attributes = append(root.Attributes, str(attrProvider, rec.Provider))
	}
	if code, ok := asInt(rec.Facts["run.exit_code"]); ok {
		root.Attributes = append(root.Attributes, num(attrExit, int64(code)))
		if code != 0 {
			root.Status.Code = statusError
		}
	}
	for _, f := range []struct{ fact, attr string }{{"run.empty_plan", attrEmptyPlan}, {"run.cancelled", attrCancelled}} {
		if b, ok := rec.Facts[f.fact].(bool); ok {
			root.Attributes = append(root.Attributes, flag(f.attr, b))
		}
	}
	spans = append(spans, root)
	for _, s := range steps {
		spans = append(spans, stepSpan(cfg, traceID, rootID, s, runStart))
	}

	return otlpRequest{ResourceSpans: []resourceSpans{{
		Resource: resource{Attributes: []keyValue{str("service.name", cfg.Telemetry.OTLP.Service())}},
		ScopeSpans: []scopeSpans{{
			Scope: scope{Name: "glassops"},
			Spans: spans,
		}},
	}}}
}

func stepSpan(cfg *config.Config, traceID, parent string, s stepRecord, runStart time.Time) span {
	start := s.Started
	if start.IsZero() {
		start = runStart
	}
	end := start.Add(time.Duration(s.DurationMS) * time.Millisecond)
	sp := span{
		TraceID: traceID, SpanID: newSpanID(), ParentSpanID: parent,
		Name: "glassops " + s.Step, Kind: kindInternal,
		StartTimeUnixNano: nano(start), EndTimeUnixNano: nano(end),
		Attributes: []keyValue{
			str(attrStep, s.Step),
			num(attrIndex, int64(s.Index)),
			str(attrWhen, s.When),
			num(attrDuration, s.DurationMS),
			flag(attrSkipped, s.Skipped),
		},
		Status: status{Code: statusUnset},
	}
	outcome := "skipped"
	if !s.Skipped && s.ExitCode != nil {
		sp.Attributes = append(sp.Attributes, num(attrExit, int64(*s.ExitCode)))
		switch *s.ExitCode {
		case 0:
			outcome = "pass"
		case 1:
			outcome = "fail"
			sp.Status.Code = statusError
		default:
			outcome = "error"
			sp.Status.Code = statusError
		}
	}
	sp.Attributes = append(sp.Attributes, str(attrOutcome, outcome))
	// A target is named only when the config defines it, so a stray argument value (an
	// alias, a login) is never exported as a name.
	if name := targetOf(s.Argv); name != "" {
		if t, ok := cfg.Targets[name]; ok {
			sp.Attributes = append(sp.Attributes, str(attrTargetName, name))
			if t.Tier != "" {
				sp.Attributes = append(sp.Attributes, str(attrTargetTier, t.Tier))
			}
		}
	}
	return sp
}

// targetOf reads the target name from a step's argv: --target X or --target=X.
func targetOf(argv []string) string {
	for i, a := range argv {
		switch {
		case a == "--target":
			if i+1 < len(argv) {
				return argv[i+1]
			}
		case strings.HasPrefix(a, "--target="):
			return strings.TrimPrefix(a, "--target=")
		}
	}
	return ""
}

func asInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	}
	return 0, false
}

func nano(t time.Time) string { return strconv.FormatInt(t.UnixNano(), 10) }
