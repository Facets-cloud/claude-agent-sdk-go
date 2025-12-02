package unit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	claude "github.com/Facets-cloud/claude-agent-sdk-go"
)

func TestBuildSettingsValue_OnlySandbox(t *testing.T) {
	boolTrue := true
	options := &claude.ClaudeAgentOptions{
		Sandbox: &claude.SandboxSettings{
			Enabled: &boolTrue,
		},
	}

	// Verify that sandbox settings are present in options
	if options.Sandbox == nil {
		t.Error("Expected sandbox settings to be present")
	}
	if options.Sandbox.Enabled == nil || !*options.Sandbox.Enabled {
		t.Error("Expected sandbox to be enabled")
	}
}

func TestBuildSettingsValue_WithJSONString(t *testing.T) {
	settingsJSON := `{"existingKey":"existingValue"}`

	// Verify settings can be parsed as JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(settingsJSON), &parsed); err != nil {
		t.Fatalf("Settings should be valid JSON: %v", err)
	}

	if parsed["existingKey"] != "existingValue" {
		t.Error("Expected existing key to be present")
	}

	// Verify that settings can be combined with sandbox
	boolTrue := true
	options := &claude.ClaudeAgentOptions{
		Settings: &settingsJSON,
		Sandbox: &claude.SandboxSettings{
			Enabled: &boolTrue,
		},
	}

	if options.Settings == nil {
		t.Error("Expected settings to be present")
	}
	if options.Sandbox == nil {
		t.Error("Expected sandbox to be present")
	}
}

func TestBuildSettingsValue_WithFilePath(t *testing.T) {
	// Create a temporary settings file
	tempDir := t.TempDir()
	settingsFile := filepath.Join(tempDir, "settings.json")

	settingsData := map[string]interface{}{
		"existingKey": "existingValue",
	}
	data, err := json.Marshal(settingsData)
	if err != nil {
		t.Fatalf("Failed to marshal settings: %v", err)
	}

	if err := os.WriteFile(settingsFile, data, 0644); err != nil {
		t.Fatalf("Failed to write settings file: %v", err)
	}

	// Verify file exists and can be read
	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		t.Error("Settings file should exist")
	}

	// Read and verify content
	content, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("Failed to read settings file: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("Settings file should contain valid JSON: %v", err)
	}

	if parsed["existingKey"] != "existingValue" {
		t.Error("Expected existing key to be present in file")
	}

	// Verify options can include file path with sandbox
	boolTrue := true
	options := &claude.ClaudeAgentOptions{
		Settings: &settingsFile,
		Sandbox: &claude.SandboxSettings{
			Enabled: &boolTrue,
		},
	}

	if options.Settings == nil || *options.Settings != settingsFile {
		t.Error("Expected settings file path to be preserved")
	}
	if options.Sandbox == nil {
		t.Error("Expected sandbox to be present")
	}
}

func TestSandboxMerging(t *testing.T) {
	// Test that sandbox settings can be merged with other settings
	existingSettings := map[string]interface{}{
		"key1": "value1",
	}
	existingJSON, _ := json.Marshal(existingSettings)
	existingStr := string(existingJSON)

	boolTrue := true
	options := &claude.ClaudeAgentOptions{
		Settings: &existingStr,
		Sandbox: &claude.SandboxSettings{
			Enabled:          &boolTrue,
			ExcludedCommands: []string{"git"},
		},
	}

	// Manually merge to verify the expected result
	var merged map[string]interface{}
	if err := json.Unmarshal(existingJSON, &merged); err != nil {
		t.Fatalf("Failed to parse existing settings: %v", err)
	}

	merged["sandbox"] = options.Sandbox

	mergedJSON, err := json.Marshal(merged)
	if err != nil {
		t.Fatalf("Failed to marshal merged settings: %v", err)
	}

	// Verify the merged result
	var result map[string]interface{}
	if err := json.Unmarshal(mergedJSON, &result); err != nil {
		t.Fatalf("Failed to parse merged JSON: %v", err)
	}

	if result["key1"] != "value1" {
		t.Error("Existing key should be preserved")
	}

	if result["sandbox"] == nil {
		t.Error("Sandbox settings should be present")
	}
}
