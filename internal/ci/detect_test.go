package ci

import (
	"bytes"
	"testing"
)

func TestDetectBitbucketAndGitLab(t *testing.T) {
	cases := []struct {
		name  string
		force string
		env   map[string]string
		want  string
	}{
		{"bitbucket auto", "", map[string]string{"BITBUCKET_BUILD_NUMBER": "12"}, "bitbucket"},
		{"gitlab auto", "auto", map[string]string{"GITLAB_CI": "true"}, "gitlab"},
		{"gitlab needs true", "", map[string]string{"GITLAB_CI": "1"}, "local"},
		{"github wins", "", map[string]string{"GITHUB_ACTIONS": "true", "GITLAB_CI": "true", "BITBUCKET_BUILD_NUMBER": "3"}, "github"},
		{"bitbucket before gitlab", "", map[string]string{"GITLAB_CI": "true", "BITBUCKET_BUILD_NUMBER": "3"}, "bitbucket"},
		{"forced bitbucket", "bitbucket", nil, "bitbucket"},
		{"forced gitlab", "GitLab", nil, "gitlab"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := Detect(tc.force, mapEnv(tc.env), &bytes.Buffer{}, t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if c.Name() != tc.want {
				t.Fatalf("Detect = %s, want %s", c.Name(), tc.want)
			}
		})
	}
}

func TestBitbucketAndGitLabAreStubs(t *testing.T) {
	for _, c := range []CI{newBitbucket(mapEnv(nil), &bytes.Buffer{}, ""), newGitLab(mapEnv(nil), &bytes.Buffer{}, "")} {
		if c.Context().CI != c.Name() {
			t.Fatalf("%s context = %+v", c.Name(), c.Context())
		}
		if err := c.SetOutput("status", "pass"); err == nil {
			t.Fatalf("%s SetOutput did not fail", c.Name())
		}
		if err := c.Summary("x"); err == nil {
			t.Fatalf("%s Summary did not fail", c.Name())
		}
	}
	var out bytes.Buffer
	newGitLab(mapEnv(nil), &out, "").Annotate(Warning, "careful")
	if out.String() != "warning: careful\n" {
		t.Fatalf("annotate = %q", out.String())
	}
}
