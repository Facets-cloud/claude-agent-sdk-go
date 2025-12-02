package unit

import (
	"encoding/json"
	"testing"

	claude "github.com/Facets-cloud/claude-agent-sdk-go"
)

func TestSandboxSettingsJSON(t *testing.T) {
	// Test basic sandbox settings serialization
	boolTrue := true
	boolFalse := false
	sandbox := &claude.SandboxSettings{
		Enabled:                  &boolTrue,
		AutoAllowBashIfSandboxed: &boolTrue,
		ExcludedCommands:         []string{"git", "npm"},
		AllowUnsandboxedCommands: &boolFalse,
	}

	data, err := json.Marshal(sandbox)
	if err != nil {
		t.Fatalf("Failed to marshal sandbox settings: %v", err)
	}

	// Verify JSON structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["enabled"] != true {
		t.Errorf("Expected enabled=true, got %v", result["enabled"])
	}

	if result["autoAllowBashIfSandboxed"] != true {
		t.Errorf("Expected autoAllowBashIfSandboxed=true, got %v", result["autoAllowBashIfSandboxed"])
	}
}

func TestSandboxNetworkConfig(t *testing.T) {
	boolTrue := true
	network := &claude.SandboxNetworkConfig{
		Enabled:        &boolTrue,
		AllowedDomains: []string{"*.github.com", "api.anthropic.com"},
		BlockedDomains: []string{"malicious.com"},
	}

	data, err := json.Marshal(network)
	if err != nil {
		t.Fatalf("Failed to marshal network config: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["enabled"] != true {
		t.Errorf("Expected enabled=true, got %v", result["enabled"])
	}

	allowedDomains := result["allowedDomains"].([]interface{})
	if len(allowedDomains) != 2 {
		t.Errorf("Expected 2 allowed domains, got %d", len(allowedDomains))
	}
}

func TestClaudeAgentOptionsWithSandbox(t *testing.T) {
	boolTrue := true
	model := "sonnet"
	options := &claude.ClaudeAgentOptions{
		Model: &model,
		Sandbox: &claude.SandboxSettings{
			Enabled:          &boolTrue,
			ExcludedCommands: []string{"docker"},
			Network: &claude.SandboxNetworkConfig{
				Enabled:        &boolTrue,
				AllowedDomains: []string{"*.safe.com"},
			},
		},
	}

	// This test verifies that sandbox settings can be included in ClaudeAgentOptions
	if options.Sandbox == nil {
		t.Error("Sandbox settings should not be nil")
	}

	if options.Sandbox.Network == nil {
		t.Error("Network config should not be nil")
	}

	if len(options.Sandbox.ExcludedCommands) != 1 {
		t.Errorf("Expected 1 excluded command, got %d", len(options.Sandbox.ExcludedCommands))
	}
}

func TestSandboxIgnoreViolations(t *testing.T) {
	ignore := &claude.SandboxIgnoreViolations{
		Commands: []string{"trusted-cmd"},
		Paths:    []string{"/safe/path/*"},
	}

	data, err := json.Marshal(ignore)
	if err != nil {
		t.Fatalf("Failed to marshal ignore violations: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	commands := result["commands"].([]interface{})
	if len(commands) != 1 || commands[0] != "trusted-cmd" {
		t.Errorf("Expected commands=['trusted-cmd'], got %v", commands)
	}

	paths := result["paths"].([]interface{})
	if len(paths) != 1 || paths[0] != "/safe/path/*" {
		t.Errorf("Expected paths=['/safe/path/*'], got %v", paths)
	}
}
