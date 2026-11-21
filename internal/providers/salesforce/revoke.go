package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/glassops-platform/glassops/internal/providers"
)

// RevokeSession revokes the session's access token through the OAuth revoke endpoint, so a
// copy an exec step's command made of it stops working. Revoking the access token leaves the
// refresh token, and so the login file or auth URL the session came from, usable: a capture on
// a scratch org on 2026-11-21 showed the token answer 401 after the revoke and the next login
// from the same sfdx-url succeed. sf org logout is not used, because it only removes the local
// login and the token went on answering 200.
func (p *Provider) RevokeSession(ctx context.Context, s providers.Session) error {
	ss, ok := asSession(s)
	if !ok || ss.home == "" {
		return errors.New("not a Salesforce session")
	}
	env, err := p.sf(ctx, ss, "", "org", "display", "--target-org", ss.alias)
	if err != nil {
		return fmt.Errorf("read the session: %w", err)
	}
	var disp struct {
		AccessToken string `json:"accessToken"`
		InstanceURL string `json:"instanceUrl"`
	}
	if err := json.Unmarshal(env.Result, &disp); err != nil {
		return fmt.Errorf("read the session: %w", err)
	}
	token := disp.AccessToken
	// sf 2.158 and later show only a placeholder in org display and print the token through
	// org auth show-access-token. Older sf prints it in org display.
	if !looksLikeToken(token) {
		token = ""
		if env, err := p.sf(ctx, ss, "", "org", "auth", "show-access-token", "--target-org", ss.alias); err == nil {
			var shown struct {
				AccessToken string `json:"accessToken"`
			}
			if json.Unmarshal(env.Result, &shown) == nil && looksLikeToken(shown.AccessToken) {
				token = shown.AccessToken
			}
		}
	}
	if token == "" {
		return errors.New("sf did not show the session's access token, so it was not revoked")
	}
	iu, err := url.Parse(disp.InstanceURL)
	if err != nil || iu.Scheme != "https" || iu.Host == "" {
		return errors.New("the session's instance URL is not an https URL, so the token was not sent to it")
	}
	endpoint := iu.Scheme + "://" + iu.Host + "/services/oauth2/revoke"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(url.Values{"token": {token}}.Encode()))
	if err != nil {
		return fmt.Errorf("revoke: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := tokenClient.Do(req)
	if err != nil {
		// The error can name the URL but never holds the form body.
		return fmt.Errorf("revoke: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revoke: %s answered HTTP %d", iu.Host, resp.StatusCode)
	}
	return nil
}

// looksLikeToken reports whether v is shaped like a Salesforce access token, which is the org
// Id, "!", then the token. A placeholder is not.
func looksLikeToken(v string) bool {
	i := strings.IndexByte(v, '!')
	return i >= 15 && len(v) > i+20 && strings.HasPrefix(v, "00D")
}
