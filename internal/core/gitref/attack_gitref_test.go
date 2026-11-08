package gitref

// Red team tests for target names that reach git as a tag name. git runs without a shell, so
// shell metacharacters cannot run anything, but a name must also never become a git option, a
// ref outside glassops/deployed/ or a multi-line argument.

import "testing"

func TestRTDeployedTagRefusesHostileTargetNames(t *testing.T) {
	for _, name := range []string{
		"", "-", "--upload-pack=touch rt", "-f", "qa;id", "qa && id", "$(id)", "`id`", "qa|id",
		"qa\nid", "qa\r", "qa id", "../../heads/main", "qa/../../x", "qa..x", "qa.lock", ".qa", "qa.",
		"qa@{1}", "qa~1", "qa^", "qa:x", "qa?", "qa*", "qa[", "qa\\x", "::warning::x",
	} {
		if tag, err := DeployedTag(name); err == nil {
			t.Errorf("DeployedTag(%q) = %q, want an error", name, tag)
		}
	}
	if tag, err := DeployedTag("qa"); err != nil || tag != DeployedTagPrefix+"qa" {
		t.Fatalf("control: %q, %v", tag, err)
	}
}
