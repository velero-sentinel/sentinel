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
	"fmt"

	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

// Message is the message sent via a Notifier.
type Message interface {
	Message() string
	GetBackup() *v1.Backup
}

// ErrorMessage is sent to a notifier in case a backup has failed.
type ErrorMessage struct {
	Backup *v1.Backup
}

func (e ErrorMessage) Message() string {
	return fmt.Sprintf("%s is in state %s", e.Backup.Name, e.Backup.Status.Phase)
}

func (e ErrorMessage) GetBackup() *v1.Backup {
	return e.Backup
}

type WarningMessage struct {
	Backup *v1.Backup
}

func (w WarningMessage) Message() string {
	return fmt.Sprintf("%s is in state %s", w.Backup.Name, w.Backup.Status.Phase)
}

func (w WarningMessage) GetBackup() *v1.Backup {
	return w.Backup
}
