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

	"github.com/pterm/pterm"
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
					content, err := pterm.DefaultInteractiveTextInput.WithMultiLine().Show()
					if err != nil {
						return err
					}
					if content == "c:editor" {
						text, err := contentWithVim()
						if err != nil {
							return err
						}
						content = text
					}
					if err := StreamChatCompletion(cmd.Context(), c, content); err != nil {
						return err
					}
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
