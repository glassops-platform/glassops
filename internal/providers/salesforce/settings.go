package salesforce

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// A scratch org definition's settings key asks for org settings that ScratchOrgInfo has no
// field for. sf applies them by deploying Settings metadata to the new org before it hands
// the org back. env create does the same: each <name>Settings key becomes one
// <Name>.settings-meta.xml file in a throwaway source project, deployed to the new org after
// the login check. A settings deploy that fails deletes the org, so a run never goes on in an
// org that is missing part of what its definition asked for.

const (
	// metadataNS is the Metadata API namespace every Settings file declares.
	metadataNS = "http://soap.sforce.com/2006/04/metadata"
	// settingsPkgDir is the package directory of the throwaway settings project.
	settingsPkgDir = "force-app"
	// settingsWaitMinutes is how long sf waits for the settings deploy.
	settingsWaitMinutes = "30"
	// settingsCleanupTimeout bounds the org delete after a failed settings deploy. The
	// delete runs even when the run's own context has ended.
	settingsCleanupTimeout = 2 * time.Minute
)

var (
	// settingsKey is a definition settings key, such as lightningExperienceSettings.
	settingsKey = regexp.MustCompile(`^[a-z][A-Za-z0-9]*Settings$`)
	// settingsElement is a field name inside a settings block.
	settingsElement = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
)

// settingsFile is one Settings metadata file converted from a definition settings key.
type settingsFile struct {
	// Name is the metadata member name, such as LightningExperience.
	Name string
	// XML is the file content.
	XML []byte
}

// Path is the file's path in a source project, relative to the package directory.
func (f settingsFile) Path() string {
	return "settings/" + f.Name + ".settings-meta.xml"
}

// node is a JSON value that keeps the order of object keys, so the XML lists fields in the
// order the definition gives them, as sf does.
type node struct {
	// kind is "object", "array", "scalar" or "null".
	kind  string
	keys  []string
	vals  []node
	items []node
	// text is a scalar written as XML text: a string as is, true or false, or a number as
	// the definition spells it.
	text string
}

// parseNode reads one JSON value from dec.
func parseNode(dec *json.Decoder) (node, error) {
	tok, err := dec.Token()
	if err != nil {
		return node{}, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			n := node{kind: "object"}
			seen := map[string]bool{}
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return node{}, err
				}
				k, ok := kt.(string)
				if !ok {
					return node{}, errors.New("an object key is not a string")
				}
				if seen[k] {
					return node{}, fmt.Errorf("key %s appears twice", k)
				}
				seen[k] = true
				v, err := parseNode(dec)
				if err != nil {
					return node{}, err
				}
				n.keys = append(n.keys, k)
				n.vals = append(n.vals, v)
			}
			if _, err := dec.Token(); err != nil {
				return node{}, err
			}
			return n, nil
		case '[':
			n := node{kind: "array"}
			for dec.More() {
				v, err := parseNode(dec)
				if err != nil {
					return node{}, err
				}
				n.items = append(n.items, v)
			}
			if _, err := dec.Token(); err != nil {
				return node{}, err
			}
			return n, nil
		}
		return node{}, fmt.Errorf("unexpected %v", t)
	case string:
		return node{kind: "scalar", text: t}, nil
	case bool:
		return node{kind: "scalar", text: fmt.Sprint(t)}, nil
	case json.Number:
		return node{kind: "scalar", text: t.String()}, nil
	case nil:
		return node{kind: "null"}, nil
	}
	return node{}, fmt.Errorf("unexpected token %v", tok)
}

// convertSettings turns the definition's settings object into Settings metadata files, one
// per key, in the order the definition lists them. objectSettings is refused: it describes
// record types and sharing for objects, which sf applies as object metadata, not Settings.
func convertSettings(raw json.RawMessage) ([]settingsFile, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	root, err := parseNode(dec)
	if err != nil {
		return nil, fmt.Errorf("salesforce: read the definition settings: %w", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("salesforce: the definition settings hold more than one value")
	}
	if root.kind != "object" {
		return nil, errors.New("salesforce: the definition settings must be an object")
	}
	files := make([]settingsFile, 0, len(root.keys))
	for i, key := range root.keys {
		if key == "objectSettings" {
			return nil, errors.New("salesforce: env create cannot apply objectSettings, which is object metadata, not Settings")
		}
		if !settingsKey.MatchString(key) {
			return nil, fmt.Errorf("salesforce: definition settings key %q does not name a Settings type, such as lightningExperienceSettings", key)
		}
		block := root.vals[i]
		if block.kind != "object" {
			return nil, fmt.Errorf("salesforce: definition settings %s must be an object", key)
		}
		typeName := strings.ToUpper(key[:1]) + key[1:]
		var buf bytes.Buffer
		buf.WriteString(xml.Header)
		buf.WriteString("<" + typeName + ` xmlns="` + metadataNS + `">` + "\n")
		if err := writeFields(&buf, key, block, 1); err != nil {
			return nil, err
		}
		buf.WriteString("</" + typeName + ">\n")
		files = append(files, settingsFile{Name: strings.TrimSuffix(typeName, "Settings"), XML: buf.Bytes()})
	}
	return files, nil
}

// writeFields writes the fields of object n as child elements at depth. A nested object
// becomes an element with its own children, and an array becomes one element per item.
func writeFields(buf *bytes.Buffer, path string, n node, depth int) error {
	for i, k := range n.keys {
		if !settingsElement.MatchString(k) {
			return fmt.Errorf("salesforce: definition settings field %s.%s is not a valid element name", path, k)
		}
		v := n.vals[i]
		if v.kind == "array" {
			for _, it := range v.items {
				if it.kind == "array" {
					return fmt.Errorf("salesforce: definition settings field %s.%s holds a list inside a list", path, k)
				}
				if err := writeElement(buf, path+"."+k, k, it, depth); err != nil {
					return err
				}
			}
			continue
		}
		if err := writeElement(buf, path+"."+k, k, v, depth); err != nil {
			return err
		}
	}
	return nil
}

// writeElement writes one element named name holding v.
func writeElement(buf *bytes.Buffer, path, name string, v node, depth int) error {
	indent := strings.Repeat("    ", depth)
	switch v.kind {
	case "null":
		return fmt.Errorf("salesforce: definition settings field %s is null", path)
	case "scalar":
		buf.WriteString(indent + "<" + name + ">")
		if err := xml.EscapeText(buf, []byte(v.text)); err != nil {
			return fmt.Errorf("salesforce: definition settings field %s: %w", path, err)
		}
		buf.WriteString("</" + name + ">\n")
	case "object":
		buf.WriteString(indent + "<" + name + ">\n")
		if err := writeFields(buf, path, v, depth+1); err != nil {
			return err
		}
		buf.WriteString(indent + "</" + name + ">\n")
	}
	return nil
}

// settingsNames lists the member names of files, for facts.
func settingsNames(files []settingsFile) []string {
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Name
	}
	return names
}

