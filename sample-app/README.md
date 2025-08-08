# Sample Apps

This directory contains sample applications demonstrating the Traceloop Go OpenLLMetry SDK.

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