package main

import (
	"encoding/json"
	"fmt"
	"io"
	"unsafe"

	"github.com/hayride-dev/morphs/components/ai/agents/internal/gen/hayride/ai/agents"
	inferencestream "github.com/hayride-dev/morphs/components/ai/agents/internal/gen/hayride/ai/inference-stream"
	"github.com/hayride-dev/morphs/components/ai/agents/internal/gen/hayride/ai/types"
	"github.com/hayride-dev/morphs/components/ai/agents/internal/gen/wasi/nn/tensor"
	"go.bytecodealliance.org/cm"
)

const maxturn = 10

var resourceTable = resources{
	agents: make(map[cm.Rep]*agent),
}

func init() {
	agents.Exports.Agent.Constructor = constructor
	agents.Exports.Agent.Invoke = invoke
	agents.Exports.Agent.InvokeStream = invokeStream
	agents.Exports.Agent.Destructor = destructor
}

type resources struct {
	agents map[cm.Rep]*agent
}

type tensorStream cm.Resource

// Read will read the next `len` bytes from the stream
// will return empty byte slice if the stream is closed.
// blocks until the data is available
func (t tensorStream) Read(p []byte) (int, error) {
	ts := cm.Reinterpret[inferencestream.TensorStream](t)
	ts.Subscribe().Block()
	data := ts.Read(uint64(len(p)))
	if data.IsErr() {
		if data.Err().Closed() {
			return 0, nil
		}
		return 0, fmt.Errorf("%s", data.Err().String())
	}
	n := copy(p, data.OK().Slice())
	p = p[:n]
	return len(p), nil
}

type agent struct {
	name    string
	tools   agents.Tools
	context agents.Context
	format  agents.Format
	graph   agents.GraphExecutionContextStream
}

func constructor(name string, instruction string, tools_ agents.Tools, context_ agents.Context, format agents.Format, graph agents.GraphExecutionContextStream) agents.Agent {
	agent := &agent{
		name:    name,
		tools:   tools_,
		context: context_,
		format:  format,
		graph:   graph,
	}

	content := []types.Content{}
	content = append(content, types.ContentText(types.TextContent{
		Text: instruction,
	}))

	result := tools_.Capabilities()
	if result.IsErr() {
		return cm.ResourceNone
	}
	for _, t := range result.OK().Slice() {
		content = append(content, types.ContentToolSchema(t))
	}

	agent.context.Push(agents.Message{Role: types.RoleSystem, Content: cm.ToList(content)})

	key := cm.Rep(uintptr(unsafe.Pointer(agent)))
	v := agents.AgentResourceNew(key)
	resourceTable.agents[key] = agent
	return v
}

func invoke(self cm.Rep, input agents.Message) cm.Result[agents.MessageShape, agents.Message, agents.Error] {
	agent, ok := resourceTable.agents[self]
	if !ok {
		wasiErr := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
		return cm.Err[cm.Result[agents.MessageShape, agents.Message, agents.Error]](wasiErr)
	}

	result := agent.context.Push(input)
	if result.IsErr() {
		err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
		return cm.Err[cm.Result[agents.MessageShape, agents.Message, agents.Error]](err)
	}

	finalMsg := &types.Message{Role: types.RoleAssistant, Content: cm.ToList([]types.Content{types.ContentText(types.TextContent{
		Text: "agent yielded no response",
	})})}

	for i := 0; i <= maxturn; i++ {
		result := agent.context.Messages()
		if result.IsErr() {
			err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
			return cm.Err[cm.Result[agents.MessageShape, agents.Message, agents.Error]](err)
		}
		msgs := result.OK().Slice()

		encodedResult := agent.format.Encode(cm.ToList(msgs))
		if encodedResult.IsErr() {
			err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
			return cm.Err[cm.Result[agents.MessageShape, agents.Message, agents.Error]](err)
		}

		d := tensor.TensorDimensions(cm.ToList([]uint32{1}))
		td := tensor.TensorData(cm.ToList(encodedResult.OK().Slice()))
		t := tensor.NewTensor(d, tensor.TensorTypeU8, td)
		inputs := []inferencestream.NamedTensor{
			{
				F0: "user",
				F1: t,
			},
		}
		computeResult := agent.graph.Compute(cm.ToList(inputs))
		if computeResult.IsErr() {
			err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
			return cm.Err[cm.Result[agents.MessageShape, agents.Message, agents.Error]](err)
		}

		stream := computeResult.OK().F1
		ts := tensorStream(stream)
		// read the output from the stream
		text := make([]byte, 0)
		for {
			// Read up to 100 bytes from the output
			// to get any tokens that have been generated and push to socket
			p := make([]byte, 100)
			len, err := ts.Read(p)
			if len == 0 || err == io.EOF {
				break
			} else if err != nil {
				err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
				return cm.Err[cm.Result[agents.MessageShape, agents.Message, agents.Error]](err)
			}
			text = append(text, p[:len]...)
		}

		decodeResult := agent.format.Decode(cm.ToList(text))
		if decodeResult.IsErr() {
			err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
			return cm.Err[cm.Result[agents.MessageShape, agents.Message, agents.Error]](err)
		}

		msg := decodeResult.OK()
		pushResponse := agent.context.Push(*msg)
		if pushResponse.IsErr() {
			err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
			return cm.Err[cm.Result[agents.MessageShape, agents.Message, agents.Error]](err)
		}
		calledTool := false
		switch msg.Role {
		case types.RoleAssistant:
			for _, c := range msg.Content.Slice() {
				switch c.String() {
				case "tool-input":
					toolresult := agent.tools.Call(*c.ToolInput())
					if toolresult.IsErr() {
						err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
						return cm.Err[cm.Result[agents.MessageShape, agents.Message, agents.Error]](err)
					}
					calledTool = true
					// Push the tool output to the context and re-compute with the tool output
					agent.context.Push(agents.Message{Role: types.RoleTool, Content: cm.ToList([]types.Content{types.ContentToolOutput(*toolresult.OK())})})
				default:
					// If the content is not a tool input, we can just continue
					continue
				}
			}
		default:
			// the role should always be an assistant
			return cm.Err[cm.Result[agents.MessageShape, agents.Message, agents.Error]](agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError)))
		}
		if !calledTool {
			// overwrite the final message with the last message
			finalMsg = msg
			// assuming if the agent is not requesting a tool call, it is the final message
			break
		}
	}
	return cm.OK[cm.Result[agents.MessageShape, agents.Message, agents.Error]](*finalMsg)
}

