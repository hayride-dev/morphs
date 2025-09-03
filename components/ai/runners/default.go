package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hayride-dev/bindings/go/hayride/ai"
	"github.com/hayride-dev/bindings/go/hayride/ai/agents"
	"github.com/hayride-dev/bindings/go/hayride/ai/graph"
	"github.com/hayride-dev/bindings/go/hayride/ai/models"
	"github.com/hayride-dev/bindings/go/hayride/ai/runner"
	"github.com/hayride-dev/bindings/go/hayride/ai/runner/export"
	"go.bytecodealliance.org/cm"
)

var _ runner.Runner = (*defaultRunner)(nil)

type defaultRunner struct {
	options ai.RunnerOptions
}

// StreamingState tracks the progress of streaming content across different channels
type StreamingState struct {
	AccumulatedContent    string
	LastFinalContent      string
	LastAnalysisContent   string
	LastCommentaryContent string
	HasToolCall           bool
	IsComplete            bool
}

func init() {
	export.Runner(constructor)
}

func constructor(options ai.RunnerOptions) (runner.Runner, error) {
	return &defaultRunner{
		options: options,
	}, nil
}

func (r *defaultRunner) Invoke(message ai.Message, agent agents.Agent, format models.Format, model graph.GraphExecutionContextStream, writer io.Writer) ([]ai.Message, error) {
	messages := make([]ai.Message, 0)

	// If we have a writer, wrap it in a message writer with writer options
	var messageWriter *runner.Writer
	if writer != nil {
		messageWriter = runner.NewWriter(r.options.Writer, writer)
	}

	if err := agent.Push(message); err != nil {
		return nil, fmt.Errorf("failed to push message to agent: %w", err)
	}
	toolCall := false
	for i := 0; i <= int(r.options.MaxTurns); i++ {
		history, err := agent.Context()
		if err != nil {
			return nil, fmt.Errorf("failed to get context: %w", err)
		}
		// Format encode the messages
		data, err := format.Encode(history...)
		if err != nil {
			return nil, fmt.Errorf("failed to encode context messages: %w", err)
		}

		fmt.Println("Encoded Message: ", string(data))

		// Call Graph Compute
		d := graph.TensorDimensions(cm.ToList([]uint32{1}))
		td := graph.TensorData(cm.ToList(data))
		t := graph.NewTensor(d, graph.TensorTypeU8, td)
		inputs := []graph.NamedTensor{
			{
				F0: "user",
				F1: t,
			},
		}

		namedTensorStream, err := model.Compute(inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to compute graph: %w", err)
		}

		// read the output from the stream
		stream := namedTensorStream.F1
		ts := graph.TensorStream(stream)

		text := make([]byte, 0)
		part := make([]byte, 256)

		// Track streaming state for building complete messages
		streamingState := &StreamingState{
			AccumulatedContent:    "",
			LastFinalContent:      "",
			LastAnalysisContent:   "",
			LastCommentaryContent: "",
			HasToolCall:           false,
			IsComplete:            false,
		}

		// Stream structured AI messages (Ollama-style) with proper channel handling
		for {
			bytesRead, err := ts.Read(part)
			if bytesRead == 0 || err == io.EOF {
				break
			} else if err != nil {
				return nil, fmt.Errorf("failed to read from tensor stream: %w", err)
			}
			text = append(text, part[:bytesRead]...)

			// Try to process streaming content
			if messageWriter != nil {
				currentText := string(text)
				r.processStreamingContent(currentText, streamingState, format, messageWriter)
			}
		}

		// After streaming is complete, decode the final complete message
		fmt.Printf("Complete stream data: %s\n", string(text))

		completeMsg, err := format.Decode(text)
		if err != nil {
			return nil, fmt.Errorf("failed to decode complete stream: %w", err)
		}

		// Handle multiple content items as separate logical messages
		// This allows format decoders to return multiple channel contents
		contentItems := completeMsg.Content.Slice()

		if len(contentItems) > 1 {
			// Multiple content items - treat as separate messages
			for i, content := range contentItems {
				if content.String() == "text" && content.Text() != nil {
					// Create separate message for each content item
					separateMsg := ai.Message{
						Role: completeMsg.Role,
						Content: cm.ToList([]ai.MessageContent{
							ai.NewMessageContent(ai.Text(*content.Text())),
						}),
						Final: i == len(contentItems)-1, // Last item is final
					}

					if err := agent.Push(separateMsg); err != nil {
						return nil, fmt.Errorf("failed to push separate message to agent: %w", err)
					}
					messages = append(messages, separateMsg)

					// Stream each separate message
					if messageWriter != nil {
						r.streamMessage(messageWriter, separateMsg)
					}
				}
			}
		} else {
			// Single content item - handle normally
			if err := agent.Push(*completeMsg); err != nil {
				return nil, fmt.Errorf("failed to push complete message to agent: %w", err)
			}
			messages = append(messages, *completeMsg)
		}

		// Check for tool calls in the complete message
		if completeMsg.Role == ai.RoleAssistant {
			for _, c := range completeMsg.Content.Slice() {
				if c.String() == "tool-input" {
					toolResult, err := agent.Execute(*c.ToolInput())
					if err != nil {
						return nil, fmt.Errorf("failed to call tool: %w", err)
					}
					toolCallMessage := ai.Message{
						Role:    ai.RoleTool,
						Content: cm.ToList([]ai.MessageContent{ai.NewMessageContent(*toolResult)}),
					}

					if err := agent.Push(toolCallMessage); err != nil {
						return nil, fmt.Errorf("failed to push tool result to agent: %w", err)
					}
					messages = append(messages, toolCallMessage)
					toolCall = true

					// Stream tool result as structured message
					if messageWriter != nil {
						r.streamMessage(messageWriter, toolCallMessage)
					}
				}
			}
		}

		if toolCall {
			toolCall = false
			continue
		}
		break
	}
	return messages, nil
}

