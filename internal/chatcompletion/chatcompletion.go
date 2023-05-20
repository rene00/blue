package chatcompletion

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"time"

	"github.com/sashabaranov/go-openai"
)

const maxRetries = 10
const baseDelay = 10 * time.Second
const maxDelay = 60 * time.Second

type ChatCompletion struct {
	model     string
	maxTokens int
	stream    bool
	messages  []openai.ChatCompletionMessage
}

type OptFunc func(*ChatCompletion)

func WithMaxTokens(n int) OptFunc {
	return func(c *ChatCompletion) {
		c.maxTokens = n
	}
}

func WithChatCompletionMessages(m []openai.ChatCompletionMessage) OptFunc {
	return func(c *ChatCompletion) {
		c.messages = m
	}
}

func NewChatCompletion(opts ...OptFunc) ChatCompletion {
	c := ChatCompletion{
		model:     openai.GPT3Dot5Turbo,
		maxTokens: 2048,
		stream:    true,
		messages:  []openai.ChatCompletionMessage{},
	}

	for _, fn := range opts {
		fn(&c)
	}

	return c
}

func (c *ChatCompletion) Message(role, content string) error {
	switch role {
	case openai.ChatMessageRoleUser, openai.ChatMessageRoleSystem, openai.ChatMessageRoleAssistant:
	default:
		return fmt.Errorf("role unsupported: %s", role)
	}

	if content == "" {
		return fmt.Errorf("content cant be empty")
	}

	messageHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", role, content)))
	exists := false
	for _, message := range c.messages {
		existingHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", message.Role, message.Content)))
		if messageHash == existingHash {
			exists = true
		}
	}
	if !exists {
		c.messages = append(c.messages, openai.ChatCompletionMessage{Role: role, Content: content})
	}

	return nil
}

func (c ChatCompletion) Request() openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Stream:    c.stream,
		Messages:  c.messages,
	}
}

func (c ChatCompletion) Messages() []openai.ChatCompletionMessage {
	return c.messages
}

func getEBO(retries int) time.Duration {
	delay := float64(baseDelay) * math.Pow(2, float64(retries))
	jitter := rand.Float64() * 0.1 * delay
	delayWithJitter := time.Duration(delay+jitter) % maxDelay
	return delayWithJitter
}

func StreamChatCompletion(ctx context.Context, c *openai.Client, req openai.ChatCompletionRequest) (openai.ChatCompletionStreamResponse, error) {
	var err error
	chatResponse := openai.ChatCompletionStreamResponse{}

	e := &openai.APIError{}
	stream := &openai.ChatCompletionStream{}
	for retries := 0; retries < maxRetries; retries++ {
		stream, err = c.CreateChatCompletionStream(ctx, req)
		if err == nil {
			break
		}
		if err != nil {
			if errors.As(err, &e) {
				switch e.HTTPStatusCode {
				case 429:
					delay := getEBO(retries)
					time.Sleep(delay)
					continue
				}
			}
			return chatResponse, err
		}
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println()
			return chatResponse, nil
		}
		if err != nil {
			fmt.Printf("\nstream error: %v\n", err)
			return chatResponse, err
		}

		fmt.Printf(response.Choices[0].Delta.Content)
	}

	return chatResponse, nil
}
