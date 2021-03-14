package opsgenie_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/incident"
	"github.com/stretchr/testify/assert"
	"github.com/velero-sentinel/sentinel/message"
	"github.com/velero-sentinel/sentinel/notification/opsgenie"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

func OverrideURI(apiUrl string) opsgenie.Option {
	return func(n *opsgenie.OpsGenieNotifier) error {
		u, err := url.ParseRequestURI(apiUrl)
		if err != nil {
			return fmt.Errorf("parsing OpsGenie API url: %s", err)
		}
		n.URL = client.ApiUrl(u.Host)
		return nil
	}
}

type MockHandler struct {
	wg *sync.WaitGroup
}

func (m *MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := incident.AsyncResult{}
	resp.RequestId = uuid.NewString()
	resp.Result = "created"
	enc := json.NewEncoder(w)
	enc.Encode(resp)
	m.wg.Done()
}

func TestRunMessages(t *testing.T) {
	testCases := []struct {
		desc    string
		msg     message.Message
		options []opsgenie.Option
	}{
		{
			desc:    "WarningMessage",
			msg:     warning(),
			options: []opsgenie.Option{opsgenie.NotifyOnWarning(true), opsgenie.NotifyStakeholders(false)},
		},
		{
			desc:    "WarningMessageNoNotify",
			msg:     warning(),
			options: []opsgenie.Option{opsgenie.NotifyOnWarning(false)},
		},
		{
			desc:    "ErrorMessage",
			msg:     errorMessage(),
			options: []opsgenie.Option{opsgenie.NotifyOnWarning(false), opsgenie.NotifyStakeholders(false)},
		},
		{
			desc:    "ErrorMessage with Tags",
			msg:     errorMessage(),
			options: []opsgenie.Option{opsgenie.Tags([]string{"foo", "bar", "baz"})},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var wg sync.WaitGroup
			server := httptest.NewServer(&MockHandler{&wg})
			defer server.Close()
			var opts = []opsgenie.Option{OverrideURI(server.URL)}
			opts = append(opts, tC.options...)
			n, err := opsgenie.New(uuid.NewString(), opts...)
			assert.NoError(t, err)
			errChan := make(chan error)
			c := n.Run(errChan)

			c <- tC.msg
			wg.Wait()
			var returnedError error
			select {
			case returnedError = <-errChan:
			case <-time.After(100 * time.Millisecond):
			}
			assert.NoError(t, returnedError)
		})
	}
}

func errorMessage() *message.Error {
	msg := new(message.Error)
	msg.Backup = new(v1.Backup)
	msg.Backup.Labels = map[string]string{"foo": "bar"}
	msg.Backup.Annotations = map[string]string{"backup.velero.io": "baz"}
	msg.Backup.Name = "testbackup"
	msg.Backup.Status.Phase = v1.BackupPhasePartiallyFailed

	return msg
}

func warning() *message.Warning {
	msg := new(message.Warning)
	msg.Backup = new(v1.Backup)
	msg.Backup.Labels = map[string]string{"foo": "bar"}
	msg.Backup.Annotations = map[string]string{"backup.velero.io": "baz"}
	msg.Backup.Name = "testbackup"
	msg.Backup.Status.Phase = v1.BackupPhasePartiallyFailed

	return msg
}
