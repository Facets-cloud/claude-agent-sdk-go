package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultMaxBufferSize     = 1024 * 1024 // 1MB
	minimumClaudeCodeVersion = "2.0.0"
	windowsCmdLengthLimit    = 8000   // Windows command line length limit
	nonWindowsCmdLengthLimit = 100000 // Non-Windows systems have much higher limits
)

// SubprocessCLITransport implements Transport using Claude Code CLI subprocess.
type SubprocessCLITransport struct {
	prompt        interface{} // string or <-chan map[string]interface{}
	isStreaming   bool
	options       *ClaudeAgentOptions
	cliPath       string
	cwd           string
	cmd           *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	stderr        io.ReadCloser
	ready         bool
	exitError     error
	maxBufferSize int
	tempFiles     []string // Temporary files created for long command lines
	mu            sync.RWMutex
	writeMu       sync.Mutex // Serializes concurrent writes to stdin
	stderrWg      sync.WaitGroup
}

// NewSubprocessCLITransport creates a new subprocess transport.
func NewSubprocessCLITransport(prompt interface{}, options *ClaudeAgentOptions, cliPath string) (*SubprocessCLITransport, error) {
	if options == nil {
		options = &ClaudeAgentOptions{}
	}

	// Determine if streaming mode
	_, isStreaming := prompt.(<-chan map[string]interface{})

	// Find CLI if not specified
	if cliPath == "" {
		var err error
		cliPath, err = findCLI()
		if err != nil {
			return nil, err
		}
	}

	// Get working directory
	cwd := ""
	if options.Cwd != nil {
		cwd = *options.Cwd
	}

	// Get max buffer size
	maxBufferSize := defaultMaxBufferSize
	if options.MaxBufferSize != nil {
		maxBufferSize = *options.MaxBufferSize
	}

	return &SubprocessCLITransport{
		prompt:        prompt,
		isStreaming:   isStreaming,
		options:       options,
		cliPath:       cliPath,
		cwd:           cwd,
		maxBufferSize: maxBufferSize,
	}, nil
}

// findCLI locates the Claude Code CLI binary.
func findCLI() (string, error) {
	// Check PATH first (prefer user-installed version)
	if path, err := exec.LookPath("claude"); err == nil {
		return path, nil
	}

	// Check common installation locations
	homeDir, _ := os.UserHomeDir()
	locations := []string{
		filepath.Join(homeDir, ".npm-global", "bin", "claude"),
		"/usr/local/bin/claude",
		filepath.Join(homeDir, ".local", "bin", "claude"),
		filepath.Join(homeDir, "node_modules", ".bin", "claude"),
		filepath.Join(homeDir, ".yarn", "bin", "claude"),
		filepath.Join(homeDir, ".claude", "local", "claude"), // Local Claude installation
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	// Finally, check for bundled CLI binary
	if bundledPath, err := getBundledCLIPath(); err == nil && bundledPath != "" {
		return bundledPath, nil
	}

	return "", NewCLINotFoundError(
		"Claude Code CLI not found. The SDK comes with a bundled CLI, but it's not available for your platform.\n"+
			"Please install Claude Code manually:\n"+
			"  npm install -g @anthropic-ai/claude-code\n"+
			"\nIf already installed locally, try:\n"+
			`  export PATH="$HOME/node_modules/.bin:$PATH"`+
			"\n\nOr specify the path when creating transport",
		"",
	)
}

// Connect starts the subprocess and prepares for communication.
func (t *SubprocessCLITransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cmd != nil {
		return nil // Already connected
	}

	// Check Claude Code version
	if err := t.checkClaudeVersion(ctx); err != nil {
		return err
	}

	// Build command
	args, err := t.buildCommand()
	if err != nil {
		return fmt.Errorf("failed to build command: %w", err)
	}
	t.cmd = exec.CommandContext(ctx, t.cliPath, args...)

	// Set working directory
	if t.cwd != "" {
		// Check if directory exists
		if _, err := os.Stat(t.cwd); os.IsNotExist(err) {
			return NewCLIConnectionError(fmt.Sprintf("working directory does not exist: %s", t.cwd), err)
		}
		t.cmd.Dir = t.cwd
	}

	// Set environment variables
	t.cmd.Env = t.buildEnv()

	// Setup pipes
	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return NewCLIConnectionError("failed to create stdin pipe", err)
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		return NewCLIConnectionError("failed to create stdout pipe", err)
	}

	// Setup stderr if needed
	shouldPipeStderr := t.options.Stderr != nil || t.options.ExtraArgs["debug-to-stderr"] != nil
	if shouldPipeStderr {
		t.stderr, err = t.cmd.StderrPipe()
		if err != nil {
			return NewCLIConnectionError("failed to create stderr pipe", err)
		}
	}

	// Start process
	if err := t.cmd.Start(); err != nil {
		t.exitError = NewCLIConnectionError("failed to start Claude Code", err)
		return t.exitError
	}

	// Start stderr reader if needed
	if shouldPipeStderr && t.stderr != nil {
		t.stderrWg.Add(1)
		go t.handleStderr()
	}

	// For non-streaming mode, close stdin immediately
	if !t.isStreaming {
		t.stdin.Close()
	}

	t.ready = true
	return nil
}

