package cmd

import "github.com/hashicorp/go-hclog"

type hclogNotifier struct {
	logger hclog.Logger
}

func (n hclogNotifier) Warn(m Message) {
	n.logger.Warn(m.Message())
}

func (n hclogNotifier) Error(m Message) {
	n.logger.Error(m.Message())
}

// LogNotifier is a simple Notifier that logs the warnings.
func LogNotifier(logger hclog.Logger) Notifier {
	return hclogNotifier{logger: logger}
}
