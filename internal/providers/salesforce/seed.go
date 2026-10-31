package salesforce

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

// A destructive change cannot be validated on a fresh scratch org, because the components it
// deletes were never there: the validate fails with "No CustomField named ... found". Seed
// puts them there first, from the base ref, with a real deploy to the scratch org.

// seedWaitMinutes is how long sf waits for the seed deploy.
const seedWaitMinutes = "60"

// childTypes are the metadata types whose members are named <Object>.<Child> and need their
// object to exist.
var childTypes = map[string]bool{
	"CustomField":     true,
	"RecordType":      true,
	"ValidationRule":  true,
	"ListView":        true,
	"FieldSet":        true,
	"CompactLayout":   true,
	"WebLink":         true,
	"BusinessProcess": true,
}

// customObjectSuffixes mark an object that a scratch org does not have until it is deployed.
// A standard object, such as Account, already exists.
var customObjectSuffixes = []string{"__c", "__mdt", "__e"}

// missingEntity is the error a minimal seed fails with when a component needs something the
// seed did not include, such as "Entity 'Widget__c' not found."
var missingEntity = regexp.MustCompile(`Entity '([^']+)' not found`)

// Seed deploys the base state req.Plan's destructive members need to the scratch org in s.
// In minimal mode it deploys those members and their parent objects from req.BaseRef. If that
// fails because something else is missing, it deploys the whole source at req.BaseRef and
// sets Fallback. In full mode it deploys the whole source at once. Off, or a plan with no
// destructive members, seeds nothing.
func (p *Provider) Seed(ctx context.Context, s providers.Session, req providers.SeedRequest) (providers.SeedResult, error) {
	sess, ok := asSession(s)
	if !ok {
		return providers.SeedResult{}, errors.New("salesforce: seed needs a session from this provider")
	}
	none := providers.SeedResult{Mode: "none", Members: []string{}, Facts: map[string]any{}}
	switch req.Mode {
	case config.SeedOff:
		return none, nil
	case config.SeedMinimal, config.SeedFull:
	default:
		return providers.SeedResult{}, fmt.Errorf("salesforce: seed mode %q is not minimal, full or off", req.Mode)
	}
	dir, err := filepath.Abs(req.Dir)
	if err != nil {
		return providers.SeedResult{}, fmt.Errorf("salesforce: resolve project dir: %w", err)
	}
	deleted, err := destructiveMembers(dir, req.Plan.Destructive)
	if err != nil {
		return providers.SeedResult{}, err
	}
	if len(deleted) == 0 {
		return none, nil
	}
	if req.BaseRef == "" || strings.HasPrefix(req.BaseRef, "-") {
		return providers.SeedResult{}, errors.New("salesforce: seed needs the base ref the deleted components come from")
	}

	proj, pkgDirs, err := p.seedProject(ctx, dir, req.BaseRef)
	if err != nil {
		return providers.SeedResult{}, err
	}
	defer os.RemoveAll(proj)

	if req.Mode == config.SeedMinimal {
		members := withParents(deleted)
		args := []string{"project", "deploy", "start"}
		for _, m := range members {
			args = append(args, "--metadata", m)
		}
		out, err := p.seedDeploy(ctx, sess, proj, args)
		if err == nil {
			return providers.SeedResult{Mode: config.SeedMinimal, Members: members, Facts: out.facts()}, nil
		}
		var miss *missingError
		if !errors.As(err, &miss) {
			return providers.SeedResult{}, err
		}
		res, ferr := p.seedFull(ctx, sess, proj, pkgDirs)
		if ferr != nil {
			return providers.SeedResult{}, fmt.Errorf("salesforce: the minimal seed needed %s, and the full seed failed: %w", miss.entity, ferr)
		}
		res.Fallback = true
		res.Facts["salesforce.seed_missing"] = miss.entity
		return res, nil
	}
	return p.seedFull(ctx, sess, proj, pkgDirs)
}

