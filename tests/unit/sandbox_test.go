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
	httpPort := 8080
	network := &claude.SandboxNetworkConfig{
		AllowUnixSockets:    []string{"/var/run/docker.sock"},
		AllowAllUnixSockets: &boolTrue,
		AllowLocalBinding:   &boolTrue,
		HttpProxyPort:       &httpPort,
	}

	data, err := json.Marshal(network)
	if err != nil {
		t.Fatalf("Failed to marshal network config: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["allowAllUnixSockets"] != true {
		t.Errorf("Expected allowAllUnixSockets=true, got %v", result["allowAllUnixSockets"])
	}

	allowUnixSockets := result["allowUnixSockets"].([]interface{})
	if len(allowUnixSockets) != 1 {
		t.Errorf("Expected 1 unix socket, got %d", len(allowUnixSockets))
	}

	if result["httpProxyPort"] != float64(8080) {
		t.Errorf("Expected httpProxyPort=8080, got %v", result["httpProxyPort"])
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
				AllowLocalBinding: &boolTrue,
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
		File:    []string{"/safe/path/*"},
		Network: []string{"localhost:*"},
	}

	data, err := json.Marshal(ignore)
	if err != nil {
		t.Fatalf("Failed to marshal ignore violations: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	files := result["file"].([]interface{})
	if len(files) != 1 || files[0] != "/safe/path/*" {
		t.Errorf("Expected file=['/safe/path/*'], got %v", files)
	}

	network := result["network"].([]interface{})
	if len(network) != 1 || network[0] != "localhost:*" {
		t.Errorf("Expected network=['localhost:*'], got %v", network)
	}
}
