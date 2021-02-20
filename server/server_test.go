package server

import (
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	mocked "github.com/mwmahlberg/hclog-mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/velero-sentinel/sentinel/message"

	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func TestNoDownstream(t *testing.T) {
	s := server{logger: hclog.NewNullLogger()}
	err := s.Run(make(chan watch.Event))
	assert.Error(t, err)
}

func TestInvalidEventObject(t *testing.T) {

	l := &mocked.Logger{}
	l.On("Error", mock.Anything).Return(nil)
	p := make(chan message.Message)
	w := make(chan watch.Event)
	s := New(p, l)
	go func() {
		assert.NoError(t, s.Run(w))
	}()

	w <- watch.Event{
		Type:   watch.Added,
		Object: nil,
	}
	close(w)
	time.Sleep(250 * time.Millisecond)
	l.AssertNumberOfCalls(t, "Error", 1)

}
func TestDispatch(t *testing.T) {
	testCases := []struct {
		desc            string
		phase           v1.BackupPhase
		eventType       watch.EventType
		messageExpected bool
	}{
		{
			desc:            "Added",
			phase:           v1.BackupPhaseNew,
			eventType:       watch.Added,
			messageExpected: false,
		},
		{
			desc:            "Deleted",
			phase:           v1.BackupPhaseCompleted,
			eventType:       watch.Deleted,
			messageExpected: false,
		},
		{
			desc:            "Completed",
			phase:           v1.BackupPhaseCompleted,
			eventType:       watch.Modified,
			messageExpected: false,
		},
		{
			desc:            "New",
			phase:           v1.BackupPhaseNew,
			eventType:       watch.Modified,
			messageExpected: false,
		},
		{
			desc:            "InProgress",
			phase:           v1.BackupPhaseInProgress,
			eventType:       watch.Modified,
			messageExpected: false,
		},
		{
			desc:            "Deleting",
			phase:           v1.BackupPhaseDeleting,
			eventType:       watch.Modified,
			messageExpected: false,
		},
		{
			desc:            "PartiallyFailedBackup",
			phase:           v1.BackupPhasePartiallyFailed,
			eventType:       watch.Modified,
			messageExpected: true,
		},
		{
			desc:            "FailedBackup",
			phase:           v1.BackupPhaseFailed,
			eventType:       watch.Modified,
			messageExpected: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			l := &mocked.Logger{}
			if tC.messageExpected == false {
				l.On("Info", mock.Anything).Return(nil).Once()
			}
			w := make(chan watch.Event)
			p := make(chan message.Message)
			s := New(p, l)

			e := watch.Event{
				Type: tC.eventType,
				Object: &v1.Backup{
					ObjectMeta: metav1.ObjectMeta{Name: "foo"},
					Status:     v1.BackupStatus{Phase: tC.phase}},
			}
			go s.Run(w)
			w <- e
			var m message.Message
			select {
			case m = <-p:
			case <-time.After(1 * time.Second):
			}
			close(w)
			assert.True(t, (m != nil) == tC.messageExpected)
			l.AssertExpectations(t)
		})
	}
}
