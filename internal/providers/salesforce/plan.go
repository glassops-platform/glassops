package salesforce

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/glassops-platform/glassops/internal/providers"
)

// Plan layout under OutDir. sfdx-git-delta writes both manifests in delta mode. Full mode
// writes only the package manifest. sf treats destructiveChanges.xml as post-destructive.
const (
	planPackage     = "package/package.xml"
	planDestructive = "destructiveChanges/destructiveChanges.xml"
)

// Plan works out the change set with sfdx-git-delta, or from the whole source when req.Full
// is set. It needs no login. Everything it writes goes under req.OutDir, which must sit
// inside <req.Dir>/.glassops/.
func (p *Provider) Plan(ctx context.Context, req providers.PlanRequest) (providers.ChangeSet, error) {
	dir, err := filepath.Abs(req.Dir)
	if err != nil {
		return providers.ChangeSet{}, err
	}
	out, err := safePlanDir(dir, req.OutDir)
	if err != nil {
		return providers.ChangeSet{}, err
	}
	if err := os.RemoveAll(out); err != nil {
		return providers.ChangeSet{}, fmt.Errorf("clear %s: %w", out, err)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return providers.ChangeSet{}, fmt.Errorf("create %s: %w", out, err)
	}

	pkgDirs, err := packageDirectories(dir)
	if err != nil {
		return providers.ChangeSet{}, err
	}
	if req.Full {
		err = p.planFull(ctx, dir, out, pkgDirs)
	} else {
		err = p.planDelta(ctx, dir, out, req.From, req.To, pkgDirs)
	}
	if err != nil {
		return providers.ChangeSet{}, err
	}

	pkg, err := readManifest(filepath.Join(out, filepath.FromSlash(planPackage)))
	if err != nil {
		return providers.ChangeSet{}, err
	}
	var dest manifest
	destPath := filepath.Join(out, filepath.FromSlash(planDestructive))
	if _, statErr := os.Stat(destPath); statErr == nil {
		if dest, err = readManifest(destPath); err != nil {
			return providers.ChangeSet{}, err
		}
	}

	relOut := relSlash(dir, out)
	cs := providers.ChangeSet{
		Schema:      providers.PlanSchema,
		Provider:    p.Name(),
		From:        req.From,
		To:          req.To,
		Full:        req.Full,
		Dir:         relOut,
		Manifest:    relOut + "/" + planPackage,
		Destructive: []string{},
		Empty:       pkg.memberCount() == 0 && dest.memberCount() == 0,
	}
	if len(dest.Types) > 0 {
		cs.Destructive = append(cs.Destructive, relOut+"/"+planDestructive)
	}
	if cs.SHA256, err = planDigest(dir, out); err != nil {
		return providers.ChangeSet{}, err
	}
	cs.Facts = planFacts(pkg, dest)
	return cs, nil
}

func (p *Provider) planDelta(ctx context.Context, dir, out, from, to string, pkgDirs []string) error {
	args := []string{"sgd", "source", "delta", "--from", from, "--to", to, "--output-dir", out}
	// sgd does not read .forceignore on its own. Its --ignore-file takes the same syntax.
	if _, err := os.Stat(filepath.Join(dir, ".forceignore")); err == nil {
		args = append(args, "--ignore-file", ".forceignore")
	}
	for _, d := range pkgDirs {
		args = append(args, "--source-dir", d)
	}
	env, err := p.sf(ctx, nil, dir, args...)
	if err != nil {
		var sfErr *SFError
		if errors.As(err, &sfErr) {
			// sgd reports a bad ref as status 1 with the text in result.error.
			if msg := resultError(env.Result); msg != "" {
				return fmt.Errorf("sfdx-git-delta: %s", msg)
			}
		}
		return err
	}
	if msg := resultError(env.Result); msg != "" {
		return fmt.Errorf("sfdx-git-delta: %s", msg)
	}
	return nil
}

