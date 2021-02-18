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
	"github.com/velero-sentinel/sentinel/message"
	"github.com/velero-sentinel/sentinel/notification/lognotifier"
	"github.com/velero-sentinel/sentinel/notification/webhook"
)

var logDefaultConfig = map[string]interface{}{"name": "notifier", "level": "warn"}

// Channel is a notification channel to be used by Sentinel.
type Channel func(cfg map[string]interface{}) Notifier

// ErrMalformedConfiguration is returned when the configuration is syntactically incorrect.
var ErrMalformedConfiguration = errors.New("Malformed configuration")

// TODO: Make pipeline an interface for easy testing
type Pipeline struct {
	logger          hclog.Logger
	upstream        <-chan message.Message
	warningChannels []chan<- message.WarningMessage
	errorChannels   []chan<- message.ErrorMessage
	done            chan bool
}

func (p Pipeline) Run() {

	for {
		select {
		case <-p.done:
			for _, w := range p.warningChannels {
				close(w)
			}
			for _, e := range p.errorChannels {
				close(e)
			}
			return
		case m := <-p.upstream:
			// TODO: Use simlpe downstream channels and let the notifiers decide
			// wether to treat warnings ot errors differently.
			switch m.(type) {
			case message.WarningMessage:
				w := m.(message.WarningMessage)
				for _, wc := range p.warningChannels {
					wc <- w
				}
			case message.ErrorMessage:
				e := m.(message.ErrorMessage)
				for _, ec := range p.errorChannels {
					ec <- e
				}
			}
		}
	}
}

func (p *Pipeline) Stop() {
	p.logger.Warn("Received shutdown command")
	p.done <- true
	p.logger.Warn("Server shutdown complete")
}

func Configure(cfg *NotifierConfig, logger hclog.Logger) (*Pipeline, error) {
	mux := make([]Notifier, 0)

	webhookLogger := logger.Named("webhook")
	for _, hook := range cfg.Webhooks {
		n, err := webhook.New(&hook, webhookLogger.With("name", hook.Name))
		if err != nil {
			return nil, err
		}
		mux = append(mux, n)
	}

	mux = append(mux, lognotifier.New(logger.Named("log")))

	p := &Pipeline{
		upstream: make(chan message.Message),
		// TODO: Initialize with len(mux) and fill them accordingly
		warningChannels: make([]chan<- message.WarningMessage, 0),
		errorChannels:   make([]chan<- message.ErrorMessage, 0),
		done:            make(chan bool),
	}

	for _, c := range mux {
		p.warningChannels = append(p.warningChannels, c.WarningC())
		p.errorChannels = append(p.errorChannels, c.ErrorC())
	}

	return p, nil
}

// Setup creates the notification system and returns a channel to write to.
// func Setup(config *viper.Viper, logger hclog.Logger) (chan<- message.Message, error) {

// 	channels := make([]Notifier, 0)
// 	ln, _ := NewLogNotifier(logDefaultConfig)
// 	channels = append(channels, ln)

// 	// See https://stackoverflow.com/a/37425224 for the conversion.

// 	channelCfgs := make([]map[string]interface{}, 0)

// 	if config != nil {
// 		channelCfgsI := config.Get("channels")

// 		channelCfgsS, ok := channelCfgsI.([]interface{})

// 		if !ok {
// 			return nil, ErrMalformedConfiguration
// 		}

// 		for _, c := range channelCfgsS {
// 			switch cfg := c.(type) {
// 			case map[interface{}]interface{}:
// 				m := make(map[string]interface{})
// 				for k, v := range cfg {
// 					m[k.(string)] = v
// 					channelCfgs = append(channelCfgs, m)
// 				}
// 			}
// 		}
// 	}

// 	if len(channelCfgs) == 0 {
// 		logger.Warn("No notification channles configured. Logging only.")
// 	}

// 	for _, cfg := range channelCfgs {
// 		switch cfg["type"] {
// 		case "webhook":
// 			logger.Warn("Channel not implemented yet.", "type", "webhook")
// 		case "opsgenie":
// 			logger.Warn("Channel not implemented yet.", "type", "opsgenie")
// 		case "amqp":
// 			logger.Warn("Channel not implemented yet.", "type", "amqp")
// 		case "nats":
// 			logger.Warn("Channel not implemented yet.", "type", "amqp")
// 		default:
// 			logger.Error("Channel unknown", "type", cfg["type"])
// 		}
// 	}

// 	c := make(chan Message)

// 	downstreamWarn := make([]chan<- WarningMessage, 0)
// 	downstreamError := make([]chan<- ErrorMessage, 0)

// 	for _, ch := range channels {
// 		downstreamWarn = append(downstreamWarn, ch.WarningC())
// 		downstreamError = append(downstreamError, ch.ErrorC())
// 	}

// 	go func() {
// 		for {
// 			select {
// 			case msg := <-c:
// 				switch msg.(type) {
// 				case WarningMessage:
// 					for _, ch := range downstreamWarn {
// 						ch <- msg.(WarningMessage)
// 					}
// 				case ErrorMessage:
// 					for _, ch := range downstreamError {
// 						ch <- msg.(ErrorMessage)
// 					}
// 				}

// 			}
// 		}
// 	}()

// 	return c, nil
// }
