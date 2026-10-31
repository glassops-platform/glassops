package salesforce

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

const (
	testPEM   = "-----BEGIN PRIVATE KEY-----\nTESTKEYBODYTESTKEYBODY\n-----END PRIVATE KEY-----"
	testToken = "00D000000000002!TESTTOKENVALUE"
	testSfdx  = "force://PlatformCLI::TESTREFRESHTOKEN@example.my.salesforce.com"
)

func testEnv() map[string]string {
	return map[string]string{
		"T_CID":   "TESTCLIENTID",
		"T_KEY":   testPEM,
		"T_URL":   testSfdx,
		"T_TOKEN": testToken,
	}
}

func jwtTarget() providers.Target {
	return providers.Target{Name: "hub", ID: "00D000000000001", Tier: "devhub",
		Auth: config.Auth{Method: "jwt", Username: "ci-user@example.com", ClientID: "$T_CID", Key: "$T_KEY"}}
}

func sfdxTarget() providers.Target {
	return providers.Target{Name: "scratch", ID: "00D000000000002", Tier: "ephemeral",
		Auth: config.Auth{Method: "sfdx-url", URL: "$T_URL"}}
}

func tokenTarget() providers.Target {
	return providers.Target{Name: "scratch", ID: "00D000000000002", Tier: "ephemeral",
		Auth: config.Auth{Method: "access-token", Token: "$T_TOKEN", InstanceURL: "https://example.my.salesforce.com"}}
}

// withReplies adds the icacls reply Windows needs and builds the fake.
func withReplies(r map[string]fakeReply) *fakeExec {
	if r == nil {
		r = map[string]fakeReply{}
	}
	r["icacls"] = fakeReply{Stdout: []byte("ok"), ExitCode: 0}
	return newFakeExec(r)
}

func noSleep(t *testing.T) *[]time.Duration {
	t.Helper()
	var waits []time.Duration
	old := loginRetry
	loginRetry.Sleep = func(ctx context.Context, d time.Duration) error {
		waits = append(waits, d)
		return ctx.Err()
	}
	t.Cleanup(func() { loginRetry = old })
	return &waits
}

func entries(t *testing.T, dir string) []string {
	t.Helper()
	list, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range list {
		names = append(names, e.Name())
	}
	return names
}

func loginCalls(f *fakeExec) []Cmd {
	var out []Cmd
	for _, c := range f.calls {
		if c.Name == "sf" && len(c.Args) > 2 && c.Args[0] == "org" && c.Args[1] == "login" {
			out = append(out, c)
		}
	}
	return out
}

