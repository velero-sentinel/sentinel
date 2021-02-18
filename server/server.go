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

package server

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"github.com/vmware-tanzu/velero/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/velero-sentinel/sentinel/message"
)

type server struct {
	logger     hclog.Logger
	factory    client.Factory
	downstream chan<- message.Message
	done       chan bool
}

func New(logger hclog.Logger) (*server, error) {
	s := &server{logger: logger, done: make(chan bool)}
	cfg, err := client.LoadConfig()
	if err != nil {
		return nil, err
	}
	s.factory = client.NewFactory("sentinel", cfg)
	return s, nil
}

func (s *server) Run() error {
	s.logger.Info("Setting up client")
	myclient, err := s.factory.Client()

	if err != nil {
		return fmt.Errorf("setting up client: %s", err)
	}

	watcher, err := myclient.VeleroV1().Backups("velero").Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("setting up watcher: %s", err)
	}
	for {
		select {
		case <-s.done:
			return nil
		case evt := <-watcher.ResultChan():

			backup, ok := evt.Object.(*v1.Backup)

			if !ok {
				s.logger.Error("Non-Backup event registered", "evt", backup)
				continue
			}

			switch evt.Type {
			case watch.Added:
				s.logger.Info("Backup added", "name", backup.Name)
				continue
			case watch.Deleted:
				s.logger.Info("Backup deleted", "name", backup.Name)
				continue
			}

			switch backup.Status.Phase {
			case v1.BackupPhaseNew:
				// There IS a state "new"... However, I do not think this actually will come
				// up a lot, unless you have many potentially concurrent backup
				s.logger.Info("New backup detected", "name", backup.Name, "state", evt.Type)
			case v1.BackupPhaseCompleted:
				s.logger.Info("Backup completed", "name", backup.Name, "state", evt.Type)
			case v1.BackupPhaseDeleting:
				s.logger.Info("Backup deletion", "name", backup.Name, "state", evt.Type)
			case v1.BackupPhaseInProgress:
				s.logger.Info("Backup in progress", "name", backup.Name, "state", evt.Type)
			case v1.BackupPhasePartiallyFailed:
				s.downstream <- message.WarningMessage{Backup: backup}
			case v1.BackupPhaseFailed:
				s.downstream <- message.ErrorMessage{Backup: backup}
			}

		}
	}
}

func (s *server) Stop() {
	s.logger.Warn("Received shutdown command")
	s.done <- true
	s.logger.Warn("Server shutdown complete")
}
