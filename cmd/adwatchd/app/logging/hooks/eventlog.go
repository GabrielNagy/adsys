package hooks

/*import (
	"fmt"
	"os"

	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
)

// EventLogHook to send logs via windows log.
type EventLogHook struct {
	Logger service.Logger
}

type EventLogFormatter struct {
	logrus.TextFormatter
}

func (hook *EventLogHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}

	switch entry.Level {
	case logrus.PanicLevel:
		return hook.Logger.Error(line)
	case logrus.FatalLevel:
		return hook.Logger.Error(line)
	case logrus.ErrorLevel:
		return hook.Logger.Error(line)
	case logrus.WarnLevel:
		return hook.Logger.Warning(line)
	case logrus.InfoLevel:
		return hook.Logger.Info(line)
	case logrus.DebugLevel:
		return hook.Logger.Info(line)
	default:
		return nil
	}
}

func (hook *EventLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (f *EventLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message), nil
}
*/
