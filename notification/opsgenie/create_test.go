package opsgenie_test

import (
	"testing"

	validator "github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/velero-sentinel/sentinel/notification/opsgenie"
)

func TestNewNoApiKey(t *testing.T) {
	n, err := opsgenie.New("")
	assert.Nil(t, n)
	assert.Error(t, err)
	assert.IsType(t, opsgenie.MissingApiKeyError, err)
}

func TestNewInvalidApiKey(t *testing.T) {
	n, err := opsgenie.New("abcd")
	assert.False(t, validator.IsUUID("abcd"))
	assert.Error(t, err)
	assert.Nil(t, n)
}

func TestNewValid(t *testing.T) {
	n, err := opsgenie.New(
		uuid.NewString(),
		opsgenie.EU(),
		opsgenie.Tags([]string{"a", "b"}),
		opsgenie.NotifyOnWarning(false),
		opsgenie.NotifyStakeholders(true),
	)

	assert.NoError(t, err)
	assert.NotNil(t, n)

}

func TestNewInvalidRetries(t *testing.T) {
	n, err := opsgenie.New(
		uuid.NewString(),
		opsgenie.Sandbox(),
		opsgenie.Retries(-1),
	)
	assert.Error(t, err)
	assert.Nil(t, n)

}

func TestNewInvalidUrl(t *testing.T) {
	n, err := opsgenie.New(
		uuid.NewString(),
		OverrideURI("api.url.com"),
	)
	assert.Error(t, err)
	assert.Nil(t, n)
}
