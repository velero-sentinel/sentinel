package lognotifier

import (
	"io"
	"log"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/velero-sentinel/sentinel/message"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type loggerMock struct {
	mock.Mock
}

func (m *loggerMock) Name() string {
	rv := m.Called()
	return rv.String(0)
}

func (m *loggerMock) Named(n string) hclog.Logger {
	rv := m.Called()
	return rv.Get(0).(hclog.Logger)
}

func (m *loggerMock) ResetNamed(n string) hclog.Logger {
	rv := m.Called(n)
	return rv.Get(0).(hclog.Logger)
}

func (m *loggerMock) SetLevel(lv hclog.Level) {
	m.Called()
}

func (m *loggerMock) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	rv := m.Called(opts)
	return rv.Get(0).(*log.Logger)
}

func (m *loggerMock) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	rv := m.Called(opts)
	return rv.Get(0).(io.Writer)
}

func (m *loggerMock) With(args ...interface{}) hclog.Logger {
	rv := m.Called(args...)
	return rv.Get(0).(hclog.Logger)
}

func (m *loggerMock) Log(level hclog.Level, msg string, args ...interface{}) {
	m.Called()
}

func (m *loggerMock) Trace(msg string, args ...interface{}) {
	m.Called()
}

func (m *loggerMock) IsTrace() bool {
	rv := m.Called()
	return rv.Bool(0)
}

func (m *loggerMock) Debug(msg string, args ...interface{}) {
	m.Called()
}

func (m *loggerMock) IsDebug() bool {
	rv := m.Called()
	return rv.Bool(0)
}

func (m *loggerMock) Info(msg string, args ...interface{}) {
	m.Called()
}

func (m *loggerMock) IsInfo() bool {
	rv := m.Called()
	return rv.Bool(0)
}

func (m *loggerMock) Warn(msg string, args ...interface{}) {
	m.Called()
}

func (m *loggerMock) IsWarn() bool {
	rv := m.Called()
	return rv.Bool(0)
}

func (m *loggerMock) Error(msg string, args ...interface{}) {
	m.Called()
}

func (m *loggerMock) IsError() bool {
	rv := m.Called()
	return rv.Bool(0)
}

func (m *loggerMock) ImpliedArgs() []interface{} {
	rv := m.Called()
	return rv.Get(0).([]interface{})
}

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
			m := &loggerMock{}
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

// func TestLog(t *testing.T) {
// 	m := loggerMock{}

// 	n := hclogNotifier{logger: loggerMock{}}
// 	c := n.Run()
// 	defer close(c)

// }
