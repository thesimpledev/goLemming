# GoLemming

A Windows CLI tool for spawning autonomous agents that can control your desktop on your behalf.

## Project Overview

GoLemming is an agentic automation tool that combines LLM reasoning with desktop control. Users describe a goal in natural language, and GoLemming spawns agents that observe the screen, decide on actions, and execute them using mouse and keyboard input.

The tool operates in two modes:
- **File operations**: Direct file system access for editing, creating, and managing files (similar to Claude Code)
- **GUI operations**: Screen capture and input simulation for interacting with applications that have no CLI or API

## Core Concepts

### Agent Loop

Each agent follows a simple reactive loop:

1. Capture current screen state
2. Send screenshot + goal + history to LLM
3. Receive action decision from LLM
4. Execute action (click, type, scroll, file edit, etc.)
5. Wait for UI stabilization
6. Repeat until goal achieved or failure

### Multi-Agent Support

GoLemming supports spawning multiple agents simultaneously. Use cases include:
- Parallel tasks across different applications
- Coordinated workflows where agents hand off to each other
- Redundant attempts at flaky operations

Agents run in isolated goroutines with their own state and history. A supervisor coordinates lifecycle and prevents conflicts (e.g., two agents fighting over the same window).

### Agent Verification

Agents can be configured to verify each other's work. This enables:
- **Checker agents**: A secondary agent reviews the output of a primary agent and flags issues
- **Approval workflows**: Work does not proceed until a verifier agent confirms the previous step
- **Adversarial review**: One agent attempts to find flaws in another agent's output

Verification modes:
- **Post-task review**: Checker agent examines final state after primary agent reports done
- **Step-by-step review**: Checker agent validates after each significant action
- **Parallel review**: Both agents work simultaneously and results are compared

Checker agents receive:
- The original goal
- The primary agent's action history
- Current screen state
- Any files created or modified

Checker agents return:
- Approval (work is correct)
- Rejection with specific issues
- Remediation actions (fixes the checker can apply directly)

## Architecture
```
golemming/
├── cmd/
│   └── golemming/
│       └── main.go              # CLI entry point
├── internal/
│   ├── agent/
│   │   ├── agent.go             # Core agent loop
│   │   ├── state.go             # Agent state management
│   │   ├── supervisor.go        # Multi-agent coordination
│   │   └── verifier.go          # Verification agent logic
│   ├── capture/
│   │   ├── screenshot.go        # Screen capture
│   │   └── diagnostics.go       # Secure desktop detection
│   ├── input/
│   │   ├── mouse.go             # Mouse control (existing)
│   │   ├── keyboard.go          # Keyboard control (existing)
│   │   └── safety.go            # User input detection, emergency stop
│   ├── llm/
│   │   ├── client.go            # LLM API client
│   │   ├── prompt.go            # System prompts and formatting
│   │   └── parser.go            # Action response parsing
│   ├── action/
│   │   ├── action.go            # Action type definitions
│   │   ├── executor.go          # Action execution dispatcher
│   │   └── file.go              # File operation actions
│   ├── cli/
│   │   ├── cli.go               # Interactive CLI interface
│   │   ├── commands.go          # Command handlers
│   │   └── display.go           # Output formatting
│   └── config/
│       └── config.go            # Configuration management
├── pkg/
│   └── protocol/
│       └── actions.go           # Shared action schema definitions
└── go.mod
```

## Action Schema

Actions returned by the LLM follow a JSON schema:
```json
{
  "type": "click",
  "x": 500,
  "y": 300,
  "button": "left",
  "double": false
}
```
```json
{
  "type": "type",
  "text": "hello world"
}
```
```json
{
  "type": "key",
  "key": "enter"
}
```
```json
{
  "type": "scroll",
  "direction": "down",
  "amount": 3
}
```
```json
{
  "type": "file_write",
  "path": "C:\\Users\\me\\document.txt",
  "content": "file contents here"
}
```
```json
{
  "type": "file_read",
  "path": "C:\\Users\\me\\document.txt"
}
```
```json
{
  "type": "wait",
  "ms": 1000
}
```
```json
{
  "type": "done",
  "summary": "Task completed successfully"
}
```
```json
{
  "type": "failed",
  "reason": "Could not locate the save button"
}
```