// buildSettingsValue builds the settings value, merging sandbox settings if provided.
//
// Returns the settings value as either:
//   - A JSON string (if sandbox is provided or settings is JSON)
//   - A file path (if only settings path is provided without sandbox)
//   - Empty string if neither settings nor sandbox is provided
func (t *SubprocessCLITransport) buildSettingsValue() (string, error) {
	hasSettings := t.options.Settings != nil
	hasSandbox := t.options.Sandbox != nil

	if !hasSettings && !hasSandbox {
		return "", nil
	}

	// If only settings path and no sandbox, pass through as-is
	if hasSettings && !hasSandbox {
		return *t.options.Settings, nil
	}

	// If we have sandbox settings, we need to merge into a JSON object
	settingsObj := make(map[string]interface{})

	if hasSettings {
		settingsStr := strings.TrimSpace(*t.options.Settings)
		// Check if settings is a JSON string or a file path
		if strings.HasPrefix(settingsStr, "{") && strings.HasSuffix(settingsStr, "}") {
			// Parse JSON string
			if err := json.Unmarshal([]byte(settingsStr), &settingsObj); err != nil {
				// If parsing fails, treat as file path
				// Read the file
				data, readErr := os.ReadFile(settingsStr)
				if readErr == nil {
					if jsonErr := json.Unmarshal(data, &settingsObj); jsonErr != nil {
						return "", fmt.Errorf("failed to parse settings file %s: %w", settingsStr, jsonErr)
					}
				}
			}
		} else {
			// It's a file path - read and parse
			data, err := os.ReadFile(settingsStr)
			if err != nil {
				// File doesn't exist or can't be read - not an error, just skip
			} else {
				if err := json.Unmarshal(data, &settingsObj); err != nil {
					return "", fmt.Errorf("failed to parse settings file %s: %w", settingsStr, err)
				}
			}
		}
	}

	// Merge sandbox settings
	if hasSandbox {
		settingsObj["sandbox"] = t.options.Sandbox
	}

	jsonBytes, err := json.Marshal(settingsObj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal settings: %w", err)
	}

	return string(jsonBytes), nil
}

