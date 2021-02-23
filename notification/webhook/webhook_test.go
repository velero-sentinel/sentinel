package webhook

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/velero-sentinel/sentinel/message"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestInvalidUrl(t *testing.T) {

	log, _ := test.NewNullLogger()

	n, err := New(&Config{Name: "invalidURL", URL: "http://ä<<>>@!/foo"}, log)

	assert.Nil(t, n)
	assert.Error(t, err)

}

const illegalTemplate = `{{.Name`

const warnString = `
{
	"type": "warning",
	"name": "{{.Backup.Name}}",
	"phase": "{{.Backup.Status.Phase}}"
}
`

const errString = `
{
	"type": "error",
	"name": "{{.Backup.Name}}",
	"phase": "{{.Backup.Status.Phase}}"
}
`

func TestTemplates(t *testing.T) {
	testCases := []struct {
		desc               string
		template           string
		message            message.Message
		expectedBackupName string
		failParse          bool
	}{
		{
			desc:     "CustomWarning",
			template: warnString,
			message: message.Warning{
				Backup: &v1.Backup{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Status:     v1.BackupStatus{Phase: v1.BackupPhasePartiallyFailed},
				}},
			expectedBackupName: "test",
		},
		{
			desc:     "CustomError",
			template: errString,
			message: message.Error{Backup: &v1.Backup{
				ObjectMeta: metav1.ObjectMeta{Name: "error"},
				Status:     v1.BackupStatus{Phase: v1.BackupPhaseFailed},
			}},
			expectedBackupName: "error",
		},
		{
			desc:      "FailParseWarn",
			template:  illegalTemplate,
			failParse: true,
		},
		{
			desc:      "FailParseError",
			template:  illegalTemplate,
			failParse: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			processed := make(chan bool)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				assert.Equal(t, http.MethodPost, r.Method)

				buf, err := ioutil.ReadAll(r.Body)
				r.Body.Close()
				assert.NoError(t, err)
				assert.NotNil(t, buf)
				rd := bytes.NewReader(buf)
				dec := json.NewDecoder(rd)
				v := make(map[string]interface{})
				err = dec.Decode(&v)

				assert.NoError(t, err, "Error parsing: '%s'", string(buf))
				assert.NotNil(t, v["name"])
				assert.EqualValues(t, tC.expectedBackupName, v["name"].(string))
				processed <- true
			}))
			c := Config{
				Name:     tC.desc,
				URL:      srv.URL,
				Method:   http.MethodPost,
				Template: tC.template,
			}
			log, _ := test.NewNullLogger()
			w, err := New(&c, log)
			if tC.failParse {
				assert.Error(t, err)
				assert.Nil(t, w)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, w)

			d := w.Run()
			d <- tC.message
			<-processed
		})
	}
}

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
					numrequest++
					if numrequest <= tC.fails {
						http.Error(w, "expected error", http.StatusInternalServerError)
					}
				}))
			defer srv.Close()

			cfg := &Config{Name: tC.desc, URL: srv.URL, Method: http.MethodPost}
			log, _ := test.NewNullLogger()
			n, _ := New(cfg, log)
			c := n.Run()
			msg := new(message.Warning)
			msg.Backup = new(v1.Backup)
			assert.NotNil(t, msg.Backup)
			msg.Backup.Name = "testbackup"
			msg.Backup.Status.Phase = v1.BackupPhasePartiallyFailed
			c <- msg

			if tC.fails > 0 {
				time.Sleep(2 * time.Second)
			}
			assert.EqualValues(t, tC.requests, numrequest)
			close(c)
		})
	}
}

func TestWebhooks(t *testing.T) {

	testCases := []struct {
		desc  string
		kind  string
		tmpl  *template.Template
		phase v1.BackupPhase
	}{
		{
			desc:  "Warning Webhook",
			kind:  "warning",
			tmpl:  defaultTemplate,
			phase: v1.BackupPhasePartiallyFailed,
		},
		{
			desc:  "Error Webhook",
			kind:  "error",
			tmpl:  defaultTemplate,
			phase: v1.BackupPhaseFailed,
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

			cfg := &Config{Name: tC.desc, URL: srv.URL, Method: http.MethodGet}
			log, _ := test.NewNullLogger()
			n, _ := New(cfg, log)

			c := n.Run()

			switch tC.kind {
			case "warning":
				msg := new(message.Warning)
				msg.Backup = new(v1.Backup)
				assert.NotNil(t, msg.Backup)
				msg.Backup.Name = "testbackup"
				msg.Backup.Status.Phase = tC.phase
				c <- *msg
			case "error":
				msg := new(message.Error)
				msg.Backup = new(v1.Backup)
				assert.NotNil(t, msg.Backup)
				msg.Backup.Name = "testbackup"
				msg.Backup.Status.Phase = tC.phase
				c <- *msg
			}
		})
	}
}
