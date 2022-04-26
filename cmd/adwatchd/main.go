package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"

	"github.com/ubuntu/adsys/internal/cmdhandler"
	"github.com/ubuntu/adsys/internal/config"
	"github.com/ubuntu/adsys/internal/consts"
	"github.com/ubuntu/adsys/internal/i18n"
)

// App encapsulates commands and options of the daemon, which can be controlled by env variables and config files.
type App struct {
	rootCmd cobra.Command
	viper   *viper.Viper

	config appConfig
}

type appConfig struct {
	Verbose int
	Dirs    []string

	//CacheDir string `mapstructure:"cache_dir"`
}

// New registers commands and return a new App.
func New() App {
	a := App{}

	a.rootCmd = cobra.Command{
		Use:   "adwatchd [COMMAND]",
		Short: i18n.G("TODO A brief description of your application"),
		Long: i18n.G(`TODO A longer description that spans multiple lines and likely contains
	examples and usage of using your application. For example:

	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly create a Cobra application.`),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// command parsing has been successful. Returns runtime (or configuration) error now and so, don’t print usage.
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
				a.config = newConfig
				if oldVerbose != a.config.Verbose {
					config.SetVerboseMode(a.config.Verbose)
				}
				oldDirs := a.config.Dirs
				a.config = newConfig
				if !slices.Equal(oldDirs, a.config.Dirs) {
					// TODO
				}

				return nil
			})
			// Set configured verbose status for the daemon.
			config.SetVerboseMode(a.config.Verbose)
			return err
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

	return a
}

// usageError returns if the error is a command parsing or runtime one.
func (a App) usageError() bool {
	return !a.rootCmd.SilenceUsage
}

func run(a App) int {
	i18n.InitI18nDomain(consts.TEXTDOMAIN)
	//TODO: defer installSignalHandler(a)()

	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
	})

	if err := a.rootCmd.Execute(); err != nil {
		log.Error(err)

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
