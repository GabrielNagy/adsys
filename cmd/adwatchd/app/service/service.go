package service

import (
	"adwatchd/app/watcher"

	log "adwatchd/app/logging"

	"github.com/kardianos/service"
	"github.com/spf13/viper"
)

// Program structures.
//  Define Start and Stop methods.
type program struct {
	Service service.Service
	Logger  service.Logger
	exit    chan struct{}
}

type ServiceConfig struct {
	StartType   string
	DirsToCheck []string
}

func NewServiceConfig() ServiceConfig {
	var start_type string
	var dirs_to_check []string

	start_type = viper.GetString("start-type")
	dirs_to_check = viper.GetStringSlice("dirs-to-check")

	return ServiceConfig{
		StartType:   start_type,
		DirsToCheck: dirs_to_check,
	}
}

func (p *program) Start(s service.Service) error {
	p.exit = make(chan struct{})

	// Start should not block. Do the actual work async.
	go watcher.StartWatching()
	return nil
}

func (p *program) Stop(s service.Service) error {
	// Any work in Stop should be quick, usually a few seconds at most.
	log.Info("I'm Stopping!")
	close(p.exit)
	return nil
}

func newServiceConfig(config ServiceConfig) service.Config {
	var start_type string
	var delayed_auto_start bool

	start_type = config.StartType
	if start_type == "delayed" {
		delayed_auto_start = true
	} else {
		delayed_auto_start = false
	}

	return service.Config{
		Name:        "adwatchd",
		DisplayName: "Active Directory Watch Daemon",
		Description: "Monitors configured directories for changes and increases associated GPT.ini version.",
		Arguments:   []string{"run"},
		Option: map[string]interface{}{
			"DelayedAutoStart": delayed_auto_start,
			"StartType":        start_type,
		},
	}
}

func InstallService(config ServiceConfig) error {
	prg, err := NewProgram(config)
	if err != nil {
		return err
	}

	if err := prg.Service.Install(); err != nil {
		log.Fatal(err)
	}
	log.Info("Service successfully installed.")
	return nil
}

func UninstallService(config ServiceConfig) error {
	prg, err := NewProgram(config)
	if err != nil {
		return err
	}

	if err := prg.Service.Uninstall(); err != nil {
		log.Fatal(err)
	}
	log.Info("Service marked for uninstallation.")
	return nil
}

func NewProgram(config ServiceConfig) (*program, error) {
	svcConfig := newServiceConfig(config)
	prg := &program{}
	s, err := service.New(prg, &svcConfig)
	if err != nil {
		return nil, err
	}

	prg.Service = s
	return prg, nil
}

func IsInteractive() bool {
	return service.Interactive()
}
