package watchdservice

import (
	"context"

	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/adsys/cmd/adwatchd/internal/loghooks"
	"github.com/ubuntu/adsys/cmd/adwatchd/internal/watcher"
)

type WatchdService struct {
	service service.Service
	watcher *watcher.Watcher
}

func New(ctx context.Context, dirs []string) (*WatchdService, error) {

	watcher, err := watcher.New(ctx, dirs)
	if err != nil {
		return nil, err
	}

	/*svcConfig := newServiceConfig(config)
	s, err := service.New(prg, &svcConfig)
	if err != nil {
		return nil, err
	}*/

	config := service.Config{
		Name:        "adwatchd",
		DisplayName: "Active Directory Watch Daemon",
		Description: "Monitors configured directories for changes and increases associated GPT.ini version.",
		////Arguments:   []string{"run"},
	}
	s, err := service.New(watcher, &config)
	if err != nil {
		return nil, err
	}

	if !service.Interactive() {
		logger, err := s.Logger(nil)
		if err != nil {
			return nil, err
		}
		log.AddHook(&loghooks.EventLog{Logger: logger})
	}

	return &WatchdService{
		service: s,
		watcher: watcher,
	}, nil
}

func (s *WatchdService) Start() error {
	return nil
}

func (s *WatchdService) Stop() error {
	return nil
}

func (s *WatchdService) Restart() error {
	return nil
}

func (s *WatchdService) Install() error {
	return nil
}

func (s *WatchdService) Uninstall() error {
	return nil
}

func (s *WatchdService) Run() error {
	return s.service.Run()
}

/*
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
*/
