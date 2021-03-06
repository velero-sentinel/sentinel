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

package message

import (
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

// Message is the message sent via a Notifier.
type Message interface {
	GetBackup() *v1.Backup
}

// Error is sent to a notifier in case a backup has failed.
type Error struct {
	Backup *v1.Backup
}

// GetBackup returns the Velero backup that caused the error message.
//
// Satisfies the Message interface.
func (e Error) GetBackup() *v1.Backup {
	return e.Backup
}

// Warning is sent to a notifier in case a backup is partially failed.
type Warning struct {
	Backup *v1.Backup
}

// GetBackup returns the Velero backup that caused the error message.
//
// Satisfies the Message interface.
func (w Warning) GetBackup() *v1.Backup {
	return w.Backup
}
