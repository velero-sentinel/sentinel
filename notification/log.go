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

	return hclogNotifier{logger: logger}, nil
}

type hclogNotifier struct {
	logger hclog.Logger
}

func (n hclogNotifier) Warn(m Message) {
	n.logger.Warn(m.Message())
}

func (n hclogNotifier) Error(m Message) {
	n.logger.Error(m.Message())
}
