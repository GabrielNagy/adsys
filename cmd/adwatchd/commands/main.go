package commands

import (
	"context"
	"path/filepath"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	log "github.com/ubuntu/adsys/internal/grpc/logstreamer"
	"golang.org/x/exp/slices"

	"github.com/ubuntu/adsys/cmd/adwatchd/internal/watchdservice"
	"github.com/ubuntu/adsys/internal/cmdhandler"
	"github.com/ubuntu/adsys/internal/config"
	"github.com/ubuntu/adsys/internal/i18n"
)

// App encapsulates commands and options of the daemon, which can be controlled by env variables and config files.
type App struct {
	rootCmd cobra.Command
	viper   *viper.Viper

	config  appConfig
	service *watchdservice.WatchdService
	options options

	ready chan struct{}
}

type appConfig struct {
	Verbose int
	Force   bool
	Dirs    []string
}

type options struct {
	name string
}
type option func(*options) error

// WithServiceName allows setting a custom name for the daemon.
func WithServiceName(name string) func(o *options) error {
	return func(o *options) error {
		o.name = name
		return nil
	}
}

// New registers commands and return a new App.
func New(opts ...option) *App {
	// Set default options.
	args := options{
		name: "adwatchd",
	}

	// Apply given options.
	for _, o := range opts {
		o(&args)
	}

	a := App{ready: make(chan struct{})}
	a.options = args
	a.rootCmd = cobra.Command{
		Use:   "adwatchd [COMMAND]",
		Short: i18n.G("AD watch daemon"),
		Long:  i18n.G(`Watch directories for changes and bump the relevant GPT.ini versions.`),
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Command parsing has been successful. Returns runtime (or configuration) error now and so, don't print usage.
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

				// Reload verbosity and directories.
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
				close(a.ready)
				return err
			}

			// If we have a config file, pass it as an argument to the service.
			var configFile []string
			if len(a.viper.ConfigFileUsed()) > 0 {
				absPath, err := filepath.Abs(a.viper.ConfigFileUsed())
				if err != nil {
					close(a.ready)
					return err
				}
				configFile = []string{"-c", absPath}
			}

			// Create main service and attach it to the app
			service, err := watchdservice.New(
				context.Background(),
				watchdservice.WithName(a.options.name),
				watchdservice.WithDirs(a.config.Dirs),
				watchdservice.WithArgs(configFile))

			if err != nil {
				close(a.ready)
				return err
			}
			a.service = service
			close(a.ready)

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

	// Install subcommands
	a.installRun()
	a.installService()

	return &a
}

// Run executes the app.
func (a *App) Run() error {
	return a.rootCmd.Execute()
}

// UsageError returns if the error is a command parsing or runtime one.
func (a App) UsageError() bool {
	return !a.rootCmd.SilenceUsage
}

// SetArgs changes the root command args. Shouldn't be in general necessary apart for integration tests.
func (a *App) SetArgs(args []string) {
	a.rootCmd.SetArgs(args)
}

// Reset recreates the ready channel. Shouldn't be in general necessary apart
// for integration tests, where multiple commands are executed on the same
// instance.
func (a *App) Reset() {
	a.ready = make(chan struct{})
}

// Quit gracefully exits the app.
func (a *App) Quit() error {
	a.waitReady()

	if service.Interactive() {
		return a.service.StopWatch(context.Background())
	}
	return a.service.Stop(context.Background())
}

// waitReady signals when the daemon is ready
// Note: we need to use a pointer to not copy the App object before the daemon is ready, and thus, creates a data race.
func (a *App) waitReady() {
	<-a.ready
}
