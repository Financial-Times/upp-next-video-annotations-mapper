package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCheckPublicThingsAPIServiceAvailability(t *testing.T) {
	startPublicThingsAPIMock(scenarioHappy)
	defer stopService()

	result, err := checkServiceAvailability("publicThingsAPI", publicThingsAPIMock.URL + "/__health")
	assert.Equal(t, result, "Ok")
	assert.Nil(t, err)
}

func TestCheckPublicThingsAPIServiceNonAvailability(t *testing.T) {
	startPublicThingsAPIMock(scenarioNotHappy)
	defer stopService()

	_, err := checkServiceAvailability("publicThingsAPI", publicThingsAPIMock.URL + "/__health")
	assert.Error(t, err)
}