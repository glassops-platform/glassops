package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/internal/providers"
)

// Scratch org creation never runs "sf org create scratch". With a Dev Hub logged in by JWT
// through an external client app, that command signs the org up and then fails its own
// follow-up login, which leaves a live org nobody can reach. CreateEnv instead creates the
// ScratchOrgInfo record itself, waits for it, exchanges the record's auth code for a refresh
// token and logs in to the new org with an sfdx url.

const (
	// scratchClientID is the connected app the auth code is issued to. PlatformCLI is the
	// one sf itself uses.
	scratchClientID = "PlatformCLI"
	// scratchCallbackURL must match the callback the auth code was requested with.
	scratchCallbackURL = "http://localhost:1717/OauthRedirect"
	// orgNameMax is the length limit of ScratchOrgInfo.OrgName.
	orgNameMax = 80
	// maxScratchDays is the longest a scratch org can live.
	maxScratchDays = 30
	// envDirPrefix starts the private dir that holds a created org's auth file.
	envDirPrefix = "glassops-env-"
	// envAuthFile is the name of the auth file inside that dir.
	envAuthFile = "sfdx-url.txt"
	// sfDateTime is how sf prints a datetime field, such as 2026-10-31T03:39:25.000+0000.
	sfDateTime = "2006-01-02T15:04:05.000-0700"
)

// envPolicy controls signup retries and polling. Tests replace envRetry to skip the waits.
type envPolicy struct {
	// Attempts is the number of signups tried, including the first.
	Attempts int
	// Backoff is the wait before the second signup. It doubles for each later one.
	Backoff time.Duration
	// PollEvery is the wait between two reads of the ScratchOrgInfo record.
	PollEvery time.Duration
	// PollFor is how long one signup may take before it counts as failed.
	PollFor time.Duration
	// Sleep waits for d or until ctx ends.
	Sleep func(ctx context.Context, d time.Duration) error
	// Now is the clock. Polling deadlines and sweep ages use it.
	Now func() time.Time
}

var envRetry = envPolicy{
	Attempts:  3,
	Backoff:   30 * time.Second,
	PollEvery: 10 * time.Second,
	PollFor:   15 * time.Minute,
	Sleep:     sleepCtx,
	Now:       time.Now,
}

// tokenClient sends the auth code exchange. Tests replace it with a TLS test server's client.
var tokenClient = &http.Client{Timeout: time.Minute}

var (
	sfID       = regexp.MustCompile(`^[A-Za-z0-9]{15}([A-Za-z0-9]{3})?$`)
	envNameOK  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	valueBlock = regexp.MustCompile(`['"\r\n]`)
	oauthCode  = regexp.MustCompile(`^[a-z_]+$`)
)

// signupError is a signup that failed on the Dev Hub side without producing an org: an
// ErrorCode on the record, or a request refused on capacity. Another signup can succeed, so
// CreateEnv retries it.
type signupError struct{ msg string }

func (e *signupError) Error() string { return e.msg }

// capacityWords mark a refused signup request that can succeed later, such as the active
// scratch org limit being reached while another run deletes its org.
var capacityWords = []string{"limit_exceeded", "scratch org limit", "capacity"}

// CreateEnv creates a scratch org from the Dev Hub session hub. On an error after the
// ScratchOrgInfo record exists, the returned Env still holds RecordID, and ID once the org is
// known, so the caller can record them and delete the org later.
func (p *Provider) CreateEnv(ctx context.Context, hub providers.Session, req providers.EnvRequest) (providers.Env, error) {
	sess, ok := asSession(hub)
	if !ok {
		return providers.Env{}, errors.New("salesforce: env create needs a session from this provider")
	}
	values, name, facts, err := signupValues(req)
	if err != nil {
		return providers.Env{}, err
	}

	var env providers.Env
	var sfdxURL string
	wait := envRetry.Backoff
	for attempt := 1; ; attempt++ {
		facts["salesforce.signup_attempts"] = attempt
		env, sfdxURL, err = p.signup(ctx, sess, values, name)
		env.Facts = facts
		if err == nil {
			break
		}
		var se *signupError
		if !errors.As(err, &se) || attempt >= envRetry.Attempts || ctx.Err() != nil {
			return env, err
		}
		if serr := envRetry.Sleep(ctx, wait); serr != nil {
			return env, serr
		}
		wait *= 2
	}

	if env.AuthFile, err = p.writeAuthFile(ctx, sfdxURL); err != nil {
		return env, err
	}
	if err := p.checkEnvLogin(ctx, env); err != nil {
		return env, err
	}
	return env, nil
}

