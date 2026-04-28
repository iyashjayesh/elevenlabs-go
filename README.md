# ElevenLabs Go SDK (Conversational AI)

[![Go Report Card](https://goreportcard.com/badge/github.com/iyashjayesh/elevenlabs-go)](https://goreportcard.com/report/github.com/iyashjayesh/elevenlabs-go)
[![GoDoc](https://godoc.org/github.com/iyashjayesh/elevenlabs-go?status.svg)](https://pkg.go.dev/github.com/iyashjayesh/elevenlabs-go)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
![GitHub last commit](https://img.shields.io/github/last-commit/iyashjayesh/elevenlabs-go)
![Visitors](https://api.visitorbadge.io/api/visitors?path=iyashjayesh%2Felevenlabs-go%20&countColor=%23263759&style=flat)

Go SDK for interacting with the [ElevenLabs API](https://elevenlabs.io/), specifically focusing on the new **ElevenAgents (Conversational AI)** module.

## Features

- **Real-time WebSockets**: Connect to ElevenLabs Conversational AI agents via WebSocket.
- **Custom Client Tools**: Register Go functions that the agent can execute during conversations.
- **Audio Interfaces**: Easy abstractions (`AudioInterface`) for bringing your own audio input/output handlers.
- **Event Callbacks**: Handle transcripts, agent responses, audio alignments, and latency metrics via easy-to-use callbacks.

## Installation

```bash
go get github.com/iyashjayesh/elevenlabs-go
```

## Quick Start

See the `example/main.go` file for a fully working text-based example of spinning up a conversational agent and registering a custom tool.

```go
package main

import (
	"log"
	"github.com/iyashjayesh/elevenlabs-go/conversationalai"
)

func main() {
    agentID := "YOUR_AGENT_ID"
    apiKey := "YOUR_API_KEY" // Optional if not requires_auth

    // Register a tool
    tools := conversationalai.NewClientTools()
    tools.Register("get_weather", func(params map[string]interface{}) (interface{}, error) {
        return "It's sunny and 72 degrees.", nil
    })

    // Listen to callbacks
    callbacks := conversationalai.Callbacks{
        OnAgentResponse: func(resp string) {
            log.Printf("Agent: %s\n", resp)
        },
    }

    config := &conversationalai.ConversationConfig{ AgentID: agentID }

    conversation := conversationalai.NewConversation(
        agentID, apiKey, nil, tools, callbacks, config,
    )

    if err := conversation.StartSession(); err != nil {
        log.Fatal(err)
    }

    conversation.SendUserMessage("Hello, what's the weather?")
    conversation.Wait()
}
```

## License

MIT License. See `LICENSE` for more information.