func TestAuthenticateJWT(t *testing.T) {
	f := withReplies(map[string]fakeReply{"sf org login jwt": {Stdout: fixture(t, "login-jwt-devhub.json")}})
	p := newTestProvider(t, f, testEnv())
	s, err := p.Authenticate(context.Background(), jwtTarget())
	if err != nil {
		t.Fatal(err)
	}
	ss, ok := asSession(s)
	if !ok {
		t.Fatal("not a salesforce session")
	}
	if ss.alias != sessionAlias || s.Target().Name != "hub" {
		t.Errorf("alias %q target %q", ss.alias, s.Target().Name)
	}
	calls := loginCalls(f)
	if len(calls) != 1 {
		t.Fatalf("login calls = %d", len(calls))
	}
	args := strings.Join(calls[0].Args, " ")
	keyFile := filepath.Join(ss.home, "jwt.key")
	for _, want := range []string{
		"org login jwt", "--username ci-user@example.com", "--client-id TESTCLIENTID",
		"--jwt-key-file " + keyFile, "--instance-url https://login.salesforce.com", "--alias glassops-target", "--json",
	} {
		if !strings.Contains(args, want) {
			t.Errorf("args lack %q: %s", want, args)
		}
	}
	// The key stays until Close, because sf re-runs JWT with the stored path.
	data, err := os.ReadFile(keyFile)
	if err != nil {
		t.Fatalf("key file missing after login: %v", err)
	}
	if !strings.Contains(string(data), "TESTKEYBODYTESTKEYBODY") {
		t.Error("key file does not hold the key")
	}
	// The session env points sf at the private home.
	if !contains(calls[0].Env, "HOME="+ss.home) || !contains(calls[0].Env, "USERPROFILE="+ss.home) {
		t.Errorf("env %v does not isolate the home", calls[0].Env)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ss.home); !os.IsNotExist(err) {
		t.Errorf("home still exists after Close: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
}

func TestAuthenticateJWTInstanceURLAndKeyPath(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "server.key")
	if err := os.WriteFile(keyPath, []byte(testPEM+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	env := testEnv()
	env["T_KEY"] = keyPath
	f := withReplies(map[string]fakeReply{"sf org login jwt": {Stdout: fixture(t, "login-jwt-devhub.json")}})
	p := newTestProvider(t, f, env)
	tg := jwtTarget()
	tg.Auth.InstanceURL = "https://test.salesforce.com"
	s, err := p.Authenticate(context.Background(), tg)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	args := strings.Join(loginCalls(f)[0].Args, " ")
	if !strings.Contains(args, "--instance-url https://test.salesforce.com") {
		t.Errorf("instance url not used: %s", args)
	}
	ss, _ := asSession(s)
	data, _ := os.ReadFile(filepath.Join(ss.home, "jwt.key"))
	if !strings.Contains(string(data), "TESTKEYBODY") {
		t.Error("key read from a path was not copied into the home")
	}
}

func TestJWTKeyText(t *testing.T) {
	got, err := jwtKeyText(`-----BEGIN PRIVATE KEY-----\nABC\n-----END PRIVATE KEY-----`)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, `\n`) || strings.Count(got, "\n") != 3 {
		t.Errorf("flattened key not restored: %q", got)
	}
	if _, err := jwtKeyText("/no/such/key/file"); err == nil || strings.Contains(err.Error(), "/no/such") {
		t.Errorf("want an error that does not echo the value, got %v", err)
	}
}

func TestAuthenticateSfdxURL(t *testing.T) {
	var fileDuringLogin string
	var homeSeen string
	exec := func(_ context.Context, c Cmd) (Output, error) {
		if c.Name == "icacls" {
			return Output{}, nil
		}
		for i, a := range c.Args {
			if a == "--sfdx-url-file" {
				data, err := os.ReadFile(c.Args[i+1])
				if err != nil {
					t.Errorf("url file not readable during login: %v", err)
				}
				fileDuringLogin = string(data)
				if !isWindows {
					if st, _ := os.Stat(c.Args[i+1]); st.Mode().Perm() != 0o600 {
						t.Errorf("url file mode %v", st.Mode().Perm())
					}
				}
			}
		}
		for _, kv := range c.Env {
			if strings.HasPrefix(kv, "HOME=") {
				homeSeen = strings.TrimPrefix(kv, "HOME=")
			}
		}
		return Output{Stdout: fixture(t, "login-sfdx-url-scratch.json")}, nil
	}
	p := NewWithOptions(Options{Exec: exec, LookupEnv: func(k string) (string, bool) { v, ok := testEnv()[k]; return v, ok }, TempDir: t.TempDir()})
	s, err := p.Authenticate(context.Background(), sfdxTarget())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if fileDuringLogin != testSfdx {
		t.Errorf("file content = %q", fileDuringLogin)
	}
	ss, _ := asSession(s)
	if homeSeen != ss.home {
		t.Errorf("sf saw home %q, session home %q", homeSeen, ss.home)
	}
	if _, err := os.Stat(filepath.Join(ss.home, "sfdx-url.txt")); !os.IsNotExist(err) {
		t.Errorf("url file was not deleted right after login: %v", err)
	}
}

func TestAuthenticateAccessToken(t *testing.T) {
	f := withReplies(map[string]fakeReply{"sf org login access-token": {Stdout: fixture(t, "login-access-token-scratch.json")}})
	p := newTestProvider(t, f, testEnv())
	s, err := p.Authenticate(context.Background(), tokenTarget())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	calls := loginCalls(f)
	if len(calls) != 1 {
		t.Fatalf("login calls = %d", len(calls))
	}
	args := strings.Join(calls[0].Args, " ")
	for _, want := range []string{"org login access-token", "--instance-url https://example.my.salesforce.com", "--alias glassops-target", "--no-prompt", "--json"} {
		if !strings.Contains(args, want) {
			t.Errorf("args lack %q: %s", want, args)
		}
	}
	if !contains(calls[0].Env, "SF_ACCESS_TOKEN="+testToken) {
		t.Error("token not passed in SF_ACCESS_TOKEN")
	}
}

// Question 3: no secret is ever an argument, on every login path, including a failed one.
func TestNoSecretInArgs(t *testing.T) {
	secrets := []string{testPEM, "TESTKEYBODYTESTKEYBODY", testToken, "TESTTOKENVALUE", testSfdx, "TESTREFRESHTOKEN"}
	// The client id is a public identifier and the spec passes it as --client-id.
	cases := []struct {
		name   string
		target providers.Target
		prefix string
		fix    string
	}{
		{"jwt", jwtTarget(), "sf org login jwt", "login-jwt-devhub.json"},
		{"sfdx-url", sfdxTarget(), "sf org login sfdx-url", "login-sfdx-url-scratch.json"},
		{"access-token", tokenTarget(), "sf org login access-token", "login-access-token-scratch.json"},
		{"jwt fail", jwtTarget(), "sf org login jwt", "login-jwt-wrong-key.json"},
		{"sfdx-url fail", sfdxTarget(), "sf org login sfdx-url", "login-sfdx-url-invalid.json"},
		{"access-token fail", tokenTarget(), "sf org login access-token", "login-access-token-invalid.json"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			noSleep(t)
			code := 0
			if strings.HasSuffix(tc.name, "fail") {
				code = 1
			}
			f := withReplies(map[string]fakeReply{tc.prefix: {Stdout: fixture(t, tc.fix), ExitCode: code}})
			p := newTestProvider(t, f, testEnv())
			s, err := p.Authenticate(context.Background(), tc.target)
			if s != nil {
				defer s.Close()
			}
			if (err != nil) != (code == 1) {
				t.Fatalf("err = %v", err)
			}
			if len(loginCalls(f)) == 0 {
				t.Fatal("no login call recorded")
			}
			for _, c := range f.calls {
				all := strings.Join(c.Args, "\x00")
				for _, sec := range secrets {
					if strings.Contains(all, sec) {
						t.Errorf("secret %q is in the args of %s", sec, c.Name)
					}
				}
			}
			if err != nil {
				for _, sec := range secrets {
					if strings.Contains(err.Error(), sec) {
						t.Errorf("secret %q is in the error", sec)
					}
				}
			}
		})
	}
}

// Question 4: the key file is 0600 and the home is 0700.
func TestSessionModes(t *testing.T) {
	if isWindows {
		t.Skip("Unix modes do not apply on Windows, see TestWindowsHomeOwnerOnly")
	}
	f := withReplies(map[string]fakeReply{"sf org login jwt": {Stdout: fixture(t, "login-jwt-devhub.json")}})
	p := newTestProvider(t, f, testEnv())
	s, err := p.Authenticate(context.Background(), jwtTarget())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ss, _ := asSession(s)
	home, err := os.Stat(ss.home)
	if err != nil {
		t.Fatal(err)
	}
	if home.Mode().Perm() != 0o700 {
		t.Errorf("home mode = %v, want 0700", home.Mode().Perm())
	}
	key, err := os.Stat(filepath.Join(ss.home, "jwt.key"))
	if err != nil {
		t.Fatal(err)
	}
	if key.Mode().Perm() != 0o600 {
		t.Errorf("key mode = %v, want 0600", key.Mode().Perm())
	}
	if !strings.HasPrefix(filepath.Base(ss.home), "glassops-session-") || filepath.Dir(ss.home) != p.tempDir {
		t.Errorf("home %q is not a glassops-session- dir under the temp dir", ss.home)
	}
}

// On Windows the home is restricted through icacls. This runs the real tool and reads the
// ACL back: one principal, no inherited entries.
func TestWindowsHomeOwnerOnly(t *testing.T) {
	if !isWindows {
		t.Skip("icacls exists on Windows only")
	}
	p := NewWithOptions(Options{LookupEnv: func(k string) (string, bool) { v, ok := testEnv()[k]; return v, ok }, TempDir: t.TempDir(),
		Exec: func(ctx context.Context, c Cmd) (Output, error) {
			if c.Name == "icacls" {
				return osExec(ctx, c)
			}
			return Output{Stdout: fixture(t, "login-jwt-devhub.json")}, nil
		}})
	s, err := p.Authenticate(context.Background(), jwtTarget())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ss, _ := asSession(s)
	out, err := exec.Command("icacls", ss.home).CombinedOutput()
	if err != nil {
		t.Fatal(err, string(out))
	}
	var principals int
	for _, ln := range strings.Split(string(out), "\n") {
		if strings.Contains(ln, "(OI)(CI)") {
			principals++
			if strings.Contains(ln, "(I)") {
				t.Errorf("inherited entry left: %s", ln)
			}
		}
	}
	if principals != 1 {
		t.Errorf("want exactly one ACL entry, got %d:\n%s", principals, out)
	}
}

func TestUnsetVariableIsNamedNotValued(t *testing.T) {
	for _, drop := range []string{"T_CID", "T_KEY"} {
		env := testEnv()
		delete(env, drop)
		f := withReplies(map[string]fakeReply{})
		p := newTestProvider(t, f, env)
		_, err := p.Authenticate(context.Background(), jwtTarget())
		if err == nil || !strings.Contains(err.Error(), drop) {
			t.Errorf("drop %s: error %v does not name the variable", drop, err)
		}
		if len(f.calls) != 0 {
			t.Errorf("sf ran with an unresolved variable")
		}
		if got := entries(t, p.tempDir); len(got) != 0 {
			t.Errorf("session dir created before auth resolved: %v", got)
		}
	}
	env := testEnv()
	env["T_KEY"] = ""
	p := newTestProvider(t, withReplies(nil), env)
	if _, err := p.Authenticate(context.Background(), jwtTarget()); err == nil || !strings.Contains(err.Error(), "T_KEY") {
		t.Errorf("empty variable: %v", err)
	}
}

func TestResolveRules(t *testing.T) {
	p := newTestProvider(t, withReplies(nil), testEnv())
	if _, err := p.resolveLogin(providers.Target{Auth: config.Auth{Method: "sfdx-url", URL: "force://literal@secret"}}); err == nil ||
		strings.Contains(err.Error(), "literal") {
		t.Errorf("literal secret must be rejected without echoing it: %v", err)
	}
	l, err := p.resolveLogin(providers.Target{Auth: config.Auth{Method: "jwt", Username: "plain@example.com", ClientID: "$T_CID", Key: "$T_KEY"}})
	if err != nil || l.username != "plain@example.com" {
		t.Errorf("text username: %v %q", err, l.username)
	}
	if _, err := p.resolveLogin(providers.Target{Auth: config.Auth{Method: "basic"}}); err == nil {
		t.Error("unknown method accepted")
	}
	if _, err := p.resolveLogin(providers.Target{Auth: config.Auth{Method: "access-token", Token: "$T_TOKEN"}}); err == nil {
		t.Error("access-token without instance_url accepted")
	}
}

func TestFailureRemovesHomeAndStaysClean(t *testing.T) {
	noSleep(t)
	f := withReplies(map[string]fakeReply{"sf org login jwt": {Stdout: fixture(t, "login-jwt-unknown-user.json"), ExitCode: 1}})
	p := newTestProvider(t, f, testEnv())
	s, err := p.Authenticate(context.Background(), jwtTarget())
	if err == nil || s != nil {
		t.Fatalf("want a failure, got %v %v", s, err)
	}
	if got := entries(t, p.tempDir); len(got) != 0 {
		t.Errorf("home left behind after failure: %v", got)
	}
	msg := err.Error()
	for _, bad := range []string{"Example Company", "TESTKEYBODY", "TESTCLIENTID", "ci-user@example.com"} {
		if strings.Contains(msg, bad) {
			t.Errorf("error leaks %q: %s", bad, msg)
		}
	}
	if !strings.Contains(msg, "External client app is not installed") {
		t.Errorf("error lost the cause: %s", msg)
	}
	if strings.Count(msg, "External client app is not installed") != 1 {
		t.Errorf("repeated sf lines were not collapsed: %s", msg)
	}
}

func TestFailureErrorMasksEmails(t *testing.T) {
	body := `{"status":1,"name":"X","message":"no such user someone@example.com"}`
	f := withReplies(map[string]fakeReply{"sf org login jwt": {Stdout: []byte(body), ExitCode: 1}})
	p := newTestProvider(t, f, testEnv())
	_, err := p.Authenticate(context.Background(), jwtTarget())
	if err == nil || strings.Contains(err.Error(), "someone@") || !strings.Contains(err.Error(), "s***@example.com") {
		t.Errorf("email not masked: %v", err)
	}
}

// blockingExec never returns and ignores ctx, like sf stuck mid-login.
func blockingExec(started chan<- string) Exec {
	return func(_ context.Context, c Cmd) (Output, error) {
		if c.Name == "icacls" {
			return Output{}, nil
		}
		for _, kv := range c.Env {
			if v, ok := strings.CutPrefix(kv, "HOME="); ok {
				started <- v
			}
		}
		select {}
	}
}

// Question 1: a cancelled context removes the session dir even while sf is mid-flight.
func TestCancelledContextRemovesHome(t *testing.T) {
	tmp := t.TempDir()
	started := make(chan string, 1)
	p := NewWithOptions(Options{Exec: blockingExec(started), TempDir: tmp,
		LookupEnv: func(k string) (string, bool) { v, ok := testEnv()[k]; return v, ok }})
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _, _ = p.Authenticate(ctx, jwtTarget()) }()
	var home string
	select {
	case home = <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("login never started")
	}
	if _, err := os.Stat(home); err != nil {
		t.Fatalf("home should exist while sf runs: %v", err)
	}
	cancel()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(home); os.IsNotExist(err) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("home still exists after the context was cancelled")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestCancelledContextStopsAuthenticate(t *testing.T) {
	f := withReplies(map[string]fakeReply{"sf org login jwt": {Stdout: fixture(t, "login-jwt-devhub.json")}})
	p := newTestProvider(t, f, testEnv())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.Authenticate(ctx, jwtTarget()); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v", err)
	}
	if len(f.calls) != 0 || len(entries(t, p.tempDir)) != 0 {
		t.Error("a cancelled context must not start a login or leave a dir")
	}
}

