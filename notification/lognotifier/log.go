package lognotifier

import (
	"github.com/hashicorp/go-hclog"
	"github.com/velero-sentinel/sentinel/message"
)

// New returns the default Log notifier.
func New(logger hclog.Logger) *hclogNotifier {

	ln := &hclogNotifier{
		logger: logger,
	}
	return ln
}

type hclogNotifier struct {
	logger hclog.Logger
}

func (n *hclogNotifier) Run() chan<- message.Message {
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
