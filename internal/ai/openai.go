package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client *openai.Client
	model  string
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

func (c *OpenAIClient) Chat(ctx context.Context, messages []Message) (string, error) {
	apiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		apiMessages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	resp, apiErr := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: apiMessages,
	})
	if apiErr != nil {
		return "", fmt.Errorf("openai chat completion failed: %w", apiErr)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai returned empty response")
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) ChatStream(ctx context.Context, messages []Message) (<-chan string, error) {
	apiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		apiMessages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	stream, streamErr := c.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: apiMessages,
		Stream:   true,
	})
	if streamErr != nil {
		return nil, fmt.Errorf("openai chat stream creation failed: %w", streamErr)
	}

	ch := make(chan string)
	go func() {
		defer func() {
			if closeErr := stream.Close(); closeErr != nil {
				slog.Error("Failed to close openai stream", "error", closeErr)
			}
			close(ch)
		}()

		for {
			response, recvErr := stream.Recv()
			if errors.Is(recvErr, io.EOF) {
				return
			}
			if recvErr != nil {
				return
			}

			if len(response.Choices) > 0 {
				ch <- response.Choices[0].Delta.Content
			}
		}
	}()

	return ch, nil
}
