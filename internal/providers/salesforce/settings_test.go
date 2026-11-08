package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// portfolioSettings is the union of the settings blocks in the portfolio's two scratch
// definitions, config/project-scratch-def.json and config/project-scratch-def-ai.json, copied
// by hand in the definitions' own order. It is the conversion's real input.
const portfolioSettings = `{
  "lightningExperienceSettings": {
    "enableS1DesktopEnabled": true
  },
  "mobileSettings": {
    "enableS1EncryptedStoragePref2": false
  },
  "communitiesSettings": {
    "enableNetworksEnabled": true
  },
  "einsteinGptSettings": {
    "enableEinsteinGptPlatform": true
  },
  "agentPlatformSettings": {
    "enableAgentPlatform": true
  }
}`

func settingsDoc(root, body string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
		"<" + root + ` xmlns="http://soap.sforce.com/2006/04/metadata">` + "\n" + body + "</" + root + ">\n"
}

func TestConvertSettingsPortfolioKeys(t *testing.T) {
	files, err := convertSettings(json.RawMessage(portfolioSettings))
	if err != nil {
		t.Fatal(err)
	}
	want := []struct{ name, path, xml string }{
		{"LightningExperience", "settings/LightningExperience.settings-meta.xml",
			settingsDoc("LightningExperienceSettings", "    <enableS1DesktopEnabled>true</enableS1DesktopEnabled>\n")},
		{"Mobile", "settings/Mobile.settings-meta.xml",
			settingsDoc("MobileSettings", "    <enableS1EncryptedStoragePref2>false</enableS1EncryptedStoragePref2>\n")},
		{"Communities", "settings/Communities.settings-meta.xml",
			settingsDoc("CommunitiesSettings", "    <enableNetworksEnabled>true</enableNetworksEnabled>\n")},
		{"EinsteinGpt", "settings/EinsteinGpt.settings-meta.xml",
			settingsDoc("EinsteinGptSettings", "    <enableEinsteinGptPlatform>true</enableEinsteinGptPlatform>\n")},
		{"AgentPlatform", "settings/AgentPlatform.settings-meta.xml",
			settingsDoc("AgentPlatformSettings", "    <enableAgentPlatform>true</enableAgentPlatform>\n")},
	}
	if len(files) != len(want) {
		t.Fatalf("got %d files, want %d", len(files), len(want))
	}
	for i, w := range want {
		f := files[i]
		if f.Name != w.name || f.Path() != w.path {
			t.Errorf("file %d is %s at %s, want %s at %s", i, f.Name, f.Path(), w.name, w.path)
		}
		if string(f.XML) != w.xml {
			t.Errorf("%s:\n%s\nwant:\n%s", w.name, f.XML, w.xml)
		}
	}
}

