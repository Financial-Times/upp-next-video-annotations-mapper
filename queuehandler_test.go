package main

import (
	"io/ioutil"
	"testing"

	"github.com/Financial-Times/kafka-client-go/v3"
	"github.com/stretchr/testify/assert"
)

type mockMessageProducer struct {
	message    string
	sendCalled bool
}

func TestQueueConsume(t *testing.T) {
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
			newStringConceptAnnotation(t, "e2290d14-7e80-4db8-a715-949da4de9a07",
				[]annotation{{"http://api.ft.com/things/71a5efa5-e6e0-3ce1-9190-a7eac8bef325", "isClassifiedBy", defaultRelevanceScore, defaultConfidenceScore}},
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
			true,
			newStringConceptAnnotation(t, "e2290d14-7e80-4db8-a715-949da4de9a07", nil),
		},
	}

	for _, test := range tests {
		mockMsgProducer := mockMessageProducer{}
		var msgProducer = &mockMsgProducer
		h := queueHandler{
			sc:              serviceConfig{},
			messageProducer: msgProducer,
			log:             getLogger(),
		}

		msg := kafka.FTMessage{
			Headers: createHeaders(test.originSystem, test.tid),
			Body:    string(getBytes(test.fileName, t)),
		}
		h.queueConsume(msg)

		assert.Equal(t, test.expectedMsgSent, mockMsgProducer.sendCalled,
			"Message sending check is wrong. Input JSON file: %s, Origin-System-Id: %s, X-Request-Id: %s", test.fileName, test.originSystem, test.tid)
		assert.Equal(t, test.expectedContent, mockMsgProducer.message,
			"Marshalled content wrong. Input JSON file: %s, Origin-System-Id: %s, X-Request-Id: %s", test.fileName, test.originSystem, test.tid)
	}
}

func createHeaders(originSystem string, requestID string) map[string]string {
	var result = make(map[string]string)
	result["Origin-System-Id"] = originSystem
	result["X-Request-Id"] = requestID
	return result
}

func (mock *mockMessageProducer) SendMessage(message kafka.FTMessage) error {
	mock.message = message.Body
	mock.sendCalled = true
	return nil
}

func (mock *mockMessageProducer) ConnectivityCheck() (string, error) {
	// do nothing
	return "", nil
}

func getBytes(fileName string, t *testing.T) []byte {
	bytes, err := ioutil.ReadFile("test-resources/" + fileName)
	if err != nil {
		assert.Fail(t, err.Error())
		return nil
	}

	return bytes
}
