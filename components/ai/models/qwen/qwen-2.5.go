package qwen

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hayride-dev/bindings/go/hayride/ai/models"
	"github.com/hayride-dev/bindings/go/hayride/types"
	"go.bytecodealliance.org/cm"
)

var (
	// Regular expression to match tool calls in the format:
	// <tool_call>\n{"name": "function_name", "arguments": {...}}\n</tool_call>
	toolCallRegex = regexp.MustCompile(`<tool_call>\s*(\{.*?\})\s*</tool_call>`)
)

const (
	// Qwen 2.5 special tokens
	imStart = "<|im_start|>"
	imEnd   = "<|im_end|>"

	// Tool-related tokens
	toolResponse    = "<tool_response>"
	toolResponseEnd = "</tool_response>"
	toolCall        = "<tool_call>"
	toolCallEnd     = "</tool_call>"
)

var _ models.Format = (*qwen25)(nil)

func Constructor() (models.Format, error) {
	return &qwen25{}, nil
}

type qwen25 struct{}

func (m *qwen25) Decode(data []byte) (*types.Message, error) {
	content := string(data)

	// Check if this is a tool call response
	if strings.Contains(content, toolCall) {
		matches := toolCallRegex.FindStringSubmatch(content)
		if len(matches) != 2 {
			return nil, fmt.Errorf("failed to parse tool call, invalid format")
		}

		// Parse the JSON tool call
		var toolCallData struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}

		if err := json.Unmarshal([]byte(matches[1]), &toolCallData); err != nil {
			return nil, fmt.Errorf("failed to parse tool call JSON: %v", err)
		}

		// Convert arguments to [][2]string format
		var argsMap map[string]interface{}
		if err := json.Unmarshal(toolCallData.Arguments, &argsMap); err != nil {
			return nil, fmt.Errorf("failed to parse tool arguments: %v", err)
		}

		var args [][2]string
		for k, v := range argsMap {
			// Convert value to string
			var valueStr string
			switch val := v.(type) {
			case string:
				valueStr = val
			case nil:
				valueStr = ""
			default:
				// Convert other types to JSON string
				jsonBytes, _ := json.Marshal(val)
				valueStr = string(jsonBytes)
			}
			args = append(args, [2]string{k, valueStr})
		}

		return &types.Message{
			Role: types.RoleAssistant,
			Content: cm.ToList([]types.MessageContent{
				types.NewMessageContent(types.CallToolParams{
					Name:      toolCallData.Name,
					Arguments: cm.ToList(args),
				}),
			}),
		}, nil
	}

	// Regular text message
	return &types.Message{
		Role: types.RoleAssistant,
		Content: cm.ToList([]types.MessageContent{
			types.NewMessageContent(types.Text(content)),
		}),
	}, nil
}

