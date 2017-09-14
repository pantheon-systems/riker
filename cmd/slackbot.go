package cmd

import (
	"fmt"

	"github.com/pantheon-systems/riker/pkg/chat/slackbot"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var slackCmd = &cobra.Command{
	Use:   "slackbot",
	Short: "Start Slack Chat mode",
	Long: `Startup riker using the slack chat driver and options

	Examples:
   riker slackbot
`,
	RunE: startSlackBot,
}

func startSlackBot(cmd *cobra.Command, args []string) error {

	for _, v := range []string{"bot-token", "api-token", "tls-cert"} {
		if viper.GetString(v) == "" {
			return fmt.Errorf("'%s' must be specified", v)
		}
	}

	b, err := slackbot.New(
		viper.GetString("bind-address"),
		viper.GetString("bot-token"),
		viper.GetString("api-token"),
		viper.GetString("tls-cert"),
		viper.GetString("ca-cert"),
		log,
	)
	if err != nil {
		return err
	}

	b.Run()
	return nil
}

func init() {
	slackCmd.PersistentFlags().StringP(
		"api-token",
		"a",
		"",
		"The slack API token to use.",
	)

	slackCmd.PersistentFlags().StringP(
		"bot-token",
		"t",
		"",
		"The slack bot token to use.",
	)

	slackCmd.PersistentFlags().StringP(
		"tls-cert",
		"c",
		"",
		"The TLS cert + key in .pem format used to identify the server.",
	)

	slackCmd.PersistentFlags().StringP(
		"ca-cert",
		"k",
		"",
		"The CA certificate to use in addition to the os certificate chain",
	)

	// binding flags to viper
	slackCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		viper.BindPFlag(f.Name, f)
	})
}
