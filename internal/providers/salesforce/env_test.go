package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

// Synthetic values. No fixture holds a ScratchOrgInfo create result or a polled record with
// an auth code, because both carry secrets in a real capture.
const (
	synthAuthCode = "synthetic-auth-code-0001"
	synthRefresh  = "synthetic-refresh-token-0001"
	synthRecordID = "2SR000000000001AAA"
)

// seqExec answers commands by prefix like fakeExec, but each prefix has a list of replies
// used in order. The last one repeats.
type seqExec struct {
	mu      sync.Mutex
	replies map[string][]fakeReply
	calls   []Cmd
	// onCall, when set, runs for every call before the reply is chosen.
	onCall func(Cmd)
}

func (f *seqExec) run(_ context.Context, c Cmd) (Output, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, c)
	if f.onCall != nil {
		f.onCall(c)
	}
	key := strings.Join(append([]string{c.Name}, c.Args...), " ")
	best, found := "", false
	for prefix := range f.replies {
		if strings.HasPrefix(key, prefix) && len(prefix) >= len(best) {
			best, found = prefix, true
		}
	}
	if !found {
		return Output{Stdout: []byte(`{"status":1,"name":"UnexpectedCommand","message":"no reply for: ` + key + `"}`), ExitCode: 1}, nil
	}
	list := f.replies[best]
	r := list[0]
	if len(list) > 1 {
		f.replies[best] = list[1:]
	}
	return Output{Stdout: r.Stdout, ExitCode: r.ExitCode}, r.Err
}

func (f *seqExec) commands(prefix string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for _, c := range f.calls {
		k := strings.Join(append([]string{c.Name}, c.Args...), " ")
		if strings.HasPrefix(k, prefix) {
			out = append(out, k)
		}
	}
	return out
}

func okReply(stdout string) []fakeReply { return []fakeReply{{Stdout: []byte(stdout)}} }

// mustFixture is fixture for places with no *testing.T, such as a table literal.
func mustFixture(name string) []byte {
	data, err := os.ReadFile(filepath.Join("testdata", "sf", name))
	if err != nil {
		panic(err)
	}
	return data
}

const (
	cmdCreate = "sf data create record --sobject ScratchOrgInfo"
	cmdPoll   = "sf data query --query SELECT Status, ErrorCode, ScratchOrg, LoginUrl, AuthCode"
	cmdActive = "sf data query --query SELECT Id, ScratchOrg, OrgName"
	cmdDelete = "sf data delete record --sobject ActiveScratchOrg"
	cmdLogin  = "sf org login sfdx-url"
)

func createReply(id string) fakeReply {
	return fakeReply{Stdout: []byte(`{"status":0,"result":{"id":"` + id + `","success":true,"errors":[]},"warnings":[]}`)}
}

func pollReply(status, errorCode, org, loginURL string) fakeReply {
	rec := map[string]any{
		"attributes": map[string]any{"type": "ScratchOrgInfo"},
		"Status":     status, "ErrorCode": nil, "ScratchOrg": nil, "LoginUrl": nil, "AuthCode": nil,
		"ExpirationDate": "2026-11-01",
	}
	if errorCode != "" {
		rec["ErrorCode"] = errorCode
	}
	if org != "" {
		rec["ScratchOrg"] = org
		rec["LoginUrl"] = loginURL
		rec["AuthCode"] = synthAuthCode
	}
	data, _ := json.Marshal(map[string]any{"status": 0, "result": map[string]any{"records": []any{rec}, "totalSize": 1, "done": true}})
	return fakeReply{Stdout: data}
}

// tokenServer is a TLS test server that answers the auth code exchange. It points
// tokenClient at itself for the test.
func tokenServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	old := tokenClient
	tokenClient = srv.Client()
	t.Cleanup(func() { tokenClient = old; srv.Close() })
	return srv
}

