package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	claude "github.com/Facets-cloud/claude-agent-sdk-go"
)

// TestSimpleStructuredOutput tests structured output with file counting requiring tool use.
func TestSimpleStructuredOutput(t *testing.T) {
	RequireClaudeCode(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Define schema for file analysis
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_count": map[string]interface{}{
				"type": "number",
			},
			"has_tests": map[string]interface{}{
				"type": "boolean",
			},
			"test_file_count": map[string]interface{}{
				"type": "number",
			},
		},
		"required": []string{"file_count", "has_tests"},
	}

	permMode := claude.PermissionModeAcceptEdits
	options := &claude.ClaudeAgentOptions{
		OutputFormat: map[string]interface{}{
			"type":   "json_schema",
			"schema": schema,
		},
		PermissionMode: &permMode,
		Cwd:            strPtr("."),
	}

	client := claude.NewClaudeSDKClient(options)
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Agent must use Glob/Bash to count files
	msgCh, errCh := client.Query(ctx, "Count how many Go files are in the current directory (not subdirectories) and check if there are any test files. Use tools to explore the filesystem.")

	// Collect messages
	messages, err := CollectMessages(msgCh, errCh)
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}

	// Verify result
	resultMessage := GetResultMessage(messages)
	if resultMessage == nil {
		t.Fatal("No result message received")
	}

	if resultMessage.IsError {
		t.Fatalf("Query failed: %v", resultMessage.Result)
	}

	if resultMessage.Subtype != "success" {
		t.Errorf("Expected success subtype, got %s", resultMessage.Subtype)
	}

	// Verify structured output is present and valid
	if resultMessage.StructuredOutput == nil {
		t.Fatal("No structured output in result")
	}

	output, ok := resultMessage.StructuredOutput.(map[string]interface{})
	if !ok {
		t.Fatalf("Structured output is not a map: %T", resultMessage.StructuredOutput)
	}

	// Check required fields
	if _, ok := output["file_count"]; !ok {
		t.Error("Missing file_count in structured output")
	}

	if _, ok := output["has_tests"]; !ok {
		t.Error("Missing has_tests in structured output")
	}

	// Verify types
	if fileCount, ok := output["file_count"].(float64); ok {
		if fileCount <= 0 {
			t.Error("Expected file_count > 0, got", fileCount)
		}
	} else {
		t.Errorf("file_count is not a number: %T", output["file_count"])
	}

	if _, ok := output["has_tests"].(bool); !ok {
		t.Errorf("has_tests is not a boolean: %T", output["has_tests"])
	}
}

// TestNestedStructuredOutput tests structured output with nested objects and arrays.
func TestNestedStructuredOutput(t *testing.T) {
	RequireClaudeCode(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Define a schema with nested structure
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"analysis": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"word_count": map[string]interface{}{
						"type": "number",
					},
					"character_count": map[string]interface{}{
						"type": "number",
					},
				},
				"required": []string{"word_count", "character_count"},
			},
			"words": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"analysis", "words"},
	}

	permMode2 := claude.PermissionModeAcceptEdits
	options := &claude.ClaudeAgentOptions{
		OutputFormat: map[string]interface{}{
			"type":   "json_schema",
			"schema": schema,
		},
		PermissionMode: &permMode2,
	}

	client := claude.NewClaudeSDKClient(options)
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	msgCh, errCh := client.Query(ctx, "Analyze this text: 'Hello world'. Provide word count, character count, and list of words.")

	// Collect messages
	messages, err := CollectMessages(msgCh, errCh)
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}

	// Verify result
	resultMessage := GetResultMessage(messages)
	if resultMessage == nil {
		t.Fatal("No result message received")
	}

	if resultMessage.IsError {
		t.Fatalf("Query failed")
	}

	if resultMessage.StructuredOutput == nil {
		t.Fatal("No structured output in result")
	}

	// Check nested structure
	output, ok := resultMessage.StructuredOutput.(map[string]interface{})
	if !ok {
		t.Fatalf("Structured output is not a map: %T", resultMessage.StructuredOutput)
	}

	if _, ok := output["analysis"]; !ok {
		t.Error("Missing analysis in structured output")
	}

	if _, ok := output["words"]; !ok {
		t.Error("Missing words in structured output")
	}

	// Check analysis nested object
	analysis, ok := output["analysis"].(map[string]interface{})
	if !ok {
		t.Fatalf("analysis is not a map: %T", output["analysis"])
	}

	if wordCount, ok := analysis["word_count"].(float64); ok {
		if wordCount != 2 {
			t.Errorf("Expected word_count=2, got %f", wordCount)
		}
	} else {
		t.Errorf("word_count is not a number: %T", analysis["word_count"])
	}

	if charCount, ok := analysis["character_count"].(float64); ok {
		if charCount != 11 { // "Hello world"
			t.Errorf("Expected character_count=11, got %f", charCount)
		}
	} else {
		t.Errorf("character_count is not a number: %T", analysis["character_count"])
	}

	// Check words array
	words, ok := output["words"].([]interface{})
	if !ok {
		t.Fatalf("words is not an array: %T", output["words"])
	}

	if len(words) != 2 {
		t.Errorf("Expected 2 words, got %d", len(words))
	}
}

