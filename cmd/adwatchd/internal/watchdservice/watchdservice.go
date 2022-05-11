package watchdservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kardianos/service"
	"github.com/ubuntu/adsys/cmd/adwatchd/internal/loghooks"
	"github.com/ubuntu/adsys/internal/decorate"
	log "github.com/ubuntu/adsys/internal/grpc/logstreamer"
	"github.com/ubuntu/adsys/internal/i18n"
	"github.com/ubuntu/adsys/internal/watcher"
)

// WatchdService contains the service and watcher.
type WatchdService struct {
	service service.Service
	watcher *watcher.Watcher
}

type options struct {
	dirs        []string
	userService bool
}
type option func(*options) error

// WithDirs allows overriding default directories to watch.
func WithDirs(dirs []string) func(o *options) error {
	return func(o *options) error {
		o.dirs = dirs
		return nil
	}
}

// New returns a new WatchdService instance.
func New(ctx context.Context, opts ...option) (*WatchdService, error) {
	args := options{}
	// applied options
	for _, o := range opts {
		if err := o(&args); err != nil {
			return nil, err
		}
	}

	var w *watcher.Watcher
	var err error
	if len(args.dirs) > 0 {
		if w, err = watcher.New(ctx, args.dirs); err != nil {
			return nil, err
		}
	}

	config := service.Config{
		Name:        "adwatchd",
		DisplayName: "Active Directory Watch Daemon",
		Description: "Monitors configured directories for changes and increases the associated GPT.ini version.",
		Arguments:   []string{"run"},
	}
	s, err := service.New(w, &config)
	if err != nil {
		return nil, err
	}

	if !service.Interactive() {
		logger, err := s.Logger(nil)
		if err != nil {
			return nil, err
		}
		log.AddHook(ctx, &loghooks.EventLog{Logger: logger})
	}

	return &WatchdService{
		service: s,
		watcher: w,
	}, nil
}

// UpdateDirs updates the watcher with the new directories.
func (s *WatchdService) UpdateDirs(ctx context.Context, dirs []string) (err error) {
	decorate.OnError(&err, i18n.G("failed to change directories to watch"))
	log.Info(ctx, "Updating directories to watch")

	return s.watcher.UpdateDirs(dirs)
}

// Start starts the watcher service.
func (s *WatchdService) Start(ctx context.Context) (err error) {
	decorate.OnError(&err, i18n.G("failed to start service"))
	log.Info(ctx, "Starting service")

	if err := s.service.Start(); err != nil {
		return err
	}

	return s.waitForStatus(ctx, service.StatusRunning)
}

// Stop stops the watcher service.
func (s *WatchdService) Stop(ctx context.Context) (err error) {
	decorate.OnError(&err, i18n.G("failed to stop service"))
	log.Info(ctx, "Stopping service")

	if err := s.service.Stop(); err != nil {
		return err
	}

	return s.waitForStatus(ctx, service.StatusStopped)
}

func (s *WatchdService) waitForStatus(ctx context.Context, status service.Status) error {
	// Check that the service updated correctly.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var gotStatus bool
	for !gotStatus {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			newStatus, _ := s.service.Status()
			if newStatus != status {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			gotStatus = true
			break
		}
	}
	return nil
}

// Restart restarts the watcher service.
func (s *WatchdService) Restart(ctx context.Context) (err error) {
	decorate.OnError(&err, i18n.G("failed to restart service"))
	log.Info(ctx, "Restarting service")

	if err := s.service.Stop(); err != nil {
		return err
	}

	if err := s.service.Start(); err != nil {
		return err
	}

	return nil
}

// Status provides a status of the watcher service.
func (s *WatchdService) Status(ctx context.Context) (status string, err error) {
	decorate.OnError(&err, i18n.G("failed to retrieve status for service"))
	log.Debug(ctx, "Getting status from service")

	uninstalledState := service.Status(42)
	stat, err := s.service.Status()
	if errors.Is(err, service.ErrNotInstalled) {
		stat = uninstalledState
	} else if err != nil {
		return "", err
	}

	var serviceStatus string
	switch stat {
	case service.StatusRunning:
		serviceStatus = "running"
	case service.StatusStopped:
		serviceStatus = "stopped"
	case service.StatusUnknown:
		serviceStatus = "unknown"
	case uninstalledState:
		serviceStatus = "not installed"
	default:
		serviceStatus = "undefined"
	}

	dirs := "none"
	if s.watcher != nil {
		dirs = strings.Join(s.watcher.Dirs(), "\n -")
	}

	status = fmt.Sprintf(i18n.G(`Service status: %s
Watching directories:
 - %v`), serviceStatus, dirs)

	return status, nil
}

// Install installs the watcher service.
func (s *WatchdService) Install(ctx context.Context) (err error) {
	decorate.OnError(&err, i18n.G("failed to install service"))
	log.Info(ctx, "Installing watcher service")
	return s.service.Install()
}

// Uninstall uninstalls the watcher service.
func (s *WatchdService) Uninstall(ctx context.Context) (err error) {
	decorate.OnError(&err, i18n.G("failed to uninstall service"))
	log.Info(ctx, "Uninstalling watcher service")
	return s.service.Uninstall()
}

// Run runs the watcher service.
func (s *WatchdService) Run(ctx context.Context) (err error) {
	decorate.OnError(&err, i18n.G("failed to run service"))
	log.Info(ctx, "Running watcher service")

	return s.service.Run()
}
