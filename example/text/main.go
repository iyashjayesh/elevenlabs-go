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
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Message from/to frontend WS
type FsMessage struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

func main() {
	agentID := os.Getenv("ELEVENLABS_AGENT_ID")
	if agentID == "" {
		log.Fatal("ELEVENLABS_AGENT_ID is not set")
	}

	apiKey := os.Getenv("ELEVENLABS_API_KEY")

	// Setup Client Tools
	tools := conversationalai.NewClientTools()
	tools.Register("get_weather", func(params map[string]interface{}) (interface{}, error) {
		location, _ := params["location"].(string)
		log.Printf("Agent requested weather for location: %s", location)
		return fmt.Sprintf("The weather in %s is sunny and 72 degrees.", location), nil
	})

	// Serve the static HTML
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Handle WebSocket connections
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		defer ws.Close()

		log.Println("Frontend connected.")

		// Helper to send messages to frontend
		sendToFrontend := func(msgType, text string) {
			msg := FsMessage{Type: msgType, Text: text}
			if err := ws.WriteJSON(msg); err != nil {
				log.Println("Frontend WS write error:", err)
			}
		}

		callbacks := conversationalai.Callbacks{
			OnAgentResponse: func(response string) {
				log.Printf("Agent: %s", response)
				sendToFrontend("agent_response", response)
			},
			OnUserTranscript: func(transcript string) {
				// We don't send this back to frontend because the frontend optimistically adds the user's message
				log.Printf("User Transcript: %s", transcript)
			},
			OnLatencyMeasurement: func(latencyMs int) {
				sendToFrontend("latency", fmt.Sprintf("%d ms", latencyMs))
			},
			OnEndSession: func() {
				log.Println("ElevenLabs session ended.")
			},
		}

		config := &conversationalai.ConversationConfig{
			AgentID: agentID,
		}

		conversation := conversationalai.NewConversation(
			agentID,
			apiKey,
			nil, // Text-only for this demo
			tools,
			callbacks,
			config,
		)

		if err := conversation.StartSession(); err != nil {
			log.Println("Failed to start ElevenLabs session:", err)
			return
		}
		defer conversation.EndSession()

		// Tell frontend we are connected
		sendToFrontend("agent_response", "Hi there! I am connected and ready to chat. (ElevenLabs Agent)")

		// Read loop from frontend WS
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Println("Frontend WS read error or closed:", err)
				break
			}

			var fsMsg FsMessage
			if err := json.Unmarshal(message, &fsMsg); err != nil {
				log.Println("Invalid JSON from frontend:", err)
				continue
			}

			if fsMsg.Type == "user_message" {
				log.Printf("User: %s", fsMsg.Text)
				// Send to ElevenLabs
				conversation.SendUserMessage(fsMsg.Text)
			}
		}
	})

	log.Println("Server listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
