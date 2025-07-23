package main

import (
	"testing"

	"github.com/hayride-dev/bindings/go/hayride/types"
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
