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

// Config represents a WebhookConfiguration. For example:
//
//    webhooks:
//			- name: "test"
//				url: "http://www.example.com"
//				method: "POST"
//				template: &tmpl |
//				{{ template "slack" . }}
type Config struct {
	// Name is a unique alias for the webhook. It is mainly used for logging purposes.
	Name string

	// The URL to call as a string in the format (http|https)://(host)[:port][/path]
	URL string

	// Method is the HTTP method to be used. Generics (GET, POST, etc.) as well as extensions are supported.
	Method string

	// The template to be used. e.g:
	Template string
}