// SettingsError is a definition's settings that did not deploy to a new environment. By the
// time CreateEnv returns it, CreateEnv has tried to delete the environment.
type SettingsError struct {
	// Problems are the component failures the org reported, one line each.
	Problems []string
	// Rejected is true when the org answered the deploy and refused part of it, which is a
	// problem in the definition. It is false when the deploy could not run or be read.
	Rejected bool
	// Deleted is true when the environment was deleted afterwards.
	Deleted bool
	// DeleteErr is why the delete failed, when it did.
	DeleteErr error
	cause     error
}

func (e *SettingsError) Error() string {
	var b strings.Builder
	b.WriteString("salesforce: the definition settings did not deploy")
	switch {
	case len(e.Problems) > 0:
		b.WriteString(": " + strings.Join(e.Problems, " / "))
	case e.cause != nil:
		b.WriteString(": " + e.cause.Error())
	}
	if e.Deleted {
		b.WriteString(". The scratch org was deleted")
	} else if e.DeleteErr != nil {
		b.WriteString(". Deleting the scratch org failed too: " + e.DeleteErr.Error())
	}
	return b.String()
}

func (e *SettingsError) Unwrap() error { return e.cause }

// CheckFailed reports a failure of the definition rather than of the tooling: the org
// refused the settings and the environment is gone. Anything else stays an error, because
// either the deploy never got an answer or an org is still alive.
func (e *SettingsError) CheckFailed() bool { return e.Rejected && e.Deleted }

// deploySettings writes files into a throwaway source project and deploys it to the org in
// sess. It returns the deploy Id when the org reports one. A failure is a *SettingsError
// without the delete fields set.
func (p *Provider) deploySettings(ctx context.Context, sess *session, files []settingsFile) (string, error) {
	proj, err := os.MkdirTemp(p.tempDir, "glassops-settings-")
	if err != nil {
		return "", &SettingsError{cause: fmt.Errorf("create the settings project dir: %w", err)}
	}
	defer os.RemoveAll(proj)
	project := `{"packageDirectories":[{"path":"` + settingsPkgDir + `","default":true}]}` + "\n"
	if err := os.WriteFile(filepath.Join(proj, ProjectFile), []byte(project), 0o644); err != nil {
		return "", &SettingsError{cause: fmt.Errorf("write the settings project: %w", err)}
	}
	dir := filepath.Join(proj, settingsPkgDir, "settings")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", &SettingsError{cause: fmt.Errorf("create the settings dir: %w", err)}
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(proj, settingsPkgDir, filepath.FromSlash(f.Path())), f.XML, 0o644); err != nil {
			return "", &SettingsError{cause: fmt.Errorf("write %s: %w", f.Path(), err)}
		}
	}

	// A settings deploy runs no tests, and it ignores source tracking conflicts, which stop a
	// deploy to a scratch org otherwise.
	env, err := p.sf(ctx, sess, proj, "project", "deploy", "start", "--source-dir", settingsPkgDir,
		"--test-level", "NoTestRun", "--ignore-conflicts", "--target-org", sess.alias, "--wait", settingsWaitMinutes)
	var sfErr *SFError
	if err != nil && !errors.As(err, &sfErr) {
		return "", &SettingsError{cause: errors.New(scrub(err.Error(), nil))}
	}
	var body deployBody
	if len(env.Result) == 0 || json.Unmarshal(env.Result, &body) != nil || body.Status == "" {
		if err != nil {
			return "", &SettingsError{cause: errors.New(scrub(err.Error(), nil))}
		}
		return "", &SettingsError{cause: errors.New("could not read the settings deploy result")}
	}
	if err == nil && body.Status == "Succeeded" {
		if body.CheckOnly != nil && *body.CheckOnly {
			return body.ID, &SettingsError{cause: errors.New("the settings deploy was check-only, so nothing was applied")}
		}
		return body.ID, nil
	}
	var problems []string
	for _, f := range body.Details.ComponentFailures {
		problems = append(problems, scrub(fmt.Sprintf("%s %s: %s", f.ComponentType, f.FullName, oneLine(f.Problem)), nil))
	}
	if len(problems) == 0 {
		return body.ID, &SettingsError{cause: fmt.Errorf("the settings deploy ended with status %s", body.Status)}
	}
	return body.ID, &SettingsError{Problems: problems, Rejected: true}
}