func goodToken(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/oauth2/token" || r.Method != http.MethodPost {
			http.Error(w, "wrong endpoint", http.StatusNotFound)
			return
		}
		_ = r.ParseForm()
		want := map[string]string{"grant_type": "authorization_code", "code": synthAuthCode, "client_id": "PlatformCLI", "redirect_uri": scratchCallbackURL}
		for k, v := range want {
			if r.PostForm.Get(k) != v {
				t.Errorf("token form %s = %q, want %q", k, r.PostForm.Get(k), v)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"x","refresh_token":%q,"instance_url":"https://example.my.salesforce.com","id":"x"}`, synthRefresh)
	}
}

// fastEnv replaces envRetry with no waits and a clock that moves a minute per sleep. It
// returns the waits asked for.
func fastEnv(t *testing.T) *[]time.Duration {
	t.Helper()
	old := envRetry
	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	var waits []time.Duration
	envRetry.Now = func() time.Time { return now }
	envRetry.Sleep = func(ctx context.Context, d time.Duration) error {
		waits = append(waits, d)
		now = now.Add(time.Minute)
		return ctx.Err()
	}
	t.Cleanup(func() { envRetry = old })
	noSleep(t)
	return &waits
}

func envProvider(t *testing.T, f *seqExec) (*Provider, *session) {
	t.Helper()
	if f.replies["icacls"] == nil {
		f.replies["icacls"] = okReply("ok")
	}
	p := NewWithOptions(Options{Exec: f.run, TempDir: t.TempDir(), LookupEnv: func(string) (string, bool) { return "", false }})
	return p, newTestSession(t, providers.Target{Name: "devhub", Tier: "devhub"})
}

func envReq(t *testing.T) providers.EnvRequest {
	return providers.EnvRequest{Name: "glassops-repo-12-345", Dir: t.TempDir(), TTL: 24 * time.Hour}
}

// happyReplies is a signup that turns Active on the second read.
func happyReplies(loginURL string) map[string][]fakeReply {
	return map[string][]fakeReply{
		cmdCreate: {createReply(synthRecordID)},
		cmdPoll:   {pollReply("Creating", "", "", ""), pollReply("Active", "", "00D000000000002", loginURL)},
		cmdLogin:  {{Stdout: mustFixture("login-sfdx-url-scratch.json")}},
	}
}

func TestCreateEnvSignsUpPollsAndWritesAuthFile(t *testing.T) {
	waits := fastEnv(t)
	srv := tokenServer(t, goodToken(t))
	f := &seqExec{replies: happyReplies(srv.URL + "/")}
	var loginFile string
	f.onCall = func(c Cmd) {
		if c.Name == "sf" && len(c.Args) > 2 && c.Args[2] == "sfdx-url" {
			loginFile = c.Args[4]
		}
	}
	p, hub := envProvider(t, f)

	env, err := p.CreateEnv(context.Background(), hub, envReq(t))
	if err != nil {
		t.Fatal(err)
	}
	if env.ID != "00D000000000002" || env.RecordID != synthRecordID || env.Name != "glassops-repo-12-345" {
		t.Fatalf("env %+v", env)
	}
	if !env.ExpiresAt.Equal(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expires %v", env.ExpiresAt)
	}
	if env.Facts["salesforce.signup_attempts"] != 1 || env.Facts["salesforce.edition"] != "Developer" || env.Facts["salesforce.shape"] != false {
		t.Fatalf("facts %v", env.Facts)
	}
	data, err := os.ReadFile(env.AuthFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "force://PlatformCLI::"+synthRefresh+"@example.my.salesforce.com\n" {
		t.Fatalf("auth file %q", data)
	}
	if loginFile != env.AuthFile {
		t.Fatalf("logged in with %q, auth file %q", loginFile, env.AuthFile)
	}
	if !strings.HasPrefix(filepath.Base(filepath.Dir(env.AuthFile)), envDirPrefix) {
		t.Fatalf("auth file dir %s", env.AuthFile)
	}
	if !isWindows {
		fi, _ := os.Stat(env.AuthFile)
		di, _ := os.Stat(filepath.Dir(env.AuthFile))
		if fi.Mode().Perm() != 0o600 || di.Mode().Perm() != 0o700 {
			t.Fatalf("modes %v %v", fi.Mode(), di.Mode())
		}
	}
	create := f.commands(cmdCreate)
	if len(create) != 1 {
		t.Fatalf("creates %v", create)
	}
	for _, want := range []string{"ConnectedAppCallbackUrl=http://localhost:1717/OauthRedirect", "ConnectedAppConsumerKey=PlatformCLI",
		"DurationDays=1", "Edition=Developer", "OrgName=glassops-repo-12-345", "--target-org glassops-target"} {
		if !strings.Contains(create[0], want) {
			t.Fatalf("create %q lacks %q", create[0], want)
		}
	}
	if len(f.commands(cmdPoll)) != 2 || len(*waits) != 1 || (*waits)[0] != envRetry.PollEvery {
		t.Fatalf("polls %d, waits %v", len(f.commands(cmdPoll)), *waits)
	}
	// Neither secret reaches an argument, and the facts hold neither.
	for _, c := range f.calls {
		joined := strings.Join(c.Args, " ")
		if strings.Contains(joined, synthRefresh) || strings.Contains(joined, synthAuthCode) {
			t.Fatalf("secret in argv: %s", joined)
		}
	}
	if b, _ := json.Marshal(env.Facts); strings.Contains(string(b), synthRefresh) {
		t.Fatalf("secret in facts")
	}
}

func TestCreateEnvRetriesSignupErrors(t *testing.T) {
	cases := []struct {
		name   string
		create []fakeReply
		poll   []fakeReply
		// creates is how many signups were requested.
		creates int
	}{
		{"error code on the record", []fakeReply{createReply("2SR000000000001AAA"), createReply("2SR000000000002AAA")},
			[]fakeReply{pollReply("Error", "C-1033", "", ""), pollReply("Active", "", "00D000000000002", "")}, 2},
		{"capacity refusal on create", []fakeReply{
			{Stdout: []byte(`{"status":1,"name":"LIMIT_EXCEEDED","message":"The signup request failed because this organization has reached its active scratch org limit"}`), ExitCode: 1},
			createReply("2SR000000000002AAA")},
			[]fakeReply{pollReply("Active", "", "00D000000000002", "")}, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			waits := fastEnv(t)
			srv := tokenServer(t, goodToken(t))
			for i := range tc.poll {
				if tc.poll[i].Stdout != nil && strings.Contains(string(tc.poll[i].Stdout), `"Active"`) {
					tc.poll[i] = pollReply("Active", "", "00D000000000002", srv.URL)
				}
			}
			f := &seqExec{replies: map[string][]fakeReply{cmdCreate: tc.create, cmdPoll: tc.poll,
				cmdLogin: {{Stdout: fixture(t, "login-sfdx-url-scratch.json")}}}}
			p, hub := envProvider(t, f)
			env, err := p.CreateEnv(context.Background(), hub, envReq(t))
			if err != nil {
				t.Fatal(err)
			}
			if got := len(f.commands(cmdCreate)); got != tc.creates {
				t.Fatalf("creates %d, want %d", got, tc.creates)
			}
			if env.RecordID != "2SR000000000002AAA" || env.Facts["salesforce.signup_attempts"] != 2 {
				t.Fatalf("env %+v", env)
			}
			if len(*waits) != 1 || (*waits)[0] != 30*time.Second {
				t.Fatalf("waits %v", *waits)
			}
		})
	}
}

func TestCreateEnvGivesUpAfterAttempts(t *testing.T) {
	waits := fastEnv(t)
	f := &seqExec{replies: map[string][]fakeReply{
		cmdCreate: {createReply("2SR000000000001AAA"), createReply("2SR000000000002AAA"), createReply("2SR000000000003AAA")},
		cmdPoll:   {pollReply("Error", "C-1033", "", "")},
	}}
	p, hub := envProvider(t, f)
	env, err := p.CreateEnv(context.Background(), hub, envReq(t))
	if err == nil || !strings.Contains(err.Error(), "ErrorCode C-1033") {
		t.Fatalf("err %v", err)
	}
	if len(f.commands(cmdCreate)) != 3 || env.RecordID != "2SR000000000003AAA" {
		t.Fatalf("creates %d, env %+v", len(f.commands(cmdCreate)), env)
	}
	if want := []time.Duration{30 * time.Second, 60 * time.Second}; fmt.Sprint(*waits) != fmt.Sprint(want) {
		t.Fatalf("waits %v, want %v", *waits, want)
	}
}

func TestCreateEnvDoesNotRetryOtherFailures(t *testing.T) {
	cases := []struct {
		name   string
		create fakeReply
		poll   []fakeReply
		want   string
	}{
		{"create refused for another reason", fakeReply{Stdout: []byte(`{"status":1,"name":"INVALID_FIELD","message":"No such column 'Bogus__c'"}`), ExitCode: 1},
			nil, "create the ScratchOrgInfo record"},
		{"signup still not active at the deadline", createReply(synthRecordID),
			[]fakeReply{pollReply("Creating", "", "", "")}, "not active after"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fastEnv(t)
			r := map[string][]fakeReply{cmdCreate: {tc.create}}
			if tc.poll != nil {
				r[cmdPoll] = tc.poll
			}
			f := &seqExec{replies: r}
			p, hub := envProvider(t, f)
			_, err := p.CreateEnv(context.Background(), hub, envReq(t))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err %v, want %q", err, tc.want)
			}
			if n := len(f.commands(cmdCreate)); n != 1 {
				t.Fatalf("creates %d, want 1", n)
			}
		})
	}
}

// The Phase 2 defect: an Id check that accepted only one length dropped the record Id before
// saving it. The Id is now kept whatever it looks like, then checked.
func TestCreateEnvRecordIDKeptBeforeCheck(t *testing.T) {
	cases := []struct {
		id      string
		wantErr bool
	}{
		{"2SR000000000001", false},
		{"2SR000000000001AAA", false},
		{"2SR0000001", true},
		{"2SR00000000000!AAA", true},
		{"", true},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%q", tc.id), func(t *testing.T) {
			fastEnv(t)
			srv := tokenServer(t, goodToken(t))
			f := &seqExec{replies: map[string][]fakeReply{
				cmdCreate: {createReply(tc.id)},
				cmdPoll:   {pollReply("Active", "", "00D000000000002", srv.URL)},
				cmdLogin:  {{Stdout: fixture(t, "login-sfdx-url-scratch.json")}},
			}}
			p, hub := envProvider(t, f)
			env, err := p.CreateEnv(context.Background(), hub, envReq(t))
			if (err != nil) != tc.wantErr {
				t.Fatalf("err %v, want error %v", err, tc.wantErr)
			}
			if env.RecordID != tc.id {
				t.Fatalf("RecordID %q, want %q", env.RecordID, tc.id)
			}
			if tc.wantErr && len(f.commands(cmdPoll)) != 0 {
				t.Fatalf("polled a bad Id")
			}
			if !tc.wantErr && !strings.Contains(f.commands(cmdPoll)[0], "WHERE Id = '"+tc.id+"'") {
				t.Fatalf("poll %v", f.commands(cmdPoll))
			}
		})
	}
}

func TestCreateEnvTokenExchangeFailureKeepsOrgID(t *testing.T) {
	fastEnv(t)
	srv := tokenServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"invalid_grant","error_description":"expired authorization code %s"}`, synthAuthCode)
	})
	f := &seqExec{replies: map[string][]fakeReply{
		cmdCreate: {createReply(synthRecordID)},
		cmdPoll:   {pollReply("Active", "", "00D000000000002", srv.URL)},
	}}
	p, hub := envProvider(t, f)
	env, err := p.CreateEnv(context.Background(), hub, envReq(t))
	if err == nil || !strings.Contains(err.Error(), "status 400, invalid_grant") {
		t.Fatalf("err %v", err)
	}
	if strings.Contains(err.Error(), synthAuthCode) || strings.Contains(err.Error(), "127.0.0.1") {
		t.Fatalf("error leaks: %v", err)
	}
	if env.ID != "00D000000000002" || env.RecordID != synthRecordID || env.AuthFile != "" {
		t.Fatalf("env %+v", env)
	}
	if n := len(f.commands(cmdCreate)); n != 1 {
		t.Fatalf("a failed exchange was retried as a new signup: %d creates", n)
	}
}

