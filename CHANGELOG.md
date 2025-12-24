# Changelog

All notable changes to the Claude Agent SDK for Go will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.18] - 2025-12-24

### Added - Complete Parity with Python SDK v0.1.18

This update brings the Go SDK to parity with Python SDK v0.1.18.

#### UserMessage UUID Field (from Python v0.1.17)
- **`UUID`** field added to `UserMessage` struct - Provides access to message identifiers needed for file checkpointing
- **Parser updated** - Extracts UUID from CLI response data
- **Improved developer experience** - Makes it easier to use `RewindFiles()` by providing direct access to message UUIDs

**Usage with file checkpointing:**
```go
options := &claude.ClaudeAgentOptions{
    EnableFileCheckpointing: true,
}
client := claude.NewClaudeSDKClient(options)

// During response processing
for msg := range client.ReceiveResponse(ctx) {
    if userMsg, ok := msg.(*claude.UserMessage); ok {
        if userMsg.UUID != nil {
            checkpointID := *userMsg.UUID  // Save for later use with RewindFiles()
        }
    }
}

// Later, rewind to that checkpoint
client.RewindFiles(ctx, checkpointID)
```

### Changed
- **CLI version updated:** 2.0.62 → 2.0.76
  - `BundledCLIVersion` updated to "2.0.76"
  - `RecommendedCLIVersion` updated to "2.0.76"
- **SDK version:** Updated to v0.1.18 for parity with Python SDK

### Added
- Unit tests for UserMessage UUID parsing (2 test cases)

---

## [0.1.17] - 2025-12-24

### Added
- **UserMessage UUID field** - Added `uuid` field to `UserMessage` response type

### Changed
- **CLI version:** 2.0.62 → 2.0.70

---

## [0.1.16] - 2025-12-24

### Fixed
- **Rate limit detection** - Error field parsing in AssistantMessage was already implemented in Go SDK

### Changed
- **CLI version:** 2.0.62 → 2.0.68

---

## [0.1.15] - 2025-12-12

### Added - File Checkpointing and Rewind Support

This update brings file checkpointing support, matching Python SDK v0.1.15.

#### File Checkpointing Feature (from Python v0.1.15)
- **`EnableFileCheckpointing`** field in `ClaudeAgentOptions` - Enable tracking of file changes during the session
- **`RewindFiles(userMessageID)`** method on `ClaudeSDKClient` - Rewind tracked files to their state at a specific checkpoint
- **`RewindFiles(userMessageID)`** method on `queryHandler` - Internal implementation for file rewinding
- **Environment variable support** - Sets `CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING=true` when enabled
- **Control protocol support** - Added `rewind_files` control request subtype with `user_message_id` parameter

**Use cases:**
- Explore different implementation approaches without losing previous work
- Recover from unwanted file modifications
- Create checkpoints during multi-step operations
- Test variations safely with ability to rollback

**Usage:**
```go
options := &claude.ClaudeAgentOptions{
    EnableFileCheckpointing: true,
}
client := claude.NewClaudeSDKClient(options)
ctx := context.Background()

if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Close()

// Send a query that modifies files
if err := client.SendMessage(ctx, "Refactor the authentication module"); err != nil {
    log.Fatal(err)
}

var checkpointID string
for msg := range client.ReceiveResponse(ctx) {
    if userMsg, ok := msg.(*UserMessage); ok {
        checkpointID = userMsg.UUID // Save checkpoint for later
    }
}

// Try another approach
if err := client.SendMessage(ctx, "Actually, use a different pattern"); err != nil {
    log.Fatal(err)
}

// Rewind back to the first checkpoint if needed
if err := client.RewindFiles(ctx, checkpointID); err != nil {
    log.Fatal(err)
}
```

### Changed
- **CLI version updated:** 2.0.60 → 2.0.62
  - `BundledCLIVersion` updated to "2.0.62"
  - `RecommendedCLIVersion` updated to "2.0.62"
- **SDK version:** Updated to v0.1.15 for parity with Python SDK

### Added
- Unit tests for file checkpointing feature (3 test cases)
- Comprehensive documentation and examples for file rewind functionality

---

