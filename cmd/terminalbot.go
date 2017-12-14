package cmd

import (
	"strings"

	"github.com/pantheon-systems/riker/pkg/chat/terminalbot"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

func init() {
	terminalCmd.PersistentFlags().String(
		"nickname",
		"some-nickname",
		"The nickname to send messages as",
	)

	terminalCmd.PersistentFlags().String(
		"groups",
		"infra,eng",
		"The groups that the sending user belongs",
	)
	// binding flags to viper
	terminalCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		viper.BindPFlag(f.Name, f)
	})

}

func startTerminalBot(cmd *cobra.Command, args []string) {
	nickname := viper.GetString("nickname")
	groups := strings.Split(viper.GetString("groups"), ",")
	b := terminalbot.New(viper.GetString("bind-address"), nickname, groups)
	b.Run()
}
