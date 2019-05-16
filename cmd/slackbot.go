package cmd

import (
	"fmt"

	"github.com/pantheon-systems/riker/pkg/chat/slackbot"
	"github.com/pantheon-systems/riker/pkg/healthz"
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
		viper.GetString("name"),
		viper.GetString("bind-address"),
		viper.GetString("bot-token"),
		viper.GetString("api-token"),
		viper.GetString("tls-cert"),
		viper.GetString("ca-cert"),
		viper.GetStringSlice("allowed-ou"),
		log,
	)
	if err != nil {
		return err
	}

	healthServer, err := healthz.New(healthz.Config{
		BindPort: viper.GetInt("healthz-port"),
		BindAddr: viper.GetString("healthz-address"),
		Logger:   log,
		Providers: []healthz.ProviderInfo{
			{
				Type:        "Slackbot",
				Description: "Check slackbot health.",
				Check:       b,
			},
		},
	})
	if err != nil {
		return err
	}

	go healthServer.StartHealthz()
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
		"/etc/ssl/cacerts/ca-bundle.crt",
		"The CA certificate chain to use.",
	)

	slackCmd.PersistentFlags().StringSliceP(
		"allowed-ou",
		"o",
		[]string{
			"riker",
			"riker-server",
			"riker-redshirt",
		},
		"Allowed Cert ous for redhirts to register commands to the server. Can be specified multiple times.",
	)

	slackCmd.PersistentFlags().Int32P(
		"healthz-port",
		"p",
		8080,
		"The port that the healthz HTTP server will listen on.",
	)

	slackCmd.PersistentFlags().StringP(
		"healthz-address",
		"l",
		"0.0.0.0",
		"The address that the healthz HTTP server will listen on.",
	)

	// binding flags to viper
	slackCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		viper.BindPFlag(f.Name, f)
	})
}
