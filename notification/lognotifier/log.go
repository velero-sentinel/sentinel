package lognotifier

import (
	"github.com/hashicorp/go-hclog"
	"github.com/velero-sentinel/sentinel/message"
)

// NewLogNotifier returns the default Log notifier.
func New(logger hclog.Logger) *hclogNotifier {

	ln := &hclogNotifier{
		logger:   logger,
		warnings: make(chan message.WarningMessage),
		errors:   make(chan message.ErrorMessage),
	}
	go ln.Run()
	return ln
}

type hclogNotifier struct {
	logger   hclog.Logger
	warnings chan message.WarningMessage
	errors   chan message.ErrorMessage
	done     chan bool
}

func (n *hclogNotifier) WarningC() chan<- message.WarningMessage {
	return n.warnings
}

func (n *hclogNotifier) ErrorC() chan<- message.ErrorMessage {
	return n.errors
}

func (n *hclogNotifier) Run() {
	for {
		select {
		case <-n.done:
			return
		case warning := <-n.warnings:
			n.logger.Warn("Problem with backup detected", "state", warning.GetBackup().Status.Phase, "message", warning.Message())
		case err := <-n.errors:
			n.logger.Error("Backup error", "state", err.GetBackup().Status.Phase, "message", err.Message())
		}
	}
}

func (n *hclogNotifier) Stop() {
	n.logger.Warn("Received shutdown command")
	n.done <- true
	n.logger.Info("Shutdown complete")
}
