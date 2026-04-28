package conversationalai

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

// Callbacks for the conversation events
type Callbacks struct {
	OnAgentResponse           func(response string)
	OnAgentResponseCorrection func(original, corrected string)
	OnAgentChatResponsePart   func(text string, partType AgentChatResponsePartType)
	OnUserTranscript          func(transcript string)
	OnLatencyMeasurement      func(latencyMs int)
	OnAudioAlignment          func(alignment *AudioEventAlignment)
	OnEndSession              func()
}

type Conversation struct {
	agentID         string
	userID          string
	requiresAuth    bool
	config          *ConversationConfig
	clientTools     *ClientTools
	audioInterface  AudioInterface
	callbacks       Callbacks
	apiKey          string

	conn     *websocket.Conn
	done     chan struct{}
	endOnce  sync.Once
	writeCh  chan interface{}
	
	lastInterruptID int64
}

// NewConversation creates a new conversational session.
func NewConversation(
	agentID string,
	apiKey string,
	audioInterface AudioInterface,
	clientTools *ClientTools,
	callbacks Callbacks,
	config *ConversationConfig,
) *Conversation {
	if config == nil {
		config = &ConversationConfig{AgentID: agentID}
	}
	if clientTools == nil {
		clientTools = NewClientTools()
	}

	return &Conversation{
		agentID:        agentID,
		apiKey:         apiKey,
		audioInterface: audioInterface,
		clientTools:    clientTools,
		callbacks:      callbacks,
		config:         config,
		done:           make(chan struct{}),
		writeCh:        make(chan interface{}, 100),
	}
}

// StartSession connects to the websocket and starts the conversation.
func (c *Conversation) StartSession() error {
	wsURL := c.getWSSUrl()

	headers := http.Header{}
	if c.apiKey != "" {
		headers.Add("xi-api-key", c.apiKey)
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %w", err)
	}
	c.conn = conn

	// Send initiation data
	initData := ConversationInitiationClientData{
		Type:                       "conversation_initiation_client_data",
		CustomLLMExtraBody:         c.config.ExtraBody,
		ConversationConfigOverride: c.config.ConversationConfigOverride,
		DynamicVariables:           c.config.DynamicVariables,

		UserID: c.config.UserID,
	}

	if err := c.conn.WriteJSON(initData); err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to send initiation data: %w", err)
	}

	// Start Audio interface if present
	if c.audioInterface != nil {
		err := c.audioInterface.Start(func(audio []byte) error {
			// Convert audio to base64
			encoded := base64.StdEncoding.EncodeToString(audio)
			chunk := AudioChunkEvent{UserAudioChunk: encoded}
			select {
			case c.writeCh <- chunk:
			case <-c.done:
			}
			return nil
		})
		if err != nil {
			c.conn.Close()
			return fmt.Errorf("failed to start audio interface: %w", err)
		}
	}

	go c.readPump()
	go c.writePump()

	return nil
}

// EndSession stops the conversation. It is safe to call from multiple
// goroutines and multiple times; only the first call has an effect.
func (c *Conversation) EndSession() {
	c.endOnce.Do(func() {
		close(c.done)
		if c.audioInterface != nil {
			c.audioInterface.Stop()
		}
		if c.callbacks.OnEndSession != nil {
			c.callbacks.OnEndSession()
		}
	})
}

// Wait blocks until the session is finished.
func (c *Conversation) Wait() {
	<-c.done
}

// SendUserMessage sends a text message from user to the agent.
func (c *Conversation) SendUserMessage(text string) {
	msg := UserMessageEvent{
		Type: "user_message",
		Text: text,
	}
	c.writeCh <- msg
}

func (c *Conversation) getWSSUrl() string {
	u := url.URL{
		Scheme:   "wss",
		Host:     "api.elevenlabs.io",
		Path:     "/v1/convai/conversation",
		RawQuery: fmt.Sprintf("agent_id=%s", c.agentID),
	}
	if c.config.Environment != "" {
		u.RawQuery += "&environment=" + c.config.Environment
	}
	return u.String()
}

func (c *Conversation) readPump() {
	defer c.EndSession()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("websocket read error: %v", err)
			break
		}

		var raw RawServerMessage
		if err := json.Unmarshal(message, &raw); err != nil {
			log.Printf("error decoding json: %v, payload: %s", err, string(message))
			continue
		}

		c.handleServerMessage(raw)
	}
}

func (c *Conversation) writePump() {
	defer c.conn.Close()
	for {
		select {
		case msg := <-c.writeCh:
			if err := c.conn.WriteJSON(msg); err != nil {
				log.Printf("websocket write error: %v", err)
				return
			}
		case <-c.done:
			c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
	}
}

func (c *Conversation) handleServerMessage(raw RawServerMessage) {
	switch raw.Type {
	case "audio":
		if raw.AudioEvent != nil {
			// Check interruption ID if event is too old
			// For simplicity, converting eventID string to int is needed, skipping for now
			audioBytes, err := base64.StdEncoding.DecodeString(raw.AudioEvent.AudioBase64)
			if err == nil && c.audioInterface != nil {
				c.audioInterface.Output(audioBytes)
			}
			if c.callbacks.OnAudioAlignment != nil && raw.AudioEvent.Alignment != nil {
				c.callbacks.OnAudioAlignment(raw.AudioEvent.Alignment)
			}
		}

	case "agent_response":
		if raw.AgentResponseEvent != nil && c.callbacks.OnAgentResponse != nil {
			c.callbacks.OnAgentResponse(raw.AgentResponseEvent.AgentResponse)
		}

	case "agent_chat_response_part":
		if raw.TextResponsePart != nil && c.callbacks.OnAgentChatResponsePart != nil {
			c.callbacks.OnAgentChatResponsePart(raw.TextResponsePart.Text, AgentChatResponsePartType(raw.TextResponsePart.Type))
		}

	case "user_transcript":
		if raw.UserTranscriptionEvent != nil && c.callbacks.OnUserTranscript != nil {
			c.callbacks.OnUserTranscript(raw.UserTranscriptionEvent.UserTranscript)
		}

	case "interruption":
		if raw.InterruptionEvent != nil {
			atomic.StoreInt64(&c.lastInterruptID, int64(raw.InterruptionEvent.EventID))
			if c.audioInterface != nil {
				c.audioInterface.Interrupt()
			}
		}

	case "ping":
		if raw.PingEvent != nil {
			// Reply with pong
			c.writeCh <- PongEvent{Type: "pong", EventID: raw.PingEvent.EventID}
			if raw.PingEvent.PingMs != nil && c.callbacks.OnLatencyMeasurement != nil {
				c.callbacks.OnLatencyMeasurement(*raw.PingEvent.PingMs)
			}
		}

	case "client_tool_call":
		if raw.ClientToolCall != nil {
			go c.executeClientTool(*raw.ClientToolCall)
		}
	}
}

func (c *Conversation) executeClientTool(call ClientToolCallEvent) {
	result, err := c.clientTools.Handle(call.ToolName, call.Parameters)
	
	resp := ClientToolResultEvent{
		Type:       "client_tool_result",
		ToolCallID: call.ToolCallID,
		Result:     result,
		IsError:    err != nil,
	}

	if err != nil {
		resp.Result = err.Error()
	}

	// Send result back
	c.writeCh <- resp
}
