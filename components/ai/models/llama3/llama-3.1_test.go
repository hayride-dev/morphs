package llama3

import (
	"strings"
	"testing"

	"github.com/hayride-dev/bindings/go/hayride/types"
	"go.bytecodealliance.org/cm"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name         string
		input        []byte
		wantText     string
		wantToolName string
		wantToolArgs [][2]string
		wantRole     types.Role
		wantErr      bool
	}{
		{
			name:     "basic decode",
			input:    []byte("Hello, world!"),
			wantText: "Hello, world!",
			wantRole: types.RoleAssistant,
			wantErr:  false,
		},
		{
			name:         "decode with function call",
			input:        []byte("<function=example_function_name>{\"example_name\": \"example_value\"}</function>"),
			wantToolName: "example_function_name",
			wantToolArgs: [][2]string{{"example_name", "example_value"}},
			wantRole:     types.RoleAssistant,
			wantErr:      false,
		},
	}

	model := llama3{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := model.Decode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if msg.Content.Len() == 0 {
				t.Error("Decoded message content is empty")
			}
			if msg.Role != tt.wantRole {
				t.Errorf("Expected role '%s', got '%s'", tt.wantRole, msg.Role)
			}

			for _, content := range msg.Content.Slice() {
				switch content.String() {
				case "text":
					// Check if the text content matches the expected content
					if *content.Text() != tt.wantText {
						t.Errorf("Expected content '%s', got '%s'", tt.wantText, *content.Text())
					}
				case "blob":
					// Check if the blob content matches the expected content
					if content.Blob() == nil || string(content.Blob().Slice()) != tt.wantText {
						t.Errorf("Expected blob content '%s', got '%v'", tt.wantText, content.Blob())
					}
				case "tool-input":
					// Check if the tool input matches the expected content
					if content.ToolInput() == nil {
						t.Error("Expected tool input, got nil")
					} else {
						if content.ToolInput().Name != tt.wantToolName {
							t.Errorf("Expected tool input name '%s', got '%s'", tt.wantToolName, content.ToolInput().Name)
						}
						for i, arg := range content.ToolInput().Arguments.Slice() {
							if arg != tt.wantToolArgs[i] {
								t.Errorf("Expected tool input argument '%s', got '%s'", tt.wantToolArgs[i], arg)
							}
						}
					}
				default:
					t.Errorf("Unexpected content type: %s", content.String())
				}
			}
		})
	}
}

func TestEncode_BasicConversation(t *testing.T) {
	model := &llama3{}

	// Create a basic conversation: system + user message
	messages := []types.Message{
		{
			Role: types.RoleSystem,
			Content: cm.ToList([]types.MessageContent{
				types.NewMessageContent(types.Text("You are a helpful assistant.")),
			}),
		},
		{
			Role: types.RoleUser,
			Content: cm.ToList([]types.MessageContent{
				types.NewMessageContent(types.Text("Hello, how are you?")),
			}),
		},
	}

	encoded, err := model.Encode(messages...)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	result := string(encoded)
	t.Logf("Encoded output:\n%s", result)

	// Validate the structure
	if !strings.Contains(result, "<|start_header_id|>system<|end_header_id|>") {
		t.Error("Missing system header")
	}

	if !strings.Contains(result, "Environment: ipython") {
		t.Error("Missing environment token for tools")
	}

	if !strings.Contains(result, "You are a helpful assistant.") {
		t.Error("Missing system message content")
	}

	if !strings.Contains(result, "<|eot_id|>") {
		t.Error("Missing end of turn token")
	}

	if !strings.Contains(result, "<|start_header_id|>user<|end_header_id|>") {
		t.Error("Missing user header")
	}

	if !strings.Contains(result, "Hello, how are you?") {
		t.Error("Missing user message content")
	}

	if !strings.Contains(result, "<|start_header_id|>assistant<|end_header_id|>") {
		t.Error("Missing assistant header at end")
	}

	// Validate the exact expected format
	expectedParts := []string{
		"<|start_header_id|>system<|end_header_id|>",
		"Environment: ipython",
		"You are a helpful assistant.",
		"<|eot_id|>",
		"<|start_header_id|>user<|end_header_id|>",
		"Hello, how are you?",
		"<|eot_id|>",
		"<|start_header_id|>assistant<|end_header_id|>",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Missing expected part: %s", part)
		}
	}
}

func TestEncode_WithAssistantResponse(t *testing.T) {
	model := &llama3{}

	// Create conversation with assistant response
	messages := []types.Message{
		{
			Role: types.RoleUser,
			Content: cm.ToList([]types.MessageContent{
				types.NewMessageContent(types.Text("Hello")),
			}),
		},
		{
			Role: types.RoleAssistant,
			Content: cm.ToList([]types.MessageContent{
				types.NewMessageContent(types.Text("Hi there!")),
			}),
		},
	}

	encoded, err := model.Encode(messages...)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	result := string(encoded)
	t.Logf("Encoded output with assistant:\n%s", result)

	// Should NOT have assistant header at the end since last message is from assistant
	if strings.HasSuffix(strings.TrimSpace(result), "<|start_header_id|>assistant<|end_header_id|>") {
		t.Error("Should not add assistant header when last message is from assistant")
	}

	// Should contain the assistant message properly formatted
	expectedParts := []string{
		"<|start_header_id|>user<|end_header_id|>",
		"Hello",
		"<|start_header_id|>assistant<|end_header_id|>",
		"Hi there!",
		"<|eot_id|>",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Missing expected part: %s", part)
		}
	}
}

func TestEncode_ToolCall(t *testing.T) {
	model := &llama3{}

	// Create a tool call scenario
	messages := []types.Message{
		{
			Role: types.RoleUser,
			Content: cm.ToList([]types.MessageContent{
				types.NewMessageContent(types.Text("What's the weather?")),
			}),
		},
		{
			Role: types.RoleAssistant,
			Content: cm.ToList([]types.MessageContent{
				types.NewMessageContent(types.CallToolParams{
					Name: "get_weather",
					Arguments: cm.ToList([][2]string{
						{"location", "New York"},
					}),
				}),
			}),
		},
		{
			Role: types.RoleTool,
			Content: cm.ToList([]types.MessageContent{
				types.NewMessageContent(types.Text("It's sunny, 72°F")),
			}),
		},
	}

	encoded, err := model.Encode(messages...)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	result := string(encoded)
	t.Logf("Encoded tool call output:\n%s", result)

	// Validate tool call format
	expectedParts := []string{
		"<|start_header_id|>user<|end_header_id|>",
		"What's the weather?",
		"<|start_header_id|>assistant<|end_header_id|>",
		"<function=get_weather>",
		`{"location":"New York"}`,
		"</function>",
		"<|eom_id|>", // Should use end of message for tool calls
		"<|start_header_id|>ipython<|end_header_id|>",
		"It's sunny, 72°F",
		"<|eot_id|>",
		"<|start_header_id|>assistant<|end_header_id|>", // Should add assistant header at end
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Missing expected part: %s", part)
		}
	}
}
