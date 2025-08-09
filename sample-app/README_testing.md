# Testing Tool Calling Without API Keys

This directory contains testing for tool calling functionality that works in CI environments without requiring actual API keys.

## Testing Approach

### VCR Recording (`tool_calling_test.go`)

Uses `go-vcr` to record real API interactions and replay them in tests. **API keys are automatically sanitized** from recordings.

```bash
# First run with real API key to record (local only)
OPENAI_API_KEY=your_key go test -v -run TestToolCallingWithMock

# Subsequent runs use recorded cassettes (works in CI)
go test -v -run TestToolCallingWithMock
```

**Security Features:**
- 🔒 Authorization headers are automatically removed from cassettes
- 🔒 Request bodies are sanitized to prevent accidental key leakage
- 🔒 Cassettes are gitignored by default for extra safety

**Pros:**
- ✅ Uses real API responses
- ✅ Accurate representation of actual data  
- ✅ Works in CI without API keys (after recording)
- ✅ Automatic sanitization of sensitive data

**Cons:**
- ❌ Requires initial recording with real API key
- ❌ Additional dependency

### Integration Tests (Optional)

Real API calls for full integration testing.

```bash
# Run integration tests (requires API keys)
OPENAI_API_KEY=your_key INTEGRATION_TEST=1 go test -v -run TestToolCallingIntegration
```

## CI Configuration

For GitHub Actions or other CI systems:

```yaml
- name: Run Tests
  run: |
    # Run VCR tests with pre-sanitized cassettes (no API keys needed)
    go test -v -run TestToolCallingWithMock
    
    # Skip integration tests in CI (or use secrets for API keys)
```

## Mock Data Structure

The VCR cassette contains realistic OpenAI API responses:

```json
{
  "choices": [{
    "message": {
      "role": "assistant",
      "tool_calls": [{
        "id": "call_YkIfypBQrmpUpxsKuS9aNdKg",
        "type": "function", 
        "function": {
          "name": "get_weather",
          "arguments": "{\"location\":\"San Francisco, CA\"}"
        }
      }]
    }
  }],
  "usage": {
    "prompt_tokens": 82,
    "completion_tokens": 17,
    "total_tokens": 99
  }
}
```

This ensures our tracing code gets realistic data to work with and validates that all span attributes are set correctly.