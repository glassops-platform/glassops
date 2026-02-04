package permit

import (
	"encoding/json"
	"os"
	"testing"
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

	actor := Identity{
		Subject:    "test-actor",
		Provider:   "github",
		ProviderID: "github:test-actor",
		Verified:   true,
	}
	evaluation := PolicyEvaluation{
		Allowed:   true,
		Evaluated: []string{"FreezeCheck"},
	}
	runtimeID := "test-runtime-id"
	instanceURL := "https://test.salesforce.com"

	// Execute
	path, err := Generate(runtimeID, actor, evaluation, instanceURL)
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

	if p.PermitID != runtimeID {
		t.Errorf("PermitID = %s, want %s", p.PermitID, runtimeID)
	}
	if p.Actor.Subject != "test-actor" {
		t.Errorf("Actor.Subject = %s, want test-actor", p.Actor.Subject)
	}
	if p.Policies.Allowed != true {
		t.Errorf("Policies.Allowed = %v, want true", p.Policies.Allowed)
	}
	if p.Inputs["instance_url"] != instanceURL {
		t.Errorf("Inputs.InstanceURL = %s, want %s", p.Inputs["instance_url"], instanceURL)
	}
}
