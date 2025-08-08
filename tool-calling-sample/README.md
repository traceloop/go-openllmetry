# Tool Calling Sample with OpenAI Go SDK

This sample demonstrates how to use tool calling with the official OpenAI Go SDK and the Traceloop Go OpenLLMetry SDK for comprehensive observability.

## Features

- **Tool Calling**: Demonstrates tool calling with weather function
- **Function Definitions**: Shows how to define tools with proper schemas
- **Tool Execution**: Implements local function execution for tool calls
- **Traceloop Integration**: Traces both request tools and response tool calls
- **Multi-turn Conversations**: Handles the complete tool calling flow

## Available Tools

1. **get_weather**: Get weather information for a location

## Setup

1. Set your environment variables:
   ```bash
   export OPENAI_API_KEY="your-openai-api-key"
   export TRACELOOP_API_KEY="your-traceloop-api-key"
   export TRACELOOP_BASE_URL="https://api.traceloop.com"  # Optional
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the sample:
   ```bash
   go run main.go
   ```

## What Gets Traced

The sample traces:

### Request Tools
- Tool function names, descriptions, and parameters
- Logged with `llm.request.functions.{i}.*` attributes

### Response Tool Calls  
- Tool call IDs, types, and function calls
- Logged with `llm.completions.{i}.tool_calls.{j}.*` attributes

### Complete Conversation Flow
- Initial user message
- Assistant response with tool calls
- Tool execution results
- Final assistant response

## Example Output

```
User: What's the weather like in San Francisco?