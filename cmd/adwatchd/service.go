package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ubuntu/adsys/internal/i18n"
)

func (a *App) installService() {

	cmd := &cobra.Command{
		Use:   "service",
		Short: i18n.G("Manages the adwatchd service"),
		Long: i18n.G(`The service command allows the user to interact with the adwatchd service. It can be used to start, stop, and restart the service.
Additionally, it can be used to check the status of the service and also install/uninstall it.
`),
	}

	cmd.AddCommand(a.serviceStart())
	cmd.AddCommand(a.serviceStop())
	cmd.AddCommand(a.serviceRestart())
	cmd.AddCommand(a.serviceStatus())
	cmd.AddCommand(a.serviceInstall())
	cmd.AddCommand(a.serviceUninstall())

	a.rootCmd.AddCommand(cmd)
}

func (a *App) serviceStart() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: i18n.G("Starts the service"),
		Long:  i18n.G("Starts the adwatchd service."),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.service.Start(context.Background())
		},
	}
}

func (a *App) serviceStop() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: i18n.G("Stops the service"),
		Long:  i18n.G("Stops the adwatchd service."),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.service.Stop(context.Background())
		},
	}
}

func (a *App) serviceRestart() *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: i18n.G("Restarts the service"),
		Long:  i18n.G("Restarts the adwatchd service."),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.service.Restart(context.Background())
		},
	}
}

func (a *App) serviceStatus() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: i18n.G("Returns service status"),
		Long:  i18n.G("Returns the status of the adwatchd service."),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(a.service.Status(context.Background()))
		},
	}
}

func (a *App) serviceInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: i18n.G("Installs the service"),
		Long: i18n.G(`Installs the adwatchd service.
		
The service will be installed as a Windows service.
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.service.Install(context.Background())
		},
	}
	cmd.Flags().String("start-type", "automatic", i18n.G("the start type of the service (automatic, delayed, manual, disabled)"))
	a.viper.BindPFlag("start-type", cmd.Flags().Lookup("start-type"))

	return cmd
}

func (a *App) serviceUninstall() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: i18n.G("Uninstalls the service"),
		Long:  i18n.G("Uninstalls the adwatchd service."),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.service.Uninstall(context.Background())
		},
	}
}