func (p *Provider) planFull(ctx context.Context, dir, out string, pkgDirs []string) error {
	pkgOut := filepath.Join(out, "package")
	if err := os.MkdirAll(pkgOut, 0o755); err != nil {
		return err
	}
	args := []string{"project", "generate", "manifest", "--output-dir", pkgOut, "--name", "package"}
	for _, d := range pkgDirs {
		args = append(args, "--source-dir", d)
	}
	_, err := p.sf(ctx, nil, dir, args...)
	return err
}

// resultError reads result.error from an sgd envelope.
func resultError(raw json.RawMessage) string {
	var r struct {
		Error string `json:"error"`
	}
	if len(raw) == 0 || json.Unmarshal(raw, &r) != nil {
		return ""
	}
	return strings.TrimSpace(r.Error)
}

// safePlanDir returns outDir as an absolute path, and refuses unless it is a directory
// strictly inside <dir>/.glassops and not a symlink. Plan clears it before writing.
func safePlanDir(dir, outDir string) (string, error) {
	if outDir == "" {
		return "", errors.New("plan: no output directory")
	}
	out, err := filepath.Abs(outDir)
	if err != nil {
		return "", err
	}
	root := filepath.Join(dir, ".glassops")
	rel, err := filepath.Rel(root, out)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("plan: refusing to clear %s, which is not inside %s", out, root)
	}
	if fi, err := os.Lstat(out); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("plan: refusing to clear %s, which is a symlink", out)
	}
	return out, nil
}

// packageDirectories lists the package directory paths in sfdx-project.json.
func packageDirectories(dir string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(dir, ProjectFile))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", ProjectFile, err)
	}
	var proj struct {
		PackageDirectories []struct {
			Path string `json:"path"`
		} `json:"packageDirectories"`
	}
	if err := json.Unmarshal(data, &proj); err != nil {
		return nil, fmt.Errorf("parse %s: %w", ProjectFile, err)
	}
	var dirs []string
	for _, d := range proj.PackageDirectories {
		if d.Path != "" {
			dirs = append(dirs, d.Path)
		}
	}
	if len(dirs) == 0 {
		return nil, fmt.Errorf("%s lists no package directories", ProjectFile)
	}
	return dirs, nil
}

type manifest struct {
	Types []manifestType `xml:"types"`
}

type manifestType struct {
	Name    string   `xml:"name"`
	Members []string `xml:"members"`
}

func (m manifest) memberCount() int {
	n := 0
	for _, t := range m.Types {
		n += len(t.Members)
	}
	return n
}

func (m manifest) members(typeName string) []string {
	var out []string
	for _, t := range m.Types {
		if t.Name == typeName {
			out = append(out, t.Members...)
		}
	}
	sort.Strings(out)
	return out
}

func readManifest(path string) (manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	return parseManifest(data, path)
}

func parseManifest(data []byte, name string) (manifest, error) {
	var m manifest
	if err := xml.Unmarshal(data, &m); err != nil {
		return manifest{}, fmt.Errorf("parse manifest %s: %w", filepath.Base(name), err)
	}
	return m, nil
}

// planDigest is the plan_sha256. It hashes every file under out, in order of its path
// relative to dir with forward slashes. For each file it writes the relative path, a newline
// and the file content. Phase 3 recomputes it the same way before a deploy.
func planDigest(dir, out string) (string, error) {
	var files []string
	err := filepath.WalkDir(out, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			files = append(files, relSlash(dir, path))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	h := sha256.New()
	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return "", err
		}
		h.Write([]byte(rel))
		h.Write([]byte("\n"))
		h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func relSlash(base, path string) string {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

// planFacts summarises the manifests. A deploy replaces a permission set or profile whole,
// so the ones listed here are overwritten in full, not merged.
func planFacts(pkg, dest manifest) map[string]any {
	counts := map[string]int{}
	for _, t := range pkg.Types {
		counts[t.Name] += len(t.Members)
	}
	list := func(xs []string) []string {
		if xs == nil {
			return []string{}
		}
		return xs
	}
	return map[string]any{
		"salesforce.counts":            counts,
		"salesforce.destructive_count": dest.memberCount(),
		"salesforce.apex_classes":      list(pkg.members("ApexClass")),
		"salesforce.permission_sets":   list(pkg.members("PermissionSet")),
		"salesforce.profiles":          list(pkg.members("Profile")),
	}
}
