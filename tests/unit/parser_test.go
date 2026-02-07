package unit

import (
	"testing"

	claude "github.com/Facets-cloud/claude-agent-sdk-go"
)

func TestParseUserMessage(t *testing.T) {
	t.Run("simple string content", func(t *testing.T) {
		data := map[string]interface{}{
			"type": "user",
			"message": map[string]interface{}{
				"role":    "user",
				"content": "Hello Claude",
			},
		}

		msg, err := claude.ParseMessage(data)
		if err != nil {
			t.Fatalf("ParseMessage failed: %v", err)
		}

		userMsg, ok := msg.(*claude.UserMessage)
		if !ok {
			t.Fatalf("expected *UserMessage, got %T", msg)
		}

		if userMsg.Content != "Hello Claude" {
			t.Errorf("expected content 'Hello Claude', got %v", userMsg.Content)
		}
	})

	t.Run("content blocks", func(t *testing.T) {
		data := map[string]interface{}{
			"type": "user",
			"message": map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Hello",
					},
					map[string]interface{}{
						"type":        "tool_result",
						"tool_use_id": "tool_123",
						"content":     "result",
					},
				},
			},
		}

		msg, err := claude.ParseMessage(data)
		if err != nil {
			t.Fatalf("ParseMessage failed: %v", err)
		}

		userMsg, ok := msg.(*claude.UserMessage)
		if !ok {
			t.Fatalf("expected *UserMessage, got %T", msg)
		}

		blocks, ok := userMsg.Content.([]claude.ContentBlock)
		if !ok {
			t.Fatalf("expected []ContentBlock, got %T", userMsg.Content)
		}

		if len(blocks) != 2 {
			t.Fatalf("expected 2 blocks, got %d", len(blocks))
		}

		textBlock, ok := blocks[0].(claude.TextBlock)
		if !ok {
			t.Errorf("expected TextBlock, got %T", blocks[0])
		}
		if textBlock.Text != "Hello" {
			t.Errorf("expected text 'Hello', got %s", textBlock.Text)
		}

		toolResultBlock, ok := blocks[1].(claude.ToolResultBlock)
		if !ok {
			t.Errorf("expected ToolResultBlock, got %T", blocks[1])
		}
		if toolResultBlock.ToolUseID != "tool_123" {
			t.Errorf("expected tool_use_id 'tool_123', got %s", toolResultBlock.ToolUseID)
		}
	})

	t.Run("with uuid field", func(t *testing.T) {
		// The uuid field is needed for file checkpointing with rewind_files()
		data := map[string]interface{}{
			"type": "user",
			"uuid": "msg-abc123-def456",
			"message": map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Hello",
					},
				},
			},
		}

		msg, err := claude.ParseMessage(data)
		if err != nil {
			t.Fatalf("ParseMessage failed: %v", err)
		}

		userMsg, ok := msg.(*claude.UserMessage)
		if !ok {
			t.Fatalf("expected *UserMessage, got %T", msg)
		}

		if userMsg.UUID == nil {
			t.Fatal("expected UUID to be set, got nil")
		}
		if *userMsg.UUID != "msg-abc123-def456" {
			t.Errorf("expected UUID 'msg-abc123-def456', got %s", *userMsg.UUID)
		}
	})

	t.Run("without uuid field", func(t *testing.T) {
		data := map[string]interface{}{
			"type": "user",
			"message": map[string]interface{}{
				"role":    "user",
				"content": "Hello",
			},
		}

		msg, err := claude.ParseMessage(data)
		if err != nil {
			t.Fatalf("ParseMessage failed: %v", err)
		}

		userMsg, ok := msg.(*claude.UserMessage)
		if !ok {
			t.Fatalf("expected *UserMessage, got %T", msg)
		}

		if userMsg.UUID != nil {
			t.Errorf("expected UUID to be nil, got %s", *userMsg.UUID)
		}
	})
}

