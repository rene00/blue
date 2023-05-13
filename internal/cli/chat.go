package cli

import (
	"blue/internal/editor"
	"blue/internal/prompt"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

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

func StreamChatCompletion(ctx context.Context, c *openai.Client, req openai.ChatCompletionRequest) (openai.ChatCompletionStreamResponse, error) {
	var err error
	chatResponse := openai.ChatCompletionStreamResponse{}

	/*
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
	*/

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
			pmptOpts := []prompt.OptFunc{}
			pmptOpts = append(pmptOpts, prompt.WithTextAreaModel())
			pmpt := prompt.NewPrompt(pmptOpts...)

			if len(args) == 0 && flags.editor {
				editor := editor.NewEditor()
				content, err = editor.Edit()
				if err != nil {
					return err
				}
			} else if len(args) == 0 {
				for {

					if _, err := pmpt.Run(); err != nil {
						return err
					}

					content = pmpt.Input()
					if content == "" {
						break
					}

					pmpt.Message("user", strings.TrimRight(content, "\n"))
					if err = pmpt.Commands(); err != nil {
						fmt.Println(err)
					}

					if pmpt.Ready() {
						resp, err := StreamChatCompletion(cmd.Context(), c, pmpt.ChatCompletion())
						if err != nil {
							return err
						}
						pmpt.Response(resp)
						pmpt.Reset()
					}

					fmt.Scanln()
				}
				return nil
			} else {
				content = strings.Join(args, " ")
			}

			pmpt.Message("user", content)
			resp, err := StreamChatCompletion(cmd.Context(), c, pmpt.ChatCompletion())
			if err != nil {
				return err
			}
			pmpt.Response(resp)

			return nil
		},
	}

	cmd.Flags().IntVar(&flags.maxTokens, "max-tokens", 2048, "Max tokens")
	cmd.Flags().BoolVar(&flags.editor, "editor", false, "Editor")
	return cmd
}
