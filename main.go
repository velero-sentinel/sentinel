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
	"github.com/hashicorp/go-hclog"
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

type debug bool

func (d debug) BeforeApply() error {
	appLogger.SetLevel(hclog.Debug)
	return nil
}

type cfgPath string

func (p cfgPath) AfterApply() error {
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
	appLogger hclog.Logger = hclog.New(&hclog.LoggerOptions{Name: "sentinel", Color: hclog.ForceColor})

	notifiers notification.NotifierConfig
	config    struct {
		Namespace      string  `kong:"required,short='n',default='velero'"`
		Debug          debug   `kong:"help='Enable debug logging',group='Logging'"`
		Bind           string  `kong:"required,default=':8080',group='Network',help='address to bind to'"`
		NotifierConfig cfgPath `kong:"type:'existingfile'"`
	}
)

func main() {

	ctx := kong.Parse(&config, kong.Bind(appLogger))

	ctx.Printf("Starting up...\n")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	p, err := pipeline.New(&notifiers, appLogger.Named("notification"))
	ctx.FatalIfErrorf(err)
	ctx.Printf("Configured pipeline\n")
	watcher, _ := setupClient()
	evts := watcher.ResultChan()
	s := server.New(p.Run(), appLogger.Named("server"))
	s.Run(evts)

	ctx.Printf("Started server")

	sig := <-sigs
	appLogger.Info("Received signal", "signal", sig.String())
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
