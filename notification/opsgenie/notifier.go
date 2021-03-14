package opsgenie

import (
	"context"
	"errors"
	"fmt"

	validator "github.com/asaskevich/govalidator"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/incident"
	"github.com/sirupsen/logrus"
	"github.com/velero-sentinel/sentinel/message"
)

type OpsGenieNotifier struct {
	Client             *incident.Client
	NotifyOnWarning    bool
	NotifyStakeholders bool
	Tags               []string
	logger             *logrus.Logger
	Retries            int
	ApiKey             string        `valid:"uuid"`
	URL                client.ApiUrl `valid:"url"`
}

var MissingApiKeyError = errors.New("Missing API Key")

func New(apiKey string, options ...Option) (*OpsGenieNotifier, error) {

	if apiKey == "" {
		return nil, MissingApiKeyError
	}

	n := &OpsGenieNotifier{
		ApiKey:             apiKey,
		NotifyOnWarning:    false,
		NotifyStakeholders: false,
	}

	var (
		err   error
		valid bool
	)

	// Apply functional options
	for _, opt := range options {
		if err := opt(n); err != nil {
			return nil, fmt.Errorf("configuring notifier: %s", err)
		}
	}

	// Ensure that the notifier satisfies the minimum requirements

	if valid, err = validator.ValidateStruct(n); !valid {
		return nil, fmt.Errorf("validating notifier: %s", err)
	}

	cfg := client.Config{
		ApiKey:     n.ApiKey,
		RetryCount: n.Retries,
	}

	if n.Client, err = incident.NewClient(&cfg); err != nil {
		return nil, fmt.Errorf("creating OpsGenie incident client: %s", err)
	}

	return n, nil
}

func (n OpsGenieNotifier) Run(errChan chan<- error) chan<- message.Message {
	in := make(chan message.Message)

	go func() {
	messages:
		for m := range in {
			r := &incident.CreateRequest{
				Details: make(map[string]string),
			}
			if n.Tags != nil {
				r.Tags = n.Tags
			}
			r.NotifyStakeholders = &n.NotifyStakeholders

			if m.GetBackup().GetAnnotations() != nil {
				r.Details = m.GetBackup().GetAnnotations()
			}

			for k, v := range m.GetBackup().Labels {
				r.Details[k] = v
			}
			r.ServiceId = "test"

			switch m.(type) {
			case *message.Warning, message.Warning:
				if n.NotifyOnWarning == false {
					continue messages
				}
				r.Priority = incident.P5
				r.Message = fmt.Sprintf("%s failed partially", m.GetBackup().Name)
			case *message.Error, message.Error:
				r.Priority = incident.P1
				r.Message = fmt.Sprintf("%s failed", m.GetBackup().Name)
			}

			_, err := n.Client.Create(context.Background(), r)
			if err != nil {
				errChan <- err
			}
		}
	}()
	return in
}
