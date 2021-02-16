package notification

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
)

// ErrIncompleteConfiguration is returned when requried keys in the configuration of a notifier is missing.
type ErrIncompleteConfiguration struct {
	Key string
}

func (e ErrIncompleteConfiguration) Error() string {
	return fmt.Sprintf("Key '%s' is missing", e.Key)
}

// NewLogNotifier returns the default Log notifier.
func NewLogNotifier(cfg map[string]interface{}) (Notifier, error) {

	for _, key := range []string{"name", "level"} {
		if _, ok := cfg[key]; !ok {
			return nil, ErrIncompleteConfiguration{Key: key}
		}
	}

	logger := hclog.New(
		&hclog.LoggerOptions{
			Name:  cfg["name"].(string),
			Level: hclog.LevelFromString(cfg["level"].(string)),
		})

	ln := &hclogNotifier{
		logger:   logger,
		warnings: make(chan WarningMessage),
		errors:   make(chan ErrorMessage),
	}
	go ln.run()
	return ln, nil
}

type hclogNotifier struct {
	logger   hclog.Logger
	warnings chan WarningMessage
	errors   chan ErrorMessage
}

func (n *hclogNotifier) WarningC() chan<- WarningMessage {
	return n.warnings
}

func (n *hclogNotifier) ErrorC() chan<- ErrorMessage {
	return n.errors
}

func (n *hclogNotifier) run() {
	for {
		select {
		case warning := <-n.warnings:
			n.logger.Warn("Problem with backup detected", "state", warning.GetBackup().Status.Phase, "message", warning.Message())
		case err := <-n.errors:
			n.logger.Error("Backup error", "state", err.GetBackup().Status.Phase, "message", err.Message())
		}
	}
}
