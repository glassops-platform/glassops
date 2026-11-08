package salesforce

import (
	"errors"
	"strings"
	"testing"
)

func TestVersionLess(t *testing.T) {
	cases := []struct {
		a, b string
		less bool
	}{
		{"8.20.1", "8.20.1", false},
		{"8.20.0", "8.20.1", true},
		{"8.20.1", "8.20.0", false},
		{"8.9.9", "8.20.0", true},
		{"7.99.99", "8", true},
		{"8", "8.0.0", false},
		{"8.20", "8.20.1", true},
		{"v8.21.0", "8.20.1", false},
		{"8.20.1-beta.1", "8.20.1", true},
		{"8.20.1", "8.20.1-beta.1", false},
		{"8.20.1-alpha", "8.20.1-beta", true},
		{"8.20.1+build.5", "8.20.1", false},
	}
	for _, tc := range cases {
		got, err := versionLess(tc.a, tc.b)
		if err != nil {
			t.Fatalf("versionLess(%q, %q): %v", tc.a, tc.b, err)
		}
		if got != tc.less {
			t.Errorf("versionLess(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.less)
		}
	}
	for _, bad := range []string{"", "latest", "8.x", "1.2.3.4", "-1.0", "8..1"} {
		if _, err := versionLess(bad, "8.20.1"); err == nil {
			t.Errorf("versionLess(%q) accepted a non-version", bad)
		}
	}
}

// localPlugins keeps the name and version of entries "sf plugins --json" printed in
// glassops-hardis:local (sf 2.155.7, sfdx-hardis 8.20.1). Every other field is left out.
const localPlugins = `[{"name":"@salesforce/cli","version":"2.155.7"},{"name":"sfdx-git-delta","version":"6.45.1"},{"name":"sfdx-hardis","version":"8.20.1"}]`

func TestCheckHardisPlugins(t *testing.T) {
	got, err := CheckHardisPlugins([]byte(localPlugins), "")
	if err != nil || got.Version != "8.20.1" || got.Warning != "" {
		t.Fatalf("tested version: %+v, %v", got, err)
	}

	newer := strings.Replace(localPlugins, `"sfdx-hardis","version":"8.20.1"`, `"sfdx-hardis","version":"8.23.0"`, 1)
	got, err = CheckHardisPlugins([]byte(newer), "8.20")
	if err != nil || !strings.Contains(got.Warning, "newer than the tested version") {
		t.Fatalf("newer: %+v, %v", got, err)
	}

	_, err = CheckHardisPlugins([]byte(localPlugins), "8.21.0")
	if !errors.Is(err, ErrHardisTooOld) || !strings.Contains(err.Error(), HardisInstallSteps) {
		t.Fatalf("too old: %v", err)
	}

	wrapped := `{"status":0,"result":` + strings.Replace(localPlugins, "sfdx-hardis", "other", 1) + `}`
	_, err = CheckHardisPlugins([]byte(wrapped), "")
	if !errors.Is(err, ErrHardisMissing) || !strings.Contains(err.Error(), "sf plugins install sfdx-hardis@"+HardisTestedVersion) {
		t.Fatalf("missing: %v", err)
	}

	if _, err := CheckHardisPlugins([]byte("not json"), ""); err == nil || errors.Is(err, ErrHardisMissing) {
		t.Fatalf("unreadable output: %v", err)
	}
}
