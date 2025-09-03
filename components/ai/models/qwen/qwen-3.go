package qwen

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hayride-dev/bindings/go/hayride/ai"
	"github.com/hayride-dev/bindings/go/hayride/ai/models"
	"github.com/hayride-dev/bindings/go/hayride/mcp"
	"go.bytecodealliance.org/cm"
)

var (
	// Regular expression to match tool calls in the format:
	// <tool_call>\n{"name": "function_name", "arguments": {...}}\n</tool_call>
	qwen3ToolCallRegex = regexp.MustCompile(`<tool_call>\s*(\{.*?\})\s*</tool_call>`)
)

const (
	// Qwen 3 special tokens
	qwen3ImStart = "<|im_start|>"
	qwen3ImEnd   = "<|im_end|>"

	// Tool-related tokens
	qwen3ToolResponse    = "<tool_response>"
	qwen3ToolResponseEnd = "</tool_response>"
	qwen3ToolCall        = "<tool_call>"
	qwen3ToolCallEnd     = "</tool_call>"

	// Thinking tokens for reasoning
	qwen3Think    = "<think>"
	qwen3ThinkEnd = "</think>"
)

var _ models.Format = (*qwen3)(nil)

func ConstructorQwen_3() (models.Format, error) {
	return &qwen3{}, nil
}

type qwen3 struct{}

func (m *qwen3) Decode(data []byte) (*ai.Message, error) {
	content := string(data)

	// Remove thinking tags if present - these are for internal reasoning
	if strings.Contains(content, qwen3Think) && strings.Contains(content, qwen3ThinkEnd) {
		start := strings.Index(content, qwen3Think)
		end := strings.Index(content, qwen3ThinkEnd) + len(qwen3ThinkEnd)
		if start != -1 && end != -1 && end > start {
			content = content[:start] + content[end:]
			content = strings.TrimSpace(content)
		}
	}

	// Check if this is a tool call response
	if strings.Contains(content, qwen3ToolCall) {
		matches := qwen3ToolCallRegex.FindStringSubmatch(content)
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

		return &ai.Message{
			Role: ai.RoleAssistant,
			Content: cm.ToList([]ai.MessageContent{
				ai.NewMessageContent(mcp.CallToolParams{
					Name:      toolCallData.Name,
					Arguments: cm.ToList(args),
				}),
			}),
		}, nil
	}

	// Regular text message
	return &ai.Message{
		Role: ai.RoleAssistant,
		Content: cm.ToList([]ai.MessageContent{
			ai.NewMessageContent(ai.Text(content)),
		}),
	}, nil
}

func (m *qwen3) Encode(messages ...ai.Message) ([]byte, error) {
	builder := &strings.Builder{}

	// Track if we have tools available
	var tools []mcp.Tool
	hasTools := false

	// First pass: collect tools from system message
	for _, msg := range messages {
		if msg.Role == ai.RoleSystem {
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

	// Find the last query index for multi-step tool detection
	lastQueryIndex := len(messages) - 1

	// Iterate backwards to find the last real user query
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role == ai.RoleUser {
			// Check if this is a tool response wrapper
			isToolResponse := false
			for _, content := range msg.Content.Slice() {
				if content.String() == "text" {
					text := *content.Text()
					if strings.HasPrefix(text, qwen3ToolResponse) && strings.HasSuffix(text, qwen3ToolResponseEnd) {
						isToolResponse = true
						break
					}
				}
			}
			if !isToolResponse {
				lastQueryIndex = i
				break
			}
		}
	}

	// Process messages
	for i, msg := range messages {
		switch msg.Role {
		case ai.RoleSystem:
			// Handle tools section first if available
			if hasTools {
				builder.WriteString(fmt.Sprintf("%ssystem\n", qwen3ImStart))

				// Add system content
				systemContent := ""
				for _, content := range msg.Content.Slice() {
					if content.String() == "text" {
						c := content.Text()
						systemContent = *c
						break
					}
				}

				if systemContent != "" {
					builder.WriteString(fmt.Sprintf("%s\n\n", systemContent))
				}

				// Add tools section
				builder.WriteString("# Tools\n\nYou may call one or more functions to assist with the user query.\n\nYou are provided with function signatures within <tools></tools> XML tags:\n<tools>")

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
				builder.WriteString(fmt.Sprintf("%s\n", qwen3ImEnd))
			} else {
				// Regular system message without tools
				builder.WriteString(fmt.Sprintf("%ssystem\n", qwen3ImStart))
				for _, content := range msg.Content.Slice() {
					if content.String() == "text" {
						c := content.Text()
						builder.WriteString(*c)
					}
				}
				builder.WriteString(fmt.Sprintf("%s\n", qwen3ImEnd))
			}

		case ai.RoleUser:
			builder.WriteString(fmt.Sprintf("%suser\n", qwen3ImStart))

			for _, content := range msg.Content.Slice() {
				if content.String() == "text" {
					c := content.Text()
					builder.WriteString(*c)
				}
			}

			builder.WriteString(fmt.Sprintf("%s\n", qwen3ImEnd))

		case ai.RoleAssistant:
			isLastMessage := i == len(messages)-1
			shouldAddThinking := i > lastQueryIndex && (isLastMessage || !isLastMessage)

			builder.WriteString(fmt.Sprintf("%sassistant", qwen3ImStart))

			if shouldAddThinking {
				builder.WriteString(fmt.Sprintf("\n%s\n\n%s\n\n", qwen3Think, qwen3ThinkEnd))
			}

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
					builder.WriteString(fmt.Sprintf("\n%s\n%s\n%s", qwen3ToolCall, string(toolCallBytes), qwen3ToolCallEnd))
				}
			}

			builder.WriteString(fmt.Sprintf("%s\n", qwen3ImEnd))

		case ai.RoleTool:
			// Check if this is the first tool message or if the previous message was not a tool
			if i == 0 || messages[i-1].Role != ai.RoleTool {
				builder.WriteString(fmt.Sprintf("%suser", qwen3ImStart))
			}

			builder.WriteString(fmt.Sprintf("\n%s\n", qwen3ToolResponse))

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

			builder.WriteString(fmt.Sprintf("\n%s", qwen3ToolResponseEnd))

			// Check if this is the last tool message or if the next message is not a tool
			if i == len(messages)-1 || messages[i+1].Role != ai.RoleTool {
				builder.WriteString(fmt.Sprintf("%s\n", qwen3ImEnd))
			}

		default:
			return nil, fmt.Errorf("unsupported message role: %v", msg.Role)
		}
	}

	// Add generation prompt if the last message is not from assistant
	if len(messages) > 0 && messages[len(messages)-1].Role != ai.RoleAssistant {
		builder.WriteString(fmt.Sprintf("%sassistant\n", qwen3ImStart))
		// Add thinking tags for new responses
		builder.WriteString(fmt.Sprintf("%s\n\n%s\n\n", qwen3Think, qwen3ThinkEnd))
	}

	return []byte(builder.String()), nil
}
