package ai

// ToolDef defines a function-calling tool for the LLM.
type ToolDef struct {
	Type     string       `json:"type"` // always "function"
	Function ToolFunction `json:"function"`
}

// ToolFunction describes a tool's schema.
type ToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"` // JSON schema object
}

// ToolCall represents a tool invocation returned by the LLM.
type ToolCall struct {
	ID       string       `json:"id"`
	Index    int          `json:"index"`
	Type     string       `json:"type"` // always "function"
	Function ToolFunctionCall `json:"function"`
}

// ToolFunctionCall is the function name + accumulated arguments.
type ToolFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// StreamCallbacks holds all possible streaming callbacks.
// Any nil callback is simply not invoked.
type StreamCallbacks struct {
	// OnContent is called for text content deltas.
	OnContent func(text string)

	// OnReasoning is called for thinking/reasoning deltas (if the model supports it).
	OnReasoning func(text string)

	// OnToolCall is called when a tool call starts or accumulates arguments.
	// The ToolCall contains the accumulated state (name + partial args).
	OnToolCall func(tc ToolCall)

	// OnFinish is called when the stream ends with a finish_reason.
	// Possible values: "stop", "tool_calls", "length", etc.
	OnFinish func(reason string)
}