func TestCloseDetachesCleanup(t *testing.T) {
	f := withReplies(map[string]fakeReply{"sf org login jwt": {Stdout: fixture(t, "login-jwt-devhub.json")}})
	p := newTestProvider(t, f, testEnv())
	ctx, cancel := context.WithCancel(context.Background())
	s, err := p.Authenticate(ctx, jwtTarget())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	cancel() // must not panic or fail after Close
}

// Question 2: a process that gets SIGTERM leaves no session dir. The parent re-runs the test
// binary as a helper that authenticates with a login that never finishes, and signals it.
func TestSIGTERMLeavesNoSessionDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows cannot send SIGTERM to another process. CI runs Linux")
	}
	tmp := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperAuthProcess$")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "HELPER_TEMPDIR="+tmp)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()

	lines := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			if v, ok := strings.CutPrefix(sc.Text(), "HOME="); ok {
				lines <- v
				return
			}
		}
		close(lines)
	}()
	var home string
	select {
	case home = <-lines:
	case <-time.After(20 * time.Second):
		t.Fatal("helper never started its login")
	}
	if home == "" || filepath.Dir(home) != tmp {
		t.Fatalf("helper home %q is not under %q", home, tmp)
	}
	if _, err := os.Stat(home); err != nil {
		t.Fatalf("home should exist before the signal: %v", err)
	}
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(home); os.IsNotExist(err) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("session dir still exists after SIGTERM")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if left := entries(t, tmp); len(left) != 0 {
		t.Errorf("temp dir not empty: %v", left)
	}
}