// processStreamingContent handles incremental streaming updates based on Harmony format channels
func (r *defaultRunner) processStreamingContent(currentText string, state *StreamingState, format models.Format, messageWriter *runner.Writer) {
	// Update accumulated content
	state.AccumulatedContent = currentText

	// Try to decode the current content to extract structured information
	partialMsg, err := format.Decode([]byte(currentText))
	if err != nil {
		// If decode fails, check if we have new raw content to stream
		if len(currentText) > len(state.LastFinalContent) {
			newContent := currentText[len(state.LastFinalContent):]
			if newContent != "" {
				deltaMessage := ai.Message{
					Role: ai.RoleAssistant,
					Content: cm.ToList([]ai.MessageContent{
						ai.NewMessageContent(ai.Text(newContent)),
					}),
				}
				r.streamMessage(messageWriter, deltaMessage)
				state.LastFinalContent = currentText
			}
		}
		return
	}

	// Successfully decoded - process different content types
	if partialMsg.Role == ai.RoleAssistant {
		// Check for tool calls first
		for _, content := range partialMsg.Content.Slice() {
			if content.String() == "tool-input" {
				if !state.HasToolCall {
					// Stream the complete tool call message
					r.streamMessage(messageWriter, *partialMsg)
					state.HasToolCall = true
				}
				return
			}
		}

		// Handle text content streaming
		var currentContent string
		for _, content := range partialMsg.Content.Slice() {
			if content.String() == "text" {
				currentContent = *content.Text()
				break
			}
		}

		// Check if this is final content or intermediate content
		if partialMsg.Final {
			// This is final channel content
			if len(currentContent) > len(state.LastFinalContent) {
				newContent := currentContent[len(state.LastFinalContent):]
				if newContent != "" {
					deltaMessage := ai.Message{
						Role: ai.RoleAssistant,
						Content: cm.ToList([]ai.MessageContent{
							ai.NewMessageContent(ai.Text(newContent)),
						}),
					}
					r.streamMessage(messageWriter, deltaMessage)
					state.LastFinalContent = currentContent
				}
			}
		} else {
			// This might be analysis or commentary content
			// For now, we stream incremental updates but can be more selective
			if len(currentContent) > len(state.LastAnalysisContent) {
				newContent := currentContent[len(state.LastAnalysisContent):]
				if newContent != "" && !state.HasToolCall {
					// Only stream analysis/commentary if it's substantial and not a tool call
					if len(newContent) > 10 { // Arbitrary threshold to avoid noise
						deltaMessage := ai.Message{
							Role: ai.RoleAssistant,
							Content: cm.ToList([]ai.MessageContent{
								ai.NewMessageContent(ai.Text(newContent)),
							}),
						}
						r.streamMessage(messageWriter, deltaMessage)
					}
					state.LastAnalysisContent = currentContent
				}
			}
		}
	}
}

// streamMessage sends a structured AI message through the writer (Ollama-style)
func (r *defaultRunner) streamMessage(messageWriter *runner.Writer, message ai.Message) {
	// Always send structured JSON messages, never raw text
	data, err := json.Marshal(message)
	if err == nil {
		messageWriter.Write(data)
	}
}

func main() {}
