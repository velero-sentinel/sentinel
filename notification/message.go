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

// Message is the message sent via a Notifier.
type Message interface {
	Message() string
}

type ErrorMessage string

func (e ErrorMessage) Message() string {
	return string(e)
}

type WarningMessage string

func (w WarningMessage) Message() string {
	return string(w)
}