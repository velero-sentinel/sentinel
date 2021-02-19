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
	"github.com/hashicorp/go-hclog"
	"github.com/velero-sentinel/sentinel/message"
	"github.com/velero-sentinel/sentinel/notification/lognotifier"
	"github.com/velero-sentinel/sentinel/notification/webhook"
)

func Configure(cfg *NotifierConfig, logger hclog.Logger) (*Pipeline, error) {
	logger.Debug("Setting up notifiers")
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
		upstream:   make(chan message.Message),
		downstream: make([]chan<- message.Message, len(mux)),
		done:       make(chan bool),
		logger:     logger.Named("pipeline"),
	}

	for n, c := range mux {
		logger.Debug("Setting up notifier", "number", n)
		p.downstream[n] = c.Run()
		logger.Debug("Set up notifier", "number", n)
	}

	return p, nil
}
