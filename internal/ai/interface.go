package ai

import "context"

type Message struct {
	Role    string
	Content string
}

type AIClient interface {
	Chat(ctx context.Context, messages []Message) (string, error)
	ChatStream(ctx context.Context, messages []Message) (<-chan string, error)
}