// TestHelperAuthProcess is the child of TestSIGTERMLeavesNoSessionDir. It is not a test of
// its own. It handles signals the way cli.Execute does.
func TestHelperAuthProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	started := make(chan string, 1)
	go func() {
		os.Stdout.WriteString("HOME=" + <-started + "\n")
	}()
	p := NewWithOptions(Options{Exec: blockingExec(started), TempDir: os.Getenv("HELPER_TEMPDIR"),
		LookupEnv: func(k string) (string, bool) { v, ok := testEnv()[k]; return v, ok }})
	_, _ = p.Authenticate(ctx, jwtTarget())
	// Authenticate only returns if the login ended. The parent kills this process.
	os.Exit(3)
}

// Synthetic value for the auth file env create writes. It differs from testSfdx, so a test
// can tell which one a login used.
const testEnvSfdx = "force://PlatformCLI::TESTENVREFRESH@example.my.salesforce.com"

// isolateTemp points the OS temp dir at a fresh dir and returns it, so a test can place an
// auth file inside or outside it.
func isolateTemp(t *testing.T) string {
	t.Helper()
	tmp := filepath.Join(t.TempDir(), "tmp")
	if err := os.MkdirAll(tmp, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(k, tmp)
	}
	return tmp
}

func writeAuth(t *testing.T, path, content string, mode os.FileMode) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
	return path
}

