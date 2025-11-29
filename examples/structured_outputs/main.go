package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	claude "github.com/Facets-cloud/claude-agent-sdk-go"
)

/*
Structured Outputs Example

This example demonstrates how to use structured outputs with the Claude SDK.
Structured outputs allow you to get responses in a specific JSON schema format,
making it easier to integrate Claude's responses into your application logic.

Prerequisites:
- ANTHROPIC_API_KEY environment variable must be set
- Claude Code CLI must be installed (npm install -g @anthropic-ai/claude-code)

Run this example:
  go run main.go
*/

func main() {
	// Check for API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	fmt.Println("=== Structured Outputs Example ===")

	// Example 1: Simple structured output
	example1SimpleStructuredOutput()

	// Example 2: Nested structured output
	example2NestedStructuredOutput()

	// Example 3: Structured output with tools
	example3StructuredOutputWithTools()
}

// example1SimpleStructuredOutput demonstrates a simple structured output request
func example1SimpleStructuredOutput() {
	fmt.Println("Example 1: Simple Structured Output")
	fmt.Println("------------------------------------")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Define a JSON schema for the expected output
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"word_count": map[string]interface{}{
				"type":        "number",
				"description": "The number of words in the text",
			},
			"sentence_count": map[string]interface{}{
				"type":        "number",
				"description": "The number of sentences in the text",
			},
			"has_question": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the text contains a question",
			},
		},
		"required": []string{"word_count", "sentence_count", "has_question"},
	}

	// Create options with structured output format
	permMode := claude.PermissionModeAcceptEdits
	options := &claude.ClaudeAgentOptions{
		OutputFormat: map[string]interface{}{
			"type":   "json_schema",
			"schema": schema,
		},
		PermissionMode: &permMode,
	}

	// Create client and connect
	client := claude.NewClaudeSDKClient(options)
	err := client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Query with a text analysis request
	prompt := "Analyze this text: 'The quick brown fox jumps over the lazy dog. What does this mean?'"
	msgCh, errCh := client.Query(ctx, prompt)

	// Collect messages
	var resultMsg *claude.ResultMessage
	for msg := range msgCh {
		if result, ok := msg.(*claude.ResultMessage); ok {
			resultMsg = result
		}
	}

	if err := <-errCh; err != nil {
		log.Fatalf("Query error: %v", err)
	}

	// Display the structured output
	if resultMsg != nil && resultMsg.StructuredOutput != nil {
		output, _ := json.MarshalIndent(resultMsg.StructuredOutput, "", "  ")
		fmt.Printf("Structured Output:\n%s\n\n", output)
	}
}

// example2NestedStructuredOutput demonstrates nested structured output
func example2NestedStructuredOutput() {
	fmt.Println("Example 2: Nested Structured Output")
	fmt.Println("------------------------------------")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Define a more complex schema with nested objects and arrays
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"summary": map[string]interface{}{
				"type":        "string",
				"description": "Brief summary of the code",
			},
			"analysis": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"language": map[string]interface{}{
						"type": "string",
					},
					"complexity": map[string]interface{}{
						"type": "string",
						"enum": []string{"low", "medium", "high"},
					},
					"line_count": map[string]interface{}{
						"type": "number",
					},
				},
				"required": []string{"language", "complexity"},
			},
			"recommendations": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "List of improvement recommendations",
			},
		},
		"required": []string{"summary", "analysis"},
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
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Analyze a code snippet
	code := `
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}
`

	prompt := fmt.Sprintf("Analyze this Go code and provide insights:\n```go\n%s\n```", code)
	msgCh, errCh := client.Query(ctx, prompt)

	// Collect messages
	var resultMsg *claude.ResultMessage
	for msg := range msgCh {
		if result, ok := msg.(*claude.ResultMessage); ok {
			resultMsg = result
		}
	}

	if err := <-errCh; err != nil {
		log.Fatalf("Query error: %v", err)
	}

	// Display the structured output
	if resultMsg != nil && resultMsg.StructuredOutput != nil {
		output, _ := json.MarshalIndent(resultMsg.StructuredOutput, "", "  ")
		fmt.Printf("Structured Output:\n%s\n\n", output)
	}
}

// example3StructuredOutputWithTools demonstrates structured output when Claude uses tools
func example3StructuredOutputWithTools() {
	fmt.Println("Example 3: Structured Output with Tool Use")
	fmt.Println("------------------------------------------")

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Define a schema for file system analysis
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"total_files": map[string]interface{}{
				"type":        "number",
				"description": "Total number of Go files found",
			},
			"test_files": map[string]interface{}{
				"type":        "number",
				"description": "Number of test files",
			},
			"has_go_mod": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether go.mod file exists",
			},
			"directory_structure": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "List of main directories",
			},
		},
		"required": []string{"total_files", "has_go_mod"},
	}

	cwd := "."
	permMode3 := claude.PermissionModeAcceptEdits
	options := &claude.ClaudeAgentOptions{
		OutputFormat: map[string]interface{}{
			"type":   "json_schema",
			"schema": schema,
		},
		PermissionMode: &permMode3,
		Cwd:            &cwd,
		AllowedTools:   []string{"Glob", "Read", "Bash"},
	}

	client := claude.NewClaudeSDKClient(options)
	err := client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Request analysis that requires tool use
	prompt := "Analyze the current Go project. Count the total Go files, test files, check if go.mod exists, and list the main directories. Use tools to explore the filesystem."
	msgCh, errCh := client.Query(ctx, prompt)

	// Collect messages and show tool usage
	var resultMsg *claude.ResultMessage
	toolUseCount := 0
	for msg := range msgCh {
		if assistantMsg, ok := msg.(*claude.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if toolUse, ok := block.(claude.ToolUseBlock); ok {
					toolUseCount++
					fmt.Printf("Tool used: %s\n", toolUse.Name)
				}
			}
		}
		if result, ok := msg.(*claude.ResultMessage); ok {
			resultMsg = result
		}
	}

	if err := <-errCh; err != nil {
		log.Fatalf("Query error: %v", err)
	}

	fmt.Printf("Total tools used: %d\n\n", toolUseCount)

	// Display the structured output
	if resultMsg != nil && resultMsg.StructuredOutput != nil {
		output, _ := json.MarshalIndent(resultMsg.StructuredOutput, "", "  ")
		fmt.Printf("Structured Output:\n%s\n\n", output)

		// Demonstrate type-safe access to structured output
		if outputMap, ok := resultMsg.StructuredOutput.(map[string]interface{}); ok {
			if totalFiles, ok := outputMap["total_files"].(float64); ok {
				fmt.Printf("Found %d Go files in the project\n", int(totalFiles))
			}
			if hasGoMod, ok := outputMap["has_go_mod"].(bool); ok {
				if hasGoMod {
					fmt.Println("✓ Project has go.mod file")
				} else {
					fmt.Println("✗ Project missing go.mod file")
				}
			}
		}
	}
}
