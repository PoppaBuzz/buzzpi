# BuzzAI

**The AI assistant that lets users manage their devices through natural language.** BuzzAI is an optional subsystem within the BuzzPi Runtime and client apps that interprets natural language requests, selects appropriate tools, and executes actions on the device. It follows a tool-use pattern similar to Anthropic's Model Context Protocol (MCP).

---

## Principles

1. **Opt-in by default** — BuzzAI is disabled until the user explicitly enables it. No AI features run without consent.
2. **Local-first processing** — Intent classification and tool selection run on-device where possible. Cloud AI is used only for complex queries that exceed local capability.
3. **Tool-grounded execution** — The AI does not generate arbitrary commands. It selects from a defined set of tools (BPP methods), each with strict parameter schemas.
4. **User confirmation for destructive actions** — Any tool that can modify system state (reboot, uninstall, write files) requires explicit user confirmation before execution.
5. **Transparent reasoning** — The AI reveals its reasoning: "I'm going to check disk usage, then recommend which large files to remove." Users can inspect, approve, or reject each step.
6. **No training on user data** — Queries and device data are never used to train AI models. Telemetry is off by default.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Client App (Android)                       │
│                                                               │
│  ┌──────────────┐  ┌──────────────────┐  ┌──────────────┐   │
│  │ Chat UI      │  │ Conversation     │  │ Tool Result  │   │
│  │ (Compose)    │◄─┤ State Manager    │──┤ Renderer     │   │
│  └──────┬───────┘  └────────┬─────────┘  └──────────────┘   │
│         │                    │                                │
│         │          ┌─────────▼─────────┐                     │
│         │          │ Intent Classifier  │                     │
│         │          │ (local: ONNX)     │                     │
│         │          └─────────┬─────────┘                     │
│         │                    │                                │
└─────────┼────────────────────┼────────────────────────────────┘
          │                    │
          │  BPP: buzzai.*     │ BPP: tool.execute
          │                    │
┌─────────▼────────────────────▼────────────────────────────────┐
│                    BuzzPi Runtime                              │
│                                                               │
│  ┌───────────────────────────────────────────────────────┐    │
│  │                 AI Service                              │    │
│  │                                                         │    │
│  │  ┌──────────┐  ┌───────────┐  ┌──────────────────┐    │    │
│  │  │ Tool     │  │ Prompt    │  │ Response         │    │    │
│  │  │ Registry │  │ Builder   │  │ Generator        │    │    │
│  │  └────┬─────┘  └─────┬─────┘  └────────┬─────────┘    │    │
│  │       │              │                  │              │    │
│  │  ┌────▼──────────────▼──────────────────▼─────────┐   │    │
│  │  │           LLM Router                            │   │    │
│  │  │  (local: llama.cpp / remote: OpenAI API)         │   │    │
│  │  └──────────────────────┬──────────────────────────┘   │    │
│  └─────────────────────────┼──────────────────────────────┘    │
│                            │                                    │
│  ┌─────────────────────────▼──────────────────────────────┐    │
│  │              Core Runtime Services                      │    │
│  │  Terminal │ Screen │ Files │ Docker │ GPIO │ Stats...   │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

---

## Tool System

BuzzAI exposes device management capabilities as tools. Each tool maps to one or more BPP methods with strict input/output schemas.

### Built-in Tools

```yaml
tools:
  - name: read_file
    description: Read contents of a file on the device
    parameters:
      path: string    # Absolute path to file
    returns: string   # File contents (truncated at 64KB)

  - name: write_file
    description: Write content to a file (requires confirmation)
    parameters:
      path: string    # Absolute path
      content: string # Content to write
      append: boolean # If true, append instead of overwrite
    confirm_required: true

  - name: execute_command
    description: Run a shell command and get output
    parameters:
      command: string    # Command to execute
      timeout_seconds: int # Max execution time (default: 30)
    confirm_required: false
    constraints:
      - No interactive commands
      - No commands that modify system packages (apt, dpkg)
      - Timeout enforced at runtime

  - name: list_directory
    description: List files and directories
    parameters:
      path: string    # Directory path
      recursive: boolean # Recursive listing (default: false)
    returns: string[]  # File/directory names

  - name: system_status
    description: Get current system status
    parameters: {}
    returns:
      cpu_percent: float
      memory_mb: int
      temperature_c: float
      disk_available_mb: int
      uptime_seconds: int

  - name: docker_ps
    description: List running Docker containers
    parameters:
      all: boolean  # Include stopped containers
    returns: string  # Formatted container list

  - name: docker_logs
    description: Get logs from a Docker container
    parameters:
      container_id: string
      tail: int     # Number of lines (default: 50)
    returns: string  # Log output

  - name: journalctl
    description: Query systemd journal logs
    parameters:
      unit: string      # Service unit name (optional)
      priority: int     # 0=emerg .. 7=debug (default: 3)
      tail: int         # Number of lines (default: 50)
    returns: string     # Log output

  - name: disk_usage
    description: Analyze disk usage
    parameters:
      path: string      # Starting path (default: /)
      depth: int        # Directory depth (default: 1)
    returns: string     # Formatted disk usage report
```

