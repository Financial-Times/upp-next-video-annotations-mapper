package main

import (
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func init() {
	logger = newAppLogger("test")
}

type mockMessageProducer struct {
	message    string
	sendCalled bool
}

func TestQueueConsume(t *testing.T) {
	startPublicThingsAPIMock(scenarioHappy)
	defer stopService()

	assert := assert.New(t)
	tests := []struct {
		fileName        string
		originSystem    string
		tid             string
		expectedMsgSent bool
		expectedContent string
	}{
		{
			"next-video-input.json",
			nextVideoOrigin,
			"1234",
			true,
			newStringConceptSuggestion(t, "e2290d14-7e80-4db8-a715-949da4de9a07",
				newSuggestion("http://api.ft.com/things/71a5efa5-e6e0-3ce1-9190-a7eac8bef325", "http://www.ft.com/ontology/Section", "isClassifiedBy", "Financials"),
			),
		},
		{
			"next-video-input.json",
			"other",
			"1234",
			false,
			"",
		},
		{
			"next-video-input.json",
			nextVideoOrigin,
			"",
			false,
			"",
		},
		{
			"invalid-format.json",
			nextVideoOrigin,
			"1234",
			false,
			"",
		},
		{
			"next-video-invalid-anns-input.json",
			nextVideoOrigin,
			"1234",
			false,
			"",
		},
		{
			"next-video-empty-anns-input.json",
			nextVideoOrigin,
			"1234",
			false,
			"",
		},
	}

	for _, test := range tests {
		mockMsgProducer, underlyingProducer := newMockMessageProducer()
		h := queueHandler{
			sc: serviceConfig{
				publicThingsURI: publicThingsAPIURLMock(),
			},
			messageProducer: &mockMsgProducer,
		}

		msg := consumer.Message{
			Headers: createHeaders(test.originSystem, test.tid),
			Body:    string(getBytes(test.fileName, t)),
		}
		h.queueConsume(msg)

		assert.Equal(test.expectedMsgSent, underlyingProducer.sendCalled,
			"Message sending check is wrong. Input JSON file: %s, Origin-System-Id: %s, X-Request-Id: %s", test.fileName, test.originSystem, test.tid)
		assert.Equal(test.expectedContent, underlyingProducer.message,
			"Marshalled content wrong. Input JSON file: %s, Origin-System-Id: %s, X-Request-Id: %s", test.fileName, test.originSystem, test.tid)
	}
}

func createHeaders(originSystem string, requestId string) map[string]string {
	var result = make(map[string]string)
	result["Origin-System-Id"] = originSystem
	result["X-Request-Id"] = requestId
	return result
}

func (mock *mockMessageProducer) SendMessage(uuid string, message producer.Message) error {
	mock.message = message.Body
	mock.sendCalled = true
	return nil
}

func getBytes(fileName string, t *testing.T) []byte {
	bytes, err := ioutil.ReadFile("test-resources/" + fileName)
	if err != nil {
		assert.Fail(t, err.Error())
		return nil
	}

	return bytes
}

func newMockMessageProducer() (producer.MessageProducer, *mockMessageProducer) {
	mock := mockMessageProducer{}
	// TODO reanalyse this; couldn't make it compile using a single return instance
	return &mock, &mock
}
