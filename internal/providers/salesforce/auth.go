package salesforce

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/glassops-platform/glassops/internal/core/redact"
	"github.com/glassops-platform/glassops/internal/providers"
)

const defaultLoginURL = "https://login.salesforce.com"

var envRef = regexp.MustCompile(`^\$[A-Z_][A-Z0-9_]*$`)

// login is a target's auth with every $VAR resolved. It holds secrets, so it never leaves
// this file and nothing in it is printed.
type login struct {
	method      string
	username    string
	clientID    string
	key         string
	sfdxURL     string
	token       string
	instanceURL string
}

// secrets lists the values that must never appear in an error.
func (l login) secrets() []string {
	var out []string
	for _, v := range []string{l.key, l.sfdxURL, l.token, l.clientID} {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

// resolve returns the value of ref. A $VAR is read through lookupEnv, and an unset or empty
// one is an error that names the variable. Any other value is returned as is when textOK.
func (p *Provider) resolve(field, ref string, textOK bool) (string, error) {
	if envRef.MatchString(ref) {
		name := ref[1:]
		v, ok := p.lookupEnv(name)
		if !ok || v == "" {
			return "", fmt.Errorf("auth.%s: environment variable %s is not set or is empty", field, name)
		}
		return v, nil
	}
	if ref == "" {
		return "", fmt.Errorf("auth.%s is required", field)
	}
	if !textOK {
		return "", fmt.Errorf("auth.%s must be a $VAR reference", field)
	}
	return ref, nil
}

func (p *Provider) resolveLogin(t providers.Target) (login, error) {
	a := t.Auth
	l := login{method: a.Method}
	var err error
	switch a.Method {
	case "jwt":
		if l.username, err = p.resolve("username", a.Username, true); err != nil {
			return l, err
		}
		if l.clientID, err = p.resolve("client_id", a.ClientID, false); err != nil {
			return l, err
		}
		if l.key, err = p.resolve("key", a.Key, false); err != nil {
			return l, err
		}
		l.instanceURL = defaultLoginURL
		if a.InstanceURL != "" {
			if l.instanceURL, err = p.resolve("instance_url", a.InstanceURL, true); err != nil {
				return l, err
			}
		}
	case "sfdx-url":
		if t.AuthFile != "" {
			if l.sfdxURL, err = p.readAuthFile(t.AuthFile); err != nil {
				return l, err
			}
			break
		}
		if l.sfdxURL, err = p.resolve("url", a.URL, false); err != nil {
			return l, err
		}
	case "access-token":
		if l.token, err = p.resolve("token", a.Token, false); err != nil {
			return l, err
		}
		if l.instanceURL, err = p.resolve("instance_url", a.InstanceURL, true); err != nil {
			return l, err
		}
	default:
		return l, fmt.Errorf("auth.method %q is not supported (use jwt, sfdx-url or access-token)", a.Method)
	}
	return l, nil
}

// Authenticate logs in to t in a new session. The session dir is private and holds sf's
// whole state for this login. It is removed on Close, when the login fails and when ctx
// ends, so a cancelled run leaves nothing behind.
func (p *Provider) Authenticate(ctx context.Context, t providers.Target) (providers.Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	l, err := p.resolveLogin(t)
	if err != nil {
		return nil, fmt.Errorf("target %s: %w", t.Name, err)
	}
	home, err := os.MkdirTemp(p.tempDir, "glassops-session-")
	if err != nil {
		return nil, fmt.Errorf("create session dir: %w", err)
	}
	s := &session{target: t, home: home, alias: sessionAlias}
	// A signal cancels ctx. Remove the dir even if sf is still running.
	s.stop = context.AfterFunc(ctx, func() { _ = os.RemoveAll(home) })
	fail := func(err error) (providers.Session, error) {
		_ = s.Close()
		return nil, err
	}
	if err := os.Chmod(home, 0o700); err != nil {
		return fail(fmt.Errorf("protect session dir: %w", err))
	}
	if isWindows {
		if err := p.restrictToOwner(ctx, home); err != nil {
			return fail(err)
		}
	}
	if err := p.login(ctx, s, l); err != nil {
		if ctx.Err() != nil {
			return fail(ctx.Err())
		}
		return fail(fmt.Errorf("%s login: %s", l.method, scrub(err.Error(), l.secrets())))
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	return s, nil
}

// restrictToOwner removes inherited access from dir and grants the current user alone.
// Windows ignores Unix modes, so this is its equivalent of 0700.
func (p *Provider) restrictToOwner(ctx context.Context, dir string) error {
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("find current user for session dir ACL: %w", err)
	}
	out, err := p.exec(ctx, Cmd{Name: "icacls", Args: []string{dir, "/inheritance:r", "/grant:r", "*" + u.Uid + ":(OI)(CI)F"}})
	if err != nil {
		return fmt.Errorf("restrict session dir to its owner: %w", err)
	}
	if out.ExitCode != 0 {
		return fmt.Errorf("restrict session dir to its owner: icacls exited %d", out.ExitCode)
	}
	return nil
}

func (p *Provider) login(ctx context.Context, s *session, l login) error {
	switch l.method {
	case "jwt":
		return p.loginJWT(ctx, s, l)
	case "sfdx-url":
		return p.loginSfdxURL(ctx, s, l)
	default:
		return p.loginAccessToken(ctx, s, l)
	}
}

// loginJWT writes the key into the session dir and keeps it until Close, because sf runs
// the JWT grant again with the stored key path when an access token expires.
func (p *Provider) loginJWT(ctx context.Context, s *session, l login) error {
	key, err := jwtKeyText(l.key)
	if err != nil {
		return err
	}
	keyFile := filepath.Join(s.home, "jwt.key")
	if err := os.WriteFile(keyFile, []byte(key), 0o600); err != nil {
		return fmt.Errorf("write key file: %w", err)
	}
	args := []string{"org", "login", "jwt",
		"--username", l.username,
		"--client-id", l.clientID,
		"--jwt-key-file", keyFile,
		"--instance-url", l.instanceURL,
		"--alias", s.alias,
	}
	return p.loginAttempts(ctx, func() error {
		_, err := p.sf(ctx, s, "", args...)
		return err
	})
}

// jwtKeyText returns the PEM text for v. A value that starts with a PEM header is the key
// itself, and anything else is the path of a key file. Literal "\n" sequences, which some CI
// systems produce when a multi-line secret is flattened, are turned back into newlines.
func jwtKeyText(v string) (string, error) {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, "-----BEGIN") {
		if !strings.Contains(v, "\n") {
			v = strings.ReplaceAll(v, `\n`, "\n")
		}
		return v + "\n", nil
	}
	data, err := os.ReadFile(v)
	if err != nil {
		return "", errors.New("auth.key is not PEM text and the file it names could not be read")
	}
	return string(data), nil
}

// loginSfdxURL gives sf the URL as a file, never as an argument, and deletes the file as
// soon as sf is done with it.
func (p *Provider) loginSfdxURL(ctx context.Context, s *session, l login) error {
	file := filepath.Join(s.home, "sfdx-url.txt")
	args := []string{"org", "login", "sfdx-url", "--sfdx-url-file", file, "--alias", s.alias}
	return p.loginAttempts(ctx, func() error {
		if err := os.WriteFile(file, []byte(l.sfdxURL), 0o600); err != nil {
			return fmt.Errorf("write sfdx url file: %w", err)
		}
		defer os.Remove(file)
		_, err := p.sf(ctx, s, "", args...)
		return err
	})
}

// loginAccessToken passes the token to sf in SF_ACCESS_TOKEN. Arguments are visible in the
// process list and the environment is not. p.sf takes no extra environment, so this builds
// the command the same way it does.
func (p *Provider) loginAccessToken(ctx context.Context, s *session, l login) error {
	args := []string{"org", "login", "access-token", "--instance-url", l.instanceURL, "--alias", s.alias, "--no-prompt", "--json"}
	return p.loginAttempts(ctx, func() error {
		env := append(sfBaseEnv(), s.env()...)
		env = append(env, "SF_ACCESS_TOKEN="+l.token)
		out, err := p.exec(ctx, Cmd{Name: "sf", Args: args, Env: env})
		if err != nil {
			return err
		}
		e, err := parseEnvelope(out.Stdout)
		if err != nil {
			return fmt.Errorf("sf org login access-token (exit %d): %w", out.ExitCode, err)
		}
		if e.Status != 0 {
			return &SFError{Command: "org login access-token", Name: e.Name, Message: e.Message, Status: e.Status, Data: e.Data}
		}
		return nil
	})
}

// loginAttempts runs try with the login retry policy. Only transient failures are retried.
func (p *Provider) loginAttempts(ctx context.Context, try func() error) error {
	return loginRetry.do(ctx, func(err error) bool {
		var sfe *SFError
		return errors.As(err, &sfe) && isTransient(sfe.Name+" "+sfe.Message)
	}, try)
}

// scrub removes any secret value from msg, masks emails, and collapses the repeated lines
// sf prints when it retries a JWT login.
func scrub(msg string, secrets []string) string {
	for _, sec := range secrets {
		msg = strings.ReplaceAll(msg, sec, "***")
	}
	var lines []string
	for _, ln := range strings.Split(msg, "\n") {
		if ln = strings.TrimSpace(ln); ln == "" {
			continue
		}
		if n := len(lines); n > 0 && lines[n-1] == ln {
			continue
		}
		lines = append(lines, ln)
	}
	return redact.Text(strings.Join(lines, " / "))
}

// readAuthFile reads the sfdx url from an auth file env create wrote. The file must be a
// regular file inside the temp dir or a .glassops dir, and on Unix only its owner may read
// it. Its content never reaches an error.
func (p *Provider) readAuthFile(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", errors.New("auth file: the path is not absolute")
	}
	fi, err := os.Lstat(path)
	if err != nil {
		return "", fmt.Errorf("auth file: %w", err)
	}
	if !fi.Mode().IsRegular() {
		return "", errors.New("auth file: not a regular file, refusing it")
	}
	if !isWindows && fi.Mode().Perm() != 0o600 {
		return "", fmt.Errorf("auth file: mode is %04o, want 0600, refusing it", fi.Mode().Perm())
	}
	if !p.authFilePlace(path) {
		return "", errors.New("auth file: outside the temp dir and .glassops, refusing it")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("auth file: %w", err)
	}
	v := strings.TrimSpace(string(data))
	if !strings.HasPrefix(v, "force://") || strings.ContainsAny(v, "\r\n") {
		return "", errors.New("auth file: does not hold one sfdx url")
	}
	return v, nil
}

// authFilePlace reports whether path is inside the OS temp dir, the provider's temp dir, or
// a directory named .glassops. The CLI checks the project's own .glassops more strictly.
func (p *Provider) authFilePlace(path string) bool {
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	norm := func(s string) string {
		if isWindows {
			return strings.ToLower(s)
		}
		return s
	}
	for _, root := range []string{os.TempDir(), p.tempDir} {
		if root == "" {
			continue
		}
		r, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(norm(r), norm(real))
		if err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
			return true
		}
	}
	for d := filepath.Dir(real); d != filepath.Dir(d); d = filepath.Dir(d) {
		if filepath.Base(d) == ".glassops" {
			return true
		}
	}
	return false
}
