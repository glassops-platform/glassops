package permit

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/glassops-platform/glassops/packages/runtime/internal/policy"
)

func TestGenerate(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	os.Setenv("GITHUB_WORKSPACE", tmpDir)
	defer os.Unsetenv("GITHUB_WORKSPACE")

	os.Setenv("GITHUB_ACTOR", "test-actor")
	os.Setenv("GITHUB_REPOSITORY", "test/repo")
	os.Setenv("GITHUB_SHA", "test-sha")
	defer func() {
		os.Unsetenv("GITHUB_ACTOR")
		os.Unsetenv("GITHUB_REPOSITORY")
		os.Unsetenv("GITHUB_SHA")
	}()

	config := &policy.Config{}
	runtimeID := "test-runtime-id"
	orgID := "test-org-id"
	instanceURL := "https://test.salesforce.com"

	// Execute
	path, err := Generate(runtimeID, orgID, config, instanceURL)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Verify File Exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Generate() did not create file at %s", path)
	}

	// Verify Content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read permit file: %v", err)
	}

	var p Permit
	if err := json.Unmarshal(content, &p); err != nil {
		t.Errorf("Failed to unmarshal permit: %v", err)
	}

	if p.RuntimeID != runtimeID {
		t.Errorf("RuntimeID = %s, want %s", p.RuntimeID, runtimeID)
	}
	if p.OrgID != orgID {
		t.Errorf("OrgID = %s, want %s", p.OrgID, orgID)
	}
	if p.Context["actor"] != "test-actor" {
		t.Errorf("Context.Actor = %s, want test-actor", p.Context["actor"])
	}
	if p.Inputs["instance_url"] != instanceURL {
		t.Errorf("Inputs.InstanceURL = %s, want %s", p.Inputs["instance_url"], instanceURL)
	}
}