func (m *qwen25) Encode(messages ...types.Message) ([]byte, error) {
	builder := &strings.Builder{}

	// Track if we have tools available
	var tools []types.Tool
	hasTools := false

	// First pass: collect tools from system message
	for _, msg := range messages {
		if msg.Role == types.RoleSystem {
			for _, content := range msg.Content.Slice() {
				if content.String() == "tools" {
					tools = content.Tools().Slice()
					hasTools = len(tools) > 0
					break
				}
			}
			break
		}
	}

	// Process messages
	for i, msg := range messages {
		switch msg.Role {
		case types.RoleSystem:
			builder.WriteString(fmt.Sprintf("%ssystem\n", imStart))

			// Add system content
			systemContent := ""
			for _, content := range msg.Content.Slice() {
				if content.String() == "text" {
					c := content.Text()
					systemContent = *c
					break
				}
			}

			// Use default system message if none provided
			if systemContent == "" {
				systemContent = "You are Qwen, created by Alibaba Cloud. You are a helpful assistant."
			}

			builder.WriteString(systemContent)

			// Add tools section if tools are available
			if hasTools {
				builder.WriteString("\n\n# Tools\n\nYou may call one or more functions to assist with the user query.\n\nYou are provided with function signatures within <tools></tools> XML tags:\n<tools>")

				for _, tool := range tools {
					toolJSON := map[string]interface{}{
						"name":        tool.Name,
						"description": tool.Description,
						"parameters":  map[string]interface{}{},
					}

					// Add parameters if available
					if len(tool.InputSchema.Properties.Slice()) > 0 {
						params := make(map[string]interface{})
						for _, prop := range tool.InputSchema.Properties.Slice() {
							params[prop[0]] = prop[1]
						}
						toolJSON["parameters"] = params
					}

					toolBytes, _ := json.Marshal(toolJSON)
					builder.WriteString(fmt.Sprintf("\n%s", string(toolBytes)))
				}

				builder.WriteString("\n</tools>\n\nFor each function call, return a json object with function name and arguments within <tool_call></tool_call> XML tags:\n<tool_call>\n{\"name\": <function-name>, \"arguments\": <args-json-object>}\n</tool_call>")
			}

			builder.WriteString(fmt.Sprintf("%s\n", imEnd))

		case types.RoleUser:
			builder.WriteString(fmt.Sprintf("%suser\n", imStart))

			for _, content := range msg.Content.Slice() {
				if content.String() == "text" {
					c := content.Text()
					builder.WriteString(*c)
				}
			}

			builder.WriteString(fmt.Sprintf("%s\n", imEnd))

		case types.RoleAssistant:
			builder.WriteString(fmt.Sprintf("%sassistant", imStart))

			for _, content := range msg.Content.Slice() {
				switch content.String() {
				case "text":
					c := content.Text()
					builder.WriteString(fmt.Sprintf("\n%s", *c))
				case "tool-input":
					c := content.ToolInput()

					// Convert arguments to JSON object
					argsMap := make(map[string]interface{})
					for _, arg := range c.Arguments.Slice() {
						// Try to parse as JSON, otherwise use as string
						var value interface{}
						if err := json.Unmarshal([]byte(arg[1]), &value); err != nil {
							value = arg[1]
						}
						argsMap[arg[0]] = value
					}

					toolCallJSON := map[string]interface{}{
						"name":      c.Name,
						"arguments": argsMap,
					}

					toolCallBytes, _ := json.Marshal(toolCallJSON)
					builder.WriteString(fmt.Sprintf("\n%s\n%s\n%s", toolCall, string(toolCallBytes), toolCallEnd))
				}
			}

			builder.WriteString(fmt.Sprintf("%s\n", imEnd))

		case types.RoleTool:
			// Check if this is the first tool message or if the previous message was not a tool
			if i == 0 || messages[i-1].Role != types.RoleTool {
				builder.WriteString(fmt.Sprintf("%suser", imStart))
			}

			builder.WriteString(fmt.Sprintf("\n%s\n", toolResponse))

			for _, content := range msg.Content.Slice() {
				if content.String() == "tool-output" {
					output := content.ToolOutput()
					for _, c := range output.Content.Slice() {
						switch c.String() {
						case "text":
							builder.WriteString(c.Text().Text)
						case "image":
							image := c.Image()
							builder.WriteString(fmt.Sprintf("Image Data: %v", image.Data))
						case "audio":
							audio := c.Audio()
							builder.WriteString(fmt.Sprintf("Audio Data: %v", audio.Data))
						case "resource-link":
							resource := c.ResourceLink()
							builder.WriteString(fmt.Sprintf("Resource Link: %s", resource.URI))
						case "resource-content":
							content := c.ResourceContent()
							switch content.ResourceContents.String() {
							case "text":
								builder.WriteString(fmt.Sprintf("Resource Content (Text): %s", content.ResourceContents.Text()))
							case "blob":
								builder.WriteString(fmt.Sprintf("Resource Content (Blob): %v", content.ResourceContents.Blob()))
							}
						}
					}
				}
			}

			builder.WriteString(fmt.Sprintf("\n%s", toolResponseEnd))

			// Check if this is the last tool message or if the next message is not a tool
			if i == len(messages)-1 || messages[i+1].Role != types.RoleTool {
				builder.WriteString(fmt.Sprintf("%s\n", imEnd))
			}

		default:
			return nil, fmt.Errorf("unsupported message role: %v", msg.Role)
		}
	}

	// Add generation prompt if the last message is not from assistant
	if len(messages) > 0 && messages[len(messages)-1].Role != types.RoleAssistant {
		builder.WriteString(fmt.Sprintf("%sassistant\n", imStart))
	}

	return []byte(builder.String()), nil
}
