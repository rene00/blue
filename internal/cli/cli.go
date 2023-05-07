package cli

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
)

type cli struct {
	debug      bool
	configFile string
	config     config
}

type config struct {
	ChatGPTAPIURL string `json:"chatgpt_api_url"`
}

func (c *cli) init() error {
	if c.configFile == "" {
		return nil
	}

	var buf []byte
	var err error
	if buf, err = ioutil.ReadFile(c.configFile); err != nil {
		return err
	}

	if err := json.Unmarshal(buf, &c.config); err != nil {
		return err
	}

	return nil
}

func Execute() {
	cli := &cli{}
	rootCmd := buildRootCmd(cli)
	rootCmd.AddCommand(chatCmd(cli))
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.ExecuteContext(context.TODO()); err != nil {
		os.Exit(1)
	}
}

func buildRootCmd(cli *cli) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "blue",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return cli.init()
		},
	}
	rootCmd.PersistentFlags().BoolVar(&cli.debug, "debug", false, "debug")
	rootCmd.PersistentFlags().StringVar(&cli.configFile, "config-file", "", "config file path")

	return rootCmd
}
