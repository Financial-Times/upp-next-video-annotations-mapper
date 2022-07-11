package main

import (
	"errors"
	"testing"

	"net/http/httptest"

	"github.com/stretchr/testify/assert"
)

func initializeHealthCheck(isProducerConnectionHealthy bool, isConsumerConnectionHealthy bool) *HealthCheck {
	return &HealthCheck{
		consumer: &mockConsumerInstance{isConnectionHealthy: isConsumerConnectionHealthy},
		producer: &mockProducerInstance{isConnectionHealthy: isProducerConnectionHealthy},
	}
}

func TestHappyHealthCheck(t *testing.T) {
	hc := initializeHealthCheck(true, true)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Reachable","ok":true`, "Read message queue healthcheck should be happy")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Is Not Lagging","ok":true`, "Read message queue monitor check should be happy")
	assert.Contains(t, w.Body.String(), `"name":"Write Message Queue Reachable","ok":true`, "Write message queue healthcheck should be happy")
}

func TestHealthCheckWithUnhappyConsumer(t *testing.T) {
	hc := initializeHealthCheck(true, false)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Reachable","ok":false`, "Read message queue healthcheck should be unhappy")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Is Not Lagging","ok":false`, "Read message queue monitor check should be unhappy")
	assert.Contains(t, w.Body.String(), `"name":"Write Message Queue Reachable","ok":true`, "Write message queue healthcheck should be happy")
}

func TestHealthCheckWithUnhappyProducer(t *testing.T) {
	hc := initializeHealthCheck(false, true)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Reachable","ok":true`, "Read message queue healthcheck should be happy")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Is Not Lagging","ok":true`, "Read message queue monitor check should be happy")
	assert.Contains(t, w.Body.String(), `"name":"Write Message Queue Reachable","ok":false`, "Write message queue healthcheck should be unhappy")
}

func TestUnhappyHealthCheck(t *testing.T) {
	hc := initializeHealthCheck(false, false)

	req := httptest.NewRequest("GET", "http://example.com/__health", nil)
	w := httptest.NewRecorder()

	hc.Health()(w, req)

	assert.Equal(t, 200, w.Code, "It should return HTTP 200 OK")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Reachable","ok":false`, "Read message queue healthcheck should be unhappy")
	assert.Contains(t, w.Body.String(), `"name":"Read Message Queue Is Not Lagging","ok":false`, "Read message queue monitor check should be unhappy")
	assert.Contains(t, w.Body.String(), `"name":"Write Message Queue Reachable","ok":false`, "Write message queue healthcheck should be unhappy")
}

func TestGTGHappyFlow(t *testing.T) {
	hc := initializeHealthCheck(true, true)

	status := hc.GTG()
	assert.True(t, status.GoodToGo)
	assert.Empty(t, status.Message)
}

func TestGTGBrokenConsumer(t *testing.T) {
	hc := initializeHealthCheck(true, false)

	status := hc.GTG()
	assert.False(t, status.GoodToGo)
	assert.Equal(t, "error connecting to the queue", status.Message)
}

func TestGTGBrokenProducer(t *testing.T) {
	hc := initializeHealthCheck(false, true)

	status := hc.GTG()
	assert.False(t, status.GoodToGo)
	assert.Equal(t, "error connecting to the queue", status.Message)
}

type mockProducerInstance struct {
	isConnectionHealthy bool
}

func (p *mockProducerInstance) ConnectivityCheck() error {
	if p.isConnectionHealthy {
		return nil
	}
	return errors.New("error connecting to the queue")
}

type mockConsumerInstance struct {
	isConnectionHealthy bool
}

func (c *mockConsumerInstance) ConnectivityCheck() error {
	if c.isConnectionHealthy {
		return nil
	}
	return errors.New("error connecting to the queue")
}

func (c *mockConsumerInstance) MonitorCheck() error {
	if c.isConnectionHealthy {
		return nil
	}
	return errors.New("error kafka client is lagging ")
}
