package notification

import (
	"errors"

	"github.com/hashicorp/go-hclog"
	"github.com/velero-sentinel/sentinel/message"
)

// TODO: Make pipeline an interface for easy testing
type Pipeline struct {
	logger     hclog.Logger
	upstream   chan message.Message
	downstream []chan<- message.Message
	done       chan bool
}

func (p Pipeline) Run() {
	if len(p.downstream) < 1 {
		panic(errors.New("downstream is empty"))
	}
	for {
		select {
		case <-p.done:
			for _, w := range p.downstream {
				close(w)
			}
			return
		case m := <-p.upstream:
			p.logger.Info("Received message from upstream")
			for _, c := range p.downstream {
				c <- m
			}
		}
	}
}

func (p *Pipeline) Stop() {
	p.logger.Warn("Received shutdown command")
	p.done <- true
	p.logger.Warn("Server shutdown complete")
}

func (p *Pipeline) Channel() chan message.Message {
	return p.upstream
}
