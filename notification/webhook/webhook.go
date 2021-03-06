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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"text/template"

	"github.com/avast/retry-go"
	"github.com/hashicorp/go-hclog"
	"github.com/velero-sentinel/sentinel/message"
)

const (
	defaultWarnTemplateString = `
{
    "type": "warning",
    "backup":{
        "name": "{{.Name}}",
        "state": "{{.Status.Phase}}"
    }
}
`

	defaultErrTemplateString = `
{
    "type":"error",
    "backup":{
        "name": "{{.Name}}",
        "state": "{{.Status.Phase}}"
    }
}`
)

var defaultWarnTemplate = template.Must(template.New("warn").Parse(defaultWarnTemplateString))
var defaultErrTemplate = template.Must(template.New("error").Parse(defaultErrTemplateString))

type Config struct {
	Name            string
	URL             string
	Method          string
	WarningTemplate string
	ErrorTemplate   string
}

func New(cfg *Config, logger hclog.Logger) (*webhookNotifier, error) {
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %s", err)
	}
	n := &webhookNotifier{
		name:   cfg.Name,
		client: *http.DefaultClient,
		url:    u,
		method: cfg.Method,

		logger: logger,

		warnTmpl: defaultWarnTemplate,
		errTmpl:  defaultErrTemplate,
	}
	if cfg.WarningTemplate != "" {
		n.warnTmpl, err = template.New("warn").Parse(cfg.WarningTemplate)
		if err != nil {
			return nil, fmt.Errorf("parsing warningTemplate: %s", err)
		}
	}
	if cfg.ErrorTemplate != "" {
		n.errTmpl, err = template.New("error").Parse(cfg.ErrorTemplate)
		if err != nil {
			return nil, fmt.Errorf("parsing errorTemplate: %s", err)
		}
	}

	return n, nil
}

type webhookNotifier struct {
	name   string
	client http.Client
	url    *url.URL
	method string

	logger hclog.Logger

	warnTmpl *template.Template
	errTmpl  *template.Template
}

func (n *webhookNotifier) Run() chan<- message.Message {
	c := make(chan message.Message)
	go func() {
		n.logger.Info("Starting up")
		buf := bytes.NewBuffer(nil)
		for m := range c {
			buf.Reset()
			switch m.(type) {
			case *message.Warning, message.Warning:
				n.warnTmpl.Execute(buf, m.GetBackup())
			case *message.Error, message.Error:
				n.errTmpl.Execute(buf, m.GetBackup())
			}

			if err := retry.Do(
				func() error {
					return call(&n.client, n.method, n.url, buf.Bytes())
				},
				retry.OnRetry(
					func(num uint, err error) {
						n.logger.Warn("Sending webhook temporarily failed", "name", n.name, "url", n.url.String(), "error", err, "attempt", num+1)
					},
				),
				retry.Attempts(3),
				retry.LastErrorOnly(true),
			); err != nil {
				n.logger.Error("Sending webhook failed", "name", n.name, "url", n.url.String(), "error", err, "attempts", 3)
			}
		}
		n.logger.Info("Shut down")
	}()
	return c
}

func call(client *http.Client, method string, url *url.URL, b []byte) error {

	rc := ioutil.NopCloser(bytes.NewReader(b))

	req, err := http.NewRequest(method, url.String(), rc)
	if err != nil {
		return retry.Unrecoverable(err)
	}

	resp, err := client.Do(req)
	if resp == nil || resp.StatusCode > 399 {
		return errors.New("Request unsuccessful")
	}
	return err
}
