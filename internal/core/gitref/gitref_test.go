package gitref

import (
	"errors"
	"strings"
	"testing"
)

// fakeGit answers git calls by their joined arguments. A missing key is a git failure.
type fakeGit struct {
	answers map[string]string
	calls   []string
}

func (f *fakeGit) run(_ string, args ...string) (string, error) {
	key := strings.Join(args, " ")
	f.calls = append(f.calls, key)
	if out, ok := f.answers[key]; ok {
		return out, nil
	}
	return "", errors.New("exit status 1")
}

const (
	headSHA = "1111111111111111111111111111111111111111"
	tagSHA  = "2222222222222222222222222222222222222222"
	baseSHA = "3333333333333333333333333333333333333333"
	oldSHA  = "4444444444444444444444444444444444444444"
)

func with(extra map[string]string) map[string]string {
	m := map[string]string{
		"rev-parse --is-shallow-repository":        "false\n",
		"rev-parse --verify --quiet HEAD^{commit}": headSHA + "\n",
	}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

func TestResolve(t *testing.T) {
	cases := []struct {
		name      string
		answers   map[string]string
		opts      Options
		want      Refs
		wantErr   string
		wantErrIs error
		notCalled string
	}{
		{
			name:    "explicit from and to",
			answers: with(map[string]string{"rev-parse --verify --quiet HEAD~1^{commit}": oldSHA}),
			opts:    Options{From: "HEAD~1", To: "HEAD"},
			want:    Refs{From: oldSHA, To: headSHA, FromSource: "explicit"},
		},
		{
			name:    "explicit from with auto to",
			answers: with(map[string]string{"rev-parse --verify --quiet v1^{commit}": oldSHA}),
			opts:    Options{From: "v1", To: "auto"},
			want:    Refs{From: oldSHA, To: headSHA, FromSource: "explicit"},
		},
		{
			name:    "explicit from that is not a commit",
			answers: with(nil),
			opts:    Options{From: "nosuchref"},
			wantErr: "--from nosuchref",
		},
		{
			name: "target with a deployed tag wins over the branch",
			answers: with(map[string]string{
				"rev-parse --verify --quiet refs/tags/glassops/deployed/qa^{commit}": tagSHA,
				"merge-base HEAD origin/develop":                                     baseSHA,
			}),
			opts:      Options{From: "auto", To: "auto", Target: "qa", Branch: "develop"},
			want:      Refs{From: tagSHA, To: headSHA, FromSource: "tag"},
			notCalled: "merge-base",
		},
		{
			name:    "target without a tag uses the origin branch merge base",
			answers: with(map[string]string{"merge-base HEAD origin/develop": baseSHA}),
			opts:    Options{Target: "qa", Branch: "develop"},
			want:    Refs{From: baseSHA, To: headSHA, FromSource: "merge-base"},
		},
		{
			name:    "target without a tag falls back to the local branch",
			answers: with(map[string]string{"merge-base HEAD develop": baseSHA}),
			opts:    Options{Target: "qa", Branch: "develop"},
			want:    Refs{From: baseSHA, To: headSHA, FromSource: "merge-base"},
		},
		{
			name:    "target with no tag and no branch is an error",
			answers: with(nil),
			opts:    Options{Target: "qa"},
			wantErr: "Pass --from.",
		},
		{
			name:    "target with a branch that does not exist is an error",
			answers: with(nil),
			opts:    Options{Target: "qa", Branch: "develop"},
			wantErr: "no merge base with origin/develop or develop",
		},
		{
			name:    "pull request base without a target",
			answers: with(map[string]string{"merge-base HEAD origin/main": baseSHA}),
			opts:    Options{BaseBranch: "main"},
			want:    Refs{From: baseSHA, To: headSHA, FromSource: "merge-base"},
		},
		{
			name:    "pull request base falls back to the local branch",
			answers: with(map[string]string{"merge-base HEAD main": baseSHA}),
			opts:    Options{BaseBranch: "main"},
			want:    Refs{From: baseSHA, To: headSHA, FromSource: "merge-base"},
		},
		{
			name:    "pull request base that cannot be found is an error",
			answers: with(nil),
			opts:    Options{BaseBranch: "main"},
			wantErr: "no merge base with the pull request base branch main",
		},
		{
			name:    "a target wins over the pull request base",
			answers: with(map[string]string{"merge-base HEAD origin/main": baseSHA}),
			opts:    Options{Target: "qa", BaseBranch: "main"},
			wantErr: "target qa has no tag",
		},
		{
			name:    "nothing to go on is an error telling the user to pass from",
			answers: with(nil),
			opts:    Options{},
			wantErr: "Pass --from.",
		},
		{
			name:      "shallow clone is an error with the fix",
			answers:   map[string]string{"rev-parse --is-shallow-repository": "true\n"},
			opts:      Options{From: "auto", Target: "qa", Branch: "develop"},
			wantErr:   "fetch-depth: 0",
			wantErrIs: ErrShallow,
			notCalled: "merge-base",
		},
		{
			name:      "shallow clone is an error even with explicit refs",
			answers:   map[string]string{"rev-parse --is-shallow-repository": "true"},
			opts:      Options{From: "a", To: "b"},
			wantErr:   "fetch-depth: 0",
			wantErrIs: ErrShallow,
		},
		{
			name:    "not a git repository",
			answers: map[string]string{},
			opts:    Options{From: "HEAD~1"},
			wantErr: "not a git repository",
		},
		{
			name:    "to that is not a commit",
			answers: with(nil),
			opts:    Options{From: "auto", To: "nosuchref"},
			wantErr: "--to nosuchref",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeGit{answers: tc.answers}
			got, err := Resolve(f.run, tc.opts)
			switch {
			case tc.wantErr != "":
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %v, want containing %q", err, tc.wantErr)
				}
				if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
					t.Fatalf("error = %v, want %v", err, tc.wantErrIs)
				}
			case err != nil:
				t.Fatal(err)
			case got != tc.want:
				t.Fatalf("Resolve() = %+v, want %+v", got, tc.want)
			}
			if tc.notCalled != "" {
				for _, c := range f.calls {
					if strings.HasPrefix(c, tc.notCalled) {
						t.Fatalf("git %q should not have run", c)
					}
				}
			}
		})
	}
}
