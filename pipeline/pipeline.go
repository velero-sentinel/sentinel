package pipeline

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/velero-sentinel/sentinel/message"
	"github.com/velero-sentinel/sentinel/notification"
	"github.com/velero-sentinel/sentinel/notification/lognotifier"
	"github.com/velero-sentinel/sentinel/notification/webhook"
)

type pipeline struct {
	logger     hclog.Logger
	downstream []chan<- message.Message
}

func New(cfg *notification.NotifierConfig, logger hclog.Logger) (*pipeline, error) {
	p := &pipeline{
		logger:     logger,
		downstream: make([]chan<- message.Message, 0),
	}

	if len(cfg.Webhooks) > 0 {
		webhooks, err := configureWebhooks(cfg.Webhooks, logger.Named("webhook"))
		if err != nil {
			return nil, fmt.Errorf("configuring webhooks: %s", err)
		}
		p.downstream = append(p.downstream, webhooks...)
	}
	ln := lognotifier.New(logger.Named("log"))
	p.downstream = append(p.downstream, ln.Run())
	return p, nil
}

func configureWebhooks(cfg []webhook.Config, webhookLogger hclog.Logger) ([]chan<- message.Message, error) {

	hooks := make([]chan<- message.Message, len(cfg))

	for n, c := range cfg {
		h, err := webhook.New(&c, webhookLogger.With("webhook", c.Name))
		if err != nil {
			return nil, fmt.Errorf("configuring webhook '%s':%s ", c.Name, err)
		}
		hooks[n] = h.Run()
	}
	return hooks, nil
}

func (p pipeline) Run() chan<- message.Message {
	up := make(chan message.Message)
	go func() {
		for m := range up {
			p.logger.Info("Received message from upstream")
			for _, c := range p.downstream {
				c <- m
			}
		}
		p.logger.Info("Upstream was closed. Closing downstream")
		for _, c := range p.downstream {
			close(c)
		}
		p.logger.Info("Closed all downstream channels")
	}()
	return up
}