func (p *Provider) seedFull(ctx context.Context, sess *session, proj string, pkgDirs []string) (providers.SeedResult, error) {
	args := []string{"project", "deploy", "start"}
	for _, d := range pkgDirs {
		args = append(args, "--source-dir", d)
	}
	out, err := p.seedDeploy(ctx, sess, proj, args)
	if err != nil {
		return providers.SeedResult{}, err
	}
	return providers.SeedResult{Mode: config.SeedFull, Members: out.deployed(), Facts: out.facts()}, nil
}

// destructiveMembers lists the members of the destructive manifests as "Type:Name", sorted.
func destructiveMembers(dir string, manifests []string) ([]string, error) {
	seen := map[string]bool{}
	for _, rel := range manifests {
		m, err := readManifest(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return nil, fmt.Errorf("salesforce: seed: %w", err)
		}
		for _, t := range m.Types {
			for _, n := range t.Members {
				seen[t.Name+":"+n] = true
			}
		}
	}
	return sortedSet(seen), nil
}

// withParents adds the custom object each child member belongs to.
func withParents(members []string) []string {
	seen := map[string]bool{}
	for _, m := range members {
		seen[m] = true
		typ, name, _ := strings.Cut(m, ":")
		if !childTypes[typ] {
			continue
		}
		obj, _, ok := strings.Cut(name, ".")
		if !ok || !isCustomObject(obj) {
			continue
		}
		seen["CustomObject:"+obj] = true
	}
	return sortedSet(seen)
}

func isCustomObject(name string) bool {
	for _, s := range customObjectSuffixes {
		if strings.HasSuffix(name, s) {
			return true
		}
	}
	return false
}

func sortedSet(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// seedProject writes the project as it was at ref into a new temp dir: sfdx-project.json,
// .forceignore when there is one, and every package directory. It returns the dir and the
// package directories.
func (p *Provider) seedProject(ctx context.Context, dir, ref string) (string, []string, error) {
	proj, err := os.MkdirTemp(p.tempDir, "glassops-seed-")
	if err != nil {
		return "", nil, fmt.Errorf("salesforce: create the seed project dir: %w", err)
	}
	fail := func(err error) (string, []string, error) {
		_ = os.RemoveAll(proj)
		return "", nil, err
	}

	data, err := p.git(ctx, dir, "show", ref+":"+ProjectFile)
	if err != nil {
		return fail(fmt.Errorf("salesforce: read %s at the base ref: %w", ProjectFile, err))
	}
	if err := os.WriteFile(filepath.Join(proj, ProjectFile), data, 0o644); err != nil {
		return fail(err)
	}
	pkgDirs, err := packageDirectories(proj)
	if err != nil {
		return fail(fmt.Errorf("salesforce: at the base ref: %w", err))
	}
	if data, err := p.git(ctx, dir, "show", ref+":.forceignore"); err == nil {
		if err := os.WriteFile(filepath.Join(proj, ".forceignore"), data, 0o644); err != nil {
			return fail(err)
		}
	}

	args := append([]string{"archive", "--format=tar", ref, "--"}, pkgDirs...)
	archive, err := p.git(ctx, dir, args...)
	if err != nil {
		return fail(fmt.Errorf("salesforce: read the source at the base ref: %w", err))
	}
	if err := extractTar(archive, proj); err != nil {
		return fail(fmt.Errorf("salesforce: unpack the source at the base ref: %w", err))
	}
	return proj, pkgDirs, nil
}

// git runs git in dir and returns stdout. A non-zero exit is an error that carries the first
// line of stderr.
func (p *Provider) git(ctx context.Context, dir string, args ...string) ([]byte, error) {
	out, err := p.exec(ctx, Cmd{Name: "git", Args: args, Dir: dir})
	if err != nil {
		return nil, err
	}
	if out.ExitCode != 0 {
		msg, _, _ := strings.Cut(strings.TrimSpace(string(out.Stderr)), "\n")
		if msg == "" {
			msg = "no message"
		}
		return nil, fmt.Errorf("git %s exited %d: %s", args[0], out.ExitCode, msg)
	}
	return out.Stdout, nil
}

// extractTar writes the regular files and dirs in data under root. An entry that would land
// outside root, such as an absolute path or one with "..", is an error. Links are skipped.
func extractTar(data []byte, root string) error {
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		name := path.Clean(h.Name)
		if path.IsAbs(name) || name == ".." || strings.HasPrefix(name, "../") || strings.Contains(name, ":") {
			return fmt.Errorf("archive entry %q is outside the project", h.Name)
		}
		target := filepath.Join(root, filepath.FromSlash(name))
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return err
			}
			_, err = io.Copy(f, tr)
			cerr := f.Close()
			if err != nil {
				return err
			}
			if cerr != nil {
				return cerr
			}
		}
	}
}

