# Sample Apps

This directory contains sample applications demonstrating the Traceloop Go OpenLLMetry SDK.

## OpenAI SDKs

This sample app includes two different OpenAI Go SDKs for demonstration purposes:

- **Sashabaranov SDK** (`github.com/sashabaranov/go-openai`) - Used in `main.go` and workflow examples
- **Official OpenAI SDK** (`github.com/openai/openai-go`) - Used in `tool_calling.go`

## Regular Sample

Run the regular sample that demonstrates basic prompt logging:

```bash
go run .
```

## Tool Calling Sample

Run the tool calling sample that demonstrates tool calling with the OpenAI Go SDK:

```bash
go run . tool-calling
```

### Environment Variables

Set the following environment variables:

```bash
export OPENAI_API_KEY="your-openai-api-key"
export TRACELOOP_API_KEY="your-traceloop-api-key"
export TRACELOOP_BASE_URL="https://api.traceloop.com"  # Optional
```

### Tool Calling Features

The tool calling sample demonstrates:
- Request tools logging with function definitions
- Response tool calls logging with execution results
- Multi-turn conversations with tool execution
- Complete traceability of tool calling interactions