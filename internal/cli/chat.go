package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	openai "github.com/sashabaranov/go-openai"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const maxRetries = 10
const baseDelay = 10 * time.Second
const maxDelay = 60 * time.Second

func getEBO(retries int) time.Duration {
	delay := float64(baseDelay) * math.Pow(2, float64(retries))
	jitter := rand.Float64() * 0.1 * delay
	delayWithJitter := time.Duration(delay+jitter) % maxDelay
	return delayWithJitter
}

type Model struct {
	textarea textarea.Model
	err      error
	Input    string
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

type errMsg error

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyTab:
			m.Input = m.textarea.Value()
			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return fmt.Sprintf(m.textarea.View()) + "\n\n"
}

func (m Model) value() string {
	return m.Input
}

func initialModel() Model {
	ti := textarea.New()
	ti.Placeholder = "press [tab] to submit prompt"
	ti.SetWidth(72)
	ti.SetHeight(3)
	ti.ShowLineNumbers = false
	ti.Focus()
	return Model{
		textarea: ti,
		err:      nil,
	}
}

func StreamChatCompletion(ctx context.Context, c *openai.Client, content string) error {
	var err error
	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 2048,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: content,
			},
		},
		Stream: true,
	}

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
			return err
		}
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println()
			return nil
		}
		if err != nil {
			fmt.Printf("\nstream error: %v\n", err)
			return err
		}

		fmt.Printf(response.Choices[0].Delta.Content)
	}

	return nil
}

func contentWithVim() (string, error) {
	content := ""
	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return content, err
	}
	defer os.Remove(tmpFile.Name())

	c := exec.Command("vim", tmpFile.Name())
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	err = c.Run()
	if err != nil {
		return content, err
	}

	f, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return content, err
	}

	content = string(f)

	return content, nil

}

func chatCmd(cli *cli) *cobra.Command {

	var flags struct {
		maxTokens int
		model     string
		editor    bool
	}

	var cmd = &cobra.Command{
		Use: "chat",
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = viper.BindPFlag("max-tokens", cmd.Flags().Lookup("max-tokens"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var content string
			var err error

			apiKey := os.Getenv("OPENAI_API_KEY")

			if apiKey == "" {
				return fmt.Errorf("set OPENAI_API_KEY")
			}

			c := openai.NewClient(apiKey)

			if len(args) == 0 && flags.editor {
				content, err = contentWithVim()
				if err != nil {
					return err
				}
			} else if len(args) == 0 {
				for {
					p := tea.NewProgram(initialModel())
					m, err := p.Run()
					if err != nil {
						return err
					}
					a, _ := m.(Model)
					content = strings.ReplaceAll(a.Input, "\n", " ")
					if content == "" {
						break
					}
					if err := StreamChatCompletion(cmd.Context(), c, content); err != nil {
						return err
					}
					fmt.Scanln()
				}
				return nil
			} else {
				content = strings.Join(args, " ")
			}
			if err := StreamChatCompletion(cmd.Context(), c, content); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&flags.maxTokens, "max-tokens", 2048, "Max tokens")
	cmd.Flags().BoolVar(&flags.editor, "editor", false, "Editor")
	return cmd
}
