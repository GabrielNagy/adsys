/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	log "adwatchd/app/logging"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var dirs []string

// rootCmd represents the base command when called without any subcommands.
/*var rootCmd = &cobra.Command{
	Use:   "adwatchd",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig(cmd)
	},

	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {},
}*/

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	setCmdFlags(rootCmd)
}

func bindConfigFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		viper.BindPFlag(f.Name, f)
	})
}

/*func setCmdFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(
		"config",
		"c",
		`C:\vagrant\adwatchd.yaml`, // TODO update with final path
		"`path` to config file",
	)
	cmd.PersistentFlags().StringSliceVarP(
		&dirs,
		"dirs-to-check",
		"D",
		[]string{`C:\Windows\Sysvol`},
		"a `directory` to check for changes (can be specified multiple times)",
	)
	cmd.PersistentFlags().String(
		"start-type",
		"automatic",
		"start `type` of the service. Supported types are: manual, automatic, delayed, and disabled",
	)
	cmd.PersistentFlags().DurationP(
		"tree-stable-time",
		"t",
		func() time.Duration { return time.Second * 10 }(),
		"How much `time` to wait before increasing the GPT.ini version after a change is detected in a monitored directory",
	)
	cmd.PersistentFlags().BoolP(
		"debug",
		"d",
		false,
		"enable debug logging (warning: service logging can be very verbose)",
	)
}*/

func setupLogging(cmd *cobra.Command) error {
	debug := viper.GetBool("debug")

	if debug {
		log.Info("debugging")
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
	return nil
}

func initConfig(cmd *cobra.Command) error {
	cfgFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	viper.SetConfigType("yaml")
	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return err
		}
	}

	bindConfigFlags(cmd)
	setupLogging(cmd)

	return nil
}
