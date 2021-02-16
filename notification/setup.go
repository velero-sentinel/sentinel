/*
Copyright © 2021 Markus Mahlberg <markus@mahlberg.io>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package notification

import (
	"errors"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/viper"
)

var logDefaultConfig = map[string]interface{}{"name": "notifier", "level": "warn"}

// Channel is a notification channel to be used by Sentinel.
type Channel func(cfg map[string]interface{}) Notifier

// ErrMalformedConfiguration is returned when the configuration is syntactically incorrect.
var ErrMalformedConfiguration = errors.New("Malformed configuration")

// Setup creates the notification system and returns a channel to write to.
func Setup(config *viper.Viper, logger hclog.Logger) (chan<- Message, error) {

	channels := make([]Notifier, 0)
	ln, _ := NewLogNotifier(logDefaultConfig)
	channels = append(channels, ln)

	// See https://stackoverflow.com/a/37425224 for the conversion.

	channelCfgs := make([]map[string]interface{}, 0)

	if config != nil {
		channelCfgsI := config.Get("channels")

		channelCfgsS, ok := channelCfgsI.([]interface{})

		if !ok {
			return nil, ErrMalformedConfiguration
		}

		for _, c := range channelCfgsS {
			switch cfg := c.(type) {
			case map[interface{}]interface{}:
				m := make(map[string]interface{})
				for k, v := range cfg {
					m[k.(string)] = v
					channelCfgs = append(channelCfgs, m)
				}
			}
		}
	}

	if len(channelCfgs) == 0 {
		logger.Warn("No notification channles configured. Logging only.")
	}

	for _, cfg := range channelCfgs {
		switch cfg["type"] {
		case "webhook":
			logger.Warn("Channel not implemented yet.", "type", "webhook")
		case "opsgenie":
			logger.Warn("Channel not implemented yet.", "type", "opsgenie")
		case "amqp":
			logger.Warn("Channel not implemented yet.", "type", "amqp")
		case "nats":
			logger.Warn("Channel not implemented yet.", "type", "amqp")
		default:
			logger.Error("Channel unknown", "type", cfg["type"])
		}
	}

	c := make(chan Message)

	downstreamWarn := make([]chan<- WarningMessage, 0)
	downstreamError := make([]chan<- ErrorMessage, 0)

	for _, ch := range channels {
		downstreamWarn = append(downstreamWarn, ch.WarningChannel())
		downstreamError = append(downstreamError, ch.ErrorChannel())
	}

	go func() {
		for {
			select {
			case msg := <-c:
				switch msg.(type) {
				case WarningMessage:
					for _, ch := range downstreamWarn {
						ch <- msg.(WarningMessage)
					}
				case ErrorMessage:
					for _, ch := range downstreamError {
						ch <- msg.(ErrorMessage)
					}
				}

			}
		}
	}()

	return c, nil
}
