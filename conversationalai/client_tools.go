package conversationalai

import (
	"fmt"
	"sync"
)

// ToolHandler represents a function that handles a client tool call.
type ToolHandler func(parameters map[string]interface{}) (interface{}, error)

// ClientTools manages registration and execution of client-side tools.
type ClientTools struct {
	mu    sync.RWMutex
	tools map[string]ToolHandler
}

// NewClientTools creates a new ClientTools instance.
func NewClientTools() *ClientTools {
	return &ClientTools{
		tools: make(map[string]ToolHandler),
	}
}

// Register registers a new tool handler.
func (ct *ClientTools) Register(toolName string, handler ToolHandler) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if _, exists := ct.tools[toolName]; exists {
		return fmt.Errorf("tool '%s' is already registered", toolName)
	}

	ct.tools[toolName] = handler
	return nil
}

// Handle executes a registered tool with the given parameters.
func (ct *ClientTools) Handle(toolName string, parameters map[string]interface{}) (interface{}, error) {
	ct.mu.RLock()
	handler, exists := ct.tools[toolName]
	ct.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool '%s' is not registered", toolName)
	}

	return handler(parameters)
}
