package webhook_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/velero-sentinel/sentinel/message"
	"github.com/velero-sentinel/sentinel/notification"
	"github.com/velero-sentinel/sentinel/notification/webhook"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testYaml = `
webhooks:
  - name: "test"
    url: "http://www.example.com"
    method: "GET"
    warningTemplate: &tmpl |
      {
        "action": "test",
        "name": "{{ .Name }}",
        "message": "{{ .Name}} is in {{ .Status.Phase }}"
      }
    errorTemplate: *tmpl
`

type expectedJson struct {
	Action  string
	Name    string
	Message string
}

var partiallyFailedBackup = v1.Backup{
	ObjectMeta: metav1.ObjectMeta{Name: "testPartiallyFailed"},
	Status:     v1.BackupStatus{Phase: v1.BackupPhasePartiallyFailed},
}

var failedBackup = v1.Backup{
	ObjectMeta: metav1.ObjectMeta{Name: "testFailedbackup"},
	Status:     v1.BackupStatus{Phase: v1.BackupPhaseFailed},
}

var warnMessage = message.Warning{Backup: &partiallyFailedBackup}

var errorMessage = message.Error{Backup: &failedBackup}

func TestTemplateProcessingEndToEnd(t *testing.T) {
	testCases := []struct {
		desc    string
		yaml    string
		message message.Message
		payload interface{}
	}{
		{
			desc: "AnchorWarning",
			yaml: testYaml,
			// backup: partiallyFailedBackup,
			message: message.Warning{Backup: &partiallyFailedBackup},
			payload: expectedJson{
				Action:  "test",
				Name:    partiallyFailedBackup.Name,
				Message: fmt.Sprintf("%s is in %s", partiallyFailedBackup.Name, partiallyFailedBackup.Status.Phase),
			},
		},
		{
			desc:    "AnchoredError",
			yaml:    testYaml,
			message: message.Error{Backup: &failedBackup},
			payload: expectedJson{
				Action:  "test",
				Name:    failedBackup.Name,
				Message: fmt.Sprintf("%s is in %s", failedBackup.Name, failedBackup.Status.Phase),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			notifiers := notification.NotifierConfig{}
			assert.NoError(t, yaml.Unmarshal([]byte(tC.yaml), &notifiers))
			assert.EqualValues(t, 1, len(notifiers.Webhooks))

			hook := notifiers.Webhooks[0]
			assert.NotEmpty(t, hook.WarningTemplate, "config: %s", spew.Sdump(notifiers))
			assert.NotEmpty(t, hook.ErrorTemplate, "config: %s", spew.Sdump(notifiers))

			done := make(chan bool)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() { done <- true }()

				p, err := ioutil.ReadAll(r.Body)
				assert.NoError(t, err, "reading body: %s", err)

				result := reflect.New(reflect.TypeOf(tC.payload)).Interface()
				assert.NoError(t, json.Unmarshal(p, &result))

				e := fmt.Sprintf("%s", reflect.ValueOf(result).Elem())
				pl := fmt.Sprintf("%s", reflect.ValueOf(tC.payload))
				assert.EqualValues(t, pl, e, "%s", spew.Sprintf("e: %s,\npl:%s", e, pl))
			}))
			defer srv.Close()

			notifiers.Webhooks[0].URL = srv.URL
			h, err := webhook.New(&notifiers.Webhooks[0], hclog.NewNullLogger())
			assert.NoError(t, err)

			c := h.Run()
			c <- tC.message
			<-done
		})
	}
}