### Verification Actions

Checker agents have additional action types:
```json
{
  "type": "approve",
  "summary": "All requirements met"
}
```
```json
{
  "type": "reject",
  "issues": [
    "File missing required header",
    "Button click targeted wrong element"
  ]
}
```
```json
{
  "type": "remediate",
  "issues": ["Incorrect file permissions"],
  "actions": [
    {"type": "file_write", "path": "...", "content": "..."}
  ]
}
```

## Implementation Phases

### Phase 1: Foundation

- [ ] Set up project structure
- [ ] Implement screen capture using kbinani/screenshot (single screenshot on demand)
- [ ] Port input package from SimplyAuto
- [ ] Implement basic LLM client (Claude API)
- [ ] Define action schema and parser
- [ ] Build minimal agent loop (single agent, no persistence)

### Phase 2: CLI Interface

- [ ] Interactive REPL for issuing commands
- [ ] Goal input and confirmation
- [ ] Live status display during execution
- [ ] Action history display
- [ ] Emergency stop (hotkey or input detection)

### Phase 3: Multi-Agent

- [ ] Agent supervisor implementation
- [ ] Agent spawning and lifecycle management
- [ ] Conflict detection (multiple agents targeting same window)
- [ ] Agent communication/handoff protocol
- [ ] Resource limits (max concurrent agents)

### Phase 4: Agent Verification

- [ ] Verifier agent role and prompts
- [ ] Post-task verification workflow
- [ ] Step-by-step verification mode
- [ ] Approval/rejection/remediation handling
- [ ] Verification result reporting
- [ ] Configurable verification strictness levels
- [ ] Retry policies on rejection

### Phase 5: Safety and Reliability

- [ ] User input detection (pause when user is active)
- [ ] Secure desktop detection (pause on UAC, lock screen)
- [ ] Action confirmation for destructive operations
- [ ] Rollback support for file operations
- [ ] Rate limiting and cost tracking for LLM calls

### Phase 6: Quality of Life

- [ ] Configuration file support
- [ ] Multiple LLM provider support
- [ ] Task templates/presets
- [ ] Session recording and replay
- [ ] Verbose/debug logging modes

## Technical Decisions

### Screen Capture

Using `kbinani/screenshot` for cross-monitor support. Single capture on demand, not continuous recording. JPEG encoding for smaller payloads to the LLM.

### LLM Integration

Start with Claude API (vision capable). Action responses in JSON with strict schema. System prompt instructs the model on available actions and expected response format.

### Input Simulation

Using Windows SendInput API via syscall. Already proven in SimplyAuto. Virtual key codes for keyboard, absolute coordinates for mouse.

### Concurrency Model

Each agent is a goroutine with its own:
- Screen capture buffer
- Action history
- LLM conversation context
- Stop channel

Supervisor manages agent registry and coordinates shared resources.

### Verification Model

Verifier agents are standard agents with a different system prompt and additional action types. They receive a read-only view of the primary agent's history and current state. Verification can be:
- Mandatory (task blocks until approved)
- Advisory (issues logged but work continues)
- Auto-remediate (verifier fixes issues directly if possible)

## Open Questions

- Should agents share screen captures or capture independently?
- How to handle overlapping windows when multiple agents are active?
- What granularity of action confirmation is appropriate?
- Should there be a "dry run" mode that shows intended actions without executing?
- Can a verifier agent also be verified (chains of verification)?
- How to handle disagreements between verifier and primary agent?
- Should verification use the same LLM or a different one for independence?

## Dependencies

- `golang.org/x/sys/windows` - Windows API access
- `kbinani/screenshot` - Screen capture
- `anthropics/anthropic-sdk-go` or raw HTTP - LLM client
- TBD: CLI framework (or raw stdin/stdout)

## Links

- Website: https://golemming.dev
- Alternate: https://golemmings.dev (redirects to primary)
