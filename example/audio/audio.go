package main

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/iyashjayesh/elevenlabs-go/conversationalai"
)

// BrowserAudio implements conversationalai.AudioInterface by bridging the
// agent's audio I/O over a single browser WebSocket connection.
//
//   - Capture: the browser sends raw 16-bit PCM mono frames at 16 kHz as
//     binary WebSocket messages. The HTTP handler hands those bytes to
//     PushUserAudio which forwards them to the SDK input callback.
//   - Playback: agent audio is forwarded as binary WebSocket frames; the
//     browser plays them through Web Audio API.
//   - Interrupt: a JSON {"type":"interrupt"} message tells the browser to
//     drop any queued playback so the agent can be cut off cleanly.
//
// All writes are serialized through writeMu because gorilla/websocket only
// allows one concurrent writer per connection.
type BrowserAudio struct {
	ws *websocket.Conn

	writeMu  sync.Mutex
	callback conversationalai.InputCallback

	once sync.Once
	done chan struct{}
}

// BrowserMessage is the JSON envelope used for control & text messages
// between the Go server and the browser tab.
type BrowserMessage struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

func NewBrowserAudio(ws *websocket.Conn) *BrowserAudio {
	return &BrowserAudio{ws: ws, done: make(chan struct{})}
}

func (b *BrowserAudio) Start(cb conversationalai.InputCallback) error {
	b.callback = cb
	return nil
}

func (b *BrowserAudio) Stop() error {
	b.once.Do(func() { close(b.done) })
	return nil
}

func (b *BrowserAudio) Output(audio []byte) error {
	return b.writeRaw(websocket.BinaryMessage, audio)
}

func (b *BrowserAudio) Interrupt() error {
	return b.WriteJSON(BrowserMessage{Type: "interrupt"})
}

// PushUserAudio forwards a chunk of mic PCM from the browser to the agent.
func (b *BrowserAudio) PushUserAudio(audio []byte) {
	if b.callback == nil {
		return
	}
	select {
	case <-b.done:
		return
	default:
	}
	_ = b.callback(audio)
}

// WriteJSON sends a control/status message to the browser. Safe to call from
// any goroutine.
func (b *BrowserAudio) WriteJSON(v interface{}) error {
	b.writeMu.Lock()
	defer b.writeMu.Unlock()
	select {
	case <-b.done:
		return nil
	default:
	}
	return b.ws.WriteJSON(v)
}

func (b *BrowserAudio) writeRaw(messageType int, p []byte) error {
	b.writeMu.Lock()
	defer b.writeMu.Unlock()
	select {
	case <-b.done:
		return nil
	default:
	}
	return b.ws.WriteMessage(messageType, p)
}
