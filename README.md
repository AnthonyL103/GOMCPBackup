# GoMCP

A Go-based Model Context Protocol (MCP) framework for building AI agents with multi-provider LLM support and dynamic tool integration.

## Features

- 🤖 **Multi-Provider Support**: Seamlessly switch between Anthropic Claude and OpenAI models
- 🔧 **Dynamic Tool Loading**: Tools are loaded from YAML configs at runtime
- 🌐 **MCP Server Architecture**: Extensible server system for custom tool implementations
- 🔄 **Automatic Tool Chaining**: Supports sequential and parallel tool execution
- 📝 **Conversation History**: Built-in chat management with tool call tracking
- 🎯 **Type-Safe Tool Schemas**: JSON Schema validation for tool inputs with recursive nested structures

## Architecture

```
┌─────────────┐
│   Agent     │ ← Loads instructions, tools, and servers from YAML
└──────┬──────┘
       │
       ├─────► Transport Layer (Anthropic/OpenAI providers)
       │       ├─── Builds tool schemas
       │       ├─── Formats messages per provider
       │       └─── Handles tool call loops
       │
       └─────► Registry (MCP Servers)
               ├─── Weather Server (port 3000)
               ├─── Schedule Server (port 8080)
               └─── Custom Servers...
```

### Provider Differences

| Feature | Anthropic | OpenAI |
|---------|-----------|--------|
| **Tool Execution** | Sequential (one at a time) | Parallel (multiple per turn) |
| **System Message** | Separate `system` field | First message with `role: system` |
| **Tool Format** | `{name, description, input_schema}` | `{type: "function", function: {...}}` |
| **Tool Results** | User message with `tool_result` | Message with `role: "tool"` |
| **Arguments** | Direct JSON object | JSON-encoded string |

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/GoMCP.git
cd GoMCP

# Install dependencies
go mod download

# Build the project
go build -o GOMCP.exe

# Run the agent
./GOMCP.exe
```

## Configuration

### Agent Configuration (`agentconfig.yaml`)

```yaml
agent_id: "my_agent"
description: "Your agent description"
instructions: |
  You are a helpful assistant.
  You have access to various tools.

llm:
  provider: "anthropic"  # or "openai"  
  model: "claude-sonnet-4-5-20250929"  # or "gpt-4o"
  api_key: "your-api-key"
  temperature: 0.7
  max_tokens: 2048

servers:
  - config_file: "serverconfigs/server1config.yaml"
  - config_file: "serverconfigs/server2config.yaml"
```

### Server Configuration

```yaml
server_id: "weather_server"
description: "Provides weather information"

runtime:
  type: "go"
  command: "go"
  args: ['run', 'C:\path\to\weather_server.go']
  port: 3000

tools:
  - tool_id: "get_weather"
    description: "Get current weather for a location"
    handler: "get_weather"
    
    input_schema:
      properties:
        location:
          type: "string"
          description: "City name"
        units:
          type: "string"
          description: "celsius or fahrenheit"
      
      required:
        - location
        - units
```

### Nested Schema Support

For complex schemas with arrays and nested objects:

```yaml
input_schema:
  properties:
    events:
      type: "array"
      description: "List of events"
      items:
        type: "object"
        properties:
          title:
            type: "string"
            description: "Event title"
          time:
            type: "string"
            description: "Event time"
        required:
          - title
          - time
```

## Project Structure

```
GoMCP/
├── main.go                    # Entry point
├── runagent.go               # Agent runner
├── agentconfig.yaml          # Agent configuration
│
├── Agent/
│   └── agent.go              # Agent core logic
│
├── chat/
│   └── chat.go               # Chat history management
│
├── transport/
│   ├── provider.go           # Provider interface
│   ├── anthropic.go          # Anthropic implementation
│   └── openai.go             # OpenAI implementation
│
├── protocol/
│   ├── llmprotocol/
│   │   ├── executer.go       # Tool execution
│   │   └── helper.go         # Tool extraction & formatting
│   ├── parseagentprotocol/
│   │   └── parseagentconfig.go
│   └── parseserverprotocol/
│       └── parseserverconfig.go
│
├── registry/
│   └── registry.go           # Server registry
│
├── server/
│   └── server.go             # MCP server abstraction
│
├── tool/
│   └── tool.go               # Tool definitions & validation
│
├── serverconfigs/
│   ├── server1config.yaml
│   └── server2config.yaml
│
└── examples/
    ├── weather_server.go
    ├── schedule_server.go
    └── microsoft_project_server.go