// signupValues builds the --values text for the ScratchOrgInfo record and the org name.
func signupValues(req providers.EnvRequest) (string, string, map[string]any, error) {
	name := req.Name
	if len(name) > orgNameMax {
		name = strings.TrimRight(name[:orgNameMax], "-._")
	}
	if !envNameOK.MatchString(name) {
		return "", "", nil, fmt.Errorf("salesforce: env name %q may hold only letters, digits, dot, dash and underscore", req.Name)
	}
	days := int((req.TTL + 24*time.Hour - 1) / (24 * time.Hour))
	if req.TTL <= 0 || days > maxScratchDays {
		return "", "", nil, fmt.Errorf("salesforce: a scratch org lives 1 to %d days, not %s", maxScratchDays, req.TTL)
	}

	fields, err := readDefinition(req.Dir, req.DefinitionFile)
	if err != nil {
		return "", "", nil, err
	}
	facts := map[string]any{}
	if req.Shape != "" {
		shape := strings.TrimSpace(req.Shape)
		if !strings.HasPrefix(shape, "00D") || !sfID.MatchString(shape) {
			return "", "", nil, errors.New("salesforce: the org shape must be an org Id, 15 or 18 characters starting 00D")
		}
		// An org shape replaces the edition. The record names the 15-character source org.
		delete(fields, "Edition")
		fields["SourceOrg"] = shape[:15]
	} else if fields["Edition"] == "" && fields["SourceOrg"] == "" {
		fields["Edition"] = "Developer"
	}
	fields["OrgName"] = name
	fields["DurationDays"] = fmt.Sprint(days)
	fields["ConnectedAppConsumerKey"] = scratchClientID
	fields["ConnectedAppCallbackUrl"] = scratchCallbackURL
	facts["salesforce.shape"] = fields["SourceOrg"] != ""
	if e := fields["Edition"]; e != "" {
		facts["salesforce.edition"] = e
	}
	facts["salesforce.duration_days"] = days

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := fields[k]
		if valueBlock.MatchString(v) {
			return "", "", nil, fmt.Errorf("salesforce: the value for %s holds a quote or a line break", k)
		}
		if strings.ContainsAny(v, " \t") {
			v = "'" + v + "'"
		}
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, " "), name, facts, nil
}

// definitionFields maps the definition file keys env create supports to ScratchOrgInfo
// fields. orgName is skipped on purpose: the org is named after the environment, so that
// sweep and delete can find it.
var definitionFields = map[string]string{
	"edition":       "Edition",
	"features":      "Features",
	"hasSampleData": "HasSampleData",
	"language":      "Language",
	"country":       "Country",
	"description":   "Description",
	"release":       "Release",
	"sourceOrg":     "SourceOrg",
}

// readDefinition reads the scratch org definition file into ScratchOrgInfo fields. An empty
// path means no file. A key env create cannot apply, such as settings, is an error, so an
// org is never created without part of what the file asks for.
func readDefinition(dir, rel string) (map[string]string, error) {
	fields := map[string]string{}
	if rel == "" {
		return fields, nil
	}
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		return nil, fmt.Errorf("salesforce: read definition file: %w", err)
	}
	var def map[string]any
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("salesforce: parse definition file %s: %w", rel, err)
	}
	var unsupported []string
	for k, v := range def {
		if k == "orgName" {
			continue
		}
		field, ok := definitionFields[k]
		if !ok {
			unsupported = append(unsupported, k)
			continue
		}
		switch t := v.(type) {
		case string:
			fields[field] = t
		case bool:
			fields[field] = fmt.Sprint(t)
		case []any:
			items := make([]string, 0, len(t))
			for _, it := range t {
				s, ok := it.(string)
				if !ok {
					return nil, fmt.Errorf("salesforce: definition key %s must list strings", k)
				}
				items = append(items, s)
			}
			fields[field] = strings.Join(items, ";")
		default:
			return nil, fmt.Errorf("salesforce: definition key %s has an unsupported value", k)
		}
	}
	if len(unsupported) > 0 {
		sort.Strings(unsupported)
		return nil, fmt.Errorf("salesforce: env create cannot apply definition keys %s", strings.Join(unsupported, ", "))
	}
	return fields, nil
}

