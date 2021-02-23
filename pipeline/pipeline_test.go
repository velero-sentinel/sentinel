package pipeline

import (
	"net/http"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/velero-sentinel/sentinel/message"
	"github.com/velero-sentinel/sentinel/notification"
	"github.com/velero-sentinel/sentinel/notification/webhook"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewFail(t *testing.T) {
	log, _ := test.NewNullLogger()
	p, err := New(&notification.NotifierConfig{
		Webhooks: []webhook.Config{webhook.Config{
			Name:   "invalidURL",
			URL:    "http://ä<<>>@!/foo",
			Method: http.MethodGet,
		}}}, log,
	)
	assert.Nil(t, p)
	assert.Error(t, err)
}

func TestNewDownstreamSetup(t *testing.T) {
	testCases := []struct {
		desc          string
		cfg           *notification.NotifierConfig
		expectedChans int
	}{
		{
			desc:          "Without Webhooks",
			cfg:           &notification.NotifierConfig{},
			expectedChans: 1,
		},
		{
			desc: "With Webhook",
			cfg: &notification.NotifierConfig{
				Webhooks: []webhook.Config{
					webhook.Config{
						Name:   "test",
						URL:    "http://example.com",
						Method: http.MethodGet}},
			},
			expectedChans: 2,
		},
	}
	for _, tC := range testCases {
		log, _ := test.NewNullLogger()
		t.Run(tC.desc, func(t *testing.T) {
			p, err := New(tC.cfg, log)
			assert.NoError(t, err)
			assert.NotNil(t, p)
			assert.NotNil(t, p.downstream)
			assert.Equalf(t, tC.expectedChans, len(p.downstream), "channels %s", spew.Sprint(p.downstream))
		})
	}
}
func TestRun(t *testing.T) {
	c1 := make(chan message.Message)
	c2 := make(chan message.Message)
	log, _ := test.NewNullLogger()
	p := pipeline{
		logger:     log,
		downstream: []chan<- message.Message{c1, c2},
	}

	d := p.Run()
	m := message.Warning{Backup: &v1.Backup{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}}
	d <- m
	var m1, m2 message.Message
loop:
	for {
		select {
		case m1 = <-c1:
			assert.EqualValues(t, m, m1)
		case m2 = <-c2:
			assert.EqualValues(t, m, m2)
		case <-time.After(1 * time.Second):
			close(d)
			break loop
		}
	}
	assert.NotNil(t, m1)
	assert.NotNil(t, m2)
	for _, c := range []chan message.Message{c1, c2} {
		_, open := <-c
		assert.False(t, open)
	}
}