## [0.1.14] - 2025-12-12

### Changed
- **CLI version updated:** 2.0.60 → 2.0.61 (internal update)
  - Updated bundled CLI to version 2.0.61

---

## [0.1.13] - 2025-12-06

### Added - Complete Parity with Python SDK v0.1.13

This update brings the Go SDK to full feature parity with Python SDK version 0.1.13, including all changes from v0.1.11, v0.1.12, and v0.1.13.

#### Tools Option Support (from Python v0.1.12)
- **`Tools`** field in `ClaudeAgentOptions` - Base set of tools to use (separate from allowed/disallowed filtering)
  - Supports `[]string` for explicit tool list: `[]string{"Read", "Write", "Bash"}`
  - Supports `ToolsPreset` for preset configuration: `{Type: "preset", Preset: "claude_code"}`
  - Supports `map[string]interface{}` for dynamic preset: `{"type": "preset", "preset": "claude_code"}`
  - Empty array `[]string{}` explicitly sets no tools
- **`ToolsPreset`** type - Structured preset configuration
  - `Type` - Always "preset"
  - `Preset` - Preset name (e.g., "claude_code")
- **CLI argument building** - `--tools` flag generation with proper serialization
  - String array joins with commas: `--tools Read,Write,Bash`
  - Empty array maps to: `--tools ""`
  - Preset "claude_code" maps to: `--tools default`
- **Distinction from filtering** - Tools option sets base tools, AllowedTools/DisallowedTools filter from that base
- **Unit tests** - Comprehensive tests in `tests/unit/transport_test.go`

**Usage:**
```go
// Explicit tool list
options := &claude.ClaudeAgentOptions{
    Tools: []string{"Read", "Write", "Bash"},
}

// Preset configuration
options := &claude.ClaudeAgentOptions{
    Tools: claude.ToolsPreset{
        Type:   "preset",
        Preset: "claude_code",
    },
}

// Combined with filtering
options := &claude.ClaudeAgentOptions{
    Tools:           []string{"Read", "Write", "Bash", "Edit"},
    AllowedTools:    []string{"Read", "Write"},
    DisallowedTools: []string{"Bash"},
}
```

#### Beta Features Support (from Python v0.1.12)
- **`Betas`** field in `ClaudeAgentOptions` - Array of beta features to enable
- **`SdkBeta`** type - String type for beta feature identifiers
- **`SdkBetaContext1M`** constant - 1M context window beta (`"context-1m-2025-08-07"`)
- **CLI argument building** - `--betas` flag generation with comma-separated values
- **Documentation link** - See https://docs.anthropic.com/en/api/beta-headers
- **Unit tests** - Tests for single beta, multiple betas, and no betas

**Usage:**
```go
// Single beta feature
options := &claude.ClaudeAgentOptions{
    Betas: []claude.SdkBeta{claude.SdkBetaContext1M},
}

// Multiple beta features
options := &claude.ClaudeAgentOptions{
    Betas: []claude.SdkBeta{
        claude.SdkBetaContext1M,
        claude.SdkBeta("future-beta-feature"),
    },
}
```

#### CLI Error Propagation Fix (from Python v0.1.13)
- **Fast-fail mechanism** - Pending control requests now fail immediately when CLI exits with error
- **Signal all pending requests** - Propagate error to all pending control responses in `pendingControlResponses` map
- **Improved error handling** - Reduces failure time from 60 seconds (timeout) to ~3 seconds
- **Updated `routeMessages()`** - Error propagation logic in query_handler.go:95-108

**Before:**
```
CLI error → Pending control requests wait for 60s timeout → Slow failure
```

**After:**
```
CLI error → Immediately signal all pending requests → Fast failure (~3s)
```

#### Write Lock TOCTOU Race Fix (from Python v0.1.13)
- **TOCTOU vulnerability fixed** - Time-Of-Check-Time-Of-Use race between Write() and Close()/EndInput()
- **Atomic state changes** - Acquire write lock before setting ready=false in Close() and EndInput()
- **Enhanced thread safety** - Prevents Write() from proceeding after ready check but before acquiring writeMu
- **Updated methods**:
  - `Close()` - Acquire writeMu before mu lock (transport_subprocess.go:255-263)
  - `EndInput()` - Acquire writeMu before mu lock (transport_subprocess.go:270-278)

