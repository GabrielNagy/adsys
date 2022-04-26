/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"adwatchd/app/logging/hooks"
	"adwatchd/app/service"

	log "adwatchd/app/logging"

	"github.com/spf13/cobra"
)

// runCmd represents the run command.
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runApplication(cmd); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runApplication(cmd *cobra.Command) error {
	errs := make(chan error)
	config := service.NewServiceConfig()
	prg, err := service.NewProgram(config)
	if err != nil {
		return err
	}

	prg.Logger, err = prg.Service.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}

	if !service.IsInteractive() {
		log.AddHook(&hooks.EventLogHook{Logger: prg.Logger})
		log.SetFormatter(&hooks.EventLogFormatter{})
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Error(err)
			}
		}
	}()

	if err = prg.Service.Run(); err != nil {
		log.Error(err)
	}

	return nil
}