func TestCreateEnvRefusesPlainHTTPLoginURL(t *testing.T) {
	fastEnv(t)
	f := &seqExec{replies: map[string][]fakeReply{
		cmdCreate: {createReply(synthRecordID)},
		cmdPoll:   {pollReply("Active", "", "00D000000000002", "http://example.my.salesforce.com")},
	}}
	p, hub := envProvider(t, f)
	if _, err := p.CreateEnv(context.Background(), hub, envReq(t)); err == nil || !strings.Contains(err.Error(), "not an https url") {
		t.Fatalf("err %v", err)
	}
}

func TestCreateEnvLoginMustReachTheNewOrg(t *testing.T) {
	fastEnv(t)
	srv := tokenServer(t, goodToken(t))
	f := &seqExec{replies: map[string][]fakeReply{
		cmdCreate: {createReply(synthRecordID)},
		cmdPoll:   {pollReply("Active", "", "00D000000000009", srv.URL)},
		cmdLogin:  {{Stdout: fixture(t, "login-sfdx-url-scratch.json")}},
	}}
	p, hub := envProvider(t, f)
	env, err := p.CreateEnv(context.Background(), hub, envReq(t))
	if err == nil || !strings.Contains(err.Error(), "not the new scratch org") {
		t.Fatalf("err %v", err)
	}
	if env.ID != "00D000000000009" || env.AuthFile == "" {
		t.Fatalf("env %+v", env)
	}
}