func TestParseAssistantMessage(t *testing.T) {
	data := map[string]interface{}{
		"type": "assistant",
		"message": map[string]interface{}{
			"role":  "assistant",
			"model": "claude-sonnet-4-5",
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Hello!",
				},
				map[string]interface{}{
					"type":      "thinking",
					"thinking":  "Let me think...",
					"signature": "sig123",
				},
				map[string]interface{}{
					"type":  "tool_use",
					"id":    "tool_456",
					"name":  "Read",
					"input": map[string]interface{}{"path": "/test"},
				},
			},
		},
	}

	msg, err := claude.ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	assistantMsg, ok := msg.(*claude.AssistantMessage)
	if !ok {
		t.Fatalf("expected *AssistantMessage, got %T", msg)
	}

	if assistantMsg.Model != "claude-sonnet-4-5" {
		t.Errorf("expected model 'claude-sonnet-4-5', got %s", assistantMsg.Model)
	}

	if len(assistantMsg.Content) != 3 {
		t.Fatalf("expected 3 content blocks, got %d", len(assistantMsg.Content))
	}

	// Check text block
	textBlock, ok := assistantMsg.Content[0].(claude.TextBlock)
	if !ok {
		t.Errorf("expected TextBlock, got %T", assistantMsg.Content[0])
	}
	if textBlock.Text != "Hello!" {
		t.Errorf("expected text 'Hello!', got %s", textBlock.Text)
	}

	// Check thinking block
	thinkingBlock, ok := assistantMsg.Content[1].(claude.ThinkingBlock)
	if !ok {
		t.Errorf("expected ThinkingBlock, got %T", assistantMsg.Content[1])
	}
	if thinkingBlock.Thinking != "Let me think..." {
		t.Errorf("expected thinking 'Let me think...', got %s", thinkingBlock.Thinking)
	}

	// Check tool use block
	toolUseBlock, ok := assistantMsg.Content[2].(claude.ToolUseBlock)
	if !ok {
		t.Errorf("expected ToolUseBlock, got %T", assistantMsg.Content[2])
	}
	if toolUseBlock.Name != "Read" {
		t.Errorf("expected tool name 'Read', got %s", toolUseBlock.Name)
	}
}

func TestParseImageBlock(t *testing.T) {
	// Test image block in assistant message
	data := map[string]interface{}{
		"type": "assistant",
		"message": map[string]interface{}{
			"role":  "assistant",
			"model": "claude-sonnet-4-5",
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Here's the image:",
				},
				map[string]interface{}{
					"type":     "image",
					"data":     "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
					"mimeType": "image/png",
				},
			},
		},
	}

	msg, err := claude.ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	assistantMsg, ok := msg.(*claude.AssistantMessage)
	if !ok {
		t.Fatalf("expected *AssistantMessage, got %T", msg)
	}

	if len(assistantMsg.Content) != 2 {
		t.Fatalf("expected 2 content blocks, got %d", len(assistantMsg.Content))
	}

	// Check image block
	imageBlock, ok := assistantMsg.Content[1].(claude.ImageBlock)
	if !ok {
		t.Fatalf("expected ImageBlock, got %T", assistantMsg.Content[1])
	}

	expectedData := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
	if imageBlock.Data != expectedData {
		t.Errorf("expected data '%s', got %s", expectedData, imageBlock.Data)
	}

	if imageBlock.MimeType != "image/png" {
		t.Errorf("expected mimeType 'image/png', got %s", imageBlock.MimeType)
	}
}

func TestParseResultMessage(t *testing.T) {
	data := map[string]interface{}{
		"type":            "result",
		"subtype":         "success",
		"duration_ms":     1000.0,
		"duration_api_ms": 800.0,
		"is_error":        false,
		"num_turns":       5.0,
		"session_id":      "session_123",
		"total_cost_usd":  0.05,
		"usage": map[string]interface{}{
			"input_tokens":  100.0,
			"output_tokens": 50.0,
		},
		"result": "completed",
	}

	msg, err := claude.ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	resultMsg, ok := msg.(*claude.ResultMessage)
	if !ok {
		t.Fatalf("expected *ResultMessage, got %T", msg)
	}

	if resultMsg.Subtype != "success" {
		t.Errorf("expected subtype 'success', got %s", resultMsg.Subtype)
	}
	if resultMsg.DurationMS != 1000 {
		t.Errorf("expected duration_ms 1000, got %d", resultMsg.DurationMS)
	}
	if resultMsg.SessionID != "session_123" {
		t.Errorf("expected session_id 'session_123', got %s", resultMsg.SessionID)
	}
	if resultMsg.TotalCostUSD == nil || *resultMsg.TotalCostUSD != 0.05 {
		t.Errorf("expected total_cost_usd 0.05, got %v", resultMsg.TotalCostUSD)
	}
}

func TestParseSystemMessage(t *testing.T) {
	data := map[string]interface{}{
		"type":    "system",
		"subtype": "init",
		"data": map[string]interface{}{
			"commands": []interface{}{"help", "exit"},
		},
	}

	msg, err := claude.ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	systemMsg, ok := msg.(*claude.SystemMessage)
	if !ok {
		t.Fatalf("expected *SystemMessage, got %T", msg)
	}

	if systemMsg.Subtype != "init" {
		t.Errorf("expected subtype 'init', got %s", systemMsg.Subtype)
	}
}