// signup creates one ScratchOrgInfo record, waits until it is Active and exchanges its auth
// code. It returns the sfdx url for the new org, which is secret.
func (p *Provider) signup(ctx context.Context, sess *session, values, name string) (providers.Env, string, error) {
	env := providers.Env{Name: name}
	out, err := p.sf(ctx, sess, "", "data", "create", "record", "--sobject", "ScratchOrgInfo",
		"--values", values, "--target-org", sess.alias)
	if err != nil {
		if ctx.Err() != nil {
			return env, "", ctx.Err()
		}
		var sfe *SFError
		if errors.As(err, &sfe) && retryableSignup(sfe.Name+" "+sfe.Message) {
			return env, "", &signupError{msg: "salesforce: the Dev Hub refused the signup: " + scrub(sfe.Message, nil)}
		}
		return env, "", fmt.Errorf("salesforce: create the ScratchOrgInfo record: %s", scrub(err.Error(), nil))
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(out.Result, &created); err != nil {
		return env, "", errors.New("salesforce: could not read the ScratchOrgInfo create result")
	}
	// Keep the record Id before checking it, so the caller can still record it and find the
	// org if anything below fails. Delete checks it again before it uses it.
	env.RecordID = created.ID
	if created.ID == "" {
		return env, "", errors.New("salesforce: no ScratchOrgInfo record Id came back")
	}
	if !sfID.MatchString(created.ID) {
		return env, "", errors.New("salesforce: the ScratchOrgInfo record Id is not 15 or 18 letters and digits")
	}

	deadline := envRetry.Now().Add(envRetry.PollFor)
	var lastErr error
	for {
		rec, err := p.readSignup(ctx, sess, created.ID)
		switch {
		case err != nil:
			if ctx.Err() != nil {
				return env, "", ctx.Err()
			}
			lastErr = err
		case rec.Status == "Active":
			if !strings.HasPrefix(rec.ScratchOrg, "00D") || !sfID.MatchString(rec.ScratchOrg) {
				return env, "", errors.New("salesforce: the active signup reports no valid org Id")
			}
			// Keep the org Id before the token exchange, so delete can find the org even
			// when the exchange fails.
			env.ID = rec.ScratchOrg
			if t, perr := time.Parse("2006-01-02", rec.ExpirationDate); perr == nil {
				env.ExpiresAt = t.UTC()
			}
			if rec.AuthCode == "" || rec.LoginURL == "" {
				return env, "", errors.New("salesforce: the active signup has no auth code or login url")
			}
			token, host, err := exchangeAuthCode(ctx, rec.LoginURL, rec.AuthCode)
			if err != nil {
				return env, "", err
			}
			return env, "force://" + scratchClientID + "::" + token + "@" + host, nil
		case rec.Status == "Error":
			code := rec.ErrorCode
			if code == "" {
				code = "unknown"
			}
			return env, "", &signupError{msg: "salesforce: signup failed with ErrorCode " + code}
		}
		if !envRetry.Now().Before(deadline) {
			// Not retried: this signup can still turn Active, and a second one would leave
			// two orgs where the caller records one.
			msg := fmt.Sprintf("salesforce: the scratch org was not active after %s", envRetry.PollFor)
			if lastErr != nil {
				msg += ", last read error: " + scrub(lastErr.Error(), nil)
			}
			return env, "", errors.New(msg)
		}
		if err := envRetry.Sleep(ctx, envRetry.PollEvery); err != nil {
			return env, "", err
		}
	}
}

func retryableSignup(msg string) bool {
	if isTransient(msg) {
		return true
	}
	low := strings.ToLower(msg)
	for _, w := range capacityWords {
		if strings.Contains(low, w) {
			return true
		}
	}
	return false
}

// signupRecord is the ScratchOrgInfo row CreateEnv polls. AuthCode is secret and is read
// only by the token exchange.
type signupRecord struct {
	Status         string `json:"Status"`
	ErrorCode      string `json:"ErrorCode"`
	ScratchOrg     string `json:"ScratchOrg"`
	LoginURL       string `json:"LoginUrl"`
	AuthCode       string `json:"AuthCode"`
	ExpirationDate string `json:"ExpirationDate"`
}

// readSignup reads the ScratchOrgInfo record id. id must already be checked.
func (p *Provider) readSignup(ctx context.Context, sess *session, id string) (signupRecord, error) {
	q := "SELECT Status, ErrorCode, ScratchOrg, LoginUrl, AuthCode, ExpirationDate FROM ScratchOrgInfo WHERE Id = '" + id + "'"
	out, err := p.sf(ctx, sess, "", "data", "query", "--query", q, "--target-org", sess.alias)
	if err != nil {
		return signupRecord{}, err
	}
	var res struct {
		Records []signupRecord `json:"records"`
	}
	if err := json.Unmarshal(out.Result, &res); err != nil {
		return signupRecord{}, errors.New("could not read the ScratchOrgInfo query")
	}
	if len(res.Records) != 1 {
		return signupRecord{}, fmt.Errorf("the ScratchOrgInfo query returned %d records", len(res.Records))
	}
	return res.Records[0], nil
}

// exchangeAuthCode trades the signup's auth code for a refresh token at loginURL and returns
// the token and the instance host. The code and the token never reach an error message.
func exchangeAuthCode(ctx context.Context, loginURL, code string) (string, string, error) {
	u, err := url.Parse(strings.TrimRight(loginURL, "/"))
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", "", errors.New("salesforce: the signup login url is not an https url")
	}
	form := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"client_id":    {scratchClientID},
		"redirect_uri": {scratchCallbackURL},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String()+"/services/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", errors.New("salesforce: could not build the token exchange request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	res, err := tokenClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", "", ctx.Err()
		}
		// A *url.Error names the url. Keep only the cause, so the org host stays out of logs.
		var ue *url.Error
		if errors.As(err, &ue) {
			err = ue.Err
		}
		return "", "", fmt.Errorf("salesforce: token exchange failed: %s", scrub(err.Error(), []string{code}))
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	var t struct {
		RefreshToken string `json:"refresh_token"`
		InstanceURL  string `json:"instance_url"`
		Error        string `json:"error"`
	}
	_ = json.Unmarshal(body, &t)
	if res.StatusCode != http.StatusOK || t.RefreshToken == "" || t.InstanceURL == "" {
		msg := fmt.Sprintf("salesforce: token exchange failed, status %d", res.StatusCode)
		if oauthCode.MatchString(t.Error) {
			msg += ", " + t.Error
		}
		return "", "", errors.New(msg)
	}
	iu, err := url.Parse(t.InstanceURL)
	if err != nil || iu.Scheme != "https" || iu.Host == "" {
		return "", "", errors.New("salesforce: the token exchange returned an instance url that is not https")
	}
	return t.RefreshToken, iu.Host, nil
}