### Tool Registry

```go
// ToolRegistry manages all available AI tools.
type ToolRegistry struct {
    tools map[string]Tool
}

type Tool struct {
    Name        string        `json:"name"`
    Description string        `json:"description"`
    Parameters  ToolParams    `json:"parameters"`
    ConfirmRequired bool      `json:"confirm_required"`
    Handler     ToolHandler   `json:"-"`
}

type ToolParams struct {
    Type       string                  `json:"type"` // "object"
    Properties map[string]ToolProperty `json:"properties"`
    Required   []string                `json:"required"`
}

type ToolProperty struct {
    Type        string      `json:"type"`
    Description string      `json:"description"`
    Default     interface{} `json:"default,omitempty"`
    Enum        []string    `json:"enum,omitempty"`
}

type ToolHandler func(ctx context.Context, params map[string]interface{}) (interface{}, error)

func (tr *ToolRegistry) Register(tool Tool) {
    tr.tools[tool.Name] = tool
}

func (tr *ToolRegistry) Execute(ctx context.Context, name string, params map[string]interface{}) (interface[], error) {
    tool, exists := tr.tools[name]
    if !exists {
        return nil, fmt.Errorf("unknown tool: %s", name)
    }

    // Validate parameters against schema
    if err := tr.validateParams(tool.Parameters, params); err != nil {
        return nil, err
    }

    // Check if confirmation required
    if tool.ConfirmRequired {
        // Return a confirmation request instead of executing
        return &ConfirmationRequest{
            Tool:   name,
            Params: params,
        }, nil
    }

    return tool.Handler(ctx, params)
}
```

---

## Conversation Flow

```
User: "My Pi is running slow, what's going on?"

                             ┌─────────────────────────┐
                             │  1. Receive message       │
                             │  intent: troubleshoot     │
                             └──────────┬──────────────┘
                                        │
                             ┌──────────▼──────────────┐
                             │  2. Build prompt          │
                             │  system + conversation    │
                             │  + available tools        │
                             └──────────┬──────────────┘
                                        │
                             ┌──────────▼──────────────┐
                             │  3. LLM processes         │
                             │  Decides: check system    │
                             │  status + disk usage      │
                             └──────────┬──────────────┘
                                        │
                      ┌─────────────────┼─────────────────┐
                      │                                   │
           ┌──────────▼──────────┐            ┌──────────▼──────────┐
           │  4a. Tool call:      │            │  4b. Tool call:      │
           │  system_status()     │            │  disk_usage("/")     │
           └──────────┬──────────┘            └──────────┬──────────┘
                      │                                   │
                      └──────────┬───────────────────────┘
                                 │
                      ┌──────────▼──────────┐
                      │  5. LLM synthesizes  │
                      │  "Your Pi is at 75°C │
                      │  and disk is 92%     │
                      │  full. Let me help   │
                      │  you clean up."      │
                      └──────────┬──────────┘
                                 │
                      ┌──────────▼──────────┐
                      │  6. Send response    │
                      │  + suggested next    │
                      │  actions             │
                      └─────────────────────┘
```

### BPP Methods

