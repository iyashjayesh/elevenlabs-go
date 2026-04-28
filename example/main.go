package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/iyashjayesh/elevenlabs-go/conversationalai"
)

func main() {
	agentID := os.Getenv("ELEVENLABS_AGENT_ID")
	if agentID == "" {
		log.Fatal("ELEVENLABS_AGENT_ID is not set")
	}

	apiKey := os.Getenv("ELEVENLABS_API_KEY")

	// Set up client tools
	tools := conversationalai.NewClientTools()
	tools.Register("get_weather", func(params map[string]interface{}) (interface{}, error) {
		location, _ := params["location"].(string)
		log.Printf("Tool get_weather called for location: %s", location)
		return fmt.Sprintf("The weather in %s is sunny and 72 degrees.", location), nil
	})

	// Set up callbacks
	callbacks := conversationalai.Callbacks{
		OnAgentResponse: func(response string) {
			log.Printf("Agent: %s", response)
		},
		OnUserTranscript: func(transcript string) {
			log.Printf("User: %s", transcript)
		},
		OnLatencyMeasurement: func(latencyMs int) {
			log.Printf("Latency: %d ms", latencyMs)
		},
		OnEndSession: func() {
			log.Println("Session ended.")
		},
	}

	config := &conversationalai.ConversationConfig{
		AgentID: agentID,
	}

	// We are running text-only for this simple example (no audio interface)
	conversation := conversationalai.NewConversation(
		agentID,
		apiKey,
		nil, // AudioInterface is nil for text-only
		tools,
		callbacks,
		config,
	)

	log.Println("Starting conversation session...")
	if err := conversation.StartSession(); err != nil {
		log.Fatalf("Failed to start session: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Interrupt received, ending session...")
		conversation.EndSession()
	}()

	// Since we are text-only, let's send a quick text message to start things off
	conversation.SendUserMessage("Hello, who are you and what is the weather in London?")

	// Wait for the conversation to finish
	conversation.Wait()
	log.Println("Exiting.")
}