// writeAuthFile writes the sfdx url to a 0600 file in a new private dir and returns its
// path. On Windows the dir is restricted to its owner, as the session dirs are.
func (p *Provider) writeAuthFile(ctx context.Context, sfdxURL string) (string, error) {
	if sfdxURL == "" {
		return "", errors.New("salesforce: no login material came back from the signup")
	}
	dir, err := os.MkdirTemp(p.tempDir, envDirPrefix)
	if err != nil {
		return "", fmt.Errorf("salesforce: create the auth dir: %w", err)
	}
	fail := func(err error) (string, error) {
		_ = os.RemoveAll(dir)
		return "", err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fail(fmt.Errorf("salesforce: protect the auth dir: %w", err))
	}
	if isWindows {
		if err := p.restrictToOwner(ctx, dir); err != nil {
			return fail(err)
		}
	}
	file := filepath.Join(dir, envAuthFile)
	if err := os.WriteFile(file, []byte(sfdxURL+"\n"), 0o600); err != nil {
		return fail(fmt.Errorf("salesforce: write the auth file: %w", err))
	}
	return file, nil
}

// checkEnvLogin logs in to the new org with the auth file, in a throwaway home, and checks
// that the org it reaches is the one the signup made.
func (p *Provider) checkEnvLogin(ctx context.Context, env providers.Env) error {
	home, err := os.MkdirTemp(p.tempDir, "glassops-session-")
	if err != nil {
		return fmt.Errorf("salesforce: create the login dir: %w", err)
	}
	defer os.RemoveAll(home)
	if err := os.Chmod(home, 0o700); err != nil {
		return fmt.Errorf("salesforce: protect the login dir: %w", err)
	}
	s := &session{home: home, alias: sessionAlias}
	args := []string{"org", "login", "sfdx-url", "--sfdx-url-file", env.AuthFile, "--alias", s.alias}
	var got string
	err = p.loginAttempts(ctx, func() error {
		out, err := p.sf(ctx, s, "", args...)
		if err != nil {
			return err
		}
		var r struct {
			OrgID string `json:"orgId"`
		}
		if err := json.Unmarshal(out.Result, &r); err != nil {
			return errors.New("could not read the login result")
		}
		got = r.OrgID
		return nil
	})
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("salesforce: log in to the new scratch org: %s", scrub(err.Error(), nil))
	}
	if got == "" || p.CanonicalID(got) != p.CanonicalID(env.ID) {
		return fmt.Errorf("salesforce: the auth file reaches org %q, not the new scratch org %s", got, env.ID)
	}
	return nil
}

