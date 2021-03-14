package opsgenie

import (
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type Option func(*OpsGenieNotifier) error

func EU() Option {
	return func(n *OpsGenieNotifier) error {
		n.URL = client.API_URL_EU
		return nil
	}
}

func Sandbox() Option {
	return func(n *OpsGenieNotifier) error {
		n.URL = client.API_URL_SANDBOX
		return nil
	}
}

func NotifyStakeholders(notify bool) Option {
	return func(n *OpsGenieNotifier) error {
		n.NotifyStakeholders = notify
		return nil
	}
}

func NotifyOnWarning(notify bool) Option {
	return func(n *OpsGenieNotifier) error {
		n.NotifyOnWarning = notify
		return nil
	}
}

func Tags(tags []string) Option {
	return func(n *OpsGenieNotifier) error {
		n.Tags = append(n.Tags, tags...)
		return nil
	}
}

func Retries(number int) Option {
	return func(n *OpsGenieNotifier) error {
		n.Retries = number
		return nil
	}
}
