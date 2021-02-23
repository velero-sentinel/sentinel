package lognotifier

import (
	"github.com/sirupsen/logrus"
	"github.com/velero-sentinel/sentinel/message"
)

// New returns the default Log notifier.
func New(logger *logrus.Logger) *logNotifier {

	ln := &logNotifier{
		logger: logger,
	}
	return ln
}

type logNotifier struct {
	logger *logrus.Logger
}

func (n *logNotifier) Run() chan<- message.Message {
	n.logger.Debug("Run called")
	c := make(chan message.Message)

	go func() {
		n.logger.Debug("Entering goroutine")
		for m := range c {
			n.logger.Debug("Received message from upstream")
			switch m.(type) {
			case message.Warning, *message.Warning:
				n.logger.Warn("Backup partially failed", "backup", m.GetBackup().Name)

			case message.Error, *message.Error:
				n.logger.Error("Backup failed", "backup", m.GetBackup().Name)
			}
		}
		n.logger.Debug("Leaving goroutine")
	}()

	return c
}