```yaml
buzzai.ask:
  description: Send a natural language request to BuzzAI
  request:
    message: string
    conversation_id: string (optional, for continuation)
    context: object (optional, current device context)
  response:
    response: string          # Natural language response
    conversation_id: string
    actions: Action[]         # Suggested actions (for confirmation UI)
    requires_confirmation: boolean

buzzai.confirm:
  description: Confirm or reject a pending action
  request:
    conversation_id: string
    action_id: string
    confirmed: boolean
  response:
    response: string          # Updated response after confirmation
    status: "executed" | "rejected" | "pending"

buzzai.tools.list:
  description: List available AI tools (for custom clients)
  request: {}
  response:
    tools: Tool[]
```

---

## LLM Integration

### LLM Router

```go
type LLMRouter struct {
    localProvider *LocalLLM       // llama.cpp or similar
    remoteProvider *RemoteLLM     // OpenAI API or similar
    config       LLMRouterConfig
}

type LLMRouterConfig struct {
    LocalModelPath     string  // Path to GGUF model
    RemoteAPIKey       string  // API key for remote provider
    RemoteEndpoint     string  // API endpoint URL
    ComplexityThreshold float64 // Route to remote if complexity exceeds threshold
    MaxLocalTokens      int    // 4096
}

func (r *LLMRouter) Complete(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
    if r.useLocal(req) {
        return r.localProvider.Complete(ctx, req)
    }
    return r.remoteProvider.Complete(ctx, req)
}

func (r *LLMRouter) useLocal(req LLMRequest) bool {
    // Use local LLM when:
    // 1. Remote API key not configured
    // 2. Network unavailable
    // 3. Query complexity is below threshold (simple questions)
    // 4. User explicitly requested local-only mode

    if r.remoteProvider == nil {
        return true
    }

    complexity := estimateComplexity(req.Messages)
    return complexity < r.config.ComplexityThreshold
}
```

### Prompt Template

```
You are BuzzAI, an AI assistant embedded in the BuzzPi Runtime.
You help users manage their Raspberry Pi device.

Available tools:
{{tool_descriptions}}

Rules:
1. Always explain what you're going to do before doing it.
2. Use tools to gather information before making recommendations.
3. For destructive actions (write files, reboot, package changes),
   set confirm_required=true and explain why.
4. If you don't know something, say so. Don't guess.
5. Format responses clearly with sections for readability.
6. When analyzing problems, check: CPU, memory, disk, temperature, uptime.
7. Keep responses concise but complete.

Conversation history:
{{conversation_history}}

User: {{user_message}}
Assistant:
```

---

## Local LLM Requirements

| Model | Size | RAM Required | Performance | Recommended For |
|-------|------|-------------|-------------|-----------------|
| Llama 3.2 1B Q4 | ~700MB | 1GB | Fast | Pi 5 (8GB) |
| Phi-3 Mini Q4 | ~2GB | 2.5GB | Moderate | Pi 5 (8GB) |
| TinyLlama Q4 | ~1GB | 1.5GB | Fast | Pi 4 (4GB) |
| Gemma 2 2B Q4 | ~1.3GB | 2GB | Moderate | Pi 5 (8GB) |

### Minimum Requirements for Local AI

```yaml
buzzai:
  enabled: false  # Disabled by default
  local_model: false  # Cloud-only by default

  local:
    model_path: "/var/lib/buzzpi/ai/models/llama-3.2-1b-q4.gguf"
    context_length: 4096
    max_tokens: 512
    threads: 4
    gpu_layers: 0  # No GPU acceleration on Pi

  cloud:
    provider: "openai"
    endpoint: "https://api.openai.com/v1"
    model: "gpt-4o-mini"    # Fast, cheap, good for device mgmt
    max_tokens: 1024
    temperature: 0.3         # Low temperature for deterministic tool use

  # Privacy mode: never send device data to cloud
  privacy_mode: false

  # Tools that require confirmation
  confirm_tools:
    - write_file
    - execute_command  # Only for high-risk commands
    - docker_restart
    - system_reboot
    - system_shutdown
    - package_install
```

---

## Tool Execution Sandbox

