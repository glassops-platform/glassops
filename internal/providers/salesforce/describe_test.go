package salesforce

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/providers"
)

// companyCanary is the company name the login fixtures carry. It must never appear in a fact
// or an error.
const companyCanary = "Example Company"

func describeFake(t *testing.T, queryFile, displayFile string) *fakeExec {
	t.Helper()
	return newFakeExec(map[string]fakeReply{
		"sf data query":  {Stdout: fixture(t, queryFile)},
		"sf org display": {Stdout: fixture(t, displayFile)},
	})
}

func describeProvider(t *testing.T, f *fakeExec) *Provider {
	t.Helper()
	p := newTestProvider(t, f, nil)
	p.run = func(name string, args ...string) ([]byte, error) {
		return []byte("@salesforce/cli/2.155.7 linux-x64 node-v22.23.4\n"), nil
	}
	return p
}

func describe(t *testing.T, f *fakeExec) (providers.TargetInfo, error) {
	t.Helper()
	return describeProvider(t, f).Describe(context.Background(), newTestSession(t, providers.Target{Name: "x"}))
}

func TestDescribeKinds(t *testing.T) {
	cases := []struct {
		query, display string
		kind           providers.Kind
		id             string
		sandbox        bool
	}{
		{"query-organization-devhub.json", "org-display-devhub.json", providers.KindProduction, "00D000000000001AAA", false},
		{"query-organization-scratch.json", "org-display-scratch.json", providers.KindScratch, "00D000000000002AAA", true},
	}
	for _, c := range cases {
		t.Run(string(c.kind), func(t *testing.T) {
			f := describeFake(t, c.query, c.display)
			info, err := describe(t, f)
			if err != nil {
				t.Fatal(err)
			}
			if info.ID != c.id || info.Kind != c.kind {
				t.Fatalf("info: %+v", info)
			}
			want := map[string]any{
				"salesforce.org_id":        c.id,
				"salesforce.org_type":      "Developer Edition",
				"salesforce.is_sandbox":    c.sandbox,
				"salesforce.instance_name": "XX000",
				"salesforce.api_version":   "68.0",
				"salesforce.sf_version":    "2.155.7",
				"salesforce.username":      "u***@example.com",
			}
			for k, v := range want {
				if info.Facts[k] != v {
					t.Errorf("fact %s = %v, want %v", k, info.Facts[k], v)
				}
			}
			if len(info.Facts) != len(want) {
				t.Errorf("facts: %v", info.Facts)
			}
			if len(f.commands()) != 2 {
				t.Errorf("calls: %v", f.commands())
			}
			for _, cmd := range f.commands() {
				if !strings.Contains(cmd, "--target-org "+sessionAlias) {
					t.Errorf("call without the session alias: %s", cmd)
				}
				if strings.Contains(cmd, "Name") && !strings.Contains(cmd, "InstanceName") {
					t.Errorf("command mentions Name: %s", cmd)
				}
				if strings.Contains(cmd, ", Name") || strings.Contains(cmd, "SELECT Name") {
					t.Errorf("query selects Name: %s", cmd)
				}
			}
			if !strings.Contains(f.commands()[0], "FROM Organization") {
				t.Errorf("first call: %s", f.commands()[0])
			}
		})
	}
}

func TestDescribeSandboxHasNoTrialDate(t *testing.T) {
	q := strings.Replace(string(fixture(t, "query-organization-scratch.json")), `"2026-10-31T01:43:32.000+0000"`, "null", 1)
	f := newFakeExec(map[string]fakeReply{
		"sf data query":  {Stdout: []byte(q)},
		"sf org display": {Stdout: fixture(t, "org-display-scratch.json")},
	})
	info, err := describe(t, f)
	if err != nil || info.Kind != providers.KindSandbox {
		t.Fatalf("kind %v, err %v", info.Kind, err)
	}
}

