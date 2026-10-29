package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/glassops-platform/glassops/internal/core/redact"
	"github.com/glassops-platform/glassops/internal/providers"
)

// orgQuery reads only what the guard and the facts need. It never selects Name, because the
// company name must not reach a result or a log.
const orgQuery = "SELECT Id, IsSandbox, OrganizationType, InstanceName, TrialExpirationDate FROM Organization"

type orgRecord struct {
	ID                  string  `json:"Id"`
	IsSandbox           *bool   `json:"IsSandbox"`
	OrganizationType    string  `json:"OrganizationType"`
	InstanceName        string  `json:"InstanceName"`
	TrialExpirationDate *string `json:"TrialExpirationDate"`
}

// Describe reads the logged-in org: its Id, kind and a few facts. Anything it cannot read is
// an error, never KindUnknown, so the guard never runs on a guess.
func (p *Provider) Describe(ctx context.Context, s providers.Session) (providers.TargetInfo, error) {
	ss, ok := asSession(s)
	if !ok {
		return providers.TargetInfo{}, errors.New("salesforce: session was not made by this provider")
	}

	env, err := p.sf(ctx, ss, "", "data", "query", "--query", orgQuery, "--target-org", ss.alias)
	if err != nil {
		return providers.TargetInfo{}, describeErr("query the organization", err)
	}
	var q struct {
		Records []orgRecord `json:"records"`
	}
	if err := json.Unmarshal(env.Result, &q); err != nil {
		return providers.TargetInfo{}, fmt.Errorf("salesforce: could not read the organization query: %w", err)
	}
	if len(q.Records) != 1 {
		return providers.TargetInfo{}, fmt.Errorf("salesforce: the organization query returned %d records, want 1", len(q.Records))
	}
	rec := q.Records[0]
	if rec.ID == "" || rec.IsSandbox == nil {
		return providers.TargetInfo{}, errors.New("salesforce: the organization query is missing Id or IsSandbox")
	}

	kind := providers.KindProduction
	switch {
	case !*rec.IsSandbox:
	case rec.TrialExpirationDate != nil && *rec.TrialExpirationDate != "":
		kind = providers.KindScratch
	default:
		kind = providers.KindSandbox
	}

	env, err = p.sf(ctx, ss, "", "org", "display", "--target-org", ss.alias)
	if err != nil {
		return providers.TargetInfo{}, describeErr("display the org", err)
	}
	var disp struct {
		APIVersion string `json:"apiVersion"`
		Username   string `json:"username"`
	}
	if err := json.Unmarshal(env.Result, &disp); err != nil {
		return providers.TargetInfo{}, fmt.Errorf("salesforce: could not read org display: %w", err)
	}

	facts := map[string]any{
		"salesforce.org_id":        rec.ID,
		"salesforce.org_type":      rec.OrganizationType,
		"salesforce.is_sandbox":    *rec.IsSandbox,
		"salesforce.instance_name": rec.InstanceName,
		"salesforce.api_version":   disp.APIVersion,
		"salesforce.username":      redact.Username(disp.Username),
	}
	if v, verr := p.Tools()[0].Version(); verr == nil {
		facts["salesforce.sf_version"] = v
	}
	return providers.TargetInfo{ID: rec.ID, Kind: kind, Facts: facts}, nil
}

// describeErr wraps an sf failure with what was being read, masking any email in it.
func describeErr(what string, err error) error {
	return fmt.Errorf("salesforce: could not %s: %s", what, redact.Text(err.Error()))
}
