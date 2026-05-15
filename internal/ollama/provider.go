package ollama

import (
	"context"
	"io"
)

// AIProvider is an interface for AI service providers
type AIProvider interface {
	// Generate generates a rewrite of the given text
	Generate(ctx context.Context, text, style, systemPrompt string) (string, error)
	// GenerateStream generates a rewrite and streams the response
	GenerateStream(ctx context.Context, text, style, systemPrompt string) (<-chan string, error)
	// GetName returns the provider name
	GetName() string
	// HealthCheck checks if the provider is available
	HealthCheck() error
	// GetVersion returns the provider version
	GetVersion() string
}


// StreamGenerator is a function that generates streaming responses
type StreamGenerator func(ctx context.Context, text, style, systemPrompt string) (<-chan string, error)

// ReadStream reads from an io.Reader and sends chunks to a channel
func ReadStream(reader io.Reader, outputChan chan<- string) {
	defer close(outputChan)
	
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				outputChan <- "[Stream Error: " + err.Error() + "]"
			}
			break
		}
		if n > 0 {
			outputChan <- string(buf[:n])
		}
	}
}