// TestConvertSettingsShapes covers what the portfolio does not use: nested objects, arrays of
// scalars and of objects, numbers, strings that need escaping, and key order. Synthetic input,
// shaped like the nested settings in Salesforce's scratch org definition docs.
func TestConvertSettingsShapes(t *testing.T) {
	in := `{"securitySettings":{"passwordPolicies":{"minimumPasswordLength":12,"complexity":"UpperLowerCaseNumeric","expiration":"Never"},
	"sessionSettings":{"sessionTimeout":"TwelveHours","lockSessionsToIp":false},
	"enableAdminLoginAsAnyUser":true},
	"languageSettings":{"enableTranslationWorkbench":true,"note":"a < b & \"c\""},
	"caseSettings":{"caseOwnerNotificationTemplates":["one","two"],"webToCase":{"caseOrigin":"Web"},
	"emailToCase":{"routingAddresses":[{"routingName":"Support","emailAddress":"support@example.com"},{"routingName":"Billing"}]}}}`
	files, err := convertSettings(json.RawMessage(in))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		settingsDoc("SecuritySettings", ""+
			"    <passwordPolicies>\n"+
			"        <minimumPasswordLength>12</minimumPasswordLength>\n"+
			"        <complexity>UpperLowerCaseNumeric</complexity>\n"+
			"        <expiration>Never</expiration>\n"+
			"    </passwordPolicies>\n"+
			"    <sessionSettings>\n"+
			"        <sessionTimeout>TwelveHours</sessionTimeout>\n"+
			"        <lockSessionsToIp>false</lockSessionsToIp>\n"+
			"    </sessionSettings>\n"+
			"    <enableAdminLoginAsAnyUser>true</enableAdminLoginAsAnyUser>\n"),
		settingsDoc("LanguageSettings", ""+
			"    <enableTranslationWorkbench>true</enableTranslationWorkbench>\n"+
			"    <note>a &lt; b &amp; &#34;c&#34;</note>\n"),
		settingsDoc("CaseSettings", ""+
			"    <caseOwnerNotificationTemplates>one</caseOwnerNotificationTemplates>\n"+
			"    <caseOwnerNotificationTemplates>two</caseOwnerNotificationTemplates>\n"+
			"    <webToCase>\n"+
			"        <caseOrigin>Web</caseOrigin>\n"+
			"    </webToCase>\n"+
			"    <emailToCase>\n"+
			"        <routingAddresses>\n"+
			"            <routingName>Support</routingName>\n"+
			"            <emailAddress>support@example.com</emailAddress>\n"+
			"        </routingAddresses>\n"+
			"        <routingAddresses>\n"+
			"            <routingName>Billing</routingName>\n"+
			"        </routingAddresses>\n"+
			"    </emailToCase>\n"),
	}
	names := []string{"Security", "Language", "Case"}
	if len(files) != len(want) {
		t.Fatalf("got %d files", len(files))
	}
	for i := range want {
		if files[i].Name != names[i] {
			t.Errorf("file %d is %s, want %s", i, files[i].Name, names[i])
		}
		if string(files[i].XML) != want[i] {
			t.Errorf("%s:\n%s\nwant:\n%s", names[i], files[i].XML, want[i])
		}
	}
}

func TestConvertSettingsRefuses(t *testing.T) {
	cases := []struct{ name, in, err string }{
		{"objectSettings", `{"objectSettings":{"account":{"sharingModel":"private"}}}`, "cannot apply objectSettings"},
		{"a key that is not a Settings type", `{"lightningExperience":{"x":true}}`, "does not name a Settings type"},
		{"a key with an upper case start", `{"MobileSettings":{"x":true}}`, "does not name a Settings type"},
		{"a block that is not an object", `{"mobileSettings":[true]}`, "must be an object"},
		{"settings that are not an object", `["mobileSettings"]`, "must be an object"},
		{"a null value", `{"mobileSettings":{"x":null}}`, "mobileSettings.x is null"},
		{"a list inside a list", `{"mobileSettings":{"x":[[1]]}}`, "list inside a list"},
		{"a field that is not an element name", `{"mobileSettings":{"a b":true}}`, "not a valid element name"},
		{"a nested field that is not an element name", `{"mobileSettings":{"a":{"<x>":true}}}`, "mobileSettings.a.<x>"},
		{"a key twice", `{"mobileSettings":{"x":true,"x":false}}`, "appears twice"},
		{"broken JSON", `{"mobileSettings":{"x":tru}}`, "read the definition settings"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := convertSettings(json.RawMessage(tc.in))
			if err == nil || !strings.Contains(err.Error(), tc.err) {
				t.Fatalf("err %v, want %q", err, tc.err)
			}
		})
	}
}

func TestReadDefinitionPortfolio(t *testing.T) {
	dir := t.TempDir()
	// The portfolio's AI definition, copied by hand.
	def := `{
  "orgName": "Portfolio AI",
  "edition": "Developer",
  "features": ["EnableSetPasswordInApi", "Communities", "Einstein1AIPlatform"],
  "settings": ` + portfolioSettings + `
}`
	if err := os.WriteFile(filepath.Join(dir, "def.json"), []byte(def), 0o644); err != nil {
		t.Fatal(err)
	}
	fields, files, err := readDefinition(dir, "def.json")
	if err != nil {
		t.Fatal(err)
	}
	if fields["Edition"] != "Developer" || fields["Features"] != "EnableSetPasswordInApi;Communities;Einstein1AIPlatform" || len(fields) != 2 {
		t.Fatalf("fields %v", fields)
	}
	if got := strings.Join(settingsNames(files), ","); got != "LightningExperience,Mobile,Communities,EinsteinGpt,AgentPlatform" {
		t.Fatalf("settings %s", got)
	}
}

