package claude

// BundledCLIVersion is the version of the Claude Code CLI bundled with this SDK.
// The SDK will use this bundled version if no other CLI installation is found.
const BundledCLIVersion = "2.0.76"

// RecommendedCLIVersion is the recommended version of the Claude Code CLI to use with this SDK.
// This version has been tested and is known to work well with the current SDK features.
//
// To install or update the Claude Code CLI to this version:
//
//	npm install -g @anthropic-ai/claude-code@2.0.76
//
// Or use the automatic installer:
//
//	curl -fsSL https://claude.ai/install.sh | bash -s 2.0.76
//
// For more information, visit: https://docs.claude.com/claude-code
const RecommendedCLIVersion = "2.0.76"

// MinimumCLIVersion is the minimum CLI version required for this SDK.
// Older versions may not support all SDK features (e.g., structured outputs, hook timeouts).
const MinimumCLIVersion = "2.0.50"