**Race Condition Before:**
```
Write(): Check ready=true → Release mu → [Close() sets ready=false] → Acquire writeMu → Write to closed stdin
```

**Fixed:**
```
Write(): Check ready=true → Release mu → Acquire writeMu → Close() blocked on writeMu → Safe
```

### Changed
- **CLI version updated** - v2.0.56 → v2.0.60 across all constants
  - `BundledCLIVersion` updated to "2.0.60"
  - `RecommendedCLIVersion` updated to "2.0.60"
  - Installation documentation updated
- **SDK version** - Updated to v0.1.13 for parity with Python SDK

### Fixed
- CLI error propagation to pending control requests (reduces timeout from 60s to ~3s)
- TOCTOU race condition in Close() and EndInput() methods
- Concurrent write safety with proper lock ordering

### Added
- Comprehensive unit tests for tools option (7 test cases)
- Comprehensive unit tests for beta features (3 test cases)
- Combined features test validating tools + betas together

---

## [0.1.11] - 2024-XX-XX

### Added - Complete Parity with Python SDK v0.1.10

This update brings the Go SDK to full feature parity with Python SDK version 0.1.10, including critical features from recent Python SDK updates.

#### Sandbox Settings Support (CRITICAL)
- **`SandboxSettings`** type with comprehensive sandbox configuration
  - `Enabled` - Enable/disable sandbox for bash commands
  - `AutoAllowBashIfSandboxed` - Auto-approve sandboxed commands
  - `ExcludedCommands` - Commands to run without sandbox
  - `AllowUnsandboxedCommands` - Allow excluded commands to run unsandboxed
  - `Network` - Network access configuration
  - `IgnoreViolations` - Specify violations to ignore
  - `EnableWeakerNestedSandbox` - Weaker isolation for nested sandboxes
- **`SandboxNetworkConfig`** type for network restrictions
  - `Enabled` - Enable/disable network access
  - `AllowedDomains` - Domain whitelist (supports wildcards)
  - `BlockedDomains` - Domain blacklist
- **`SandboxIgnoreViolations`** type for violation exceptions
  - `Commands` - Command patterns to ignore
  - `Paths` - File path patterns to ignore
- **`Sandbox`** field added to `ClaudeAgentOptions`
- Settings merging logic - Sandbox settings automatically merged into `--settings` CLI argument
- New example: `examples/sandbox/` with comprehensive sandbox usage
- Unit tests: `tests/unit/sandbox_test.go` and `tests/unit/settings_merge_test.go`

#### Bundled CLI Binary Support (HIGH)
- **Embedded CLI binaries** - SDK now bundles Claude Code CLI v2.0.56 for all platforms
  - darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64
- **`bundled.go`** - Automatic CLI extraction and usage
  - `getBundledCLIPath()` - Extracts bundled CLI to temp directory
  - Fallback to bundled CLI when no system installation found
- **`scripts/download_cli.go`** - Script to download and bundle CLI binaries
- **`_bundled/` directory** structure for multi-platform binaries
- **Updated `findCLI()`** - Checks bundled CLI after system paths
- **`BundledCLIVersion`** constant in `cli_version.go`
- CLI version updated to 2.0.56 (from 2.0.55)

#### SessionStart/SessionEnd Hook Events (MEDIUM)
- **`HookEventSessionStart`** - Hook fired when session starts
- **`HookEventSessionEnd`** - Hook fired when session ends
- **`SessionStartHookInput`** type - Typed input for SessionStart hooks
- **`SessionEndHookInput`** type - Typed input for SessionEnd hooks
- Complete hook lifecycle coverage

#### Settings Field Enhancement (HIGH)
- **JSON string parsing** - Settings field now accepts inline JSON strings
- **Automatic format detection** - Distinguishes between JSON strings and file paths
- **Sandbox merging** - Automatic merging of sandbox settings with existing settings
- **`buildSettingsValue()`** method - Handles settings merging logic
- Matches Python SDK's `_build_settings_value()` behavior