```

## Usage

### Interactive Mode

```bash
./GOMCP.exe
```

The agent will start and wait for user input. Type your messages and press Enter.

### Example Interactions

```
You: What's the weather in Paris?
Agent: [calls get_weather tool]
Agent: It's currently 15°C in Paris with clear skies.

You: Add an event to my schedule for tomorrow at 2pm
Agent: [calls add_event tool]
Agent: I've added the event to your schedule.
```

### Supported Models

**Anthropic:**
- `claude-opus-4-5-20251101`
- `claude-sonnet-4-5-20250929`
- `claude-haiku-4-5-20251001`

**OpenAI:**
- `gpt-4o`
- `gpt-4o-mini`
- `gpt-4-turbo`
- `o1-preview`
- `o1-mini`

## Adding a New Provider

1. Create a new file in `transport/` (e.g., `gemini.go`)
2. Implement the `Provider` interface:
   ```go
   type Provider interface {
       SendRequest(chat *chat.Chat, agent *agent.Agent, userMessage string) error
       GetProviderName() string
   }
   ```
3. Add model detection in `runagent.go`
4. Update `createProvider()` to instantiate your provider

## Adding a New MCP Server

1. Create a Go HTTP server (see `examples/weather_server.go`)
2. Implement endpoints matching your tool handlers:
   ```go
   http.HandleFunc("/execute/my_tool", handleMyTool)
   ```
3. Create a YAML config in `serverconfigs/`
4. Add the config to `agentconfig.yaml` servers list

## How It Works

### Tool Call Flow

1. **User Input** → Agent receives message
2. **Extract Tools** → Load available tools from all servers
3. **LLM Request** → Send to provider with tools array
4. **Parse Response** → Check for tool calls
5. **Execute Tools** → Call server endpoints with parameters
6. **Loop** → Continue if more tool calls needed (sequential or parallel)
7. **Final Response** → Save and display to user

### Message Storage

Chat history stores tool cycles as single messages:
```go
Message {
    Role: "assistant"
    Content: "Final response text"
    ToolCall: {ToolID, Parameters, ToolUseID}
    ToolResult: {Content, IsError, ToolUseID}
}
```

These expand to provider-specific formats when sending to APIs.

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o GOMCP.exe
```

## Roadmap

- [ ] Streaming support
- [ ] More provider integrations (Google Gemini, local models)
- [ ] Web UI for chat interface
- [ ] Tool generation from OpenAPI specs
- [ ] Persistent chat history (database storage)
- [ ] Multi-agent conversations
- [ ] Custom tool validation rules

### Tool Generation flow

1. **Tool1: GenerateServerCode** → Generates the server code, compiles it and validates the syntax
2. **Tool2: DeployeAndTestTools** → Deploy a test server and test it with agent generates params
3. **Tool3: DeployAndRegister** → Deploys the server for good and registers the tool in the GeneratedServersManager as a generated server, and registers the tool in the actual agent registrar
4. **Tool4: CleanupServerGeneration** → Gets rid of the files that are unnecessary and created during the generation process, leaves the binary for the working server.
5. **Tool5: DeleteServer** → Deletes generated servers that are unused. 


### Architecture Reasoning

This architecture was built after several iterations with using only one server gen tool, though that yielded little success as the agent would make mistakes giving no feedback to the LLM on where the process went wrong. 

You might be wondering how the process is enforced with having separate tools, each tool generation process in each instance of tool generation is stored. Each tool also has guards in place that prevent the tool at that process stage from running if the past one wasn't successful. This way the LLM gets feedback at every stage and we are validating the process in code instead of purely relying on prompt generation, and lets the LLM alter its params whether its handlercode, import params, input params, etc. per step in the process to fix its errors. 


## Credits

Built using Go and Model Context Protocol ideaology 
