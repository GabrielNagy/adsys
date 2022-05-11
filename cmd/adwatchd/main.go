package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	log "github.com/ubuntu/adsys/internal/grpc/logstreamer"
	"golang.org/x/exp/slices"

	"github.com/ubuntu/adsys/cmd/adwatchd/internal/watchdservice"
	"github.com/ubuntu/adsys/internal/cmdhandler"
	"github.com/ubuntu/adsys/internal/config"
	"github.com/ubuntu/adsys/internal/consts"
	"github.com/ubuntu/adsys/internal/i18n"
)

// App encapsulates commands and options of the daemon, which can be controlled by env variables and config files.
type App struct {
	rootCmd cobra.Command
	viper   *viper.Viper

	config  appConfig
	service *watchdservice.WatchdService
}

type appConfig struct {
	Verbose int
	Dirs    []string
}

// New registers commands and return a new App.
func New() App {
	a := App{}

	a.rootCmd = cobra.Command{
		Use:   "adwatchd [COMMAND]",
		Short: i18n.G("AD watch daemon"),
		Long:  i18n.G(`Watch directories for changes and bump the relevant GPT.ini versions.`),
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// command parsing has been successful. Returns runtime (or configuration) error now and so, don't print usage.
			cmd.SilenceUsage = true
			err := config.Init("adwatchd", a.rootCmd, a.viper, func(refreshed bool) error {
				var newConfig appConfig
				if err := config.LoadConfig(&newConfig, a.viper); err != nil {
					return err
				}

				// First run: just init configuration.
				if !refreshed {
					a.config = newConfig
					return nil
				}

				// Config reload

				// Reload necessary parts
				oldVerbose := a.config.Verbose
				oldDirs := a.config.Dirs
				a.config = newConfig
				if oldVerbose != a.config.Verbose {
					config.SetVerboseMode(a.config.Verbose)
				}
				if !slices.Equal(oldDirs, a.config.Dirs) {
					if a.service != nil {
						if err := a.service.UpdateDirs(context.Background(), a.config.Dirs); err != nil {
							log.Warningf(context.Background(), "failed to update directories: %v", err)
						}
					}
				}

				return nil
			})
			// Set configured verbose status for the daemon before getting error output.
			config.SetVerboseMode(a.config.Verbose)
			if err != nil {
				return err
			}

			// Create main service and attach it to the app
			service, err := watchdservice.New(context.Background(), watchdservice.WithDirs(a.config.Dirs))
			if err != nil {
				return err
			}
			a.service = service

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},

		// We display usage error ourselves
		SilenceErrors: true,
	}

	a.viper = viper.New()

	cmdhandler.InstallVerboseFlag(&a.rootCmd, a.viper)
	a.rootCmd.PersistentFlags().StringP(
		"config",
		"c",
		``, // TODO update with final path
		i18n.G("`path` to config file"),
	)

	// subcommands
	a.installRun()
	a.installService()

	return a
}

// usageError returns if the error is a command parsing or runtime one.
func (a App) usageError() bool {
	return !a.rootCmd.SilenceUsage
}

func run(a App) int {
	i18n.InitI18nDomain(consts.TEXTDOMAIN)
	//TODO: defer installSignalHandler(a)()

	// log.SetFormatter(&log.TextFormatter{
	// 	DisableLevelTruncation: true,
	// 	DisableTimestamp:       true,
	// })

	if err := a.rootCmd.Execute(); err != nil {
		log.Error(context.Background(), err)

		if a.usageError() {
			return 2
		}
		return 1
	}

	return 0
}

func main() {
	app := New()
	os.Exit(run(app))
}
