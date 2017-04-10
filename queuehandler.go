package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/satori/go.uuid"
	"net/http"
	"time"
)

const nextVideoOrigin = "http://cmdb.ft.com/systems/next-video-editor"
const dateFormat = "2006-01-02T03:04:05.000Z0700"
const generatedMsgType = "concept-suggestions"

type queueHandler struct {
	sc              serviceConfig
	consumerConfig  consumer.QueueConfig
	producerConfig  producer.MessageProducerConfig
	messageConsumer *consumer.MessageConsumer
	messageProducer *producer.MessageProducer
}

func (h *queueHandler) init() {
	messageProducer := producer.NewMessageProducer(h.producerConfig)
	h.messageProducer = &messageProducer

	httpCl := http.Client{}
	messageConsumer := consumer.NewConsumer(h.consumerConfig, h.queueConsume, &httpCl)
	h.messageConsumer = &messageConsumer
}

func (h *queueHandler) queueConsume(m consumer.Message) {
	if m.Headers["Origin-System-Id"] != nextVideoOrigin {
		logger.messageEvent(h.consumerConfig.Topic, fmt.Sprintf("Ignoring message with different Origin-System-Id: %v", m.Headers["Origin-System-Id"]))
		return
	}
	vm := videoMapper{sc: h.sc, strContent: m.Body, tid: m.Headers["X-Request-Id"]}
	marshalledEvent, videoUUID, err := h.mapNextVideoAnnotationsMessage(&vm)
	if err != nil {
		logger.warnMessageEvent(queueEvent{h.sc.serviceName, h.consumerConfig.Queue, h.consumerConfig.Topic, vm.tid}, videoUUID, err,
			"Error mapping the message from queue")
		return
	}

	if marshalledEvent == nil {
		return
	}

	headers := createHeader(m.Headers)
	msgToSend := string(marshalledEvent)
	err = (*h.messageProducer).SendMessage("", producer.Message{Headers: headers, Body: msgToSend})
	if err != nil {
		logger.warnMessageEvent(queueEvent{h.sc.serviceName, h.producerConfig.Queue, h.producerConfig.Topic, vm.tid}, videoUUID, err,
			"Error sending transformed message to queue")
		return
	}
	logger.messageSentEvent(queueEvent{h.sc.serviceName, h.producerConfig.Queue, h.producerConfig.Topic, vm.tid}, videoUUID,
		fmt.Sprintf("Mapped and sent: [%v]", msgToSend))
}

func (h *queueHandler) mapNextVideoAnnotationsMessage(vm *videoMapper) ([]byte, string, error) {
	logger.messageEvent(h.consumerConfig.Topic, "Start mapping next video message.")
	if err := json.Unmarshal([]byte(vm.strContent), &vm.unmarshalled); err != nil {
		return nil, "", fmt.Errorf("Video JSON from Next couldn't be unmarshalled: %v. Skipping invalid JSON: %v", err.Error(), vm.strContent)
	}
	if vm.tid == "" {
		return nil, "", errors.New("X-Request-Id not found in kafka message headers. Skipping message")
	}
	return vm.mapNextVideoAnnotations()
}

func createHeader(origMsgHeaders map[string]string) map[string]string {
	return map[string]string{
		"X-Request-Id":      origMsgHeaders["X-Request-Id"],
		"Message-Timestamp": time.Now().Format(dateFormat),
		"Message-Id":        uuid.NewV4().String(),
		"Message-Type":      generatedMsgType,
		"Content-Type":      "application/json",
		"Origin-System-Id":  origMsgHeaders["Origin-System-Id"],
	}
}