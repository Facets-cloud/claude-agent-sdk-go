# Changelog

All notable changes to the Claude Agent SDK for Go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added - Sync with Python SDK v0.1.10

This release brings the Go SDK feature parity with Python SDK version 0.1.10.

#### Structured Outputs Support (from Python v0.1.7)
- **`OutputFormat`** field in `ClaudeAgentOptions` - Define JSON schemas for structured responses
- **`StructuredOutput`** field in `ResultMessage` - Access parsed structured output from queries
- New example: `examples/structured_outputs/` with three comprehensive examples
- E2E tests: `tests/e2e/structured_output_test.go` with 4 test cases
- Unit tests for parsing structured output in `tests/unit/parser_test.go`
- Documentation in README.md with usage examples

**Usage:**
```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "file_count": map[string]interface{}{"type": "number"},
        "has_tests":  map[string]interface{}{"type": "boolean"},
    },
    "required": []string{"file_count", "has_tests"},
}

options := &claude.ClaudeAgentOptions{
    OutputFormat: map[string]interface{}{
        "type":   "json_schema",
        "schema": schema,
    },
}
```

#### AssistantMessageError Type (from Python v0.1.9)
- **`AssistantMessageError`** type with 6 error constants:
  - `AssistantMessageErrorAuthenticationFailed`
  - `AssistantMessageErrorBillingError`
  - `AssistantMessageErrorRateLimit`
  - `AssistantMessageErrorInvalidRequest`
  - `AssistantMessageErrorServerError`
  - `AssistantMessageErrorUnknown`
- **`Error`** field in `AssistantMessage` to capture API-level errors
- Parser support for parsing error field from message JSON
- Unit tests for all error types in `tests/unit/parser_test.go`

**Usage:**
```go
for msg := range msgCh {
    if assistantMsg, ok := msg.(*claude.AssistantMessage); ok {
        if assistantMsg.Error != nil {
            fmt.Printf("API Error: %s\n", *assistantMsg.Error)
        }
    }
}
```

#### HookMatcher Timeout (from Python v0.1.8)
- **`Timeout`** field in `HookMatcher` - Set timeout in seconds for hook execution
- Automatic timeout propagation to Claude Code CLI via control protocol
- Documentation in README.md with timeout usage example

**Usage:**
```go
timeout := 5.0
options := &claude.ClaudeAgentOptions{
    Hooks: map[claude.HookEvent][]claude.HookMatcher{
        claude.HookEventPreToolUse: {
            {
                Matcher: "Bash",
                Hooks:   []claude.HookCallback{myHook},
                Timeout: &timeout,
            },
        },
    },
}
```

#### Infrastructure & Tooling
- **CLI Version Tracking**: New `cli_version.go` file tracking recommended CLI version (2.0.55)
  - `RecommendedCLIVersion` constant for tested CLI version
  - `MinimumCLIVersion` constant for minimum required version
- **GitHub Actions CI/CD**:
  - `.github/workflows/test.yml` - Automated testing on Ubuntu, macOS, and Windows
  - `.github/workflows/lint.yml` - Code linting with golangci-lint, gofmt, and go vet
- **Examples**: New `examples/structured_outputs/` with 3 comprehensive examples

#### Documentation
- Updated README.md with:
  - Recommended Claude Code CLI version (2.0.55) in prerequisites
  - New "Structured Outputs" section with usage examples
  - Updated "Hooks" section demonstrating timeout usage
  - Updated "Configuration Options" to include `OutputFormat`
- New CHANGELOG.md documenting all changes

### Changed
- `types.go:473-477` - Added `Timeout` field to `HookMatcher` struct
- `types.go:99-116` - Added `Error` field to `AssistantMessage` struct
- `types.go:128-142` - Added `StructuredOutput` field to `ResultMessage` struct
- `types.go:39-49` - Added `AssistantMessageError` type and constants
- `types.go:565-567` - Added `OutputFormat` field to `ClaudeAgentOptions` struct
- `parser.go:111-115` - Updated AssistantMessage parser to handle error field
- `parser.go:264-266` - Updated ResultMessage parser to handle structured_output field
- `helpers.go:6-9` - Added `Timeout` field to `hookMatcherInternal` struct
- `helpers.go:26-30` - Updated `convertHooksToInternal` to include timeout
- `query_handler.go:148-155` - Updated hook configuration to pass timeout to CLI

### Fixed
- Enhanced parser tests with comprehensive coverage for new fields
- Added helper constants for CLI version management
- Improved type safety with new error constants

## Version History

This is the first changelog entry. The Go SDK was originally ported from Python SDK v0.1.6.

## Migration Notes

**Upgrading from previous versions:**

1. **No breaking changes** - All new features are additive
2. **Optional features** - Structured outputs, error handling, and hook timeouts are optional
3. **Recommended**: Update Claude Code CLI to version 2.0.55:
   ```bash
   npm install -g @anthropic-ai/claude-code@2.0.55
   ```

**New capabilities:**

1. **Structured Outputs** - Perfect for extracting structured data from queries
2. **Error Handling** - Better visibility into API-level errors
3. **Hook Timeouts** - More control over hook execution time

## Compatibility

- **Go**: 1.21+ (unchanged)
- **Claude Code CLI**: 2.0.50+ (minimum), 2.0.55 (recommended)
- **Python SDK Parity**: v0.1.10

## Links

- [Claude Agent SDK Documentation](https://docs.anthropic.com/en/docs/claude-code/sdk)
- [Python SDK Repository](https://github.com/anthropics/claude-agent-sdk-python)
- [Claude Code CLI](https://www.npmjs.com/package/@anthropic-ai/claude-code)
