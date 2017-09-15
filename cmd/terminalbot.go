package cmd

import (
	"github.com/pantheon-systems/riker/pkg/chat/terminalbot"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var terminalCmd = &cobra.Command{
	Use:   "terminal",
	Short: "Start Terminal chat mode",
	Long: `Startup riker using terminal driver for testing

	Examples:
   riker terminal
`,
	Run: startTerminalBot,
}

func startTerminalBot(cmd *cobra.Command, args []string) {
	b := terminalbot.New(viper.GetString("bind-address"))
	b.Run()
}
