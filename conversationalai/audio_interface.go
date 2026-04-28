package conversationalai

// InputCallback is called regularly with input audio chunks from the user.
// The audio should be in 16-bit PCM mono format at 16kHz. Recommended
// chunk size is 4000 samples (250 milliseconds).
type InputCallback func(audio []byte) error

// AudioInterface provides an abstraction for handling audio input and output.
type AudioInterface interface {
	// Start begins the audio interface.
	// It is called one time before the conversation starts.
	// Implementations should capture audio and send it using the callback.
	Start(callback InputCallback) error

	// Stop cleans up any resources used by the audio interface
	// and stops any audio streams.
	Stop() error

	// Output plays audio received from the user/agent.
	// The audio input is in 16-bit PCM mono format at 16kHz.
	// Implementations can choose to do additional buffering.
	// This method should return quickly and not block.
	Output(audio []byte) error

	// Interrupt is a signal to stop any audio output immediately.
	// It is called when the user has interrupted the agent.
	Interrupt() error
}