const cmdSettingsDeploy = "sf project deploy start --source-dir force-app"

// settingsReq writes a definition with settings into a new dir and returns the dir and the
// definition path in it.
func settingsReq(t *testing.T, settings string) (req struct{ dir, file string }) {
	t.Helper()
	req.dir = t.TempDir()
	req.file = "config/def.json"
	if err := os.MkdirAll(filepath.Join(req.dir, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	def := `{"edition":"Developer","settings":` + settings + `}`
	if err := os.WriteFile(filepath.Join(req.dir, "config", "def.json"), []byte(def), 0o644); err != nil {
		t.Fatal(err)
	}
	return req
}

func TestCreateEnvDeploysSettingsToTheNewOrg(t *testing.T) {
	fastEnv(t)
	srv := tokenServer(t, goodToken(t))
	replies := happyReplies(srv.URL + "/")
	replies[cmdSettingsDeploy] = []fakeReply{{Stdout: mustFixture("deploy-start-pass.json")}}
	f := &seqExec{replies: replies}
	var order []string
	var deployed map[string]string
	var project string
	var deployHome, loginHome string
	f.onCall = func(c Cmd) {
		k := strings.Join(append([]string{c.Name}, c.Args...), " ")
		home := ""
		for _, e := range c.Env {
			if strings.HasPrefix(e, "HOME=") {
				home = strings.TrimPrefix(e, "HOME=")
			}
		}
		switch {
		case strings.HasPrefix(k, cmdLogin):
			order = append(order, "login")
			loginHome = home
		case strings.HasPrefix(k, cmdSettingsDeploy):
			order = append(order, "settings")
			deployHome = home
			data, _ := os.ReadFile(filepath.Join(c.Dir, ProjectFile))
			project = string(data)
			deployed = map[string]string{}
			_ = filepath.Walk(filepath.Join(c.Dir, settingsPkgDir), func(p string, fi os.FileInfo, err error) error {
				if err == nil && !fi.IsDir() {
					rel, _ := filepath.Rel(c.Dir, p)
					b, _ := os.ReadFile(p)
					deployed[filepath.ToSlash(rel)] = string(b)
				}
				return nil
			})
		}
	}
	p, hub := envProvider(t, f)
	r := settingsReq(t, `{"lightningExperienceSettings":{"enableS1DesktopEnabled":true},"mobileSettings":{"enableS1EncryptedStoragePref2":false}}`)
	req := envReq(t)
	req.Dir, req.DefinitionFile = r.dir, r.file

	env, err := p.CreateEnv(context.Background(), hub, req)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(order, ",") != "login,settings" {
		t.Fatalf("order %v", order)
	}
	if deployHome == "" || deployHome != loginHome || deployHome == hub.home {
		t.Fatalf("settings deployed with home %q, login home %q, hub home %q", deployHome, loginHome, hub.home)
	}
	if !strings.Contains(project, `"path":"force-app"`) {
		t.Fatalf("project %s", project)
	}
	if len(deployed) != 2 ||
		deployed["force-app/settings/LightningExperience.settings-meta.xml"] != settingsDoc("LightningExperienceSettings", "    <enableS1DesktopEnabled>true</enableS1DesktopEnabled>\n") ||
		deployed["force-app/settings/Mobile.settings-meta.xml"] != settingsDoc("MobileSettings", "    <enableS1EncryptedStoragePref2>false</enableS1EncryptedStoragePref2>\n") {
		t.Fatalf("deployed %v", deployed)
	}
	cmds := f.commands(cmdSettingsDeploy)
	if len(cmds) != 1 {
		t.Fatalf("settings deploys %v", cmds)
	}
	for _, w := range []string{"--test-level NoTestRun", "--ignore-conflicts", "--target-org glassops-target", "--wait 30", "--json"} {
		if !strings.Contains(cmds[0], w) {
			t.Fatalf("settings deploy %q lacks %q", cmds[0], w)
		}
	}
	if env.AuthFile == "" || env.ID != "00D000000000002" {
		t.Fatalf("env %+v", env)
	}
	if env.Facts["salesforce.settings_deployed"] != true || env.Facts["salesforce.settings_deploy_id"] == "" ||
		strings.Join(env.Facts["salesforce.settings"].([]string), ",") != "LightningExperience,Mobile" {
		t.Fatalf("facts %v", env.Facts)
	}
	if len(f.commands(cmdDelete)) != 0 {
		t.Fatal("deleted an org whose settings deployed")
	}
	// The login home and the settings project are both gone.
	if _, err := os.Stat(deployHome); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("login home left behind: %v", err)
	}
}

func TestCreateEnvWithoutSettingsDeploysNothing(t *testing.T) {
	fastEnv(t)
	srv := tokenServer(t, goodToken(t))
	f := &seqExec{replies: happyReplies(srv.URL + "/")}
	p, hub := envProvider(t, f)
	env, err := p.CreateEnv(context.Background(), hub, envReq(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(f.commands("sf project deploy")) != 0 {
		t.Fatal("deployed settings with no settings in the definition")
	}
	if _, ok := env.Facts["salesforce.settings_deployed"]; ok {
		t.Fatalf("facts %v", env.Facts)
	}
}

func TestCreateEnvSettingsFailureDeletesTheOrg(t *testing.T) {
	deleted := fakeReply{Stdout: []byte(`{"status":0,"result":{"id":"2AS000000000001AAA"},"warnings":[]}`)}
	deleteFails := fakeReply{Stdout: []byte(`{"status":1,"name":"INSUFFICIENT_ACCESS","message":"insufficient access rights"}`), ExitCode: 1}
	cases := []struct {
		name   string
		deploy fakeReply
		delete fakeReply
		// rejected is true when the org refused components, deleted when the delete worked.
		rejected, deleted bool
		msg               string
	}{
		// The compile failure fixture is a real deploy start body with componentFailures. A
		// refused Settings component has the same shape [unverified, step 0.4 captures it].
		{"org refuses the settings", fakeReply{Stdout: mustFixture("deploy-start-compile-fail.json"), ExitCode: 1}, deleted,
			true, true, "ApexClass Greeter: Missing '<EOF>' at 'public'"},
		{"org refuses and the delete fails", fakeReply{Stdout: mustFixture("deploy-start-compile-fail.json"), ExitCode: 1}, deleteFails,
			true, false, "Deleting the scratch org failed too"},
		{"sf fails without a result", fakeReply{Stdout: []byte(`{"status":1,"name":"NoOrgFound","message":"No org configuration found"}`), ExitCode: 1}, deleted,
			false, true, "No org configuration found"},
		{"sf prints nothing", fakeReply{Stdout: []byte(``), ExitCode: 1}, deleted, false, true, "did not deploy"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fastEnv(t)
			srv := tokenServer(t, goodToken(t))
			replies := happyReplies(srv.URL + "/")
			replies[cmdSettingsDeploy] = []fakeReply{tc.deploy}
			replies[cmdActive] = []fakeReply{{Stdout: mustFixture("query-active-scratch-org.json")}}
			replies[cmdDelete] = []fakeReply{tc.delete}
			f := &seqExec{replies: replies}
			p, hub := envProvider(t, f)
			r := settingsReq(t, `{"communitiesSettings":{"enableNetworksEnabled":true}}`)
			req := envReq(t)
			req.Dir, req.DefinitionFile = r.dir, r.file

			env, err := p.CreateEnv(context.Background(), hub, req)
			var se *SettingsError
			if !errors.As(err, &se) {
				t.Fatalf("err %v, want a *SettingsError", err)
			}
			if se.Rejected != tc.rejected || se.Deleted != tc.deleted || se.CheckFailed() != (tc.rejected && tc.deleted) {
				t.Fatalf("settings error %+v", se)
			}
			if !strings.Contains(err.Error(), tc.msg) {
				t.Fatalf("err %q lacks %q", err, tc.msg)
			}
			// The org is still described, so env.json can record the attempt.
			if env.ID != "00D000000000002" || env.RecordID != synthRecordID {
				t.Fatalf("env %+v", env)
			}
			q := f.commands(cmdActive)
			if len(q) != 1 || !strings.Contains(q[0], "WHERE ScratchOrg = '00D000000000002'") {
				t.Fatalf("active queries %v", q)
			}
			if d := f.commands(cmdDelete); len(d) != 1 || !strings.Contains(d[0], "--record-id 2AS000000000001AAA") {
				t.Fatalf("deletes %v", d)
			}
			// The delete runs from the Dev Hub session, not the new org's.
			for _, c := range f.calls {
				if strings.HasPrefix(strings.Join(append([]string{c.Name}, c.Args...), " "), cmdDelete) {
					if !contains(c.Env, "HOME="+hub.home) {
						t.Fatalf("delete env %v, want the hub home", c.Env)
					}
				}
			}
			if env.Facts["salesforce.settings_deployed"] != false || env.Facts["salesforce.settings_org_deleted"] != tc.deleted {
				t.Fatalf("facts %v", env.Facts)
			}
			if tc.deleted {
				if env.AuthFile != "" {
					t.Fatalf("auth file %q kept for a deleted org", env.AuthFile)
				}
			} else if _, err := os.Stat(env.AuthFile); err != nil {
				t.Fatalf("auth file for a live org: %v", err)
			}
			if b, _ := json.Marshal(env.Facts); strings.Contains(string(b)+err.Error(), synthRefresh) {
				t.Fatal("secret in facts or error")
			}
		})
	}
}

func TestCreateEnvSettingsCleanupOutlivesTheRun(t *testing.T) {
	fastEnv(t)
	srv := tokenServer(t, goodToken(t))
	replies := happyReplies(srv.URL + "/")
	replies[cmdSettingsDeploy] = []fakeReply{{Stdout: mustFixture("deploy-start-compile-fail.json"), ExitCode: 1}}
	replies[cmdActive] = []fakeReply{{Stdout: mustFixture("query-active-scratch-org.json")}}
	replies[cmdDelete] = []fakeReply{{Stdout: []byte(`{"status":0,"result":{"id":"2AS000000000001AAA"}}`)}}
	f := &seqExec{replies: replies}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// The run is stopped while the settings deploy is out.
	f.onCall = func(c Cmd) {
		if strings.HasPrefix(strings.Join(append([]string{c.Name}, c.Args...), " "), cmdSettingsDeploy) {
			cancel()
		}
	}
	p, hub := envProvider(t, f)
	// Like a real process run, a call on an ended context fails without running.
	p.exec = func(ctx context.Context, c Cmd) (Output, error) {
		if err := ctx.Err(); err != nil {
			return Output{}, err
		}
		return f.run(ctx, c)
	}
	r := settingsReq(t, `{"mobileSettings":{"enableS1EncryptedStoragePref2":false}}`)
	req := envReq(t)
	req.Dir, req.DefinitionFile = r.dir, r.file

	_, err := p.CreateEnv(ctx, hub, req)
	var se *SettingsError
	if !errors.As(err, &se) || !se.Deleted {
		t.Fatalf("err %v", err)
	}
	if len(f.commands(cmdDelete)) != 1 {
		t.Fatal("no delete after the run was stopped")
	}
}

func TestCreateEnvBadSettingsNeverSignsUp(t *testing.T) {
	fastEnv(t)
	f := &seqExec{replies: map[string][]fakeReply{}}
	p, hub := envProvider(t, f)
	r := settingsReq(t, `{"objectSettings":{"account":{"sharingModel":"private"}}}`)
	req := envReq(t)
	req.Dir, req.DefinitionFile = r.dir, r.file
	if _, err := p.CreateEnv(context.Background(), hub, req); err == nil || !strings.Contains(err.Error(), "objectSettings") {
		t.Fatalf("err %v", err)
	}
	if len(f.calls) != 0 {
		t.Fatalf("calls %v", f.calls)
	}
}
