package main

import (
	"context"
	"fmt"
	"log"
	"time"

	claude "github.com/Facets-cloud/claude-agent-sdk-go"
)

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}

func main() {
	fmt.Println("Claude Agent SDK - Sandbox Example")
	fmt.Println("===================================")
	fmt.Println()

	// Example 1: Basic Sandbox Configuration
	fmt.Println("Example 1: Basic Sandbox with Auto-Allow")
	basicSandboxExample()

	// Example 2: Sandbox with Network Restrictions
	fmt.Println("\nExample 2: Sandbox with Network Restrictions")
	networkSandboxExample()

	// Example 3: Sandbox with Excluded Commands
	fmt.Println("\nExample 3: Sandbox with Excluded Commands")
	excludedCommandsExample()

	// Example 4: Sandbox with Ignore Violations
	fmt.Println("\nExample 4: Sandbox with Ignore Violations")
	ignoreViolationsExample()
}

func basicSandboxExample() {
	options := &claude.ClaudeAgentOptions{
		Sandbox: &claude.SandboxSettings{
			Enabled:                  boolPtr(true),
			AutoAllowBashIfSandboxed: boolPtr(true),
		},
	}

	fmt.Printf("Sandbox enabled: %v\n", *options.Sandbox.Enabled)
	fmt.Printf("Auto-allow bash: %v\n", *options.Sandbox.AutoAllowBashIfSandboxed)
	fmt.Println("Note: In a real scenario, this would execute bash commands in a sandboxed environment")

	// In a real application, you would use:
	// ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	// defer cancel()
	// messages, errors, err := claude.Query(ctx, "List files in current directory", options, nil)
}

func networkSandboxExample() {
	options := &claude.ClaudeAgentOptions{
		Sandbox: &claude.SandboxSettings{
			Enabled: boolPtr(true),
			Network: &claude.SandboxNetworkConfig{
				Enabled:        boolPtr(true),
				AllowedDomains: []string{"*.github.com", "api.anthropic.com"},
				BlockedDomains: []string{"malicious.com", "dangerous.net"},
			},
		},
	}

	fmt.Printf("Network enabled: %v\n", *options.Sandbox.Network.Enabled)
	fmt.Printf("Allowed domains: %v\n", options.Sandbox.Network.AllowedDomains)
	fmt.Printf("Blocked domains: %v\n", options.Sandbox.Network.BlockedDomains)
	fmt.Println("Note: These settings control network access within the sandbox")
}

func excludedCommandsExample() {
	options := &claude.ClaudeAgentOptions{
		Sandbox: &claude.SandboxSettings{
			Enabled:                  boolPtr(true),
			ExcludedCommands:         []string{"git", "npm", "docker"},
			AllowUnsandboxedCommands: boolPtr(true),
		},
	}

	fmt.Printf("Excluded commands: %v\n", options.Sandbox.ExcludedCommands)
	fmt.Printf("Allow unsandboxed: %v\n", *options.Sandbox.AllowUnsandboxedCommands)
	fmt.Println("Note: These commands will run without sandbox isolation")
}

func ignoreViolationsExample() {
	options := &claude.ClaudeAgentOptions{
		Sandbox: &claude.SandboxSettings{
			Enabled: boolPtr(true),
			IgnoreViolations: &claude.SandboxIgnoreViolations{
				Commands: []string{"ls", "cat"},
				Paths:    []string{"/tmp/*", "/var/log/*"},
			},
		},
	}

	if options.Sandbox.IgnoreViolations != nil {
		fmt.Printf("Ignored commands: %v\n", options.Sandbox.IgnoreViolations.Commands)
		fmt.Printf("Ignored paths: %v\n", options.Sandbox.IgnoreViolations.Paths)
		fmt.Println("Note: Violations for these commands/paths will be ignored")
	}
}

// completeExample demonstrates a complete working example with Query.
// Uncomment the call in main() to run this example with actual Claude queries.
// Note: This requires a working Claude CLI installation and will make actual API calls.
func completeExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	options := &claude.ClaudeAgentOptions{
		Sandbox: &claude.SandboxSettings{
			Enabled:                  boolPtr(true),
			AutoAllowBashIfSandboxed: boolPtr(true),
			ExcludedCommands:         []string{"git"},
			AllowUnsandboxedCommands: boolPtr(false),
			Network: &claude.SandboxNetworkConfig{
				Enabled:        boolPtr(true),
				AllowedDomains: []string{"*.github.com"},
			},
		},
		MaxTurns: intPtr(10),
	}

	messages, errors, err := claude.Query(
		ctx,
		"List the files in the current directory and check if git is available",
		options,
		nil,
	)

	if err != nil {
		log.Fatalf("Failed to start query: %v", err)
	}

	// Process messages
	for msg := range messages {
		switch m := msg.(type) {
		case *claude.UserMessage:
			fmt.Println("Received user message")
		case *claude.AssistantMessage:
			fmt.Println("Received assistant message")
		case *claude.ResultMessage:
			fmt.Printf("Query completed in %d ms\n", m.DurationMS)
		case *claude.SystemMessage:
			fmt.Printf("System message: %s\n", m.Subtype)
		default:
			fmt.Printf("Received message of type: %T\n", msg)
		}
	}

	// Check for errors
	if err := <-errors; err != nil {
		log.Printf("Query error: %v", err)
	}
}

func intPtr(i int) *int {
	return &i
}
