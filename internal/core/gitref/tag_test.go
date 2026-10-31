package gitref

import (
	"errors"
	"strings"
	"testing"
)

func TestDeployedTagName(t *testing.T) {
	cases := []struct {
		target string
		want   string
		ok     bool
	}{
		{"qa", "glassops/deployed/qa", true},
		{"prod-eu", "glassops/deployed/prod-eu", true},
		{"uat_2", "glassops/deployed/uat_2", true},
		{"v1.2", "glassops/deployed/v1.2", true},
		{"", "", false},
		{"../main", "", false},
		{"a/b", "", false},
		{"a..b", "", false},
		{"qa.lock", "", false},
		{"-qa", "", false},
		{"q a", "", false},
		{"qa.", "", false},
	}
	for _, tc := range cases {
		got, err := DeployedTag(tc.target)
		if (err == nil) != tc.ok || got != tc.want {
			t.Errorf("DeployedTag(%q) = %q, %v, want %q ok=%v", tc.target, got, err, tc.want, tc.ok)
		}
	}
}

func TestLastDeployed(t *testing.T) {
	cases := []struct {
		name      string
		answers   map[string]string
		wantSHA   string
		wantOK    bool
		wantErrIs error
		wantErr   bool
	}{
		{
			name:    "tag present",
			answers: with(map[string]string{"rev-parse --verify --quiet refs/tags/glassops/deployed/qa^{commit}": tagSHA + "\n"}),
			wantSHA: tagSHA, wantOK: true,
		},
		{name: "no tag", answers: with(nil)},
		{
			name:      "shallow clone is an error, never a missing tag",
			answers:   map[string]string{"rev-parse --is-shallow-repository": "true\n"},
			wantErrIs: ErrShallow, wantErr: true,
		},
		{name: "not a repository", answers: map[string]string{}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := &fakeGit{answers: tc.answers}
			sha, ok, err := LastDeployed(g.run, "/repo", "qa")
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
				t.Fatalf("err = %v, want %v", err, tc.wantErrIs)
			}
			if sha != tc.wantSHA || ok != tc.wantOK {
				t.Fatalf("got %q ok=%v, want %q ok=%v", sha, ok, tc.wantSHA, tc.wantOK)
			}
		})
	}
}

func TestLastDeployedRejectsBadTarget(t *testing.T) {
	g := &fakeGit{answers: with(nil)}
	if _, _, err := LastDeployed(g.run, "/repo", "../x"); err == nil {
		t.Fatal("a target name with a path in it was accepted")
	}
	if len(g.calls) != 0 {
		t.Fatalf("git ran for a bad target name: %v", g.calls)
	}
}

func TestMoveAndPushDeployedTag(t *testing.T) {
	g := &fakeGit{answers: map[string]string{
		"tag -f glassops/deployed/qa " + headSHA:        "",
		"push -f origin refs/tags/glassops/deployed/qa": "",
	}}
	if err := MoveDeployedTag(g.run, "/repo", "qa", headSHA); err != nil {
		t.Fatal(err)
	}
	if err := PushDeployedTag(g.run, "/repo", "", "qa"); err != nil {
		t.Fatal(err)
	}
	want := []string{"tag -f glassops/deployed/qa " + headSHA, "push -f origin refs/tags/glassops/deployed/qa"}
	if strings.Join(g.calls, "|") != strings.Join(want, "|") {
		t.Fatalf("calls = %v, want %v", g.calls, want)
	}
}

func TestMoveAndPushDeployedTagErrors(t *testing.T) {
	g := &fakeGit{answers: map[string]string{}}
	if err := MoveDeployedTag(g.run, "/repo", "qa", headSHA); err == nil || !strings.Contains(err.Error(), "glassops/deployed/qa") {
		t.Fatalf("move error = %v", err)
	}
	if err := MoveDeployedTag(g.run, "/repo", "qa", ""); err == nil {
		t.Fatal("an empty commit was accepted")
	}
	if err := PushDeployedTag(g.run, "/repo", "upstream", "qa"); err == nil || !strings.Contains(err.Error(), "upstream") {
		t.Fatalf("push error = %v", err)
	}
	if err := PushDeployedTag(g.run, "/repo", "", "a/b"); err == nil {
		t.Fatal("a bad target name was pushed")
	}
}