// TestStructuredOutputWithEnum tests structured output with enum constraints.
func TestStructuredOutputWithEnum(t *testing.T) {
	RequireClaudeCode(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"has_tests": map[string]interface{}{
				"type": "boolean",
			},
			"test_framework": map[string]interface{}{
				"type": "string",
				"enum": []string{"pytest", "unittest", "nose", "go-test", "unknown"},
			},
			"test_count": map[string]interface{}{
				"type": "number",
			},
		},
		"required": []string{"has_tests", "test_framework"},
	}

	permMode3 := claude.PermissionModeAcceptEdits
	options := &claude.ClaudeAgentOptions{
		OutputFormat: map[string]interface{}{
			"type":   "json_schema",
			"schema": schema,
		},
		PermissionMode: &permMode3,
		Cwd:            strPtr("."),
	}

	client := claude.NewClaudeSDKClient(options)
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	msgCh, errCh := client.Query(ctx, "Search for test files in the tests/ directory. Determine which test framework is being used (go-test for Go, pytest for Python, etc.) and count how many test files exist. Use Grep to search for test patterns.")

	// Collect messages
	messages, err := CollectMessages(msgCh, errCh)
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}

	// Verify result
	resultMessage := GetResultMessage(messages)
	if resultMessage == nil {
		t.Fatal("No result message received")
	}

	if resultMessage.IsError {
		t.Fatalf("Query failed")
	}

	if resultMessage.StructuredOutput == nil {
		t.Fatal("No structured output in result")
	}

	// Check enum values are valid
	output, ok := resultMessage.StructuredOutput.(map[string]interface{})
	if !ok {
		t.Fatalf("Structured output is not a map: %T", resultMessage.StructuredOutput)
	}

	framework, ok := output["test_framework"].(string)
	if !ok {
		t.Fatalf("test_framework is not a string: %T", output["test_framework"])
	}

	validFrameworks := map[string]bool{
		"pytest":   true,
		"unittest": true,
		"nose":     true,
		"go-test":  true,
		"unknown":  true,
	}

	if !validFrameworks[framework] {
		t.Errorf("test_framework has invalid enum value: %s", framework)
	}

	hasTests, ok := output["has_tests"].(bool)
	if !ok {
		t.Fatalf("has_tests is not a boolean: %T", output["has_tests"])
	}

	// This repo uses Go testing
	if !hasTests {
		t.Error("Expected has_tests to be true")
	}

	if framework != "go-test" && framework != "unknown" {
		t.Errorf("Expected test_framework to be 'go-test' or 'unknown', got %s", framework)
	}
}

// TestStructuredOutputWithTools tests structured output when agent uses tools.
func TestStructuredOutputWithTools(t *testing.T) {
	RequireClaudeCode(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Schema for file analysis
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_count": map[string]interface{}{
				"type": "number",
			},
			"has_readme": map[string]interface{}{
				"type": "boolean",
			},
		},
		"required": []string{"file_count", "has_readme"},
	}

	// Use temp directory for cross-platform compatibility
	tempDir := os.TempDir()

	permMode4 := claude.PermissionModeAcceptEdits
	options := &claude.ClaudeAgentOptions{
		OutputFormat: map[string]interface{}{
			"type":   "json_schema",
			"schema": schema,
		},
		PermissionMode: &permMode4,
		Cwd:            &tempDir,
	}

	client := claude.NewClaudeSDKClient(options)
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	msgCh, errCh := client.Query(ctx, "Count how many files are in the current directory and check if there's a README file. Use tools as needed.")

	// Collect messages
	messages, err := CollectMessages(msgCh, errCh)
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}

	// Verify result
	resultMessage := GetResultMessage(messages)
	if resultMessage == nil {
		t.Fatal("No result message received")
	}

	if resultMessage.IsError {
		t.Fatalf("Query failed")
	}

	if resultMessage.StructuredOutput == nil {
		t.Fatal("No structured output in result")
	}

	// Check structure
	output, ok := resultMessage.StructuredOutput.(map[string]interface{})
	if !ok {
		t.Fatalf("Structured output is not a map: %T", resultMessage.StructuredOutput)
	}

	if _, ok := output["file_count"]; !ok {
		t.Error("Missing file_count in structured output")
	}

	if _, ok := output["has_readme"]; !ok {
		t.Error("Missing has_readme in structured output")
	}

	if fileCount, ok := output["file_count"].(float64); ok {
		if fileCount < 0 {
			t.Errorf("file_count should be non-negative, got %f", fileCount)
		}
	} else {
		t.Errorf("file_count is not a number: %T", output["file_count"])
	}

	if _, ok := output["has_readme"].(bool); !ok {
		t.Errorf("has_readme is not a boolean: %T", output["has_readme"])
	}

	// Verify tools were used
	if !HasToolUse(messages) {
		t.Error("Expected agent to use tools to analyze filesystem")
	}
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}
