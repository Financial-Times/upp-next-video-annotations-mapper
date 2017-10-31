package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Financial-Times/go-logger"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/satori/go.uuid"
)

const (
	nextVideoOrigin  = "http://cmdb.ft.com/systems/next-video-editor"
	dateFormat       = "2006-01-02T15:04:05.000Z0700"
	generatedMsgType = "concept-annotations"
	mapEvent         = "Map"
	contentType      = "Annotations"
)

type queueHandler struct {
	sc              serviceConfig
	httpCl          *http.Client
	consumerConfig  consumer.QueueConfig
	producerConfig  producer.MessageProducerConfig
	messageConsumer consumer.MessageConsumer
	messageProducer producer.MessageProducer
}

func (h *queueHandler) init() {
	h.messageProducer = producer.NewMessageProducer(h.producerConfig)

	h.messageConsumer = consumer.NewConsumer(h.consumerConfig, h.queueConsume, h.httpCl)
}

func (h *queueHandler) queueConsume(m consumer.Message) {
	if m.Headers["Origin-System-Id"] != nextVideoOrigin {
		logger.WithField("queue_topic", h.consumerConfig.Topic).Infof("Ignoring message with different Origin-System-Id: %v", m.Headers["Origin-System-Id"])
		return
	}
	vm := videoMapper{sc: h.sc, strContent: m.Body, tid: m.Headers["X-Request-Id"]}
	marshalledEvent, videoUUID, err := h.mapNextVideoAnnotationsMessage(&vm)
	if err != nil {
		logger.WithMonitoringEvent(mapEvent, vm.tid, contentType).
			WithValidFlag(false).
			WithUUID(videoUUID).
			WithFields(map[string]interface{}{"queue_name": h.consumerConfig.Queue, "queue_topic": h.consumerConfig.Topic}).
			WithError(err).
			Warnf("Error mapping the message from queue")
		return
	}

	headers := createHeader(m.Headers)
	msgToSend := string(marshalledEvent)
	err = h.messageProducer.SendMessage("", producer.Message{Headers: headers, Body: msgToSend})
	if err != nil {
		logger.WithMonitoringEvent(mapEvent, vm.tid, contentType).
			WithValidFlag(true).
			WithUUID(videoUUID).
			WithFields(map[string]interface{}{"queue_name": h.consumerConfig.Queue, "queue_topic": h.consumerConfig.Topic}).
			WithError(err).
			Warnf("Error sending transformed message to queue")
		return
	}

	logger.WithMonitoringEvent(mapEvent, vm.tid, contentType).
		WithValidFlag(true).
		WithUUID(videoUUID).
		WithFields(map[string]interface{}{"queue_name": h.consumerConfig.Queue, "queue_topic": h.consumerConfig.Topic}).
		Info("Mapped and sent.")
}

func (h *queueHandler) mapNextVideoAnnotationsMessage(vm *videoMapper) ([]byte, string, error) {
	logger.WithTransactionID(vm.tid).
		WithFields(map[string]interface{}{"queue_name": h.consumerConfig.Queue, "queue_topic": h.consumerConfig.Topic}).
		Info("Start mapping next video message.")

	if err := json.Unmarshal([]byte(vm.strContent), &vm.unmarshalled); err != nil {
		return nil, "", fmt.Errorf("Video JSON from Next couldn't be unmarshalled: %v. Skipping invalid JSON with tid: %s.", err.Error(), vm.tid)
	}
	if vm.tid == "" {
		return nil, "", fmt.Errorf("X-Request-Id not found in kafka message headers. Skipping message with tid %s", vm.tid)
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