```go
// ToolSandbox enforces execution constraints on AI-triggered commands.
type ToolSandbox struct {
    timeout       time.Duration
    maxOutputSize int
    blockedPrefixes []string
}

func (s *ToolSandbox) ExecuteCommand(ctx context.Context, command string) (string, error) {
    // Check blocked commands
    for _, prefix := range s.blockedPrefixes {
        if strings.HasPrefix(command, prefix) {
            return "", fmt.Errorf("command blocked for safety: %s", prefix)
        }
    }

    // Apply timeout
    ctx, cancel := context.WithTimeout(ctx, s.timeout)
    defer cancel()

    cmd := exec.CommandContext(ctx, "bash", "-c", command)

    // Run with reduced privileges (nobody user)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Credential: &syscall.Credential{
            Uid: 65534,  // nobody
            Gid: 65534,
        },
    }

    output, err := cmd.Output()
    if len(output) > s.maxOutputSize {
        output = output[:s.maxOutputSize]
    }

    return string(output), err
}
```

### Disallowed Commands

```go
var dangerousPrefixes = []string{
    "rm -rf /",
    "dd if=",
    "mkfs.",
    "fdisk",
    "> /dev/",
    "chmod 777 /",
    "apt remove",
    "dpkg --purge",
    ":(){ :|:& };:",         // Fork bomb
    "wget", "curl",          // No external downloads
    "sudo",                  // No privilege escalation
    "reboot", "shutdown",    // Use specific tool for system control
    "passwd",                // No password changes
}
```

---

## Response Safety

```go
// SafetyFilter checks AI responses for dangerous content before sending to user.
type SafetyFilter struct {
    blockedPatterns []regexp.Regexp
}

func (f *SafetyFilter) FilterResponse(response string) (string, bool) {
    for _, pattern := range f.blockedPatterns {
        if pattern.MatchString(response) {
            return "I cannot provide that information.", false
        }
    }

    return response, true
}
```

---

## Conversation Persistence

```go
type ConversationStore struct {
    db  *sql.DB
    ttl time.Duration // 24 hours — conversations auto-expire
}

type Conversation struct {
    ID        string
    Messages  []Message
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Message struct {
    Role    string // "user" | "assistant" | "tool"
    Content string
    ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

func (cs *ConversationStore) Save(msg Message, convID string) error {
    // Append message to conversation history in SQLite
    // Expire conversations older than ttl
}

func (cs *ConversationStore) Load(convID string) (*Conversation, error) {
    // Load full conversation history
    // Apply context window limits (trim oldest messages if too long)
}
```

---

## Client Integration (Android)

```kotlin
@Composable
fun BuzzAiScreen(viewModel: BuzzAiViewModel) {
    val messages by viewModel.messages.collectAsState()
    val input by viewModel.input.collectAsState()
    val isProcessing by viewModel.isProcessing.collectAsState()

    Column {
        // Conversation history
        LazyColumn(modifier = Modifier.weight(1f)) {
            items(messages) { message ->
                when (message.role) {
                    "user" -> UserMessageBubble(message.content)
                    "assistant" -> AiMessageBubble(
                        content = message.content,
                        actions = message.suggestedActions,
                        onConfirm = { viewModel.confirmAction(it) },
                        onReject = { viewModel.rejectAction(it) }
                    )
                    "tool" -> ToolCallBubble(message.content)
                }
            }
        }

        // Loading indicator
        if (isProcessing) {
            LinearProgressIndicator(modifier = Modifier.fillMaxWidth())
        }

        // Input area
        Row(modifier = Modifier.padding(8.dp)) {
            OutlinedTextField(
                value = input,
                onValueChange = viewModel::setInput,
                modifier = Modifier.weight(1f),
                placeholder = { Text("Ask about your device...") },
                enabled = !isProcessing
            )
            IconButton(
                onClick = { viewModel.send() },
                enabled = input.isNotBlank() && !isProcessing
            ) {
                Icon(Icons.Default.Send, contentDescription = "Send")
            }
        }
    }
}
```

---

## Testing Strategy

| Test | Scenario | Expectation |
|------|----------|-------------|
| Tool invocation | "Check disk usage" | Calls disk_usage tool, returns results |
| Multi-step reasoning | "Why is my Pi slow?" | Calls system_status, disk_usage, returns analysis |
| Confirmation flow | "Delete large log file" | Returns confirmation request, executes on confirm |
| Destructive prevention | "Delete everything" | Safety filter blocks dangerous commands |
| Conversation context | "What about temperature?" (after asking about CPU) | Understands context from previous messages |
| Local-only mode | No network available | Falls back to local LLM (if configured) |
| Tool error handling | Tool returns error | AI explains error, suggests alternative |
| Privacy mode | Privacy mode enabled | All requests processed locally, no cloud calls |
