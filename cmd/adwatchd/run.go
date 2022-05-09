package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ubuntu/adsys/cmd/adwatchd/internal/watchdservice"
	"github.com/ubuntu/adsys/internal/i18n"
)

func (a *App) installRun() {

	cmd := &cobra.Command{
		Use:   "run",
		Short: i18n.G("Starts the watchd service"),
		Long:  i18n.G(`Can be run through the service manager a service or interactively as a standalone application.`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(a.config.Dirs) < 1 {
				return fmt.Errorf(i18n.G("run commands needs at least one directory to watch either with --dirs or in the configuration file"))
			}
			fmt.Println(a.config.Verbose)
			service, err := watchdservice.New(context.Background(), a.config.Dirs)
			if err != nil {
				return err
			}
			return service.Run()
		},
	}

	var dirs []string
	cmd.Flags().StringSliceVarP(
		&dirs,
		"dirs",
		"d",
		[]string{`C:\Windows\Sysvol`},
		i18n.G("a `directory` to check for changes (can be specified multiple times)"),
	)
	a.viper.BindPFlag("dirs", cmd.Flags().Lookup("dirs"))

	a.rootCmd.AddCommand(cmd)
}
