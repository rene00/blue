package prompt

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type command struct {
	name string
}

type Input struct {
	role    string
	content string
}

type Prompt struct {
	ready    bool
	messages []openai.ChatCompletionMessage
	// The chat completion responses from openai.
	responses []openai.ChatCompletionStreamResponse
	inputs    []Input
	Opts
}

type Opts struct {
	// The  openai model.
	model string
	// The openai max tokens.
	maxTokens int
	// The openai stream.
	stream bool
}

type OptFunc func(*Opts)

func withMaxTokens(n int) OptFunc {
	return func(opts *Opts) {
		opts.maxTokens = n
	}
}

func defaultOpts() Opts {
	return Opts{
		model:     openai.GPT3Dot5Turbo,
		maxTokens: 2048,
		stream:    true,
	}
}

func NewPrompt(opts ...OptFunc) *Prompt {
	o := defaultOpts()
	for _, fn := range opts {
		fn(&o)
	}

	return &Prompt{
		Opts: o,
	}
}

// Ready returns a bool and indicates whether the prompt is ready to send.
func (p *Prompt) Ready() bool {
	return p.ready
}

func (p *Prompt) Message(role, content string) error {
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
	for _, message := range p.messages {
		existingHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", message.Role, message.Content)))
		if messageHash == existingHash {
			exists = true
		}
	}
	if !exists {
		p.messages = append(p.messages, openai.ChatCompletionMessage{Role: role, Content: content})
	}

	return nil
}

func (p *Prompt) removeCommandFromMessage(m openai.ChatCompletionMessage, idx int, command string) {
	// m.Content = strings.ReplaceAll(m.Content, match[0], "")
	m.Content = strings.ReplaceAll(m.Content, fmt.Sprintf("%s", command), "")

	// Delete original message which contains the command.
	p.messages = append(p.messages[:idx], p.messages[idx+1:]...)

	// Append the new message which is the original message without the command.
	p.messages = append(p.messages, m)
}

func (p *Prompt) Commands() error {

	re := regexp.MustCompile(`c:([a-z]{1,})`)

	ready := true

	// Update p.commands which is a list of commands to execute.
	for idx, m := range p.messages {
		match := re.FindStringSubmatch(m.Content)
		if len(match) == 2 {
			c := command{name: match[1]}
			switch strings.ToLower(c.name) {

			// c:print will print all messages. The messages will not be sent.
			case "print":
				p.removeCommandFromMessage(m, idx, match[0])
				for _, i := range p.messages {
					data, err := json.Marshal(i)
					if err != nil {
						return err
					}
					fmt.Println(string(data))
				}
				ready = false
				p.Reset()
			default:
				return fmt.Errorf("command %s unsupported", c.name)
			}

		}
	}

	p.ready = ready

	return nil
}

func (p *Prompt) Reset() {
	p.messages = []openai.ChatCompletionMessage{}
}

func (p *Prompt) Inputs() {
	re := regexp.MustCompile(`i:(.*):(.*)`)
	for _, m := range p.messages {
		match := re.FindStringSubmatch(m.Content)
		if len(match) == 3 {
			input := Input{role: match[1], content: match[2]}
			p.inputs = append(p.inputs, input)

			// remove input from message
			substr := fmt.Sprintf("i:%s:%s", match[1], match[2])
			m.Content = strings.ReplaceAll(m.Content, substr, "")
		}
	}

	for _, i := range p.inputs {
		p.Message(i.role, i.content)
	}
}

func (p *Prompt) ChatCompletion() openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model:     p.Opts.model,
		MaxTokens: p.Opts.maxTokens,
		Messages:  p.messages,
		Stream:    p.Opts.stream,
	}

}

// Response accepts a openai chat stream response.
func (p *Prompt) Response(r openai.ChatCompletionStreamResponse) {
	p.responses = append(p.responses, r)
}