func TestDescribeUnreadableIsAnError(t *testing.T) {
	display := fixture(t, "org-display-devhub.json")
	cases := map[string]string{
		"no records":      `{"status":0,"result":{"records":[],"totalSize":0,"done":true}}`,
		"two records":     `{"status":0,"result":{"records":[{"Id":"a","IsSandbox":false},{"Id":"b","IsSandbox":false}]}}`,
		"missing sandbox": `{"status":0,"result":{"records":[{"Id":"00D000000000001AAA"}]}}`,
		"missing id":      `{"status":0,"result":{"records":[{"IsSandbox":false}]}}`,
		"bad result":      `{"status":0,"result":"nope"}`,
		"sf error":        `{"status":1,"name":"INVALID_SESSION_ID","message":"Session expired for user@example.com"}`,
		"no json":         `Error: boom`,
	}
	for name, q := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFakeExec(map[string]fakeReply{
				"sf data query":  {Stdout: []byte(q)},
				"sf org display": {Stdout: display},
			})
			info, err := describe(t, f)
			if err == nil {
				t.Fatalf("no error, info %+v", info)
			}
			if info.ID != "" || info.Kind != "" {
				t.Fatalf("partial info returned with an error: %+v", info)
			}
			if strings.Contains(err.Error(), "user@example.com") {
				t.Errorf("email not masked: %v", err)
			}
		})
	}
}

func TestDescribeDisplayFailures(t *testing.T) {
	for name, display := range map[string]string{
		"sf error": `{"status":1,"name":"NamedOrgNotFoundError","message":"nope"}`,
		"bad body": `{"status":0,"result":[1]}`,
	} {
		t.Run(name, func(t *testing.T) {
			f := newFakeExec(map[string]fakeReply{
				"sf data query":  {Stdout: fixture(t, "query-organization-devhub.json")},
				"sf org display": {Stdout: []byte(display)},
			})
			if _, err := describe(t, f); err == nil {
				t.Fatal("no error")
			}
		})
	}
}

func TestDescribeExecError(t *testing.T) {
	f := newFakeExec(map[string]fakeReply{"sf data query": {Err: errors.New("exec failed")}})
	if _, err := describe(t, f); err == nil {
		t.Fatal("no error")
	}
}

type foreignSession struct{}

func (foreignSession) Target() providers.Target { return providers.Target{} }
func (foreignSession) Close() error             { return nil }

func TestDescribeRejectsForeignSession(t *testing.T) {
	p := describeProvider(t, newFakeExec(nil))
	if _, err := p.Describe(context.Background(), foreignSession{}); err == nil {
		t.Fatal("no error")
	}
}

func TestDescribeToleratesMissingSfVersion(t *testing.T) {
	f := describeFake(t, "query-organization-devhub.json", "org-display-devhub.json")
	p := describeProvider(t, f)
	p.run = func(string, ...string) ([]byte, error) { return nil, errors.New("no sf") }
	info, err := p.Describe(context.Background(), newTestSession(t, providers.Target{}))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := info.Facts["salesforce.sf_version"]; ok {
		t.Fatal("sf_version set without a version")
	}
}

func TestDescribeNeverLeaksCompanyName(t *testing.T) {
	login := string(fixture(t, "login-jwt-devhub.json"))
	if !strings.Contains(login, companyCanary) {
		t.Fatal("canary missing from the login fixture")
	}
	// Replies that carry the company name in the fields a careless parser would pick up.
	f := newFakeExec(map[string]fakeReply{
		"sf data query":  {Stdout: []byte(`{"status":0,"result":{"records":[{"Id":"00D000000000001AAA","IsSandbox":false,"OrganizationType":"Developer Edition","Name":"` + companyCanary + `"}]}}`)},
		"sf org display": {Stdout: []byte(`{"status":0,"result":{"apiVersion":"68.0","username":"u@example.com","name":"` + companyCanary + `"}}`)},
	})
	info, err := describe(t, f)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(info.ID, companyCanary) {
		t.Error("id leaks the company name")
	}
	for k, v := range info.Facts {
		if strings.Contains(fmt.Sprint(v), companyCanary) {
			t.Errorf("fact %s leaks the company name", k)
		}
	}
	// A failure whose reply holds the name outside the message.
	f = newFakeExec(map[string]fakeReply{
		"sf data query": {Stdout: []byte(`{"status":0,"result":{"records":[],"name":"` + companyCanary + `"}}`)},
	})
	if _, err = describe(t, f); err == nil || strings.Contains(err.Error(), companyCanary) {
		t.Fatalf("error: %v", err)
	}
}