func invokeStream(self cm.Rep, message agents.Message, writer agents.OutputStream) cm.Result[agents.Error, struct{}, agents.Error] {
	agent, ok := resourceTable.agents[self]
	if !ok {
		wasiErr := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
		return cm.Err[cm.Result[agents.Error, struct{}, agents.Error]](wasiErr)
	}

	result := agent.context.Push(message)
	if result.IsErr() {
		err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
		return cm.Err[cm.Result[agents.Error, struct{}, agents.Error]](err)
	}

	finalMsg := &types.Message{Role: types.RoleAssistant, Content: cm.ToList([]types.Content{types.ContentText(types.TextContent{
		Text: "agent yielded no response",
	})})}

	for i := 0; i <= maxturn; i++ {
		result := agent.context.Messages()
		if result.IsErr() {
			err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
			return cm.Err[cm.Result[agents.Error, struct{}, agents.Error]](err)
		}
		msgs := result.OK().Slice()

		encodedResult := agent.format.Encode(cm.ToList(msgs))
		if encodedResult.IsErr() {
			err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
			return cm.Err[cm.Result[agents.Error, struct{}, agents.Error]](err)
		}

		d := tensor.TensorDimensions(cm.ToList([]uint32{1}))
		td := tensor.TensorData(cm.ToList(encodedResult.OK().Slice()))
		t := tensor.NewTensor(d, tensor.TensorTypeU8, td)
		inputs := []inferencestream.NamedTensor{
			{
				F0: "user",
				F1: t,
			},
		}
		computeResult := agent.graph.Compute(cm.ToList(inputs))
		if computeResult.IsErr() {
			err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
			return cm.Err[cm.Result[agents.Error, struct{}, agents.Error]](err)
		}

		stream := computeResult.OK().F1
		ts := tensorStream(stream)
		// read the output from the stream
		text := make([]byte, 0)
		for {
			// Read up to 100 bytes from the output
			// to get any tokens that have been generated and push to socket
			p := make([]byte, 100)
			len, err := ts.Read(p)
			if len == 0 || err == io.EOF {
				break
			} else if err != nil {
				err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
				return cm.Err[cm.Result[agents.Error, struct{}, agents.Error]](err)
			}
			text = append(text, p[:len]...)

			// TODO:: Optionally write RAW output to the writer
			// this would result in data getting back to the client faster
			// additionally once the full message is read in, we will decode it
			// and write the full typed message.
			// For this to work cleanly, we need a new message content type, potentially role type as well.
		}

		decodeResult := agent.format.Decode(cm.ToList(text))
		if decodeResult.IsErr() {
			err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
			return cm.Err[cm.Result[agents.Error, struct{}, agents.Error]](err)
		}

		msg := decodeResult.OK()
		pushResponse := agent.context.Push(*msg)
		if pushResponse.IsErr() {
			err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
			return cm.Err[cm.Result[agents.Error, struct{}, agents.Error]](err)
		}
		calledTool := false
		switch msg.Role {
		case types.RoleAssistant:
			for _, c := range msg.Content.Slice() {
				switch c.String() {
				case "tool-input":
					toolresult := agent.tools.Call(*c.ToolInput())
					if toolresult.IsErr() {
						err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
						return cm.Err[cm.Result[agents.Error, struct{}, agents.Error]](err)
					}
					calledTool = true
					// Push the tool output to the context and re-compute with the tool output
					agent.context.Push(agents.Message{Role: types.RoleTool, Content: cm.ToList([]types.Content{types.ContentToolOutput(*toolresult.OK())})})
				default:
					// If the content is not a tool input, we can just continue
					continue
				}
			}
		default:
			// the role should always be an assistant
			err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
			return cm.Err[cm.Result[agents.Error, struct{}, agents.Error]](err)
		}
		if !calledTool {
			// overwrite the final message with the last message
			finalMsg = msg
			// Write full message to the output stream
			bytes, err := json.Marshal(finalMsg)
			if err != nil {
				err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
				return cm.Err[cm.Result[agents.Error, struct{}, agents.Error]](err)
			}
			result := writer.Write(cm.ToList(bytes))
			if result.IsErr() {
				err := agents.ErrorResourceNew(cm.Rep(agents.ErrorCodeInvokeError))
				return cm.Err[cm.Result[agents.Error, struct{}, agents.Error]](err)
			}
			// assuming if the agent is not requesting a tool call, it is the final message
			break
		}
	}

	return cm.OK[cm.Result[agents.Error, struct{}, agents.Error]](struct{}{})
}

func destructor(self cm.Rep) {
	delete(resourceTable.agents, self)
}

func main() {}
