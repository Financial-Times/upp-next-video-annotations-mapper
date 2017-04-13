package main

import (
	"encoding/json"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	_            = iota
	statusOK int = 1 + iota
	statusNA
	statusMissingTopics
)

const consumerTopic = "NativeCmsPublicationEvents"
const producerTopic = "V1ConceptAnnotations"

var queueServerMock *httptest.Server

func TestCheckMessageQueueAvailability(t *testing.T) {
	assert := assert.New(t)

	startQueueServerMock(statusOK)
	defer queueServerMock.Close()

	h := healthCheck{
		client:       http.Client{},
		consumerConf: newConsumerConfig(queueServerMock.URL),
		producerConf: newProducerConfig(queueServerMock.URL),
	}

	result, err := h.checkAggregateMessageQueueProxiesReachable()
	assert.Nil(err, "Error not expected.")
	assert.Equal("Ok", result, "Message queue availability status is wrong")
}

func TestCheckMessageQueueNonAvailability(t *testing.T) {
	assert := assert.New(t)

	startQueueServerMock(statusNA)
	defer queueServerMock.Close()

	h := healthCheck{
		client:       http.Client{},
		consumerConf: newConsumerConfig(queueServerMock.URL),
		producerConf: newProducerConfig(queueServerMock.URL),
	}

	_, err := h.checkAggregateMessageQueueProxiesReachable()
	assert.Equal(true, err != nil, "Error was expected.")
}

func TestCheckMessageQueueMissingTopic(t *testing.T) {
	assert := assert.New(t)

	startQueueServerMock(statusMissingTopics)
	defer queueServerMock.Close()

	h := healthCheck{
		client:       http.Client{},
		consumerConf: newConsumerConfig(queueServerMock.URL),
		producerConf: newProducerConfig(queueServerMock.URL),
	}

	_, err := h.checkAggregateMessageQueueProxiesReachable()
	assert.Equal(true, err != nil, "Error was expected.")
}

func TestCheckMessageQueueWrongQueueURL(t *testing.T) {
	assert := assert.New(t)

	startQueueServerMock(statusOK)
	defer queueServerMock.Close()

	tests := []struct {
		consumerConfig consumer.QueueConfig
		producerConfig producer.MessageProducerConfig
	}{
		{
			newConsumerConfig("wrong url"),
			newProducerConfig(queueServerMock.URL),
		},
		{
			newConsumerConfig(queueServerMock.URL),
			newProducerConfig("wrong url"),
		},
	}

	for _, test := range tests {
		h := healthCheck{
			client:       http.Client{},
			consumerConf: test.consumerConfig,
			producerConf: test.producerConfig,
		}

		_, err := h.checkAggregateMessageQueueProxiesReachable()
		assert.Equal(true, err != nil, "Error was expected for input consumer [%v], producer [%v]", test.consumerConfig, test.producerConfig)
	}
}

func startQueueServerMock(status int) {
	router := mux.NewRouter()
	var getContent http.HandlerFunc

	switch status {
	case statusOK:
		getContent = statusOKHandler
	case statusNA:
		getContent = internalErrorHandler
	case statusMissingTopics:
		getContent = statusMissingTopicsHandler
	}

	router.Path("/topics").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(getContent)})

	queueServerMock = httptest.NewServer(router)
}

func statusOKHandler(w http.ResponseWriter, r *http.Request) {
	writeTopics(w, consumerTopic, producerTopic)
}

func statusMissingTopicsHandler(w http.ResponseWriter, r *http.Request) {
	writeTopics(w, "OtherTopic")
}

func writeTopics(w http.ResponseWriter, topics ...string) {
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(topics)
	if err != nil {
		panic("Unexpected error during response topics write")
	}
	w.Write(b)
}

func internalErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func newConsumerConfig(addr string) consumer.QueueConfig {
	return consumer.QueueConfig{
		Addrs:            []string{addr},
		Topic:            consumerTopic,
		Queue:            "queue",
		AuthorizationKey: "auth",
	}
}

func newProducerConfig(addr string) producer.MessageProducerConfig {
	return producer.MessageProducerConfig{
		Addr:          addr,
		Topic:         producerTopic,
		Queue:         "queue",
		Authorization: "auth",
	}
}
