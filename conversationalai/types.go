package conversationalai

type AgentChatResponsePartType string

const (
	AgentChatResponsePartTypeStart AgentChatResponsePartType = "start"
	AgentChatResponsePartTypeDelta AgentChatResponsePartType = "delta"
	AgentChatResponsePartTypeStop  AgentChatResponsePartType = "stop"
)

// Client events
type ClientEvent struct {
	Type string `json:"type"`
}

type UserMessageEvent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type UserActivityEvent struct {
	Type string `json:"type"`
}

type ContextualUpdateEvent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ClientToolResultEvent struct {
	Type       string      `json:"type"`
	ToolCallID string      `json:"tool_call_id"`
	Result     interface{} `json:"result"`
	IsError    bool        `json:"is_error"`
}

type AudioChunkEvent struct {
	UserAudioChunk string `json:"user_audio_chunk"`
}

type ConversationInitiationClientData struct {
	Type                         string                 `json:"type"`
	CustomLLMExtraBody           map[string]interface{} `json:"custom_llm_extra_body,omitempty"`
	ConversationConfigOverride   map[string]interface{} `json:"conversation_config_override,omitempty"`
	DynamicVariables             map[string]interface{} `json:"dynamic_variables,omitempty"`
	SourceInfo                   SourceInfo             `json:"source_info"`
	UserID                       string                 `json:"user_id,omitempty"`
}

type SourceInfo struct {
	Source  string `json:"source"`
	Version string `json:"version"`
}

// Server Events
type ServerEvent struct {
	Type string `json:"type"`
}

type AudioEventAlignment struct {
	Chars             []string `json:"chars"`
	CharStartTimesMs  []int    `json:"char_start_times_ms"`
	CharDurationsMs   []int    `json:"char_durations_ms"`
}

type AudioEvent struct {
	AudioBase64 string               `json:"audio_base_64"`
	EventID     string               `json:"event_id"`
	Alignment   *AudioEventAlignment `json:"alignment,omitempty"`
}

type AgentResponseEvent struct {
	AgentResponse string `json:"agent_response"`
}

type TextResponsePart struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type AgentChatResponsePartEvent struct {
	TextResponsePart TextResponsePart `json:"text_response_part"`
}

type UserTranscriptionEvent struct {
	UserTranscript string `json:"user_transcript"`
}

type InterruptionEvent struct {
	EventID string `json:"event_id"`
}

type PingEvent struct {
	PingMs *int `json:"ping_ms,omitempty"`
}

type ClientToolCallEvent struct {
	ToolCallID string                 `json:"tool_call_id"`
	ToolName   string                 `json:"tool_name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// Wrapping struct for raw message decoding
type RawServerMessage struct {
	Type                               string                         `json:"type"`
	ConversationInitiationMetadataEvent map[string]interface{}       `json:"conversation_initiation_metadata_event,omitempty"`
	AudioEvent                         *AudioEvent                    `json:"audio_event,omitempty"`
	AgentResponseEvent                 *AgentResponseEvent            `json:"agent_response_event,omitempty"`
	TextResponsePart                   *TextResponsePart              `json:"text_response_part,omitempty"`
	UserTranscriptionEvent             *UserTranscriptionEvent        `json:"user_transcription_event,omitempty"`
	InterruptionEvent                  *InterruptionEvent             `json:"interruption_event,omitempty"`
	PingEvent                          *PingEvent                     `json:"ping_event,omitempty"`
	ClientToolCall                     *ClientToolCallEvent           `json:"client_tool_call,omitempty"`
}

// Config for initiating the conversation
type ConversationConfig struct {
	AgentID                    string
	UserID                     string
	RequiresAuth               bool
	Environment                string
	ExtraBody                  map[string]interface{}
	ConversationConfigOverride map[string]interface{}
	DynamicVariables           map[string]interface{}
}
