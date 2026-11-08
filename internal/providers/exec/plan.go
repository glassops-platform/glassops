package exec

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/glassops-platform/glassops/internal/providers"
)

// The files plan writes into the plan dir. Each lists one path per line, relative to the
// project root with forward slashes, sorted.
const (
	// ChangesFile lists the files added or changed under exec.paths. It is the manifest.
	ChangesFile = "changes.txt"
	// DeletedFile lists the files deleted under exec.paths. It is written only when there
	// are any, and it is the change set's destructive manifest.
	DeletedFile = "deleted.txt"
)

// Plan lists the files under exec.paths that changed between req.From and req.To, or every
// tracked file under them at req.To (HEAD when empty) for a full plan. It writes them to
// changes.txt and deleted.txt in req.OutDir, runs exec.plan when set so it can add more
// files there, and hashes everything in the plan dir with providers.PlanDigest. The PERMIT
// binds that digest and the commit, so it binds the file list and the content at the commit.
func (p *Provider) Plan(ctx context.Context, req providers.PlanRequest) (providers.ChangeSet, error) {
	dir, err := filepath.Abs(req.Dir)
	if err != nil {
		return providers.ChangeSet{}, err
	}
	if p.ws.Dir == "" {
		p.ws.Dir = dir
	}
	out, err := planDir(dir, req.OutDir)
	if err != nil {
		return providers.ChangeSet{}, err
	}
	paths, err := cleanPaths(p.cfg.Paths)
	if err != nil {
		return providers.ChangeSet{}, err
	}
	if err := os.RemoveAll(out); err != nil {
		return providers.ChangeSet{}, fmt.Errorf("clear %s: %w", out, err)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return providers.ChangeSet{}, fmt.Errorf("create %s: %w", out, err)
	}

	var changed, deleted []string
	if req.Full {
		ref := req.To
		if ref == "" {
			ref = "HEAD"
		}
		changed, err = p.listTree(ctx, dir, ref, paths)
	} else {
		if req.From == "" || req.To == "" {
			return providers.ChangeSet{}, fmt.Errorf("a delta plan needs both refs, got from %q and to %q", req.From, req.To)
		}
		changed, deleted, err = p.diff(ctx, dir, req.From, req.To, paths)
	}
	if err != nil {
		return providers.ChangeSet{}, err
	}
	if err := writeList(filepath.Join(out, ChangesFile), changed); err != nil {
		return providers.ChangeSet{}, err
	}
	relOut := relSlash(dir, out)
	cs := providers.ChangeSet{
		Schema:      providers.PlanSchema,
		Provider:    p.Name(),
		From:        req.From,
		To:          req.To,
		Full:        req.Full,
		Dir:         relOut,
		Manifest:    relOut + "/" + ChangesFile,
		Destructive: []string{},
	}
	if len(deleted) > 0 {
		if err := writeList(filepath.Join(out, DeletedFile), deleted); err != nil {
			return providers.ChangeSet{}, err
		}
		cs.Destructive = []string{relOut + "/" + DeletedFile}
	}

	facts := map[string]any{}
	if strings.TrimSpace(p.cfg.Plan) != "" {
		extra := []string{EnvFrom + "=" + req.From, EnvTo + "=" + req.To}
		o, err := p.step(ctx, stepPlan, p.cfg.Plan, dir, nil, relOut, extra)
		if err != nil {
			return providers.ChangeSet{}, err
		}
		if !o.passed {
			return providers.ChangeSet{}, fmt.Errorf("exec.plan failed: %s", strings.Join(o.problems, "; "))
		}
		for k, v := range o.facts {
			facts[k] = v
		}
	}

	others, err := otherFiles(dir, out)
	if err != nil {
		return providers.ChangeSet{}, err
	}
	digest, err := providers.PlanDigest(dir, out)
	if err != nil {
		return providers.ChangeSet{}, fmt.Errorf("hash the plan files: %w", err)
	}
	cs.SHA256 = digest
	cs.Empty = len(changed) == 0 && len(deleted) == 0 && len(others) == 0
	facts["exec.paths"] = paths
	facts["exec.changed_count"] = len(changed)
	facts["exec.destructive_count"] = len(deleted)
	facts["exec.plan_files"] = len(others)
	facts["exec.plan_command"] = strings.TrimSpace(p.cfg.Plan) != ""
	cs.Facts = facts
	return cs, nil
}