func TestCreateEnvLoginFailureIsScrubbed(t *testing.T) {
	fastEnv(t)
	srv := tokenServer(t, goodToken(t))
	f := &seqExec{replies: map[string][]fakeReply{
		cmdCreate: {createReply(synthRecordID)},
		cmdPoll:   {pollReply("Active", "", "00D000000000002", srv.URL)},
		cmdLogin:  {{Stdout: fixture(t, "login-sfdx-url-invalid.json"), ExitCode: 1}},
	}}
	p, hub := envProvider(t, f)
	_, err := p.CreateEnv(context.Background(), hub, envReq(t))
	if err == nil || !strings.Contains(err.Error(), "log in to the new scratch org") || strings.Contains(err.Error(), synthRefresh) {
		t.Fatalf("err %v", err)
	}
}

func TestSignupValues(t *testing.T) {
	long := "glassops-" + strings.Repeat("a", 100)
	cases := []struct {
		name    string
		req     providers.EnvRequest
		def     string
		want    []string
		notWant []string
		err     string
	}{
		{"defaults", providers.EnvRequest{Name: "glassops-r-1-2", TTL: 24 * time.Hour}, "",
			[]string{"Edition=Developer", "DurationDays=1", "OrgName=glassops-r-1-2"}, []string{"SourceOrg"}, ""},
		{"ttl rounds up to whole days", providers.EnvRequest{Name: "n", TTL: 25 * time.Hour}, "",
			[]string{"DurationDays=2"}, nil, ""},
		{"ttl zero", providers.EnvRequest{Name: "n"}, "", nil, nil, "1 to 30 days"},
		{"ttl over 30 days", providers.EnvRequest{Name: "n", TTL: 31 * 24 * time.Hour}, "", nil, nil, "1 to 30 days"},
		{"shape replaces edition with the 15 character org Id", providers.EnvRequest{Name: "n", TTL: time.Hour, Shape: "00D000000000009AAA"},
			`{"edition":"Enterprise"}`, []string{"SourceOrg=00D000000000009 "}, []string{"Edition="}, ""},
		{"shape that is not an org Id", providers.EnvRequest{Name: "n", TTL: time.Hour, Shape: "001000000000009"}, "", nil, nil, "org shape must be an org Id"},
		{"name truncated to the field limit", providers.EnvRequest{Name: long, TTL: time.Hour}, "",
			[]string{"OrgName=" + long[:orgNameMax] + " "}, nil, ""},
		{"name with a space", providers.EnvRequest{Name: "glassops x", TTL: time.Hour}, "", nil, nil, "may hold only"},
		{"definition fields", providers.EnvRequest{Name: "n", TTL: time.Hour},
			`{"orgName":"Ignored","edition":"Partner Developer","features":["EnableSetPasswordInApi","MultiCurrency"],"hasSampleData":false}`,
			[]string{"Edition='Partner Developer'", "Features=EnableSetPasswordInApi;MultiCurrency", "HasSampleData=false", "OrgName=n "}, []string{"Ignored"}, ""},
		{"definition settings are refused", providers.EnvRequest{Name: "n", TTL: time.Hour},
			`{"edition":"Developer","settings":{"lightningExperienceSettings":{"enableS1DesktopEnabled":true}}}`, nil, nil, "cannot apply definition keys settings"},
		{"definition value with a quote", providers.EnvRequest{Name: "n", TTL: time.Hour}, `{"description":"it's"}`, nil, nil, "quote"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.Dir = t.TempDir()
			if tc.def != "" {
				tc.req.DefinitionFile = "config/def.json"
				if err := os.MkdirAll(filepath.Join(tc.req.Dir, "config"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(tc.req.Dir, "config", "def.json"), []byte(tc.def), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			values, _, _, err := signupValues(tc.req)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("err %v, want %q", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			values += " "
			for _, w := range tc.want {
				if !strings.Contains(values, w) {
					t.Fatalf("values %q lack %q", values, w)
				}
			}
			for _, w := range tc.notWant {
				if strings.Contains(values, w) {
					t.Fatalf("values %q hold %q", values, w)
				}
			}
		})
	}
}

func activeReply(rows ...map[string]any) fakeReply {
	recs := make([]any, len(rows))
	for i, r := range rows {
		recs[i] = r
	}
	data, _ := json.Marshal(map[string]any{"status": 0, "result": map[string]any{"records": recs, "totalSize": len(rows), "done": true}})
	return fakeReply{Stdout: data}
}

func authFileIn(t *testing.T, p *Provider) string {
	t.Helper()
	path, err := p.writeAuthFile(context.Background(), "force://PlatformCLI::x@example.my.salesforce.com")
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDeleteEnv(t *testing.T) {
	deleted := fakeReply{Stdout: []byte(`{"status":0,"result":{"id":"2AS000000000001AAA"},"warnings":[]}`)}
	cases := []struct {
		name    string
		env     providers.Env
		replies map[string][]fakeReply
		// where is the expected ActiveScratchOrg filter, empty for no query.
		where   string
		deletes int
		err     string
	}{
		{"by 18 character org Id", providers.Env{ID: "00D000000000002AAA"},
			map[string][]fakeReply{cmdActive: {{Stdout: mustFixture("query-active-scratch-org.json")}}, cmdDelete: {deleted}},
			"WHERE ScratchOrg = '00D000000000002'", 1, ""},
		{"by 15 character org Id", providers.Env{ID: "00D000000000002"},
			map[string][]fakeReply{cmdActive: {{Stdout: mustFixture("query-active-scratch-org.json")}}, cmdDelete: {deleted}},
			"WHERE ScratchOrg = '00D000000000002'", 1, ""},
		{"already gone", providers.Env{ID: "00D000000000002AAA"},
			map[string][]fakeReply{cmdActive: {activeReply()}}, "WHERE ScratchOrg = '00D000000000002'", 0, ""},
		{"two rows match", providers.Env{ID: "00D000000000002AAA"},
			map[string][]fakeReply{cmdActive: {activeReply(map[string]any{"Id": "2AS000000000001AAA"}, map[string]any{"Id": "2AS000000000002AAA"})}},
			"WHERE ScratchOrg = '00D000000000002'", 0, "2 active scratch orgs matched"},
		{"falls back to the signup record", providers.Env{RecordID: "2SR000000000001AAA"},
			map[string][]fakeReply{cmdPoll: {{Stdout: mustFixture("query-scratch-org-info.json")}},
				cmdActive: {{Stdout: mustFixture("query-active-scratch-org.json")}}, cmdDelete: {deleted}},
			"WHERE ScratchOrg = '00D000000000002'", 1, ""},
		{"signup that never produced an org", providers.Env{RecordID: "2SR000000000001"},
			map[string][]fakeReply{cmdPoll: {pollReply("Error", "C-1033", "", "")}}, "", 0, ""},
		{"by name", providers.Env{Name: "glassops-repo-12-345"},
			map[string][]fakeReply{cmdActive: {{Stdout: mustFixture("query-active-scratch-org.json")}}, cmdDelete: {deleted}},
			"WHERE OrgName = 'glassops-repo-12-345'", 1, ""},
		{"org Id that is not an org Id", providers.Env{ID: "001000000000002AAA"}, map[string][]fakeReply{}, "", 0, "refusing"},
		{"org Id of the wrong length", providers.Env{ID: "00D0000000000020"}, map[string][]fakeReply{}, "", 0, "refusing"},
		{"record Id with a quote", providers.Env{RecordID: "2SR00000000001'AA"}, map[string][]fakeReply{}, "", 0, "refusing"},
		{"name with a quote", providers.Env{Name: "x' OR Name != '"}, map[string][]fakeReply{}, "", 0, "refusing"},
		{"nothing to go on", providers.Env{}, map[string][]fakeReply{}, "", 0, "refusing"},
		{"delete fails", providers.Env{ID: "00D000000000002AAA"},
			map[string][]fakeReply{cmdActive: {{Stdout: mustFixture("query-active-scratch-org.json")}},
				cmdDelete: {{Stdout: []byte(`{"status":1,"name":"INSUFFICIENT_ACCESS","message":"insufficient access rights"}`), ExitCode: 1}}},
			"WHERE ScratchOrg = '00D000000000002'", 1, "insufficient access"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := &seqExec{replies: tc.replies}
			p, hub := envProvider(t, f)
			tc.env.AuthFile = authFileIn(t, p)
			err := p.DeleteEnv(context.Background(), hub, tc.env)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("err %v, want %q", err, tc.err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			q := f.commands(cmdActive)
			if tc.where == "" && len(q) != 0 || tc.where != "" && (len(q) != 1 || !strings.Contains(q[0], tc.where)) {
				t.Fatalf("queries %v, want %q", q, tc.where)
			}
			d := f.commands(cmdDelete)
			if len(d) != tc.deletes {
				t.Fatalf("deletes %v", d)
			}
			if tc.deletes == 1 && !strings.Contains(d[0], "--record-id 2AS000000000001AAA") {
				t.Fatalf("delete %v", d)
			}
			_, statErr := os.Stat(tc.env.AuthFile)
			if gone := errors.Is(statErr, os.ErrNotExist); gone != (tc.err == "") {
				t.Fatalf("auth file removed %v, err %v", gone, err)
			}
		})
	}
}

func TestRemoveAuthFileLeavesOtherPathsAlone(t *testing.T) {
	p := NewWithOptions(Options{TempDir: t.TempDir()})
	other := filepath.Join(t.TempDir(), envAuthFile)
	if err := os.WriteFile(other, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	p.removeAuthFile(other)
	if _, err := os.Stat(other); err != nil {
		t.Fatalf("removed a file outside an env dir: %v", err)
	}
}

func TestSweepEnvs(t *testing.T) {
	fastEnv(t) // now is 2026-10-31T12:00Z
	row := func(id, org, name, created string) map[string]any {
		return map[string]any{"Id": id, "ScratchOrg": org, "OrgName": name, "CreatedDate": created,
			"ExpirationDate": "2026-11-01", "ScratchOrgInfoId": "2SR000000000001AAA"}
	}
	f := &seqExec{replies: map[string][]fakeReply{
		cmdActive: {activeReply(
			row("2AS000000000001AAA", "00D000000000001", "glassops-repo-1-1", "2026-10-29T10:00:00.000+0000"),
			row("2AS000000000002AAA", "00D000000000002", "glassops-repo-2-2", "2026-10-31T11:00:00.000+0000"),
			row("2AS000000000003AAA", "00D000000000003", "GlassOps E2E", "2026-10-01T10:00:00.000+0000"),
			row("2AS000000000004AAA", "00D000000000004", "glassopsx-repo-4-4", "2026-10-01T10:00:00.000+0000"),
			row("2AS000000000005AAA", "00D000000000005", "glassops-repo-5-5", "yesterday"),
		)},
		cmdDelete: {{Stdout: []byte(`{"status":0,"result":{},"warnings":[]}`)}},
	}}
	p, hub := envProvider(t, f)
	got, err := p.SweepEnvs(context.Background(), hub, "glassops", 24*time.Hour)
	if err == nil || !strings.Contains(err.Error(), "glassops-repo-5-5 has a CreatedDate that cannot be read") {
		t.Fatalf("err %v", err)
	}
	if len(got) != 1 || got[0].ID != "00D000000000001" || got[0].Name != "glassops-repo-1-1" || got[0].RecordID != "2SR000000000001AAA" {
		t.Fatalf("deleted %+v", got)
	}
	d := f.commands(cmdDelete)
	if len(d) != 1 || !strings.Contains(d[0], "--record-id 2AS000000000001AAA") {
		t.Fatalf("deletes %v", d)
	}
	q := f.commands(cmdActive)
	if len(q) != 1 || !strings.Contains(q[0], "SELECT Id, ScratchOrg, OrgName, CreatedDate, ExpirationDate, ScratchOrgInfoId FROM ActiveScratchOrg") ||
		strings.Contains(q[0], "SignupUsername") {
		t.Fatalf("query %v", q)
	}
}

func TestSweepEnvsReadsTheFixture(t *testing.T) {
	fastEnv(t)
	f := &seqExec{replies: map[string][]fakeReply{cmdActive: {{Stdout: fixture(t, "query-active-scratch-org.json")}}}}
	p, hub := envProvider(t, f)
	got, err := p.SweepEnvs(context.Background(), hub, "glassops", time.Hour)
	if err != nil || len(got) != 0 || len(f.commands(cmdDelete)) != 0 {
		t.Fatalf("a row named %q was swept: %v %v", "GlassOps E2E", got, err)
	}
	// The fixture's CreatedDate parses with the layout sweep uses.
	if _, err := time.Parse(sfDateTime, "2026-10-31T03:39:25.000+0000"); err != nil {
		t.Fatal(err)
	}
}

func TestSweepEnvsRefusesUnsafeArguments(t *testing.T) {
	cases := []struct {
		prefix string
		maxAge time.Duration
	}{
		{"", time.Hour},
		{"glassops", 0},
		{"glassops", -time.Hour},
		{"glass' OR '", time.Hour},
	}
	for _, tc := range cases {
		f := &seqExec{replies: map[string][]fakeReply{}}
		p, hub := envProvider(t, f)
		if _, err := p.SweepEnvs(context.Background(), hub, tc.prefix, tc.maxAge); err == nil {
			t.Fatalf("prefix %q max age %s: no error", tc.prefix, tc.maxAge)
		}
		if len(f.calls) != 0 {
			t.Fatalf("ran %v", f.calls)
		}
	}
}

func TestEnvMethodsNeedAProviderSession(t *testing.T) {
	p := NewWithOptions(Options{TempDir: t.TempDir()})
	var other providers.Session = otherSession{}
	if _, err := p.CreateEnv(context.Background(), other, providers.EnvRequest{}); err == nil {
		t.Fatal("create")
	}
	if err := p.DeleteEnv(context.Background(), other, providers.Env{ID: "00D000000000002"}); err == nil {
		t.Fatal("delete")
	}
	if _, err := p.SweepEnvs(context.Background(), other, "glassops", time.Hour); err == nil {
		t.Fatal("sweep")
	}
	if _, err := p.Seed(context.Background(), other, providers.SeedRequest{Mode: config.SeedMinimal}); err == nil {
		t.Fatal("seed")
	}
}

type otherSession struct{}

func (otherSession) Target() providers.Target { return providers.Target{} }
func (otherSession) Close() error             { return nil }
