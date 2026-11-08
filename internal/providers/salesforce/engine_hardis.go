package salesforce

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

// HardisTestedVersion is the sfdx-hardis version the hardis engine was tested with. It is also
// the oldest version accepted when salesforce.hardis.min_version is not set.
const HardisTestedVersion = "8.20.1"

// hardisPlugin is the sf plugin name of sfdx-hardis.
const hardisPlugin = "sfdx-hardis"

// HardisInstallSteps tells the user how to add sfdx-hardis. GlassOps never installs it.
const HardisInstallSteps = "add it to the image you run GlassOps in: " +
	"sf plugins install " + hardisPlugin + "@" + HardisTestedVersion +
	" (see examples/hardis/Dockerfile). GlassOps never installs it"

// ErrHardisMissing means sfdx-hardis is not installed as an sf plugin.
var ErrHardisMissing = errors.New("sfdx-hardis is not installed")

// ErrHardisTooOld means the installed sfdx-hardis is older than the minimum version.
var ErrHardisTooOld = errors.New("sfdx-hardis is older than the minimum version")

// HardisCheck is what CheckHardisPlugins found.
type HardisCheck struct {
	// Version is the installed version.
	Version string
	// Warning is set when the version is newer than HardisTestedVersion.
	Warning string
}

// CheckHardisPlugins reads the output of "sf plugins --json" and checks that sfdx-hardis is
// installed at min or later. An empty min means HardisTestedVersion. A missing plugin is
// ErrHardisMissing and an old one ErrHardisTooOld, both with the install steps.
func CheckHardisPlugins(out []byte, min string) (HardisCheck, error) {
	if min == "" {
		min = HardisTestedVersion
	}
	plugins, err := parsePlugins(out)
	if err != nil {
		return HardisCheck{}, fmt.Errorf("could not read sf plugins output: %w", err)
	}
	for _, pl := range plugins {
		if pl.Name != hardisPlugin {
			continue
		}
		c := HardisCheck{Version: pl.Version}
		older, err := versionLess(pl.Version, min)
		if err != nil {
			return c, fmt.Errorf("sfdx-hardis version: %w", err)
		}
		if older {
			return c, fmt.Errorf("%w: %s is installed, %s or later is needed, %s", ErrHardisTooOld, pl.Version, min, HardisInstallSteps)
		}
		if newer, _ := versionLess(HardisTestedVersion, pl.Version); newer {
			c.Warning = fmt.Sprintf("sfdx-hardis %s is newer than the tested version %s", pl.Version, HardisTestedVersion)
		}
		return c, nil
	}
	return HardisCheck{}, fmt.Errorf("%w, %s", ErrHardisMissing, HardisInstallSteps)
}

// version is a parsed major.minor.patch with an optional prerelease suffix.
type version struct {
	nums [3]int
	pre  string
}

// parseVersion reads "8", "8.20", "8.20.1" or "8.20.1-beta.1". Build metadata after "+" is
// ignored.
func parseVersion(s string) (version, error) {
	var v version
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	s, _, _ = strings.Cut(s, "+")
	core, pre, _ := strings.Cut(s, "-")
	v.pre = pre
	parts := strings.Split(core, ".")
	if core == "" || len(parts) > 3 {
		return v, fmt.Errorf("not a version: %q", s)
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return v, fmt.Errorf("not a version: %q", s)
		}
		v.nums[i] = n
	}
	return v, nil
}

// versionLess reports whether a is older than b. A prerelease is older than its release.
func versionLess(a, b string) (bool, error) {
	va, err := parseVersion(a)
	if err != nil {
		return false, err
	}
	vb, err := parseVersion(b)
	if err != nil {
		return false, err
	}
	for i := range va.nums {
		if va.nums[i] != vb.nums[i] {
			return va.nums[i] < vb.nums[i], nil
		}
	}
	switch {
	case va.pre == vb.pre, va.pre == "":
		return false, nil
	case vb.pre == "":
		return true, nil
	}
	return va.pre < vb.pre, nil
}

// errHardisNotImplemented is what the hardis engine returns until it lands. Commands report
// it as a tool error, exit 2.
var errHardisNotImplemented = errors.New("salesforce engine hardis is not implemented yet")

// hardisEngine runs plan, validate and deploy through sfdx-hardis, which the user installs.
// It is not implemented yet.
type hardisEngine struct {
	p   *Provider
	cfg config.Hardis
}

func newHardisEngine(p *Provider, cfg config.Hardis) Engine {
	return hardisEngine{p: p, cfg: cfg}
}

func (e hardisEngine) Name() string { return config.EngineHardis }

func (e hardisEngine) Plan(context.Context, providers.PlanRequest) (providers.ChangeSet, error) {
	return providers.ChangeSet{}, errHardisNotImplemented
}

func (e hardisEngine) Validate(context.Context, providers.Session, providers.ChangeSet, providers.ValidateOptions) (providers.ValidateResult, error) {
	return providers.ValidateResult{}, errHardisNotImplemented
}

func (e hardisEngine) Deploy(context.Context, providers.Session, providers.ChangeSet, providers.DeployOptions) (providers.DeployResult, error) {
	return providers.DeployResult{}, errHardisNotImplemented
}