// buildCommand constructs CLI arguments from options.
func (t *SubprocessCLITransport) buildCommand() ([]string, error) {
	args := []string{"--output-format", "stream-json", "--verbose"}

	// System prompt
	if t.options.SystemPrompt == nil {
		// Explicitly pass empty string to avoid using CLI's default system prompt
		args = append(args, "--system-prompt", "")
	} else {
		switch sp := t.options.SystemPrompt.(type) {
		case string:
			args = append(args, "--system-prompt", sp)
		case map[string]interface{}:
			if sp["type"] == "preset" && sp["append"] != nil {
				args = append(args, "--append-system-prompt", sp["append"].(string))
			}
		case SystemPromptPreset:
			if sp.Append != nil {
				args = append(args, "--append-system-prompt", *sp.Append)
			}
		}
	}

	// Tools option (base set of tools)
	if t.options.Tools != nil {
		switch tools := t.options.Tools.(type) {
		case []string:
			if len(tools) == 0 {
				args = append(args, "--tools", "")
			} else {
				args = append(args, "--tools", strings.Join(tools, ","))
			}
		case map[string]interface{}:
			// Preset object - 'claude_code' preset maps to 'default'
			if tools["type"] == "preset" {
				args = append(args, "--tools", "default")
			}
		case ToolsPreset:
			// Preset object - 'claude_code' preset maps to 'default'
			args = append(args, "--tools", "default")
		}
	}

	// Tool restrictions (filter from base tools)
	if len(t.options.AllowedTools) > 0 {
		args = append(args, "--allowedTools", strings.Join(t.options.AllowedTools, ","))
	}
	if len(t.options.DisallowedTools) > 0 {
		args = append(args, "--disallowedTools", strings.Join(t.options.DisallowedTools, ","))
	}

	// Limits
	if t.options.MaxTurns != nil {
		args = append(args, "--max-turns", fmt.Sprintf("%d", *t.options.MaxTurns))
	}

	// Model
	if t.options.Model != nil {
		args = append(args, "--model", *t.options.Model)
	}
	if t.options.FallbackModel != nil {
		args = append(args, "--fallback-model", *t.options.FallbackModel)
	}

	// Beta features
	if len(t.options.Betas) > 0 {
		betaStrs := make([]string, len(t.options.Betas))
		for i, beta := range t.options.Betas {
			betaStrs[i] = string(beta)
		}
		args = append(args, "--betas", strings.Join(betaStrs, ","))
	}

	// Budget and token control
	if t.options.MaxBudgetUSD != nil {
		args = append(args, "--max-budget-usd", fmt.Sprintf("%.2f", *t.options.MaxBudgetUSD))
	}
	if t.options.MaxThinkingTokens != nil {
		args = append(args, "--max-thinking-tokens", fmt.Sprintf("%d", *t.options.MaxThinkingTokens))
	}

	// Permission settings
	if t.options.PermissionMode != nil {
		args = append(args, "--permission-mode", string(*t.options.PermissionMode))
	}
	if t.options.PermissionPromptToolName != nil {
		args = append(args, "--permission-prompt-tool", *t.options.PermissionPromptToolName)
	}

	// Conversation continuation
	if t.options.ContinueConversation {
		args = append(args, "--continue")
	}
	if t.options.Resume != nil {
		args = append(args, "--resume", *t.options.Resume)
	}
	if t.options.ForkSession {
		args = append(args, "--fork-session")
	}

	// Settings (merge with sandbox if needed)
	if settingsValue, err := t.buildSettingsValue(); err != nil {
		return nil, err
	} else if settingsValue != "" {
		args = append(args, "--settings", settingsValue)
	}

	// Additional directories
	for _, dir := range t.options.AddDirs {
		args = append(args, "--add-dir", dir)
	}

	// MCP servers
	if len(t.options.McpServers) > 0 {
		serversForCLI := make(map[string]interface{})
		for name, config := range t.options.McpServers {
			if sdkConfig, ok := config.(McpSdkServerConfig); ok {
				// For SDK servers, pass everything except instance
				serversForCLI[name] = map[string]interface{}{
					"type": sdkConfig.Type,
					"name": sdkConfig.Name,
				}
			} else {
				serversForCLI[name] = config
			}
		}
		if len(serversForCLI) > 0 {
			mcpJSON, _ := json.Marshal(map[string]interface{}{"mcpServers": serversForCLI})
			args = append(args, "--mcp-config", string(mcpJSON))
		}
	}

	// Partial messages
	if t.options.IncludePartialMessages {
		args = append(args, "--include-partial-messages")
	}

	// Agents
	if len(t.options.Agents) > 0 {
		agentsJSON, _ := json.Marshal(t.options.Agents)
		args = append(args, "--agents", string(agentsJSON))
	}

	// Setting sources
	if t.options.SettingSources != nil {
		sources := make([]string, len(t.options.SettingSources))
		for i, src := range t.options.SettingSources {
			sources[i] = string(src)
		}
		args = append(args, "--setting-sources", strings.Join(sources, ","))
	} else {
		args = append(args, "--setting-sources", "")
	}

	// Plugins
	if len(t.options.Plugins) > 0 {
		for _, plugin := range t.options.Plugins {
			if plugin.Type == "local" {
				args = append(args, "--plugin-dir", plugin.Path)
			}
			// Note: Other plugin types can be added in the future
		}
	}

	// Extra args
	for flag, value := range t.options.ExtraArgs {
		if value == nil {
			args = append(args, "--"+flag)
		} else {
			args = append(args, "--"+flag, *value)
		}
	}

	// Input mode
	if t.isStreaming {
		args = append(args, "--input-format", "stream-json")
	} else {
		// String prompt
		args = append(args, "--print", "--", t.prompt.(string))
	}

	// Check if command line is too long (Windows limitation)
	// This optimization helps when large agent definitions would exceed command line limits
	cmdStr := strings.Join(args, " ")
	cmdLengthLimit := nonWindowsCmdLengthLimit
	if isWindows() {
		cmdLengthLimit = windowsCmdLengthLimit
	}

	if len(cmdStr) > cmdLengthLimit && len(t.options.Agents) > 0 {
		// Command is too long - use temp file for agents
		// Find the --agents argument and replace its value with @filepath
		for i, arg := range args {
			if arg == "--agents" && i+1 < len(args) {
				agentsJSONValue := args[i+1]

				// Create a temporary file
				tempFile, err := os.CreateTemp("", "claude-agents-*.json")
				if err != nil {
					// Log warning but continue - the command might still work
					fmt.Fprintf(os.Stderr, "Warning: Failed to create temp file for long command: %v\n", err)
					break
				}

				// Write the agents JSON to the file
				if _, err := tempFile.WriteString(agentsJSONValue); err != nil {
					tempFile.Close()
					os.Remove(tempFile.Name())
					fmt.Fprintf(os.Stderr, "Warning: Failed to write to temp file: %v\n", err)
					break
				}
				tempFile.Close()

				// Track for cleanup
				t.tempFiles = append(t.tempFiles, tempFile.Name())

				// Replace agents JSON with @filepath reference
				args[i+1] = "@" + tempFile.Name()

				fmt.Fprintf(os.Stderr, "Command line length (%d) exceeds limit (%d). Using temp file for --agents: %s\n",
					len(cmdStr), cmdLengthLimit, tempFile.Name())
				break
			}
		}
	}

	return args, nil
}