#### MCP Stdin Wait Logic (MEDIUM)
- **First result tracking** - Wait for first result before closing stdin
- **`firstResultChan`** - Channel to signal first result received
- **`firstResultOnce`** - Ensure signal only sent once
- **Updated `StreamInput()`** - Wait logic when SDK MCP servers or hooks present
- 60-second timeout matching Python SDK default
- Fixes race condition with SDK MCP bidirectional communication
- Prevents premature stdin closure

#### System Prompt Default Handling (MEDIUM)
- **Explicit empty string** - Pass `--system-prompt ""` when SystemPrompt is nil
- Prevents using CLI's default system prompt when not intended
- Matches Python SDK behavior exactly

#### Concurrent Write Safety (LOW)
- **`writeMu` mutex** - Dedicated mutex for serializing stdin writes
- Prevents race conditions with concurrent Write() calls
- Enhanced Write() method with proper lock separation

#### Version Management
- **`version.go`** - New file for SDK version constant
- **`SDKVersion`** constant (0.1.0) - Centralized version tracking
- Extracted from transport_subprocess.go for better organization

### Changed
- **Updated `buildCommand()`** signature - Now returns `([]string, error)` for error propagation
- **Enhanced concurrent safety** - Separate read/write locks for state vs I/O operations
- **Improved error messages** - Better CLI not found error with bundled CLI information

### Fixed
- Settings merging with sandbox configuration
- Race condition with SDK MCP servers closing stdin too early
- System prompt default behavior matching Python SDK
- Concurrent write safety to stdin

### Added
- New download script for bundling CLI binaries across platforms
- Unit tests for sandbox types and settings merging
- Example program demonstrating sandbox configuration
- Documentation updates for all new features

---

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

### Added
- **CLI Version Tracking**: New `cli_version.go` file tracking recommended CLI version (2.0.55)
  - `RecommendedCLIVersion` constant for tested CLI version
  - `MinimumCLIVersion` constant for minimum required version
- **GitHub Actions CI/CD**:
  - `.github/workflows/test.yml` - Automated testing on Ubuntu, macOS, and Windows
  - `.github/workflows/lint.yml` - Code linting with golangci-lint, gofmt, and go vet
- **Examples**: New `examples/structured_outputs/` with 3 comprehensive examples
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

**Upgrading to v0.1.13 from v0.1.11:**

1. **No breaking changes** - All new features are additive
2. **Optional features** - Tools option and beta features are optional
3. **Recommended**: Update Claude Code CLI to version 2.0.60:
   ```bash
   npm install -g @anthropic-ai/claude-code@2.0.60
   ```

**New capabilities in v0.1.13:**

1. **Tools Option** - Base set of tools to use, separate from allowed/disallowed filtering
2. **Beta Features** - Enable API beta features like 1M context window
3. **Faster Error Handling** - CLI errors now propagate immediately (~3s vs 60s)
4. **Enhanced Thread Safety** - Fixed TOCTOU race condition in Close()/EndInput()

**Upgrading from older versions:**

1. **No breaking changes** - All features since v0.1.6 are additive
2. **Optional features** - Structured outputs, error handling, hook timeouts, tools, and betas are optional
3. **Recommended**: Update Claude Code CLI to version 2.0.60

**All capabilities:**

1. **Structured Outputs** - Perfect for extracting structured data from queries
2. **Error Handling** - Better visibility into API-level errors
3. **Hook Timeouts** - More control over hook execution time
4. **Tools Option** - Explicit base tool set configuration
5. **Beta Features** - Enable bleeding-edge API features
6. **Sandbox Settings** - Comprehensive security configuration for bash commands

## Compatibility

- **Go**: 1.21+ (unchanged)
- **Claude Code CLI**: 2.0.50+ (minimum), 2.0.76 (recommended)
- **Python SDK Parity**: v0.1.18

## Links

- [Claude Agent SDK Documentation](https://docs.anthropic.com/en/docs/claude-code/sdk)
- [Python SDK Repository](https://github.com/anthropics/claude-agent-sdk-python)
- [Claude Code CLI](https://www.npmjs.com/package/@anthropic-ai/claude-code)
