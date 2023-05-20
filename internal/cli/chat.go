package cli

import (
	"blue/internal/chatcompletion"
	"blue/internal/editor"
	"blue/internal/model"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	openai "github.com/sashabaranov/go-openai"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
				editor := editor.NewEditor()
				content, err = editor.Edit()
				if err != nil {
					return err
				}
			} else if len(args) == 0 {
				i := ""
				for {
					m := model.NewModel(i)
					if _, err := tea.NewProgram(m).Run(); err != nil {
						return err
					}

					if m.UserInput.Input() == "" {
						break
					}

					if m.Ready() {
						chatCompletion := chatcompletion.NewChatCompletion()
						if err := chatCompletion.Message("user", m.UserInput.FilteredInput()); err != nil {
							return err
						}
						_, err := chatcompletion.StreamChatCompletion(cmd.Context(), c, chatCompletion.Request())
						if err != nil {
							return err
						}
						fmt.Println()
						m.Reset()
					} else {
						i = m.UserInput.FilteredInput()
					}
				}
				return nil
			} else {
				content = strings.Join(args, " ")
			}

			chatCompletion := chatcompletion.NewChatCompletion()
			if err := chatCompletion.Message("user", content); err != nil {
				return err
			}
			if _, err := chatcompletion.StreamChatCompletion(cmd.Context(), c, chatCompletion.Request()); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&flags.maxTokens, "max-tokens", 2048, "Max tokens")
	cmd.Flags().BoolVar(&flags.editor, "editor", false, "Editor")
	return cmd
}
