// Copyright © 2017 Pantheon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const appName = "riker"

var (
	version = "development"
	log     *logrus.Logger
	cfgFile string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Report version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "riker",
	Short: "micro chat-bot framework",
	Long: `Riker is a chat service gateway that uses gRPC to communicate with
chat command providers. These  'redshirts' are simple clients that register
a command path to riker with their required authentication.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.AddCommand(slackCmd)
	RootCmd.AddCommand(terminalCmd)
	RootCmd.AddCommand(versionCmd)

	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringP(
		"bot-name",
		"n",
		"riker",
		"The name of your server in chat",
	)

	RootCmd.PersistentFlags().StringP(
		"bind-address",
		"b",
		":6000",
		"The ip and port riker should listen for redshirts on",
	)

	RootCmd.PersistentFlags().StringVarP(
		&cfgFile,
		"config-file",
		"f",
		"",
		"Configuration file to use")

	RootCmd.PersistentFlags().BoolP(
		"debug",
		"d",
		false,
		"Enable debug logging")

	RootCmd.PersistentFlags().BoolP(
		"json-log",
		"j",
		false,
		"Enable json output formatted logging",
	)

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	dashToUs := strings.NewReplacer("-", "_")
	viper.SetConfigName("." + appName)           // name of config file (without extension)
	viper.AddConfigPath(".")                     // cwd is highest (preferred) config path
	viper.AddConfigPath("$HOME")                 // home directory as second search path
	viper.SetEnvPrefix(strings.ToUpper(appName)) // environment variable prefix
	viper.SetEnvKeyReplacer(dashToUs)            // convert environment variable keys from - to _
	viper.AutomaticEnv()                         // read in environment variables that match

	// Bind all cobra command flags to viper.
	RootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		viper.BindPFlag(f.Name, f)
	})

	log = logrus.New()
	log.Out = os.Stdout

	if viper.GetBool("json-log") {
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.999Z07:00", // RFC3339 at millisecond precision
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime: "@timestamp",
				logrus.FieldKeyMsg:  "message",
			},
		}
		log.Info("Enabling JSON logging")
	}

	log.Level = logrus.InfoLevel
	if viper.GetBool("debug") {
		log.Level = logrus.DebugLevel
		log.Debug("debugging enabled")
	}

	// Allows us to specify a config file via flag
	if cfgFile != "" {
		log.Debug("Using config from file: ", cfgFile)
		viper.SetConfigFile(cfgFile)
	}

	// If a configuration file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil && cfgFile != "" {
		log.Warn("Error reading config", err)
	}
}
