# Sandbox Settings Example

This example demonstrates how to configure sandbox settings for the Claude Agent SDK to control bash command execution and network access.

## Features

The sandbox feature provides security by isolating bash commands executed by Claude:

1. **Basic Sandboxing**: Enable sandbox to isolate bash commands
2. **Auto-Allow**: Automatically approve sandboxed bash commands
3. **Network Control**: Restrict network access with allow/block lists
4. **Command Exclusions**: Specify commands to run without sandbox
5. **Violation Ignoring**: Ignore specific sandbox violations

## Running the Example

```bash
cd examples/sandbox
go run main.go
```

## Sandbox Configuration Options

### Basic Configuration

```go
options := &claude.ClaudeAgentOptions{
    Sandbox: &claude.SandboxSettings{
        Enabled: boolPtr(true),
        AutoAllowBashIfSandboxed: boolPtr(true),
    },
}
```

### Network Restrictions

```go
Sandbox: &claude.SandboxSettings{
    Enabled: boolPtr(true),
    Network: &claude.SandboxNetworkConfig{
        Enabled: boolPtr(true),
        AllowedDomains: []string{"*.github.com", "api.anthropic.com"},
        BlockedDomains: []string{"malicious.com"},
    },
}
```

### Excluded Commands

```go
Sandbox: &claude.SandboxSettings{
    Enabled: boolPtr(true),
    ExcludedCommands: []string{"git", "npm"},
    AllowUnsandboxedCommands: boolPtr(true),
}
```

### Ignore Violations

```go
Sandbox: &claude.SandboxSettings{
    Enabled: boolPtr(true),
    IgnoreViolations: &claude.SandboxIgnoreViolations{
        Commands: []string{"ls", "cat"},
        Paths: []string{"/tmp/*"},
    },
}
```

## Use Cases

1. **Development Environment**: Enable sandbox with auto-allow for testing
2. **Production**: Enable sandbox without auto-allow for security
3. **CI/CD**: Use excluded commands for git/docker operations
4. **Restricted Network**: Control which domains can be accessed

## Security Notes

- Sandbox provides an extra layer of security but should not be the only security measure
- Excluded commands run with full system access - use with caution
- Network restrictions help prevent unauthorized external communication
- Review sandbox violations to identify potentially dangerous operations

## Related Examples

- [Basic Usage](../quickstart/) - Getting started with the SDK
- [Hooks](../hooks/) - Intercept and control execution
- [Tool Permissions](../tool_permission_callback/) - Programmatic permission control
