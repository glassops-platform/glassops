package salesforce

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

// Verify check names. Each is one entry in VerifyResult.Checks.
const (
	CheckVerifyStatus    = "verify.status"
	CheckVerifyCheckOnly = "verify.check_only"
	CheckVerifyMembers   = "verify.members"
	CheckVerifyTests     = "verify.tests"
	CheckVerifyFlows     = "verify.flows"
	CheckVerifyForeign   = "verify.foreign_deploys"
)

// fullDeployIDPattern is a whole deploy Id, 15 or 18 characters with the 0Af key prefix. An
// Id goes into a SOQL literal only after it matches.
var fullDeployIDPattern = regexp.MustCompile(`^0Af[A-Za-z0-9]{12}(?:[A-Za-z0-9]{3})?$`)

// flowNamePattern is a Flow API name that is safe inside a SOQL string literal.
var flowNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

// recordIDPattern is a Salesforce record Id that is safe inside a SOQL string literal.
var recordIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{15}(?:[A-Za-z0-9]{3})?$`)

// testLevelsNeedingTests are the deploy test levels that run Apex tests.
var testLevelsNeedingTests = map[string]bool{
	"RunSpecifiedTests": true,
	"RunLocalTests":     true,
	"RunAllTestsInOrg":  true,
	"RunRelevantTests":  true,
}

// deployRequestFields are read from the Tooling API DeployRequest by Id. Every field exists
// in the capture tooling-deployrequest-pass.json.
const deployRequestFields = "Id, Status, CheckOnly, NumberComponentsDeployed, NumberComponentsTotal, " +
	"NumberComponentErrors, NumberTestsCompleted, NumberTestErrors, NumberTestsTotal, RunTestsEnabled, " +
	"StartDate, CompletedDate, CreatedById"

// windowFields are read for every deploy in the window, as in tooling-deployrequest-window.json.
const windowFields = "Id, Status, CheckOnly, CreatedDate, StartDate, CompletedDate, CreatedById, " +
	"NumberComponentsDeployed, RunTestsEnabled"

// deployRequestQuery reads one DeployRequest. id must match fullDeployIDPattern.
func deployRequestQuery(id string) string {
	return "SELECT " + deployRequestFields + " FROM DeployRequest WHERE Id = '" + id + "'"
}

// windowQuery reads the deploys that are not check-only created between from and to.
func windowQuery(from, to time.Time) string {
	return "SELECT " + windowFields + " FROM DeployRequest WHERE CheckOnly = false AND CreatedDate >= " +
		soqlTime(from) + " AND CreatedDate <= " + soqlTime(to) + " ORDER BY CreatedDate"
}

// flowDefinitionQuery reads the FlowDefinition of each name. Every name must match
// flowNamePattern.
func flowDefinitionQuery(names []string) string {
	return "SELECT Id, DeveloperName, ActiveVersionId, LatestVersionId FROM FlowDefinition WHERE DeveloperName IN (" +
		quoteList(names) + ")"
}

// flowVersionQuery reads every Flow version of the given definitions. Every Id must match
// recordIDPattern.
func flowVersionQuery(definitionIDs []string) string {
	return "SELECT Id, VersionNumber, Status, LastModifiedDate, DefinitionId FROM Flow WHERE DefinitionId IN (" +
		quoteList(definitionIDs) + ") ORDER BY VersionNumber"
}

func quoteList(xs []string) string {
	q := make([]string, len(xs))
	for i, x := range xs {
		q[i] = "'" + x + "'"
	}
	return strings.Join(q, ", ")
}

// soqlTime is t as a SOQL dateTime literal, in UTC to the second.
func soqlTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

// parseSFTime reads the two date forms the captures hold: "2026-11-08T18:33:11.000+0000"
// from the Tooling API and "2026-11-08T18:33:11.000Z" from deploy report.
func parseSFTime(s string) (time.Time, bool) {
	for _, layout := range []string{"2006-01-02T15:04:05.000-0700", time.RFC3339Nano} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// deployRequestRow is a Tooling API DeployRequest record. Pointers stay nil when a field is
// missing, so a missing field is never read as false or zero.
type deployRequestRow struct {
	ID                       string   `json:"Id"`
	Status                   string   `json:"Status"`
	CheckOnly                *bool    `json:"CheckOnly"`
	NumberComponentsDeployed *flexInt `json:"NumberComponentsDeployed"`
	NumberComponentsTotal    *flexInt `json:"NumberComponentsTotal"`
	NumberTestsCompleted     *flexInt `json:"NumberTestsCompleted"`
	NumberTestErrors         *flexInt `json:"NumberTestErrors"`
	NumberTestsTotal         *flexInt `json:"NumberTestsTotal"`
	RunTestsEnabled          *bool    `json:"RunTestsEnabled"`
	CreatedDate              string   `json:"CreatedDate"`
	StartDate                string   `json:"StartDate"`
	CreatedByID              string   `json:"CreatedById"`
}

type componentSuccess struct {
	ComponentType string `json:"componentType"`
	FullName      string `json:"fullName"`
	Changed       *bool  `json:"changed"`
	Deleted       *bool  `json:"deleted"`
}

type coverageWarning struct {
	Message string `json:"message"`
}

// reportBody is the part of a deploy report verify reads.
type reportBody struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Details struct {
		ComponentSuccesses flexList[componentSuccess] `json:"componentSuccesses"`
		RunTestResult      struct {
			CodeCoverageWarnings flexList[coverageWarning] `json:"codeCoverageWarnings"`
		} `json:"runTestResult"`
	} `json:"details"`
}

type flowDefinitionRow struct {
	ID              string `json:"Id"`
	DeveloperName   string `json:"DeveloperName"`
	ActiveVersionID string `json:"ActiveVersionId"`
	LatestVersionID string `json:"LatestVersionId"`
}

type flowVersionRow struct {
	ID               string  `json:"Id"`
	VersionNumber    flexInt `json:"VersionNumber"`
	Status           string  `json:"Status"`
	LastModifiedDate string  `json:"LastModifiedDate"`
	DefinitionID     string  `json:"DefinitionId"`
}

// Verify reads the org after a deploy and checks it against the plan. It never deploys. The
// checks are:
//
//   - the Tooling API DeployRequest for the deploy Id says Succeeded and not check-only
//   - every plan member, and every destructive member, is in the deploy's component successes
//   - tests ran when the test level needs them
//   - each Flow in the plan whose source says Active runs as its latest version, and that
//     version was modified after the deploy started (verify.flows: fail, warn or off)
//   - no other deploy that is not check-only was created in the window (verify.foreign_deploys)
//
// Anything that cannot be read, such as a query that fails, is an error, never a pass.
func (p *Provider) Verify(ctx context.Context, s providers.Session, req providers.VerifyRequest) (providers.VerifyResult, error) {
	sess, ok := asSession(s)
	if !ok {
		return providers.VerifyResult{}, errors.New("salesforce: verify needs a session from this provider")
	}
	dir, err := filepath.Abs(req.Dir)
	if err != nil {
		return providers.VerifyResult{}, fmt.Errorf("salesforce: resolve project dir: %w", err)
	}
	id := req.Deploy.ID
	if !fullDeployIDPattern.MatchString(id) {
		return providers.VerifyResult{}, fmt.Errorf("salesforce: %q is not a deploy Id", id)
	}
	if err := checkPlanFiles(dir, req.Plan); err != nil {
		return providers.VerifyResult{}, err
	}
	pkg, err := readManifest(filepath.Join(dir, filepath.FromSlash(req.Plan.Manifest)))
	if err != nil {
		return providers.VerifyResult{}, fmt.Errorf("salesforce: %w", err)
	}
	var destructive []manifest
	for _, d := range req.Plan.Destructive {
		m, err := readManifest(filepath.Join(dir, filepath.FromSlash(d)))
		if err != nil {
			return providers.VerifyResult{}, fmt.Errorf("salesforce: %w", err)
		}
		destructive = append(destructive, m)
	}

	dr, err := p.readDeployRequest(ctx, sess, id)
	if err != nil {
		return providers.VerifyResult{}, err
	}
	rep, err := p.readDeployReport(ctx, sess, dir, id)
	if err != nil {
		return providers.VerifyResult{}, err
	}

	v := &verifyRun{res: providers.VerifyResult{Facts: map[string]any{}}}
	v.res.Facts["verify.deploy_id"] = dr.ID
	v.res.Facts["verify.deploy_status"] = dr.Status
	v.statusCheck(dr, rep)
	v.checkOnlyCheck(dr)
	v.membersCheck(dr, rep, pkg, destructive)
	v.testsCheck(dr, req.Deploy, hasApex(pkg, destructive))

	started := req.Deploy.Started
	if t, ok := parseSFTime(dr.StartDate); ok {
		// The org's own clock, the same one Flow LastModifiedDate is on.
		started = t
	}
	if err := p.flowsCheck(ctx, sess, dir, v, req.Flows, pkg, rep, started); err != nil {
		return providers.VerifyResult{}, err
	}
	if err := p.foreignCheck(ctx, sess, v, req, dr); err != nil {
		return providers.VerifyResult{}, err
	}

	v.res.Passed = true
	for _, c := range v.res.Checks {
		if c.Status == providers.VerifyFail {
			v.res.Passed = false
		}
	}
	return v.res, nil
}

// verifyRun collects the checks of one Verify call.
type verifyRun struct {
	res providers.VerifyResult
}

func (v *verifyRun) add(name string, status providers.VerifyStatus, detail string) {
	v.res.Checks = append(v.res.Checks, providers.VerifyCheck{Name: name, Status: status, Detail: detail})
}

// readDeployRequest reads the deploy's Tooling API DeployRequest. The record whose Id matches
// is used. No matching record is an error.
func (p *Provider) readDeployRequest(ctx context.Context, sess *session, id string) (deployRequestRow, error) {
	rows, err := toolingQuery[deployRequestRow](ctx, p, sess, deployRequestQuery(id))
	if err != nil {
		return deployRequestRow{}, fmt.Errorf("salesforce: read DeployRequest %s: %w", id, err)
	}
	for _, r := range rows {
		if sameID(r.ID, id) {
			return r, nil
		}
	}
	return deployRequestRow{}, fmt.Errorf("salesforce: the target has no DeployRequest %s", id)
}

// readDeployReport reads the deploy's result body with deploy report, for its component
// successes. sf can exit non-zero for a failed deploy and still print the body.
func (p *Provider) readDeployReport(ctx context.Context, sess *session, dir, id string) (reportBody, error) {
	env, err := p.sf(ctx, sess, dir, "project", "deploy", "report", "--job-id", id, "--target-org", sess.alias)
	var body reportBody
	if len(env.Result) == 0 || json.Unmarshal(env.Result, &body) != nil || body.ID == "" {
		if err != nil {
			return reportBody{}, fmt.Errorf("salesforce: read deploy report %s: %w", id, err)
		}
		return reportBody{}, fmt.Errorf("salesforce: deploy report for %s has no result body", id)
	}
	if !sameID(body.ID, id) {
		return reportBody{}, fmt.Errorf("salesforce: deploy report for %s answered for %s", id, body.ID)
	}
	return body, nil
}

// toolingQuery runs a Tooling API query and returns its records.
func toolingQuery[T any](ctx context.Context, p *Provider, sess *session, q string) ([]T, error) {
	env, err := p.sf(ctx, sess, "", "data", "query", "--use-tooling-api", "--query", q, "--target-org", sess.alias)
	if err != nil {
		return nil, describeErr("run a Tooling API query", err)
	}
	var out struct {
		Records []T   `json:"records"`
		Done    *bool `json:"done"`
	}
	if err := json.Unmarshal(env.Result, &out); err != nil {
		return nil, fmt.Errorf("salesforce: could not read the Tooling API query result: %w", err)
	}
	if out.Done != nil && !*out.Done {
		// More rows than one page holds. verify reads small sets, so a partial answer is
		// refused instead of read as complete.
		return nil, errors.New("salesforce: the Tooling API query returned only part of its records")
	}
	return out.Records, nil
}

// sameID compares two Salesforce Ids by their 15 character form.
func sameID(a, b string) bool {
	return a != "" && shortID(a) == shortID(b)
}

func shortID(s string) string {
	if len(s) > 15 {
		return s[:15]
	}
	return s
}

func (v *verifyRun) statusCheck(dr deployRequestRow, rep reportBody) {
	if dr.Status == "Succeeded" {
		v.add(CheckVerifyStatus, providers.VerifyPass, fmt.Sprintf("deploy %s is Succeeded in the target", dr.ID))
		return
	}
	detail := fmt.Sprintf("deploy %s is %s in the target, not Succeeded", dr.ID, orUnknown(dr.Status))
	var reasons []string
	for _, w := range rep.Details.RunTestResult.CodeCoverageWarnings {
		if m := oneLine(w.Message); m != "" {
			reasons = append(reasons, m)
		}
	}
	if len(reasons) > 0 {
		detail += ": " + strings.Join(reasons, " ")
	}
	v.add(CheckVerifyStatus, providers.VerifyFail, detail)
}

func (v *verifyRun) checkOnlyCheck(dr deployRequestRow) {
	switch {
	case dr.CheckOnly == nil:
		v.add(CheckVerifyCheckOnly, providers.VerifyFail, "the DeployRequest does not say whether it was check-only")
	case *dr.CheckOnly:
		v.res.Facts["verify.check_only"] = true
		v.add(CheckVerifyCheckOnly, providers.VerifyFail, fmt.Sprintf("deploy %s was check-only, so it changed nothing", dr.ID))
	default:
		v.res.Facts["verify.check_only"] = false
		v.add(CheckVerifyCheckOnly, providers.VerifyPass, "not check-only")
	}
}

// reportedAs maps a manifest member to the type and name the deploy reports for it. A folder
// is named in the manifest as its content type, and the deploy reports it as a folder type:
//   - Dashboard "Folder/" (trailing slash) is DashboardFolder "Folder", and Report the same.
//   - EmailTemplate "Folder" (no slash in the name) is EmailFolder "Folder".
//   - Settings "Name" is "NameSettings" with the full name "Name".
//
// All three are [confirmed] by deploy-report-folders-settings.json. Document "Folder/" is
// DocumentFolder "Folder" by the same folder rule, [unverified]: it was not captured.
// Every other member, and every member that does not fit these shapes, is matched as is, so a
// different type with the same name still fails.
func reportedAs(typ, name string) (string, string) {
	switch typ {
	case "Dashboard", "Report", "Document": // Document is [unverified]
		if base, ok := strings.CutSuffix(name, "/"); ok && base != "" && !strings.Contains(base, "/") {
			return typ + "Folder", base
		}
	case "EmailTemplate":
		if name != "" && name != "*" && !strings.Contains(name, "/") {
			return "EmailFolder", name
		}
	case "Settings":
		if name != "" && name != "*" {
			return name + "Settings", name
		}
	}
	return typ, name
}

// memberKey is how a plan member and a component success are matched: the type as is, the
// name URL-decoded and folded to lower case.
func memberKey(typ, name string) string {
	if dec, err := url.PathUnescape(name); err == nil {
		name = dec
	}
	return typ + ":" + strings.ToLower(name)
}

// membersCheck finds every plan member in the component successes (T1). The counts are
// recorded as facts, never compared, because sf counts expand and vary.
func (v *verifyRun) membersCheck(dr deployRequestRow, rep reportBody, pkg manifest, destructive []manifest) {
	deployed := map[string]bool{}
	deleted := map[string]bool{}
	byType := map[string]int{}
	successes := 0
	for _, c := range rep.Details.ComponentSuccesses {
		if c.ComponentType == "" {
			continue // package.xml itself
		}
		successes++
		k := memberKey(c.ComponentType, c.FullName)
		if c.Deleted != nil && *c.Deleted {
			deleted[k] = true
		} else {
			deployed[k] = true
			byType[c.ComponentType]++
		}
	}
	var missing []string
	members := 0
	for _, t := range pkg.Types {
		for _, m := range t.Members {
			members++
			switch {
			case m == "*":
				if byType[t.Name] == 0 {
					missing = append(missing, t.Name+":*")
				}
			case !deployed[memberKey(reportedAs(t.Name, m))]:
				missing = append(missing, t.Name+":"+m)
			}
		}
	}
	var notDeleted []string
	destructiveCount := 0
	for _, d := range destructive {
		for _, t := range d.Types {
			for _, m := range t.Members {
				destructiveCount++
				if !deleted[memberKey(reportedAs(t.Name, m))] {
					notDeleted = append(notDeleted, t.Name+":"+m)
				}
			}
		}
	}
	sort.Strings(missing)
	sort.Strings(notDeleted)
	f := v.res.Facts
	f["verify.plan_members"] = members
	f["verify.destructive_members"] = destructiveCount
	f["verify.component_successes"] = successes
	f["verify.members_missing"] = nonNil(missing)
	f["verify.destructive_missing"] = nonNil(notDeleted)
	if dr.NumberComponentsDeployed != nil {
		f["verify.components_deployed"] = int(*dr.NumberComponentsDeployed)
	}
	if dr.NumberComponentsTotal != nil {
		f["verify.components_total"] = int(*dr.NumberComponentsTotal)
	}
	if len(missing) == 0 && len(notDeleted) == 0 {
		detail := fmt.Sprintf("all %d plan members are in the deploy's component successes", members)
		if destructiveCount > 0 {
			detail += fmt.Sprintf(", and all %d destructive members are listed as deleted", destructiveCount)
		}
		v.add(CheckVerifyMembers, providers.VerifyPass, detail)
		return
	}
	var parts []string
	if len(missing) > 0 {
		parts = append(parts, fmt.Sprintf("%d of %d plan members are not in the deploy's component successes: %s", len(missing), members, listSome(missing)))
	}
	if len(notDeleted) > 0 {
		parts = append(parts, fmt.Sprintf("%d of %d destructive members are not listed as deleted: %s", len(notDeleted), destructiveCount, listSome(notDeleted)))
	}
	v.add(CheckVerifyMembers, providers.VerifyFail, strings.Join(parts, ". "))
}

// hasApex reports whether the plan's manifest or any destructive manifest names an Apex class
// or trigger.
func hasApex(pkg manifest, destructive []manifest) bool {
	for _, m := range append([]manifest{pkg}, destructive...) {
		if len(m.members("ApexClass")) > 0 || len(m.members("ApexTrigger")) > 0 {
			return true
		}
	}
	return false
}

// testsCheck checks that tests ran when the deploy's test level needs them. planHasApex says
// whether the plan deploys or deletes any Apex class or trigger.
func (v *verifyRun) testsCheck(dr deployRequestRow, d providers.DeployRecord, planHasApex bool) {
	f := v.res.Facts
	f["verify.test_level"] = d.TestLevel
	if dr.RunTestsEnabled != nil {
		f["verify.run_tests_enabled"] = *dr.RunTestsEnabled
	}
	total, completed, errs := intOr(dr.NumberTestsTotal, -1), intOr(dr.NumberTestsCompleted, -1), intOr(dr.NumberTestErrors, -1)
	if total >= 0 {
		f["verify.tests_total"] = total
	}
	if completed >= 0 {
		f["verify.tests_completed"] = completed
	}
	if errs >= 0 {
		f["verify.tests_failed"] = errs
	}
	switch {
	case d.Quick:
		// [unverified] No quick deploy DeployRequest was captured. The tests of a quick
		// deploy ran in the validation it deploys, which the PERMIT binds.
		v.add(CheckVerifyTests, providers.VerifySkipped, "a quick deploy reuses the tests of the validation it deploys")
		return
	case d.TestLevel == "":
		v.add(CheckVerifyTests, providers.VerifySkipped, "no test level is set for this target, so verify requires no tests")
		return
	case !testLevelsNeedingTests[d.TestLevel]:
		v.add(CheckVerifyTests, providers.VerifySkipped, "test level "+d.TestLevel+" requires no tests")
		return
	}
	known := dr.RunTestsEnabled != nil && total >= 0 && completed >= 0 && errs >= 0
	if d.TestLevel == "RunRelevantTests" && !planHasApex && known && (!*dr.RunTestsEnabled || total == 0) {
		// [unverified] No RunRelevantTests DeployRequest was captured. Salesforce picks the
		// tests from the Apex in the deploy, so a plan with no Apex can rightly run none.
		v.add(CheckVerifyTests, providers.VerifySkipped, "test level RunRelevantTests picks tests from the Apex in the deploy, and the plan has no Apex, so no tests ran")
		return
	}
	switch {
	case !known:
		v.add(CheckVerifyTests, providers.VerifyFail, "the DeployRequest does not say whether tests ran")
	case !*dr.RunTestsEnabled:
		v.add(CheckVerifyTests, providers.VerifyFail, fmt.Sprintf("tests did not run, and test level %s requires them", d.TestLevel))
	case total == 0:
		v.add(CheckVerifyTests, providers.VerifyFail, fmt.Sprintf("no tests ran, and test level %s requires them", d.TestLevel))
	case errs > 0:
		v.add(CheckVerifyTests, providers.VerifyFail, fmt.Sprintf("%d of %d tests failed", errs, total))
	case completed < total:
		v.add(CheckVerifyTests, providers.VerifyFail, fmt.Sprintf("only %d of %d tests completed", completed, total))
	default:
		v.add(CheckVerifyTests, providers.VerifyPass, fmt.Sprintf("%d of %d tests ran at test level %s", completed, total, d.TestLevel))
	}
}

// flowsCheck checks each Flow in the plan whose source says Active: the target must run it as
// its latest version, and, when the deploy changed the Flow, that version must be modified
// after the deploy started.
func (p *Provider) flowsCheck(ctx context.Context, sess *session, dir string, v *verifyRun, mode string, pkg manifest, rep reportBody, started time.Time) error {
	if mode == config.VerifyOff {
		v.add(CheckVerifyFlows, providers.VerifySkipped, "verify.flows is off")
		return nil
	}
	names := pkg.members("Flow")
	if len(names) == 0 {
		v.add(CheckVerifyFlows, providers.VerifySkipped, "the plan has no Flows")
		return nil
	}
	if contains(names, "*") {
		all, err := sourceFlowNames(dir)
		if err != nil {
			return err
		}
		names = all
	}
	var active, inactive []string
	for _, n := range names {
		status, err := sourceFlowStatus(dir, n)
		if err != nil {
			return err
		}
		if status == "Active" {
			active = append(active, n)
		} else {
			inactive = append(inactive, n)
		}
	}
	v.res.Facts["verify.flows_active_in_source"] = nonNil(active)
	v.res.Facts["verify.flows_not_active_in_source"] = nonNil(inactive)
	if len(active) == 0 {
		v.add(CheckVerifyFlows, providers.VerifySkipped, "no Flow in the plan is Active in source: "+strings.Join(inactive, ", "))
		return nil
	}
	for _, n := range active {
		if !flowNamePattern.MatchString(n) {
			return fmt.Errorf("salesforce: %q is not a Flow name verify can query", n)
		}
	}
	defs, err := toolingQuery[flowDefinitionRow](ctx, p, sess, flowDefinitionQuery(active))
	if err != nil {
		return fmt.Errorf("salesforce: read FlowDefinition: %w", err)
	}
	byName := map[string]flowDefinitionRow{}
	var defIDs []string
	for _, d := range defs {
		if !recordIDPattern.MatchString(d.ID) {
			return fmt.Errorf("salesforce: FlowDefinition %s has an Id verify cannot read", d.DeveloperName)
		}
		byName[strings.ToLower(d.DeveloperName)] = d
		defIDs = append(defIDs, d.ID)
	}
	versions := map[string]flowVersionRow{}
	if len(defIDs) > 0 {
		rows, err := toolingQuery[flowVersionRow](ctx, p, sess, flowVersionQuery(defIDs))
		if err != nil {
			return fmt.Errorf("salesforce: read Flow versions: %w", err)
		}
		for _, r := range rows {
			versions[shortID(r.ID)] = r
		}
	}
	changed := map[string]bool{}
	for _, c := range rep.Details.ComponentSuccesses {
		if c.ComponentType == "Flow" {
			// A success without the changed field counts as changed, so the time check runs.
			changed[strings.ToLower(c.FullName)] = c.Changed == nil || *c.Changed
		}
	}

	var problems, notLive []string
	for _, n := range active {
		d, found := byName[strings.ToLower(n)]
		if msg := flowProblem(n, d, found, versions, changed[strings.ToLower(n)], started); msg != "" {
			problems = append(problems, msg)
			notLive = append(notLive, n)
		}
	}
	v.res.Facts["verify.flows_not_live"] = nonNil(notLive)
	if len(problems) == 0 {
		v.add(CheckVerifyFlows, providers.VerifyPass, fmt.Sprintf("%d active Flows run their latest version: %s", len(active), strings.Join(active, ", ")))
		return nil
	}
	status := providers.VerifyFail
	if mode == config.VerifyWarn {
		status = providers.VerifyWarn
	}
	v.add(CheckVerifyFlows, status, strings.Join(problems, ". "))
	return nil
}

// flowProblem says what is wrong with one Flow that is Active in source, or "".
func flowProblem(name string, d flowDefinitionRow, found bool, versions map[string]flowVersionRow, changed bool, started time.Time) string {
	if !found {
		return fmt.Sprintf("Flow %s is not in the target", name)
	}
	if d.ActiveVersionID == "" {
		return fmt.Sprintf("Flow %s is Active in source but has no active version in the target", name)
	}
	act, okAct := versions[shortID(d.ActiveVersionID)]
	if !sameID(d.ActiveVersionID, d.LatestVersionID) {
		latest := "a later version is not active"
		if lat, ok := versions[shortID(d.LatestVersionID)]; ok && d.LatestVersionID != "" {
			latest = fmt.Sprintf("the latest version %d is %s", lat.VersionNumber, orUnknown(lat.Status))
		}
		if !okAct {
			return fmt.Sprintf("Flow %s does not run its latest version in the target, and %s", name, latest)
		}
		return fmt.Sprintf("Flow %s runs version %d in the target, and %s", name, act.VersionNumber, latest)
	}
	if !okAct {
		return fmt.Sprintf("Flow %s: the target did not return its active version", name)
	}
	if !changed {
		return "" // the deploy left the Flow unchanged, so its version is older by design
	}
	mod, ok := parseSFTime(act.LastModifiedDate)
	if !ok || mod.Before(started) {
		return fmt.Sprintf("Flow %s: active version %d was last modified at %s, before the deploy started at %s",
			name, act.VersionNumber, orUnknown(act.LastModifiedDate), started.UTC().Format(time.RFC3339))
	}
	return ""
}

// sourceFlowStatus reads the <status> of Flow name from its source file in the package
// directories. A Flow the plan names with no source file is an error.
func sourceFlowStatus(dir, name string) (string, error) {
	path, err := findSourceFlow(dir, name)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("salesforce: read Flow %s: %w", name, err)
	}
	var f struct {
		Status string `xml:"status"`
	}
	if err := xml.Unmarshal(data, &f); err != nil {
		return "", fmt.Errorf("salesforce: parse Flow %s: %w", name, err)
	}
	return strings.TrimSpace(f.Status), nil
}

const flowSuffix = ".flow-meta.xml"

// findSourceFlow returns the source file of Flow name. Two files for one name are an error,
// because verify could not tell which one was deployed.
func findSourceFlow(dir, name string) (string, error) {
	var found []string
	err := walkPackageDirs(dir, func(path string) {
		if strings.EqualFold(filepath.Base(path), name+flowSuffix) {
			found = append(found, path)
		}
	})
	if err != nil {
		return "", err
	}
	switch len(found) {
	case 0:
		return "", fmt.Errorf("salesforce: Flow %s is in the plan, but no %s%s is in the package directories", name, name, flowSuffix)
	case 1:
		return found[0], nil
	}
	return "", fmt.Errorf("salesforce: Flow %s has %d source files in the package directories", name, len(found))
}

// sourceFlowNames lists every Flow in the package directories, for a plan that names Flow *.
func sourceFlowNames(dir string) ([]string, error) {
	var names []string
	err := walkPackageDirs(dir, func(path string) {
		if base := filepath.Base(path); strings.HasSuffix(base, flowSuffix) {
			names = append(names, strings.TrimSuffix(base, flowSuffix))
		}
	})
	sort.Strings(names)
	return names, err
}

func walkPackageDirs(dir string, visit func(path string)) error {
	pkgDirs, err := packageDirectories(dir)
	if err != nil {
		return fmt.Errorf("salesforce: %w", err)
	}
	for _, pd := range pkgDirs {
		root := filepath.Join(dir, filepath.FromSlash(pd))
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type().IsRegular() {
				visit(path)
			}
			return nil
		})
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("salesforce: read package directory %s: %w", pd, err)
		}
	}
	return nil
}

// foreignCheck lists the other deploys that are not check-only created in the window (T2).
func (p *Provider) foreignCheck(ctx context.Context, sess *session, v *verifyRun, req providers.VerifyRequest, ours deployRequestRow) error {
	if req.ForeignDeploys == config.VerifyOff {
		v.add(CheckVerifyForeign, providers.VerifySkipped, "verify.foreign_deploys is off")
		return nil
	}
	from, to := req.WindowStart.Truncate(time.Second), req.WindowEnd
	if req.WindowStart.IsZero() || to.IsZero() || to.Before(from) {
		return errors.New("salesforce: verify needs a deploy window with a start before its end")
	}
	// SOQL compares whole seconds, so the upper bound is the next second. Each row is
	// checked against the window again below.
	rows, err := toolingQuery[deployRequestRow](ctx, p, sess, windowQuery(from, to.Truncate(time.Second).Add(time.Second)))
	if err != nil {
		return fmt.Errorf("salesforce: read the deploys in the window: %w", err)
	}
	foreign := []map[string]any{}
	var ids []string
	for _, r := range rows {
		if sameID(r.ID, ours.ID) || (r.CheckOnly != nil && *r.CheckOnly) {
			continue
		}
		// A row with no readable CreatedDate is counted, so a shape change cannot hide one.
		if c, ok := parseSFTime(r.CreatedDate); ok && (c.Before(from) || c.After(to)) {
			continue
		}
		foreign = append(foreign, map[string]any{
			"id":        r.ID,
			"status":    r.Status,
			"created":   r.CreatedDate,
			"same_user": r.CreatedByID != "" && sameID(r.CreatedByID, ours.CreatedByID),
		})
		ids = append(ids, fmt.Sprintf("%s (%s, created %s)", r.ID, orUnknown(r.Status), orUnknown(r.CreatedDate)))
	}
	v.res.Facts["verify.foreign_deploys"] = foreign
	v.res.Facts["verify.foreign_deploys_count"] = len(foreign)
	v.res.Facts["verify.window_start"] = from.UTC().Format(time.RFC3339)
	v.res.Facts["verify.window_end"] = to.UTC().Format(time.RFC3339)
	if len(foreign) == 0 {
		v.add(CheckVerifyForeign, providers.VerifyPass, "no other deploy to the target in the window")
		return nil
	}
	status := providers.VerifyWarn
	if req.ForeignDeploys == config.VerifyFail {
		status = providers.VerifyFail
	}
	v.add(CheckVerifyForeign, status, fmt.Sprintf("%d other deploys to the target in the window: %s", len(foreign), strings.Join(ids, ", ")))
	return nil
}

func intOr(n *flexInt, def int) int {
	if n == nil {
		return def
	}
	return int(*n)
}

func orUnknown(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}

func nonNil(xs []string) []string {
	if xs == nil {
		return []string{}
	}
	return xs
}

// listSome joins up to 10 names, then says how many more there are.
func listSome(xs []string) string {
	const most = 10
	if len(xs) <= most {
		return strings.Join(xs, ", ")
	}
	return strings.Join(xs[:most], ", ") + fmt.Sprintf(" and %d more", len(xs)-most)
}
