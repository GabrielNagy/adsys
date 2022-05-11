package watchdservice

import "github.com/kardianos/service"

// AsUserService allows installing/uninstalling the service as a user.
func AsUserService() func(o *options) error {
	return func(o *options) error {
		o.userService = true
		return nil
	}
}

// WithName allows setting a custom name to the service.
func WithName(name string) func(o *options) error {
	return func(o *options) error {
		o.name = name
		return nil
	}
}

// Interactive allows setting the service to be interactive.
func Interactive(interactive bool) func(o *options) error {
	return func(o *options) error {
		o.interactive = interactive
		return nil
	}
}

// ServiceStatus queries the API for the service status.
func (w *WatchdService) ServiceStatus() (service.Status, error) {
	return w.service.Status()
}
