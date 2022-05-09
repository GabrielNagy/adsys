package loghooks

import (
	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
)

// EventLog sends logs via Windows/journal log.
type EventLog struct {
	service.Logger
}

// Fire is called when an event should be logged.
func (hook *EventLog) Fire(entry *logrus.Entry) error {
	line := entry.Message

	switch entry.Level {
	case logrus.PanicLevel:
		fallthrough
	case logrus.FatalLevel:
		fallthrough
	case logrus.ErrorLevel:
		return hook.Error(line)
	case logrus.WarnLevel:
		return hook.Warning(line)
	default:
		return hook.Info(line)
	}
}

// Levels returns the level that this hook is triggered on.
func (hook *EventLog) Levels() []logrus.Level {
	return logrus.AllLevels
}