func TestParseStreamEvent(t *testing.T) {
	data := map[string]interface{}{
		"type":       "stream_event",
		"uuid":       "event_123",
		"session_id": "session_456",
		"event": map[string]interface{}{
			"type":  "content_block_delta",
			"delta": map[string]interface{}{"text": "Hello"},
		},
	}

	msg, err := claude.ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	streamEvent, ok := msg.(*claude.StreamEvent)
	if !ok {
		t.Fatalf("expected *StreamEvent, got %T", msg)
	}

	if streamEvent.UUID != "event_123" {
		t.Errorf("expected uuid 'event_123', got %s", streamEvent.UUID)
	}
	if streamEvent.SessionID != "session_456" {
		t.Errorf("expected session_id 'session_456', got %s", streamEvent.SessionID)
	}
}

func TestParseMessageErrors(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
	}{
		{
			name: "nil data",
			data: nil,
		},
		{
			name: "missing type",
			data: map[string]interface{}{
				"message": "test",
			},
		},
		{
			name: "unknown type",
			data: map[string]interface{}{
				"type": "unknown",
			},
		},
		{
			name: "user message missing message field",
			data: map[string]interface{}{
				"type": "user",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := claude.ParseMessage(tt.data)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestParseUserMessageWithMixedContent(t *testing.T) {
	// Test parsing user messages with mixed content blocks (text + tool_result)
	// This is a common pattern and should be handled correctly
	data := map[string]interface{}{
		"type": "user",
		"message": map[string]interface{}{
			"role": "user",
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Here's the file content:",
				},
				map[string]interface{}{
					"type":        "tool_result",
					"tool_use_id": "toolu_123",
					"content":     "File contents here",
				},
				map[string]interface{}{
					"type": "text",
					"text": "What do you think?",
				},
			},
		},
	}

	msg, err := claude.ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	userMsg, ok := msg.(*claude.UserMessage)
	if !ok {
		t.Fatalf("expected *UserMessage, got %T", msg)
	}

	// Content should be array of blocks
	contentBlocks, ok := userMsg.Content.([]claude.ContentBlock)
	if !ok {
		t.Fatalf("expected content to be []ContentBlock, got %T", userMsg.Content)
	}

	if len(contentBlocks) != 3 {
		t.Errorf("expected 3 content blocks, got %d", len(contentBlocks))
	}
}

func TestParseMessagePreservesErrorData(t *testing.T) {
	// Test that parse errors contain the original data for debugging
	data := map[string]interface{}{
		"type":    "user",
		"message": "invalid", // Should be a map, not a string
	}

	_, err := claude.ParseMessage(data)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Error should be a MessageParseError
	parseErr, ok := err.(*claude.MessageParseError)
	if !ok {
		t.Fatalf("expected *MessageParseError, got %T", err)
	}

	// Error message should contain useful debugging information
	errMsg := parseErr.Error()
	if errMsg == "" {
		t.Error("error message should not be empty")
	}
}

func TestParseAssistantMessageWithError(t *testing.T) {
	// Test parsing AssistantMessage with error field
	t.Run("with error field", func(t *testing.T) {
		data := map[string]interface{}{
			"type":  "assistant",
			"error": "rate_limit",
			"message": map[string]interface{}{
				"role":  "assistant",
				"model": "claude-sonnet-4-5",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Error occurred",
					},
				},
			},
		}

		msg, err := claude.ParseMessage(data)
		if err != nil {
			t.Fatalf("ParseMessage failed: %v", err)
		}

		assistantMsg, ok := msg.(*claude.AssistantMessage)
		if !ok {
			t.Fatalf("expected *AssistantMessage, got %T", msg)
		}

		if assistantMsg.Error == nil {
			t.Fatal("expected error field to be set")
		}

		if *assistantMsg.Error != claude.AssistantMessageErrorRateLimit {
			t.Errorf("expected error 'rate_limit', got %s", *assistantMsg.Error)
		}
	})

	t.Run("without error field", func(t *testing.T) {
		data := map[string]interface{}{
			"type": "assistant",
			"message": map[string]interface{}{
				"role":  "assistant",
				"model": "claude-sonnet-4-5",
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Success",
					},
				},
			},
		}

		msg, err := claude.ParseMessage(data)
		if err != nil {
			t.Fatalf("ParseMessage failed: %v", err)
		}

		assistantMsg, ok := msg.(*claude.AssistantMessage)
		if !ok {
			t.Fatalf("expected *AssistantMessage, got %T", msg)
		}

		if assistantMsg.Error != nil {
			t.Errorf("expected error field to be nil, got %v", assistantMsg.Error)
		}
	})

	t.Run("all error types", func(t *testing.T) {
		errorTypes := []claude.AssistantMessageError{
			claude.AssistantMessageErrorAuthenticationFailed,
			claude.AssistantMessageErrorBillingError,
			claude.AssistantMessageErrorRateLimit,
			claude.AssistantMessageErrorInvalidRequest,
			claude.AssistantMessageErrorServerError,
			claude.AssistantMessageErrorUnknown,
		}

		for _, errType := range errorTypes {
			data := map[string]interface{}{
				"type":  "assistant",
				"error": string(errType),
				"message": map[string]interface{}{
					"role":    "assistant",
					"model":   "claude-sonnet-4-5",
					"content": []interface{}{},
				},
			}

			msg, err := claude.ParseMessage(data)
			if err != nil {
				t.Fatalf("ParseMessage failed for error type %s: %v", errType, err)
			}

			assistantMsg := msg.(*claude.AssistantMessage)
			if assistantMsg.Error == nil || *assistantMsg.Error != errType {
				t.Errorf("expected error type %s, got %v", errType, assistantMsg.Error)
			}
		}
	})
}

