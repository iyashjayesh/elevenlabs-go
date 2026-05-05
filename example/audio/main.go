// Browser-based voice chat with an ElevenLabs Conversational AI agent.
//
// A small Go HTTP server hosts index.html and exposes a WebSocket endpoint at
// /ws. The browser captures the microphone, downsamples to 16-bit PCM mono at
// 16 kHz and streams it as binary WS frames. This server forwards those bytes
// to the agent and pushes the agent's audio responses back as binary frames
// for the browser to play.
//
// Required environment:
//
//	ELEVENLABS_AGENT_ID  ID of the agent to talk to
//	ELEVENLABS_API_KEY   (optional) only when the agent requires auth
//
// The agent must be configured with `pcm_16000` (16 kHz, 16-bit, mono) audio
// output in the ElevenLabs dashboard. The default `mp3_44100_128` will not
// play correctly because this example pipes raw PCM straight to Web Audio.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/iyashjayesh/elevenlabs-go/conversationalai"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	agentID := os.Getenv("ELEVENLABS_AGENT_ID")
	if agentID == "" {
		log.Fatal("ELEVENLABS_AGENT_ID is not set")
	}
	apiKey := os.Getenv("ELEVENLABS_API_KEY")

	tools := conversationalai.NewClientTools()
	if err := tools.Register("get_weather", func(params map[string]interface{}) (interface{}, error) {
		location, _ := params["location"].(string)
		log.Printf("[tool] get_weather(%q)", location)
		return fmt.Sprintf("The weather in %s is sunny and 72 degrees.", location), nil
	}); err != nil {
		log.Fatalf("register tool: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleConversation(w, r, agentID, apiKey, tools)
	})

	log.Println("Server listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func handleConversation(w http.ResponseWriter, r *http.Request, agentID, apiKey string, tools *conversationalai.ClientTools) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer ws.Close()
	log.Println("Frontend connected.")

	audio := NewBrowserAudio(ws)
	defer audio.Stop()

	send := func(msg BrowserMessage) {
		_ = audio.WriteJSON(msg)
	}

	callbacks := conversationalai.Callbacks{
		OnAgentResponse: func(text string) {
			log.Printf("Agent: %s", text)
			send(BrowserMessage{Type: "agent_response", Text: text})
		},
		OnUserTranscript: func(text string) {
			log.Printf("User: %s", text)
			audio.MarkUserHasSpoken()
			send(BrowserMessage{Type: "user_transcript", Text: text})
		},
		OnLatencyMeasurement: func(ms int) {
			send(BrowserMessage{Type: "latency", Text: fmt.Sprintf("%d ms", ms)})
		},
		OnEndSession: func() {
			log.Println("ElevenLabs session ended.")
			send(BrowserMessage{Type: "status", Text: "Session ended"})
		},
	}

	conv := conversationalai.NewConversation(
		agentID, apiKey, audio, tools, callbacks,
		&conversationalai.ConversationConfig{AgentID: agentID},
	)

	if err := conv.StartSession(); err != nil {
		log.Println("start session:", err)
		send(BrowserMessage{Type: "status", Text: "Failed to start ElevenLabs session"})
		return
	}
	defer conv.EndSession()

	send(BrowserMessage{Type: "status", Text: "Connected to Agent"})

	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("frontend ws closed:", err)
			return
		}
		switch mt {
		case websocket.BinaryMessage:
			audio.PushUserAudio(message)
		case websocket.TextMessage:
			var msg BrowserMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Println("invalid frontend message:", err)
				continue
			}
			if msg.Type == "user_message" && msg.Text != "" {
				log.Printf("User (text): %s", msg.Text)
				audio.MarkUserHasSpoken()
				conv.SendUserMessage(msg.Text)
			}
		}
	}
}
