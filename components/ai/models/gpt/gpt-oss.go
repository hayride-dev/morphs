package gpt

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hayride-dev/bindings/go/hayride/ai/models"
	"github.com/hayride-dev/bindings/go/hayride/types"
	"go.bytecodealliance.org/cm"
)

var (
	// Regular expression to match tool calls in the format:
	// ```json\n{"function": "name", "arguments": {...}}\n```
	gptOSSToolCallRegex = regexp.MustCompile("```json\\s*(\\{.*?\\})\\s*```")
)

const (
	// GPT-OSS special tokens
	startToken = "<|start|>"
	endToken   = "<|end|>"

	// Message format tokens
	messageToken = "<|message|>"

	// Role definitions
	systemRole    = "system"
	userRole      = "user"
	assistantRole = "assistant"
	developerRole = "developer"

	// Channel definitions
	analysisChannel   = "analysis"
	commentaryChannel = "commentary"
	finalChannel      = "final"

	// Special formatting for tool calls
	constrainJSON = "<|constrain|>json"
	callToken     = "<|call|>"
	channelToken  = "<|channel|>"
	toToken       = " to="
)

var _ models.Format = (*gptOSS)(nil)

func ConstructorGPTOSS() (models.Format, error) {
	return &gptOSS{}, nil
}

type gptOSS struct{}