func TestParseResultMessageWithStructuredOutput(t *testing.T) {
	// Test parsing ResultMessage with structured_output field
	t.Run("with simple structured output", func(t *testing.T) {
		data := map[string]interface{}{
			"type":            "result",
			"subtype":         "success",
			"duration_ms":     1000.0,
			"duration_api_ms": 800.0,
			"is_error":        false,
			"num_turns":       5.0,
			"session_id":      "session_123",
			"structured_output": map[string]interface{}{
				"file_count": 10.0,
				"has_tests":  true,
			},
		}

		msg, err := claude.ParseMessage(data)
		if err != nil {
			t.Fatalf("ParseMessage failed: %v", err)
		}

		resultMsg, ok := msg.(*claude.ResultMessage)
		if !ok {
			t.Fatalf("expected *ResultMessage, got %T", msg)
		}

		if resultMsg.StructuredOutput == nil {
			t.Fatal("expected structured_output field to be set")
		}

		output, ok := resultMsg.StructuredOutput.(map[string]interface{})
		if !ok {
			t.Fatalf("expected structured_output to be map, got %T", resultMsg.StructuredOutput)
		}

		if fileCount, ok := output["file_count"].(float64); !ok || fileCount != 10.0 {
			t.Errorf("expected file_count=10, got %v", output["file_count"])
		}

		if hasTests, ok := output["has_tests"].(bool); !ok || !hasTests {
			t.Errorf("expected has_tests=true, got %v", output["has_tests"])
		}
	})

	t.Run("with nested structured output", func(t *testing.T) {
		data := map[string]interface{}{
			"type":            "result",
			"subtype":         "success",
			"duration_ms":     1000.0,
			"duration_api_ms": 800.0,
			"is_error":        false,
			"num_turns":       5.0,
			"session_id":      "session_123",
			"structured_output": map[string]interface{}{
				"analysis": map[string]interface{}{
					"word_count":      2.0,
					"character_count": 11.0,
				},
				"words": []interface{}{"Hello", "world"},
			},
		}

		msg, err := claude.ParseMessage(data)
		if err != nil {
			t.Fatalf("ParseMessage failed: %v", err)
		}

		resultMsg, ok := msg.(*claude.ResultMessage)
		if !ok {
			t.Fatalf("expected *ResultMessage, got %T", msg)
		}

		if resultMsg.StructuredOutput == nil {
			t.Fatal("expected structured_output field to be set")
		}

		output, ok := resultMsg.StructuredOutput.(map[string]interface{})
		if !ok {
			t.Fatalf("expected structured_output to be map, got %T", resultMsg.StructuredOutput)
		}

		// Check nested analysis
		analysis, ok := output["analysis"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected analysis to be map, got %T", output["analysis"])
		}

		if wordCount, ok := analysis["word_count"].(float64); !ok || wordCount != 2.0 {
			t.Errorf("expected word_count=2, got %v", analysis["word_count"])
		}

		// Check words array
		words, ok := output["words"].([]interface{})
		if !ok {
			t.Fatalf("expected words to be array, got %T", output["words"])
		}

		if len(words) != 2 {
			t.Errorf("expected 2 words, got %d", len(words))
		}
	})

	t.Run("without structured output", func(t *testing.T) {
		data := map[string]interface{}{
			"type":            "result",
			"subtype":         "success",
			"duration_ms":     1000.0,
			"duration_api_ms": 800.0,
			"is_error":        false,
			"num_turns":       5.0,
			"session_id":      "session_123",
		}

		msg, err := claude.ParseMessage(data)
		if err != nil {
			t.Fatalf("ParseMessage failed: %v", err)
		}

		resultMsg, ok := msg.(*claude.ResultMessage)
		if !ok {
			t.Fatalf("expected *ResultMessage, got %T", msg)
		}

		if resultMsg.StructuredOutput != nil {
			t.Errorf("expected structured_output to be nil, got %v", resultMsg.StructuredOutput)
		}
	})
}
