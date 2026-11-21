package salesforce

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/providers"
)

// revokeToken is built from parts so the source never holds a token shaped literal.
var revokeToken = "00Dxx0000001gPL" + "!" + "AQ4AQFakeTokenValueThatIsLongEnough.abc"

// revokeServer is a TLS server standing in for the org. It records the revoke request.
func revokeServer(t *testing.T, status int) (*httptest.Server, *[]url.Values) {
	t.Helper()
	var forms []url.Values
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/services/oauth2/revoke" {
			t.Errorf("request %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		v, _ := url.ParseQuery(string(body))
		forms = append(forms, v)
		w.WriteHeader(status)
	}))
	old := tokenClient
	tokenClient = srv.Client()
	t.Cleanup(func() { tokenClient = old; srv.Close() })
	return srv, &forms
}

func display(token, instance string) fakeReply {
	return fakeReply{Stdout: []byte(`{"status":0,"result":{"accessToken":"` + token + `","instanceUrl":"` + instance + `","id":"00Dxx0000001gPLEAY"}}`)}
}

func TestRevokeSessionWithANewerSF(t *testing.T) {
	srv, forms := revokeServer(t, http.StatusOK)
	f := newFakeExec(map[string]fakeReply{
		"sf org display":                {Stdout: display("[use sf org auth show-access-token to see it]", srv.URL).Stdout},
		"sf org auth show-access-token": {Stdout: []byte(`{"status":0,"result":{"accessToken":"` + revokeToken + `"}}`)},
	})
	p := newTestProvider(t, f, nil)
	if err := p.RevokeSession(context.Background(), newTestSession(t, providers.Target{Name: "qa"})); err != nil {
		t.Fatal(err)
	}
	if len(*forms) != 1 || (*forms)[0].Get("token") != revokeToken {
		t.Fatalf("revoke requests = %v", *forms)
	}
}

func TestRevokeSessionWithAnOlderSF(t *testing.T) {
	srv, forms := revokeServer(t, http.StatusOK)
	f := newFakeExec(map[string]fakeReply{"sf org display": display(revokeToken, srv.URL)})
	p := newTestProvider(t, f, nil)
	if err := p.RevokeSession(context.Background(), newTestSession(t, providers.Target{Name: "qa"})); err != nil {
		t.Fatal(err)
	}
	if len(*forms) != 1 {
		t.Fatalf("revoke requests = %v", *forms)
	}
	for _, c := range f.commands() {
		if strings.Contains(c, "show-access-token") {
			t.Fatal("an org display that showed the token still asked show-access-token")
		}
	}
}

func TestRevokeSessionRefusals(t *testing.T) {
	srv, forms := revokeServer(t, http.StatusBadRequest)
	cases := map[string]struct {
		replies map[string]fakeReply
		want    string
	}{
		"no token shown": {map[string]fakeReply{
			"sf org display":                display("hidden", srv.URL),
			"sf org auth show-access-token": {Stdout: []byte(`{"status":1,"name":"CommandNotFound"}`), ExitCode: 1},
		}, "did not show"},
		"plain http instance": {map[string]fakeReply{"sf org display": display(revokeToken, "http://example.invalid")}, "not an https URL"},
		"org refuses":         {map[string]fakeReply{"sf org display": display(revokeToken, srv.URL)}, "HTTP 400"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			p := newTestProvider(t, newFakeExec(tc.replies), nil)
			err := p.RevokeSession(context.Background(), newTestSession(t, providers.Target{Name: "qa"}))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %v, want %q", err, tc.want)
			}
			if strings.Contains(err.Error(), revokeToken) {
				t.Fatal("the token reached the error")
			}
		})
	}
	if len(*forms) != 1 {
		t.Fatalf("only the refused revoke should reach the server, got %d requests", len(*forms))
	}
}
