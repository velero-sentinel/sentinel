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

package cmd

import (
	"github.com/hashicorp/go-hclog"
	"github.com/velero-sentinel/sentinel/notification"
)

type hclogNotifier struct {
	logger hclog.Logger
}

func (n hclogNotifier) Warn(m notification.Message) {
	n.logger.Warn(m.Message())
}

func (n hclogNotifier) Error(m notification.Message) {
	n.logger.Error(m.Message())
}

// LogNotifier is a simple Notifier that logs the warnings.
func LogNotifier(logger hclog.Logger) notification.Notifier {
	return hclogNotifier{logger: logger}
}