func authFileTarget(path string) providers.Target {
	tg := sfdxTarget()
	tg.AuthFile = path
	return tg
}

func TestAuthenticateSfdxURLFromAuthFile(t *testing.T) {
	tmp := isolateTemp(t)
	path := writeAuth(t, filepath.Join(tmp, "glassops-env-1", "sfdx-url.txt"), testEnvSfdx+"\n", 0o600)
	var used string
	exec := func(_ context.Context, c Cmd) (Output, error) {
		if c.Name == "icacls" {
			return Output{}, nil
		}
		for i, a := range c.Args {
			if a == "--sfdx-url-file" {
				data, _ := os.ReadFile(c.Args[i+1])
				used = string(data)
			}
			if strings.Contains(a, "TESTENVREFRESH") {
				t.Errorf("auth file content in argv: %s", a)
			}
		}
		return Output{Stdout: fixture(t, "login-sfdx-url-scratch.json")}, nil
	}
	// T_URL is set, and must be ignored when an auth file is given.
	p := NewWithOptions(Options{Exec: exec, LookupEnv: func(k string) (string, bool) { v, ok := testEnv()[k]; return v, ok }, TempDir: tmp})
	s, err := p.Authenticate(context.Background(), authFileTarget(path))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if used != testEnvSfdx {
		t.Fatalf("login used %q, want the auth file's url", used)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != testEnvSfdx+"\n" {
		t.Fatalf("auth file changed or removed by the login: %q %v", data, err)
	}
}

func TestAuthFileInsideGlassopsDirIsAllowed(t *testing.T) {
	isolateTemp(t)
	path := writeAuth(t, filepath.Join(t.TempDir(), "proj", ".glassops", "env", "ephemeral", "auth"), testEnvSfdx, 0o600)
	p := NewWithOptions(Options{})
	if got, err := p.readAuthFile(path); err != nil || got != testEnvSfdx {
		t.Fatalf("got %q, err %v", got, err)
	}
}

func TestAuthFileRefusals(t *testing.T) {
	type tc struct {
		name string
		path func(t *testing.T, tmp string) string
		want string
	}
	cases := []tc{
		{"outside temp and .glassops", func(t *testing.T, _ string) string {
			return writeAuth(t, filepath.Join(t.TempDir(), "elsewhere", "auth"), testEnvSfdx, 0o600)
		}, "outside the temp dir"},
		{"relative path", func(*testing.T, string) string { return filepath.Join("glassops-env-1", "auth") }, "not absolute"},
		{"missing file", func(_ *testing.T, tmp string) string { return filepath.Join(tmp, "nope") }, "auth file"},
		{"a directory", func(t *testing.T, tmp string) string {
			d := filepath.Join(tmp, "adir")
			_ = os.MkdirAll(d, 0o700)
			return d
		}, "not a regular file"},
		{"not an sfdx url", func(t *testing.T, tmp string) string {
			return writeAuth(t, filepath.Join(tmp, "auth"), "SECRETVALUE-not-a-url", 0o600)
		}, "does not hold one sfdx url"},
		{"two lines", func(t *testing.T, tmp string) string {
			return writeAuth(t, filepath.Join(tmp, "auth"), testEnvSfdx+"\n"+testEnvSfdx, 0o600)
		}, "does not hold one sfdx url"},
	}
	if !isWindows {
		cases = append(cases,
			tc{"group readable", func(t *testing.T, tmp string) string {
				return writeAuth(t, filepath.Join(tmp, "auth"), testEnvSfdx, 0o640)
			}, "want 0600"},
			tc{"symlink", func(t *testing.T, tmp string) string {
				real := writeAuth(t, filepath.Join(tmp, "real"), testEnvSfdx, 0o600)
				link := filepath.Join(tmp, "link")
				if err := os.Symlink(real, link); err != nil {
					t.Skip(err)
				}
				return link
			}, "not a regular file"})
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tmp := isolateTemp(t)
			called := false
			exec := func(context.Context, Cmd) (Output, error) { called = true; return Output{}, nil }
			p := NewWithOptions(Options{Exec: exec, LookupEnv: func(k string) (string, bool) { v, ok := testEnv()[k]; return v, ok }, TempDir: tmp})
			_, err := p.Authenticate(context.Background(), authFileTarget(c.path(t, tmp)))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err %v, want %q", err, c.want)
			}
			if strings.Contains(err.Error(), "SECRETVALUE") || strings.Contains(err.Error(), "TESTENVREFRESH") {
				t.Fatalf("content in error: %v", err)
			}
			if called {
				t.Fatal("sf ran after a refused auth file")
			}
		})
	}
}