// planDir resolves out and refuses one outside <dir>/.glassops, because plan clears it.
func planDir(dir, out string) (string, error) {
	if out == "" {
		out = filepath.Join(dir, ".glassops", "plan")
	}
	abs, err := filepath.Abs(out)
	if err != nil {
		return "", err
	}
	root := filepath.Join(dir, ".glassops")
	rel, err := filepath.Rel(root, abs)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("plan dir %s must be inside %s", abs, root)
	}
	return abs, nil
}

// cleanPaths checks exec.paths: each must be a path inside the project, written relative to
// its root.
func cleanPaths(in []string) ([]string, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("exec.paths is empty, so there is nothing to plan from")
	}
	var outPaths []string
	for _, raw := range in {
		s := filepath.ToSlash(strings.TrimSpace(raw))
		c := path.Clean(s)
		if s == "" || path.IsAbs(c) || filepath.IsAbs(raw) || c == ".." || strings.HasPrefix(c, "../") || strings.HasPrefix(c, ":") {
			return nil, fmt.Errorf("exec.paths: %q must be a path inside the project, relative to its root", raw)
		}
		outPaths = append(outPaths, c)
	}
	return outPaths, nil
}

// diff lists the files under paths that changed from from to to, and the ones deleted.
// Renames count as a delete and an add, so the old path is never lost.
func (p *Provider) diff(ctx context.Context, dir, from, to string, paths []string) (changed, deleted []string, err error) {
	// --literal-pathspecs keeps a path from being read as a glob. --relative keeps the output
	// relative to the project dir when it is below the repository root.
	args := append([]string{"--literal-pathspecs", "diff", "--name-status", "--no-renames", "--relative", "-z", from, to, "--"}, paths...)
	raw, err := p.git(ctx, dir, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list the changed files: %w", err)
	}
	fields := splitNUL(raw)
	if len(fields)%2 != 0 {
		return nil, nil, fmt.Errorf("list the changed files: git diff output has an odd number of fields")
	}
	for i := 0; i < len(fields); i += 2 {
		status, file := fields[i], fields[i+1]
		if err := checkName(file); err != nil {
			return nil, nil, err
		}
		if strings.HasPrefix(status, "D") {
			deleted = append(deleted, file)
		} else {
			changed = append(changed, file)
		}
	}
	sort.Strings(changed)
	sort.Strings(deleted)
	return changed, deleted, nil
}

// listTree lists every tracked file under paths at ref.
func (p *Provider) listTree(ctx context.Context, dir, ref string, paths []string) ([]string, error) {
	// Without --full-tree, ls-tree reads paths and prints names relative to the project dir.
	args := append([]string{"--literal-pathspecs", "ls-tree", "-r", "-z", "--name-only", ref, "--"}, paths...)
	raw, err := p.git(ctx, dir, args...)
	if err != nil {
		return nil, fmt.Errorf("list the files at %s: %w", ref, err)
	}
	files := splitNUL(raw)
	for _, f := range files {
		if err := checkName(f); err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	return files, nil
}

// checkName refuses a file name the one-per-line plan files cannot hold.
func checkName(name string) error {
	if strings.ContainsAny(name, "\n\r") {
		return fmt.Errorf("file %q has a line break in its name, which the plan cannot list", name)
	}
	return nil
}

func splitNUL(raw []byte) []string {
	var out []string
	for _, f := range bytes.Split(raw, []byte{0}) {
		if len(f) > 0 {
			out = append(out, string(f))
		}
	}
	return out
}

// writeList writes one entry per line, each ending in a newline.
func writeList(file string, entries []string) error {
	var b strings.Builder
	for _, e := range entries {
		b.WriteString(e)
		b.WriteByte('\n')
	}
	if err := os.WriteFile(file, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", filepath.Base(file), err)
	}
	return nil
}

// otherFiles lists the regular files in out that plan itself did not write, which are the
// ones exec.plan added.
func otherFiles(dir, out string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(out, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(out, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel != ChangesFile && rel != DeletedFile {
			files = append(files, relSlash(dir, p))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list the plan files: %w", err)
	}
	return files, nil
}

func relSlash(dir, p string) string {
	rel, err := filepath.Rel(dir, p)
	if err != nil {
		return filepath.ToSlash(p)
	}
	return filepath.ToSlash(rel)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
