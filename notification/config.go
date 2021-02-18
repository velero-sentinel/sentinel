package notification

import "github.com/velero-sentinel/sentinel/notification/webhook"

type NotifierConfig struct {
	Webhooks []webhook.Config
}
