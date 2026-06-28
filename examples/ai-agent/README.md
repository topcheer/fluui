# Fluui AI Agent Example

A full-featured AI Agent demonstrating all major Fluui capabilities.

## Features

- **Streaming AI Chat** — Real-time LLM streaming with markdown rendering
- **StatusBar** — Live model, token rate, and context window metrics
- **TabBar** — Multi-session management (3 sessions)
- **Selection** — Mouse drag text selection with OSC52 clipboard copy
- **Thinking Blocks** — AI reasoning visualization
- **Tool Calls** — Function calling with result display

## Setup

Create a `.env` file in the project root:

```bash
FLUUI_LLM_API_KEY=your-api-key
FLUUI_LLM_BASE_URL=https://api.example.com/v1
FLUUI_LLM_MODEL=your-model-name
FLUUI_LLM_SYSTEM_PROMPT=You are a helpful assistant.
```

## Running

```bash
cp .env.example .env
# Edit .env with your API credentials
go run ./examples/ai-agent/
```

## Key Bindings

| Key | Action |
|-----|--------|
| Enter | Send message |
| Up/Down | Scroll history / Input history |
| Ctrl+C | Stop streaming (or quit if idle) |
| Ctrl+T | Switch to next tab |
| Alt+1/2/3 | Switch to tab by index |
| Ctrl+Shift+C | Copy selection to clipboard |
| Esc | Quit |

## Architecture

```
┌─────────────────────────────────────┐
│ TabBar: Session 1 │ Session 2 │ ... │
├─────────────────────────────────────┤
│                                     │
│  ChatApp (streaming AI conversation) │
│  - Thinking blocks                  │
│  - Tool calls / results             │
│  - Markdown rendering               │
│  - Diff highlighting                │
│                                     │
├─────────────────────────────────────┤
│ StatusBar: ● model │ tokens │ clock  │
└─────────────────────────────────────┘
```

This example demonstrates the full Fluui component stack:
- `app.ChatApp` — High-level chat orchestration
- `component.StatusBar` — AI metrics display
- `component.TabBar` — Session management
- `app.SelectionManager` — Text selection
- `ai.Client` — OpenAI-compatible streaming
