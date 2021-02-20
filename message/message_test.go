package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
)

func TestMessages(t *testing.T) {
	testCases := []struct {
		desc    string
		message Message
	}{
		{
			desc:    "Warning",
			message: Warning{Backup: &v1.Backup{}},
		},
		{
			desc:    "Error",
			message: Error{Backup: &v1.Backup{}},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.NotNil(t, tC.message.GetBackup())
		})
	}
}
