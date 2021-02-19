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

package main

import (
	"errors"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/go-hclog"
	"github.com/velero-sentinel/sentinel/notification"
	"github.com/velero-sentinel/sentinel/server"
	"gopkg.in/yaml.v2"
)

var notifierTypes = []string{"log", "webhook"}

type NotifierType int

func (n *NotifierType) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	for i, v := range notifierTypes {
		if v == s {
			*n = NotifierType(i)
			return nil
		}
	}
	return errors.New("Unknown type of notifier: " + s)
}

const (
	Log NotifierType = iota
	Webhook
)

type debug bool

func (d debug) BeforeApply() error {
	appLogger.SetLevel(hclog.Debug)
	return nil
}

var (
	appLogger hclog.Logger
	notifiers notification.NotifierConfig
	config    struct {
		Namespace      string `kong:"required,short='n',default='velero'"`
		Debug          debug  `kong:"help='Enable debug logging',group='Logging'"`
		Bind           string `kong:"required,default=':8080',group='Network',help='address to bind to'"`
		NotifierConfig string `kong:"type:'existingfile'"`
	}
)

func main() {

	appLogger = hclog.New(&hclog.LoggerOptions{Name: "sentinel", Color: hclog.ForceColor})

	// ctx := kong.Parse(&config, kong.Configuration(kongyaml.Loader, "./sentinel.yml"))
	_ = kong.Parse(&config, kong.Bind(appLogger))

	appLogger.Info("Starting up")

	if config.NotifierConfig != "" {
		appLogger.Debug("Reading notfier config", "file", config.NotifierConfig)
		f, _ := os.Open(config.NotifierConfig)
		dec := yaml.NewDecoder(f)

		if err := dec.Decode(&notifiers); err != nil {
			panic(err)
		}
		if appLogger.IsDebug() {
			appLogger.Debug("Read notifier configuration", "file", config.NotifierConfig)
			spew.Dump(notifiers)
		}
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	p, _ := notification.Configure(&notifiers, appLogger.Named("notification"))
	go p.Run()

	ch := p.Channel()
	if ch == nil {
		panic("Pipeline channel is nil")
	}

	appLogger.Info("Started pipeline")

	s, _ := server.New(p.Channel(), appLogger.Named("server"))
	s.Run()
	appLogger.Info("Started server")

	sig := <-sigs
	appLogger.Info("Received signal", "signal", sig.String())
	s.Stop()
	p.Stop()

	appLogger.Info("Saying goodbye!")
}
