package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ubuntu/adsys/internal/i18n"
)

func (a *App) installRun() {

	cmd := &cobra.Command{
		Use:   "run",
		Short: i18n.G("Starts the directory watch loop"),
		Long: i18n.G(`Can run as a service through the service manager or interactively as a standalone application.

The program will monitor the configured directories for changes and bump the appropriate GPT.ini versions anytime a change is detected.
If a GPT.ini file does not exist for a directory, a warning will be issued and the file will be created.
`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(a.config.Dirs) < 1 {
				return fmt.Errorf(i18n.G("run command needs at least one directory to watch either with --dirs or in the configuration file"))
			}

			return a.service.Run(context.Background())
		},
	}

	var dirs []string
	cmd.Flags().StringSliceVarP(
		&dirs,
		"dirs",
		"d",
		[]string{`C:\Windows\sysvol`},
		i18n.G("a `directory` to check for changes (can be specified multiple times)"),
	)
	a.viper.BindPFlag("dirs", cmd.Flags().Lookup("dirs"))

	a.rootCmd.AddCommand(cmd)
}
