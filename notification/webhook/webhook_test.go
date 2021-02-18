package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/velero-sentinel/sentinel/message"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

// func TestRequestValidity(t *testing.T) {

// 	r, err := http.NewRequest("asdf", "http://www.example.com", nil)

// 	if err == nil {
// 		t.Error("Invalid request method was accepted")
// 	}

// 	if r != nil {
// 		t.Error("Request was returned")
// 	}

// }

func TestRetry(t *testing.T) {
	testCases := []struct {
		desc           string
		fails          int
		invalidRequest bool
		requests       int
	}{
		{
			desc:           "Irrecoverable",
			fails:          0,
			invalidRequest: true,
			requests:       0,
		},
		{
			desc:           "Recovers",
			fails:          1,
			invalidRequest: false,
			requests:       2,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {

			numrequest := 0

			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if tC.requests == 0 {
						t.Error("Got request where there should be none")
						t.FailNow()
					}
					dec := json.NewDecoder(r.Body)
					result := make(map[string]interface{})
					assert.NoError(t, dec.Decode(&result))
					assert.NotNil(t, result["type"], "Result type should not be nil")
					assert.Equal(t, "warning", result["type"])
					numrequest++
					if numrequest <= tC.fails {
						http.Error(w, "expected error", http.StatusInternalServerError)
					}
				}))
			defer srv.Close()

			u, err := url.Parse(srv.URL)
			assert.NoError(t, err)
			assert.NotNil(t, u)

			if tC.invalidRequest {
				u = &url.URL{Host: "http://ä<<>>@!/foo"}
			}
			wc := make(chan message.WarningMessage)
			ec := make(chan message.ErrorMessage)
			done := make(chan bool)
			n := webhookNotifier{
				name:     "testWarn",
				logger:   hclog.NewNullLogger(),
				client:   *http.DefaultClient,
				warnTmpl: defaultWarnTemplate,
				errTmpl:  defaultErrTemplate,
				url:      u,
				method:   http.MethodPost,
				warnings: wc,
				errors:   ec,
				done:     done,
			}
			go func() {
				msg := new(message.WarningMessage)
				msg.Backup = new(v1.Backup)
				assert.NotNil(t, msg.Backup)
				msg.Backup.Name = "testbackup"
				msg.Backup.Status.Phase = v1.BackupPhasePartiallyFailed
				n.warnings <- *msg
				done <- true
			}()
			n.run()
			assert.EqualValues(t, tC.requests, numrequest)
		})
	}
}

func TestWebhooks(t *testing.T) {
	wc := make(chan message.WarningMessage)
	ec := make(chan message.ErrorMessage)
	testCases := []struct {
		desc  string
		kind  string
		tmpl  *template.Template
		phase v1.BackupPhase
		warnC chan message.WarningMessage
		errC  chan message.ErrorMessage
	}{
		{
			desc:  "Warning Webhook",
			kind:  "warning",
			tmpl:  defaultWarnTemplate,
			phase: v1.BackupPhasePartiallyFailed,
			warnC: wc,
			errC:  ec,
		},
		{
			desc:  "Error Webhook",
			kind:  "error",
			tmpl:  defaultErrTemplate,
			phase: v1.BackupPhaseFailed,
			warnC: wc,
			errC:  ec,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {

			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					dec := json.NewDecoder(r.Body)
					result := make(map[string]interface{})
					assert.NoError(t, dec.Decode(&result))
					assert.NotNil(t, result["type"])
					assert.Equal(t, tC.kind, result["type"])
				}))
			defer srv.Close()
			u, err := url.Parse(srv.URL)
			assert.NoError(t, err)
			assert.NotNil(t, u)

			done := make(chan bool)
			n := webhookNotifier{
				name:     tC.desc,
				logger:   hclog.NewNullLogger(),
				client:   *http.DefaultClient,
				warnTmpl: defaultWarnTemplate,
				errTmpl:  defaultErrTemplate,
				url:      u,
				method:   http.MethodPost,
				warnings: wc,
				errors:   ec,
				done:     done,
			}
			go func() {

				switch tC.kind {
				case "warning":
					msg := new(message.WarningMessage)
					msg.Backup = new(v1.Backup)
					assert.NotNil(t, msg.Backup)
					msg.Backup.Name = "testbackup"
					msg.Backup.Status.Phase = tC.phase
					n.WarningC() <- *msg
				case "error":
					msg := new(message.ErrorMessage)
					msg.Backup = new(v1.Backup)
					assert.NotNil(t, msg.Backup)
					msg.Backup.Name = "testbackup"
					msg.Backup.Status.Phase = tC.phase
					n.ErrorC() <- *msg
				}

				done <- true
			}()
			n.run()
		})
	}
}

func TestWarning(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			dec := json.NewDecoder(r.Body)
			result := make(map[string]interface{})
			assert.NoError(t, dec.Decode(&result))
			assert.NotNil(t, result["type"])
			assert.Equal(t, "warning", result["type"])
		}))

	defer srv.Close()
	u, err := url.Parse(srv.URL)
	assert.NoError(t, err)
	assert.NotNil(t, u)

	wc := make(chan message.WarningMessage)
	ec := make(chan message.ErrorMessage)
	done := make(chan bool)
	n := webhookNotifier{
		name:     "testWarn",
		logger:   hclog.NewNullLogger(),
		client:   *http.DefaultClient,
		warnTmpl: defaultWarnTemplate,
		errTmpl:  defaultErrTemplate,
		url:      u,
		method:   http.MethodPost,
		warnings: wc,
		errors:   ec,
		done:     done,
	}
	go func() {
		msg := new(message.WarningMessage)
		msg.Backup = new(v1.Backup)
		assert.NotNil(t, msg.Backup)
		msg.Backup.Name = "testbackup"
		msg.Backup.Status.Phase = v1.BackupPhasePartiallyFailed
		n.warnings <- *msg
		done <- true
	}()
	n.run()
}
