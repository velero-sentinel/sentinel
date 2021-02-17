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

package webhook

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"text/template"

	"github.com/avast/retry-go"
	"github.com/hashicorp/go-hclog"
	"github.com/velero-sentinel/sentinel/notification"
)

const defaultWarnTemplate = `
{
    "type": "warning",
    "backup":{
        "name": "{{.Name}}",
        "state": "{{.Status.Phase}}"
    }
}
`

const defaultErrTemplate = `
{
    "type":"error",
    "backup":{
        "name": "{{.Name}}",
        "state": "{{.Status.Phase}}"
    }
}
`

// func New(cfg map[string]interface{}) (Notifier, error) {

// }

type webhookNotifier struct {
	name     string
	logger   hclog.Logger
	client   http.Client
	warnTmpl *template.Template
	errTmpl  *template.Template
	url      *url.URL
	method   string
	warnings chan notification.WarningMessage
	errors   chan notification.ErrorMessage
	done     chan bool
}

func (n *webhookNotifier) WarningC() chan<- notification.WarningMessage {
	return n.warnings
}

func (n *webhookNotifier) ErrorC() chan<- notification.ErrorMessage {
	return n.errors
}

func (n *webhookNotifier) run() {
	buf := bytes.NewBuffer(nil)
	for {
		buf.Reset()
		select {
		case <-n.done:
			return
		case m := <-n.warnings:
			n.warnTmpl.Execute(buf, m.Backup)
		case m := <-n.errors:
			n.errTmpl.Execute(buf, m.Backup)
		}
		br := bytes.NewReader(buf.Bytes())
		rc := ioutil.NopCloser(br)

		err := retry.Do(
			func() error {
				req, err := http.NewRequest(n.method, n.url.String(), rc)
				if err != nil {
					return retry.Unrecoverable(err)
				}
				resp, err := n.client.Do(req)

				if resp == nil || resp.StatusCode > 399 {
					br.Seek(0, io.SeekStart)
					return errors.New("Request unsuccessful")
				}
				return err
			},
			retry.OnRetry(
				func(num uint, err error) {
					n.logger.Warn("Sending webhook temporarily failed", "name", n.name, "url", n.url.String(), "error", err, "attempt", num+1)
				},
			),

			retry.Attempts(3),
			retry.LastErrorOnly(true),
		)
		if err != nil {
			n.logger.Error("Sending webhook failed", "name", n.name, "url", n.url.String(), "error", err, "attempts", 3)
		}
	}
}