// missingError is a minimal seed that failed because a component needs entity, which the
// seed did not include.
type missingError struct {
	entity string
}

func (e *missingError) Error() string {
	return fmt.Sprintf("salesforce: the seed needs %s, which it does not include", e.entity)
}

// componentRef names one component in a deploy result.
type componentRef struct {
	ComponentType string `json:"componentType"`
	FullName      string `json:"fullName"`
}

// seedOutput is the deploy result of a seed.
type seedOutput struct {
	body      deployBody
	successes []componentRef
}

func (o seedOutput) facts() map[string]any {
	return map[string]any{
		"salesforce.seed_deploy_id":  o.body.ID,
		"salesforce.seed_components": int(o.body.NumberComponentsTotal),
	}
}

// deployed lists the components the deploy reported as "Type:Name", without package.xml.
func (o seedOutput) deployed() []string {
	seen := map[string]bool{}
	for _, c := range o.successes {
		if c.ComponentType == "" {
			continue
		}
		seen[c.ComponentType+":"+c.FullName] = true
	}
	return sortedSet(seen)
}

// seedDeploy runs the seed deploy from the temp project. A seed never runs tests, and it
// ignores source tracking conflicts, which stop a deploy to a scratch org otherwise. A
// failure that names a missing entity comes back as *missingError.
func (p *Provider) seedDeploy(ctx context.Context, sess *session, proj string, args []string) (seedOutput, error) {
	args = append(args, "--test-level", "NoTestRun", "--ignore-conflicts",
		"--target-org", sess.alias, "--wait", seedWaitMinutes)
	env, err := p.sf(ctx, sess, proj, args...)
	var sfErr *SFError
	if err != nil && !errors.As(err, &sfErr) {
		return seedOutput{}, fmt.Errorf("salesforce: seed deploy: %w", err)
	}
	var out seedOutput
	var wrap struct {
		Details struct {
			ComponentSuccesses flexList[componentRef] `json:"componentSuccesses"`
		} `json:"details"`
	}
	if len(env.Result) == 0 || json.Unmarshal(env.Result, &out.body) != nil || json.Unmarshal(env.Result, &wrap) != nil {
		if err != nil {
			return seedOutput{}, fmt.Errorf("salesforce: seed deploy: %s", scrub(err.Error(), nil))
		}
		return seedOutput{}, errors.New("salesforce: could not read the seed deploy result")
	}
	out.successes = wrap.Details.ComponentSuccesses
	if err == nil && out.body.Status == "Succeeded" {
		if out.body.CheckOnly != nil && *out.body.CheckOnly {
			return seedOutput{}, errors.New("salesforce: the seed deploy was check-only, so nothing was seeded")
		}
		return out, nil
	}
	for _, f := range out.body.Details.ComponentFailures {
		if m := missingEntity.FindStringSubmatch(f.Problem); m != nil {
			return out, &missingError{entity: m[1]}
		}
	}
	problems := out.body.toResult(out.body.ID).Problems
	return out, fmt.Errorf("salesforce: seed deploy %s: %s", out.body.Status, scrub(strings.Join(problems, " / "), nil))
}