// isWindows returns true if running on Windows
func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}

// buildEnv constructs environment variables.
func (t *SubprocessCLITransport) buildEnv() []string {
	env := os.Environ()

	// Add user env vars
	for k, v := range t.options.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Add SDK identifier
	env = append(env, "CLAUDE_CODE_ENTRYPOINT=sdk-go")
	env = append(env, fmt.Sprintf("CLAUDE_AGENT_SDK_VERSION=%s", SDKVersion))

	// Enable file checkpointing if requested
	if t.options.EnableFileCheckpointing {
		env = append(env, "CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING=true")
	}

	// Set PWD if cwd is specified
	if t.cwd != "" {
		env = append(env, fmt.Sprintf("PWD=%s", t.cwd))
	}

	return env
}

// handleStderr reads stderr in background.
func (t *SubprocessCLITransport) handleStderr() {
	defer t.stderrWg.Done()

	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if t.options.Stderr != nil {
			t.options.Stderr(line)
		}
	}
}

// Write sends data to stdin.
func (t *SubprocessCLITransport) Write(ctx context.Context, data string) error {
	// Use read lock to check state
	t.mu.RLock()
	if !t.ready || t.stdin == nil {
		t.mu.RUnlock()
		return NewCLIConnectionError("transport is not ready for writing", nil)
	}

	if t.cmd.ProcessState != nil {
		t.mu.RUnlock()
		return NewCLIConnectionError(fmt.Sprintf("cannot write to terminated process (exit code: %d)", t.cmd.ProcessState.ExitCode()), nil)
	}

	if t.exitError != nil {
		t.mu.RUnlock()
		return NewCLIConnectionError(fmt.Sprintf("cannot write to process that exited with error: %v", t.exitError), t.exitError)
	}
	t.mu.RUnlock()

	// Serialize concurrent writes to stdin
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	_, err := t.stdin.Write([]byte(data))
	if err != nil {
		t.mu.Lock()
		t.ready = false
		t.exitError = NewCLIConnectionError("failed to write to process stdin", err)
		t.mu.Unlock()
		return t.exitError
	}

	return nil
}

// ReadMessages reads and parses messages from stdout.
func (t *SubprocessCLITransport) ReadMessages(ctx context.Context) (<-chan map[string]interface{}, <-chan error) {
	msgCh := make(chan map[string]interface{}, 10)
	errCh := make(chan error, 1)

	go func() {
		defer close(msgCh)
		defer close(errCh)

		scanner := bufio.NewScanner(t.stdout)
		// Set initial buffer size for scanner (configurable, default 64KB)
		initialSize := 64 * 1024
		if t.options != nil && t.options.ScannerInitialBufferSize != nil && *t.options.ScannerInitialBufferSize > 0 {
			initialSize = *t.options.ScannerInitialBufferSize
		}
		buf := make([]byte, 0, initialSize)
		scanner.Buffer(buf, t.maxBufferSize)

		var jsonBuffer strings.Builder

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Split by newlines (in case multiple JSON objects on one line)
			jsonLines := strings.Split(line, "\n")

			for _, jsonLine := range jsonLines {
				jsonLine = strings.TrimSpace(jsonLine)
				if jsonLine == "" {
					continue
				}

				// Accumulate partial JSON using strings.Builder for efficiency
				jsonBuffer.WriteString(jsonLine)

				if jsonBuffer.Len() > t.maxBufferSize {
					errCh <- NewCLIJSONDecodeError(
						fmt.Sprintf("JSON message exceeded maximum buffer size of %d bytes", t.maxBufferSize),
						fmt.Errorf("buffer size %d exceeds limit %d", jsonBuffer.Len(), t.maxBufferSize),
					)
					return
				}

				// Try to parse
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(jsonBuffer.String()), &data); err == nil {
					// Successfully parsed
					jsonBuffer.Reset()
					msgCh <- data
				}
				// If parse fails, keep accumulating
			}
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			errCh <- NewCLIConnectionError("error reading from stdout", err)
			return
		}

		// Wait for process to complete
		if err := t.cmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				t.exitError = NewProcessError(
					"command failed",
					exitErr.ExitCode(),
					"check stderr output for details",
				)
				errCh <- t.exitError
			}
		}
	}()

	return msgCh, errCh
}