func (m *gptOSS) Decode(data []byte) (*types.Message, error) {
	content := string(data)

	// Remove any start/end tokens if present
	content = strings.TrimPrefix(content, startToken+assistantRole)
	content = strings.TrimSuffix(content, endToken)

	// Check for tool calls in JSON code blocks
	if strings.Contains(content, "```json") {
		matches := gptOSSToolCallRegex.FindStringSubmatch(content)
		if len(matches) == 2 {
			// Parse the JSON tool call
			var toolCallData map[string]interface{}
			if err := json.Unmarshal([]byte(matches[1]), &toolCallData); err != nil {
				return nil, fmt.Errorf("failed to parse tool call JSON: %v", err)
			}

			// Extract function name and arguments
			functionName, ok := toolCallData["function"].(string)
			if !ok {
				return nil, fmt.Errorf("tool call missing function name")
			}

			// Convert arguments to [][2]string format
			var args [][2]string
			if argsInterface, exists := toolCallData["arguments"]; exists {
				switch argsVal := argsInterface.(type) {
				case map[string]interface{}:
					for k, v := range argsVal {
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
				}
			}

			return &types.Message{
				Role: types.RoleAssistant,
				Content: cm.ToList([]types.MessageContent{
					types.NewMessageContent(types.CallToolParams{
						Name:      functionName,
						Arguments: cm.ToList(args),
					}),
				}),
			}, nil
		}
	}

	// Regular text message - strip any channel markers
	content = stripChannelMarkers(content)
	content = strings.TrimSpace(content)

	return &types.Message{
		Role: types.RoleAssistant,
		Content: cm.ToList([]types.MessageContent{
			types.NewMessageContent(types.Text(content)),
		}),
	}, nil
}

func (m *gptOSS) Encode(messages ...types.Message) ([]byte, error) {
	builder := &strings.Builder{}

	// Track system configuration
	var tools []types.Tool
	hasTools := false
	var systemContent string

	// First pass: collect system information and tools
	for _, msg := range messages {
		if msg.Role == types.RoleSystem {
			for _, content := range msg.Content.Slice() {
				switch content.String() {
				case "text":
					c := content.Text()
					systemContent = *c
				case "tools":
					tools = content.Tools().Slice()
					hasTools = len(tools) > 0
				}
			}
			break
		}
	}

	// Build system message
	builder.WriteString(fmt.Sprintf("%s%s%s", startToken, systemRole, messageToken))

	// Add base system prompt
	builder.WriteString("You are ChatGPT, a large language model trained by OpenAI.\n")
	builder.WriteString("Knowledge cutoff: 2024-06\n")
	builder.WriteString(fmt.Sprintf("Current date: %s\n", time.Now().Format("2006-01-02")))
	builder.WriteString("\nReasoning: medium\n")

	// Add tools section if available
	if hasTools {
		builder.WriteString("\n# Tools\n\n")
		builder.WriteString("## functions\n\n")
		builder.WriteString("namespace functions {\n")

		for _, tool := range tools {
			// Add tool description as comment
			if tool.Description != "" {
				builder.WriteString(fmt.Sprintf("// %s\n", tool.Description))
			}

			// Build parameter type definition
			if len(tool.InputSchema.Properties.Slice()) > 0 {
				builder.WriteString(fmt.Sprintf("type %s = (_: {\n", tool.Name))
				for _, prop := range tool.InputSchema.Properties.Slice() {
					// Add property description if available
					builder.WriteString(fmt.Sprintf("  %s: string,\n", prop[0]))
				}
				builder.WriteString("}) => any;\n")
			} else {
				builder.WriteString(fmt.Sprintf("type %s = () => any;\n", tool.Name))
			}
		}

		builder.WriteString("\n} // namespace functions\n")
		builder.WriteString("\n# Valid channels: analysis, commentary, final. Channel must be included for every message.\n")
		builder.WriteString("Calls to these tools must go to the commentary channel: 'functions'.\n")
	}

	builder.WriteString(fmt.Sprintf("%s", endToken))

	// Add developer message if we have tools or custom system content
	if hasTools || (systemContent != "" && systemContent != "You are ChatGPT, a large language model trained by OpenAI.") {
		builder.WriteString(fmt.Sprintf("%s%s%s", startToken, developerRole, messageToken))

		if hasTools {
			builder.WriteString("# Tools\n\n")
			builder.WriteString("## functions\n\n")
			builder.WriteString("namespace functions {\n")

			for _, tool := range tools {
				if tool.Description != "" {
					builder.WriteString(fmt.Sprintf("// %s\n", tool.Description))
				}

				if len(tool.InputSchema.Properties.Slice()) > 0 {
					builder.WriteString(fmt.Sprintf("type %s = (_: {\n", tool.Name))
					for _, prop := range tool.InputSchema.Properties.Slice() {
						builder.WriteString(fmt.Sprintf("  // %s\n", prop[1]))
						builder.WriteString(fmt.Sprintf("  %s: string,\n", prop[0]))
					}
					builder.WriteString("}) => any;\n")
				} else {
					builder.WriteString(fmt.Sprintf("type %s = () => any;\n", tool.Name))
				}
			}

			builder.WriteString("\n} // namespace functions\n\n")
		}

		if systemContent != "" && systemContent != "You are ChatGPT, a large language model trained by OpenAI." {
			builder.WriteString("# Instructions\n\n")
			builder.WriteString(systemContent)
			builder.WriteString("\n")
		}

		builder.WriteString(fmt.Sprintf("%s", endToken))
	}

	// Find the last user message index
	lastUserIdx := -1
	for i, msg := range messages {
		if msg.Role == types.RoleUser {
			lastUserIdx = i
		}
	}

	// Process all messages except system
	for i, msg := range messages {
		if msg.Role == types.RoleSystem {
			continue // Already handled
		}

		isLast := i == len(messages)-1

		switch msg.Role {
		case types.RoleUser:
			builder.WriteString(fmt.Sprintf("%s%s%s", startToken, userRole, messageToken))

			for _, content := range msg.Content.Slice() {
				if content.String() == "text" {
					c := content.Text()
					builder.WriteString(*c)
				}
			}

			builder.WriteString(fmt.Sprintf("%s", endToken))

		case types.RoleAssistant:
			// For analysis content after last user message
			hasThinking := false
			for _, content := range msg.Content.Slice() {
				if content.String() == "text" && i > lastUserIdx {
					hasThinking = true
					break
				}
			}

			if hasThinking && i > lastUserIdx {
				builder.WriteString(fmt.Sprintf("%s%s%s%s%s", startToken, assistantRole, channelToken, analysisChannel, messageToken))
				// Add some thinking content placeholder
				builder.WriteString("Let me analyze this request...")
				if !isLast {
					builder.WriteString(fmt.Sprintf("%s", endToken))
				}
			}

			// Process content
			for _, content := range msg.Content.Slice() {
				switch content.String() {
				case "text":
					c := content.Text()
					if !hasThinking || i <= lastUserIdx {
						builder.WriteString(fmt.Sprintf("%s%s%s%s%s", startToken, assistantRole, channelToken, finalChannel, messageToken))
						builder.WriteString(*c)
						if !isLast {
							builder.WriteString(fmt.Sprintf("%s", endToken))
						}
					}

				case "tool-input":
					c := content.ToolInput()

					// Build the tool call
					builder.WriteString(fmt.Sprintf("%s%s%s%s%sfunctions.%s %s%s",
						startToken, assistantRole, channelToken, commentaryChannel, toToken,
						c.Name, constrainJSON, messageToken))

					// Convert arguments to JSON
					argsMap := make(map[string]interface{})
					for _, arg := range c.Arguments.Slice() {
						// Try to parse as JSON, otherwise use as string
						var value interface{}
						if err := json.Unmarshal([]byte(arg[1]), &value); err != nil {
							value = arg[1]
						}
						argsMap[arg[0]] = value
					}

					argsJSON, _ := json.Marshal(argsMap)
					builder.WriteString(string(argsJSON))
					builder.WriteString(fmt.Sprintf("%s", callToken))
				}
			}

		case types.RoleTool:
			for _, content := range msg.Content.Slice() {
				if content.String() == "tool-output" {
					output := content.ToolOutput()

					builder.WriteString(fmt.Sprintf("%sfunctions.tool_result to=%s%s",
						startToken, assistantRole, messageToken))

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
							resourceContent := c.ResourceContent()
							switch resourceContent.ResourceContents.String() {
							case "text":
								builder.WriteString(fmt.Sprintf("Resource Content (Text): %s", resourceContent.ResourceContents.Text()))
							case "blob":
								builder.WriteString(fmt.Sprintf("Resource Content (Blob): %v", resourceContent.ResourceContents.Blob()))
							}
						}
					}

					builder.WriteString(fmt.Sprintf("%s", endToken))
				}
			}

		default:
			return nil, fmt.Errorf("unsupported message role: %v", msg.Role)
		}
	}

	// Add generation prompt if the last message is not from assistant
	if len(messages) > 0 && messages[len(messages)-1].Role != types.RoleAssistant {
		builder.WriteString(fmt.Sprintf("%s%s", startToken, assistantRole))
	}

	return []byte(builder.String()), nil
}

// stripChannelMarkers removes channel markers from content
func stripChannelMarkers(content string) string {
	// Remove channel markers like <|channel|>analysis<|message|>
	channelRegex := regexp.MustCompile(`<\|channel\|>[^<]*<\|message\|>`)
	content = channelRegex.ReplaceAllString(content, "")

	// Remove constrain markers
	content = strings.ReplaceAll(content, constrainJSON, "")
	content = strings.ReplaceAll(content, callToken, "")

	return content
}