// DeleteEnv deletes the scratch org env names from the Dev Hub session hub. It finds the org
// by env.ID, else through the ScratchOrgInfo record in env.RecordID, else by env.Name. An
// org that is already gone, or a signup that never produced one, is not an error.
func (p *Provider) DeleteEnv(ctx context.Context, hub providers.Session, env providers.Env) error {
	sess, ok := asSession(hub)
	if !ok {
		return errors.New("salesforce: env delete needs a session from this provider")
	}
	var where string
	switch {
	case env.ID != "":
		if !strings.HasPrefix(env.ID, "00D") || !sfID.MatchString(env.ID) {
			return errors.New("salesforce: refusing to delete, the org Id is not 15 or 18 letters and digits starting 00D")
		}
		// ActiveScratchOrg.ScratchOrg holds the 15-character org Id.
		where = "ScratchOrg = '" + env.ID[:15] + "'"
	case env.RecordID != "":
		if !sfID.MatchString(env.RecordID) {
			return errors.New("salesforce: refusing to delete, the ScratchOrgInfo record Id is not 15 or 18 letters and digits")
		}
		// The run may have stopped while the signup was still going. It can still turn
		// Active afterwards, so ask the record which org it produced.
		rec, err := p.readSignup(ctx, sess, env.RecordID)
		if err != nil {
			return fmt.Errorf("salesforce: read the ScratchOrgInfo record: %s", scrub(err.Error(), nil))
		}
		if rec.ScratchOrg == "" {
			p.removeAuthFile(env.AuthFile)
			return nil
		}
		if !strings.HasPrefix(rec.ScratchOrg, "00D") || !sfID.MatchString(rec.ScratchOrg) {
			return errors.New("salesforce: refusing to delete, the signup reports an org Id that is not valid")
		}
		where = "ScratchOrg = '" + rec.ScratchOrg[:15] + "'"
	case env.Name != "":
		if !envNameOK.MatchString(env.Name) {
			return errors.New("salesforce: refusing to delete, the env name has unexpected characters")
		}
		where = "OrgName = '" + env.Name + "'"
	default:
		return errors.New("salesforce: refusing to delete, there is no org Id, record Id or name")
	}

	rows, err := p.activeScratchOrgs(ctx, sess, where)
	if err != nil {
		return err
	}
	switch len(rows) {
	case 0:
		p.removeAuthFile(env.AuthFile)
		return nil
	case 1:
	default:
		return fmt.Errorf("salesforce: refusing to delete, %d active scratch orgs matched", len(rows))
	}
	if err := p.deleteActive(ctx, sess, rows[0].ID); err != nil {
		return err
	}
	p.removeAuthFile(env.AuthFile)
	return nil
}

// activeRow is one ActiveScratchOrg record. The query never selects SignupUsername.
type activeRow struct {
	ID               string `json:"Id"`
	ScratchOrg       string `json:"ScratchOrg"`
	OrgName          string `json:"OrgName"`
	CreatedDate      string `json:"CreatedDate"`
	ExpirationDate   string `json:"ExpirationDate"`
	ScratchOrgInfoID string `json:"ScratchOrgInfoId"`
}

