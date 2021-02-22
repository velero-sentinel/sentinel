package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebug(t *testing.T) {
	d := debug(true)
	assert.NoError(t, d.BeforeApply())
	assert.True(t, appLogger.IsDebug())
}

func TestNotifierConfig(t *testing.T) {
	testCases := []struct {
		desc          string
		path          CfgPath
		expectedError bool
	}{
		{
			desc:          "AllGood",
			path:          "_testdata/sentinel.yml",
			expectedError: false,
		},
		{
			desc:          "Non-existing file",
			path:          "foobar.yml",
			expectedError: true,
		},
		{
			desc:          "Malformed YAML",
			path:          "_testdata/malformed.yml",
			expectedError: true,
		},
		{
			desc:          "Unset",
			path:          "",
			expectedError: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := tC.path.AfterApply()
			assert.True(t, (err != nil) == tC.expectedError)
		})
	}
}
