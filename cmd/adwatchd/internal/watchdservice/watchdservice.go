package watchdservice

import (
	"context"

	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"
	"github.com/ubuntu/adsys/cmd/adwatchd/internal/loghooks"
	"github.com/ubuntu/adsys/cmd/adwatchd/internal/watcher"
)

// WatchdService contains the service and watcher.
type WatchdService struct {
	service service.Service
	watcher *watcher.Watcher
}

// New returns a new WatchdService instance.
func New(ctx context.Context, dirs []string) (*WatchdService, error) {

	watcher, err := watcher.New(ctx, dirs)
	if err != nil {
		return nil, err
	}

	config := service.Config{
		Name:        "adwatchd",
		DisplayName: "Active Directory Watch Daemon",
		Description: "Monitors configured directories for changes and increases the associated GPT.ini version.",
		Arguments:   []string{"run"},
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

// UpdateDirs updates the watcher with the new directories.
func (s *WatchdService) UpdateDirs(dirs []string) error {
	return s.watcher.UpdateDirs(dirs)
}

// Start starts the watcher service.
func (s *WatchdService) Start() error {
	return nil
}

// Stop stops the watcher service.
func (s *WatchdService) Stop() error {
	return nil
}

// Restart restarts the watcher service.
func (s *WatchdService) Restart() error {
	return nil
}

// Status provides a status of the watcher service.
func (s *WatchdService) Status() string {
	return ""
}

// Install installs the watcher service.
func (s *WatchdService) Install() error {
	return nil
}

// Uninstall uninstalls the watcher service.
func (s *WatchdService) Uninstall() error {
	return nil
}

// Run runs the watcher service.
func (s *WatchdService) Run() error {
	return s.service.Run()
}
