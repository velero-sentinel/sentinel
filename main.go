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
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
	"github.com/velero-sentinel/sentinel/notification"
	"github.com/velero-sentinel/sentinel/pipeline"
	"github.com/vmware-tanzu/velero/pkg/client"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/velero-sentinel/sentinel/server"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	veleroNamespace = "velero"
)

// Build time variables
var (
	Version   string
	GitCommit string
	BuildDate string
)

type jsonlog bool

func (j jsonlog) BeforeApply() error {
	appLogger.SetFormatter(&logrus.JSONFormatter{})
	return nil
}

type debug bool

func (d debug) BeforeApply() error {
	appLogger.SetLevel(logrus.DebugLevel)
	return nil
}

// CfgPath denotes the path to the YAML configuration file for notifiers.
type CfgPath string

// AfterApply is a kong callback. After kong verified that the path indeed an existing file,
// the YAML configuration is parsed into the notifier config.
func (p CfgPath) AfterApply() error {
	if p == "" {
		return nil
	}

	f, err := os.Open(string(p))
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewDecoder(f).Decode(&notifiers)
}

var (
	appLogger = &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.InfoLevel,
		Formatter: &logrus.TextFormatter{
			DisableSorting:         false,
			DisableLevelTruncation: true,
			DisableTimestamp:       false,
		},
	}

	notifiers notification.NotifierConfig
	config    struct {
		Namespace      string  `kong:"required,short='n',default='velero'"`
		Debug          debug   `kong:"help='Enable debug logging',group='Logging'"`
		JSON           jsonlog `kong:"help='output log in JSON',group='Logging'"`
		Bind           string  `kong:"required,default=':8080',group='Network',help='address to bind to'"`
		NotifierConfig CfgPath `kong:"type:'existingfile'"`
	}
)

var versionInfo = logrus.Fields{
	"version":    Version,
	"revision":   GitCommit,
	"build-date": BuildDate,
}

func main() {

	kong.Parse(&config, kong.Bind(appLogger))

	appLogger.WithFields(versionInfo).Info("Starting up...")

	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	p, err := pipeline.New(&notifiers, appLogger)
	if err != nil {
		appLogger.WithError(err).Fatal("Setting up pipeline")
	}

	appLogger.Info("Configured pipeline")
	watcher, err := setupClient()

	appLogger.WithError(err).Fatal("Client could not be set up")

	evts := watcher.ResultChan()
	s := server.New(p.Run(), appLogger)
	s.Run(evts)

	appLogger.Info("Started server")

	sig := <-sigs
	appLogger.WithField("signal", sig.String()).Info("Received signal")
	watcher.Stop()
	appLogger.Info("Saying goodbye!")
}

func setupClient() (watch.Interface, error) {
	cfg, err := client.LoadConfig()
	factory := client.NewFactory("sentinel", cfg)
	myclient, err := factory.Client()

	if err != nil {
		return nil, fmt.Errorf("setting up client: %s", err)
	}

	return myclient.VeleroV1().Backups(veleroNamespace).Watch(context.Background(), metav1.ListOptions{})
}