// EndInput closes stdin to signal end of input.
func (t *SubprocessCLITransport) EndInput() error {
	// Acquire write lock first to prevent TOCTOU race with Write()
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stdin != nil {
		err := t.stdin.Close()
		t.stdin = nil
		return err
	}
	return nil
}

// IsReady checks if transport is ready for communication.
//
// Returns true after successful Connect() and before Close().
// Thread-safe for concurrent access.
func (t *SubprocessCLITransport) IsReady() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.ready
}

// Close terminates the subprocess and cleans up.
func (t *SubprocessCLITransport) Close() error {
	// Acquire write lock first to prevent TOCTOU race with Write()
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	t.mu.Lock()
	defer t.mu.Unlock()

	t.ready = false

	if t.cmd == nil {
		return nil
	}

	// Close stdin to signal the process
	if t.stdin != nil {
		t.stdin.Close()
		t.stdin = nil
	}

	// Kill process if still running
	if t.cmd.Process != nil && t.cmd.ProcessState == nil {
		t.cmd.Process.Kill()
	}

	// Wait for process with timeout to avoid hanging
	if t.cmd != nil && t.cmd.Process != nil {
		done := make(chan struct{})
		go func() {
			t.cmd.Wait()
			close(done)
		}()

		// Wait up to 2 seconds for process to exit
		select {
		case <-done:
			// Process exited normally
		case <-time.After(2 * time.Second):
			// Force kill if still running
			if t.cmd.Process != nil {
				t.cmd.Process.Signal(os.Kill)
			}
		}
	}

	// Wait for stderr reader to finish (with timeout)
	stderrDone := make(chan struct{})
	go func() {
		t.stderrWg.Wait()
		close(stderrDone)
	}()
	select {
	case <-stderrDone:
	case <-time.After(1 * time.Second):
		// Stderr reader didn't finish, continue anyway
	}

	// Clean up temporary files
	for _, tempFile := range t.tempFiles {
		if err := os.Remove(tempFile); err != nil {
			// Log but don't fail on cleanup errors
			fmt.Fprintf(os.Stderr, "Warning: Failed to remove temp file %s: %v\n", tempFile, err)
		}
	}
	t.tempFiles = nil

	t.cmd = nil
	t.exitError = nil

	return nil
}

// checkClaudeVersion checks if the Claude Code CLI version meets minimum requirements.
// Returns an error if the version check fails critically, or logs a warning for outdated versions.
func (t *SubprocessCLITransport) checkClaudeVersion(ctx context.Context) error {
	// Skip version check if environment variable is set
	if os.Getenv("CLAUDE_AGENT_SDK_SKIP_VERSION_CHECK") != "" {
		return nil
	}

	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Run claude -v to get version
	cmd := exec.CommandContext(checkCtx, t.cliPath, "-v")
	output, err := cmd.Output()
	if err != nil {
		// If version check fails, log but don't block (CLI might still work)
		return nil
	}

	// Parse version from output
	versionStr := strings.TrimSpace(string(output))
	re := regexp.MustCompile(`([0-9]+\.[0-9]+\.[0-9]+)`)
	match := re.FindStringSubmatch(versionStr)

	if match == nil {
		// Couldn't parse version, skip check
		return nil
	}

	version := match[1]

	// Compare versions
	if compareVersions(version, minimumClaudeCodeVersion) < 0 {
		warning := fmt.Sprintf("Warning: Claude Code version %s is unsupported in the Agent SDK. "+
			"Minimum required version is %s. "+
			"Some features may not work correctly.", version, minimumClaudeCodeVersion)
		fmt.Fprintln(os.Stderr, warning)
	}

	return nil
}

// compareVersions compares two semantic version strings.
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < 3; i++ {
		var num1, num2 int
		if i < len(parts1) {
			num1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			num2, _ = strconv.Atoi(parts2[i])
		}

		if num1 < num2 {
			return -1
		}
		if num1 > num2 {
			return 1
		}
	}

	return 0
}