const activeFields = "Id, ScratchOrg, OrgName, CreatedDate, ExpirationDate, ScratchOrgInfoId"

// activeScratchOrgs queries ActiveScratchOrg. where must be built from checked values only.
func (p *Provider) activeScratchOrgs(ctx context.Context, sess *session, where string) ([]activeRow, error) {
	q := "SELECT " + activeFields + " FROM ActiveScratchOrg"
	if where != "" {
		q += " WHERE " + where
	}
	out, err := p.sf(ctx, sess, "", "data", "query", "--query", q, "--target-org", sess.alias)
	if err != nil {
		return nil, fmt.Errorf("salesforce: query ActiveScratchOrg: %s", scrub(err.Error(), nil))
	}
	var res struct {
		Records []activeRow `json:"records"`
	}
	if err := json.Unmarshal(out.Result, &res); err != nil {
		return nil, errors.New("salesforce: could not read the ActiveScratchOrg query")
	}
	return res.Records, nil
}

func (p *Provider) deleteActive(ctx context.Context, sess *session, id string) error {
	if !sfID.MatchString(id) {
		return errors.New("salesforce: refusing to delete, the ActiveScratchOrg record Id is not valid")
	}
	if _, err := p.sf(ctx, sess, "", "data", "delete", "record", "--sobject", "ActiveScratchOrg",
		"--record-id", id, "--target-org", sess.alias); err != nil {
		return fmt.Errorf("salesforce: delete ActiveScratchOrg %s: %s", id, scrub(err.Error(), nil))
	}
	return nil
}

// removeAuthFile removes an auth file CreateEnv wrote, with its dir. A path that is not in a
// dir CreateEnv made is left alone.
func (p *Provider) removeAuthFile(path string) {
	if path == "" || filepath.Base(path) != envAuthFile {
		return
	}
	dir := filepath.Dir(path)
	if !strings.HasPrefix(filepath.Base(dir), envDirPrefix) {
		return
	}
	_ = os.RemoveAll(dir)
}

// SweepEnvs deletes the active scratch orgs whose name starts with prefix and a dash and
// that were created more than maxAge ago, and returns the ones it deleted. A row it cannot
// read, or a delete that fails, is reported in the error after the rest are done.
func (p *Provider) SweepEnvs(ctx context.Context, hub providers.Session, prefix string, maxAge time.Duration) ([]providers.Env, error) {
	sess, ok := asSession(hub)
	if !ok {
		return nil, errors.New("salesforce: env sweep needs a session from this provider")
	}
	if !envNameOK.MatchString(prefix) {
		return nil, errors.New("salesforce: env sweep needs a name prefix of letters, digits, dot, dash or underscore")
	}
	if maxAge <= 0 {
		return nil, errors.New("salesforce: env sweep needs a max age above zero")
	}
	if !strings.HasSuffix(prefix, "-") {
		prefix += "-"
	}
	rows, err := p.activeScratchOrgs(ctx, sess, "")
	if err != nil {
		return nil, err
	}
	cutoff := envRetry.Now().Add(-maxAge)
	var deleted []providers.Env
	var errs []error
	for _, r := range rows {
		if !strings.HasPrefix(r.OrgName, prefix) {
			continue
		}
		created, perr := time.Parse(sfDateTime, r.CreatedDate)
		if perr != nil {
			errs = append(errs, fmt.Errorf("salesforce: %s has a CreatedDate that cannot be read, so it was left alone", r.OrgName))
			continue
		}
		if !created.Before(cutoff) {
			continue
		}
		if err := p.deleteActive(ctx, sess, r.ID); err != nil {
			errs = append(errs, err)
			if ctx.Err() != nil {
				break
			}
			continue
		}
		e := providers.Env{ID: r.ScratchOrg, RecordID: r.ScratchOrgInfoID, Name: r.OrgName}
		if t, perr := time.Parse("2006-01-02", r.ExpirationDate); perr == nil {
			e.ExpiresAt = t.UTC()
		}
		deleted = append(deleted, e)
	}
	return deleted, errors.Join(errs...)
}
