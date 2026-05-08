# ElevenLabs Go SDK (Conversational AI)

[![Go Report Card](https://goreportcard.com/badge/github.com/iyashjayesh/elevenlabs-go)](https://goreportcard.com/report/github.com/iyashjayesh/elevenlabs-go)
[![GoDoc](https://godoc.org/github.com/iyashjayesh/elevenlabs-go?status.svg)](https://pkg.go.dev/github.com/iyashjayesh/elevenlabs-go)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
![GitHub last commit](https://img.shields.io/github/last-commit/iyashjayesh/elevenlabs-go)
![Visitors](https://api.visitorbadge.io/api/visitors?path=iyashjayesh%2Felevenlabs-go%20&countColor=%23263759&style=flat)

Go SDK for interacting with the [ElevenLabs API](https://elevenlabs.io/), specifically focusing on the new **ElevenAgents (Conversational AI)** module.

<div align="center">
  <video autoplay muted loop playsinline width="100%">
    <source src="conversationalai_demo.mp4" type="video/mp4">
  </video>
</div>

## Features

- **Real-time WebSockets**: Connect to ElevenLabs Conversational AI agents via WebSocket.
- **Custom Client Tools**: Register Go functions that the agent can execute during conversations.
- **Audio Interfaces**: Easy abstractions (`AudioInterface`) for bringing your own audio input/output handlers.
- **Event Callbacks**: Handle transcripts, agent responses, audio alignments, and latency metrics via easy-to-use callbacks.

## Installation

```bash
go get github.com/iyashjayesh/elevenlabs-go
```

## Examples

The [`example/`](./example) directory ships with two runnable demos:

| Path                                       | What it does                                                                       |
| ------------------------------------------ | ---------------------------------------------------------------------------------- |
| [`example/text`](./example/text)           | Browser-based text chat. A small Go server proxies WebSocket messages between an HTML UI and the agent. |
| [`example/audio`](./example/audio)         | Browser-based voice chat. The browser captures the mic via Web Audio (with built-in echo cancellation), streams 16 kHz PCM over WebSocket to the Go server, and plays the agent's PCM audio back through the speakers. |

Both examples read `ELEVENLABS_AGENT_ID` and (optionally) `ELEVENLABS_API_KEY` from the environment.

```bash
export ELEVENLABS_AGENT_ID=...
export ELEVENLABS_API_KEY=...   # only required if the agent has auth enabled

# Text demo: open http://localhost:8080
cd example/text && go run .

# Voice demo: open http://localhost:8080 and tap the mic
# (the agent must be configured with pcm_16000 output in the ElevenLabs dashboard)
cd example/audio && go run .
```

The voice demo relies on the browser's built-in echo cancellation, so the laptop speaker → mic feedback loop is handled for you. The mic stream is downsampled to 16-bit mono PCM at 16 kHz inside the browser and shipped to the Go server as binary WebSocket frames, which the server simply forwards to and from the agent. Each example is its own Go module so they can be built independently.

## Quick Start

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
