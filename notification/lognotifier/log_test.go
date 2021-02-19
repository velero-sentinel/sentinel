package lognotifier

import (
	"testing"
	"time"

	mocked "github.com/mwmahlberg/hclog-mock"
	"github.com/stretchr/testify/mock"
	"github.com/velero-sentinel/sentinel/message"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLogNotifier(t *testing.T) {
	testCases := []struct {
		desc   string
		msg    message.Message
		method string
		number int
	}{
		{
			desc:   "warning once",
			msg:    &message.Warning{Backup: &v1.Backup{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			method: "Warn",
			number: 1,
		},
		{
			desc:   "error once",
			msg:    &message.Error{Backup: &v1.Backup{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			method: "Error",
			number: 1,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			m := &mocked.Logger{}
			m.On("Debug", mock.Anything).Maybe().Return(nil)
			m.On(tC.method, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Once()

			n := New(m)
			c := n.Run()
			c <- tC.msg
			time.Sleep(200 * time.Millisecond)
			defer close(c)
			m.AssertExpectations(t)
		})
	}
}
