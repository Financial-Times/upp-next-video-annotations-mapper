package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/kafka-client-go/v3"
	"github.com/google/uuid"
)

const (
	nextVideoOrigin  = "http://cmdb.ft.com/systems/next-video-editor"
	dateFormat       = "2006-01-02T15:04:05.000Z0700"
	generatedMsgType = "concept-annotations"
	mapEvent         = "Map"
	contentType      = "Annotations"
)

type messageProducer interface {
	SendMessage(kafka.FTMessage) error
}

type queueHandler struct {
	sc              serviceConfig
	messageProducer messageProducer
	log             *logger.UPPLogger
}

func newQueueHandler(sc serviceConfig, messageProducer messageProducer, log *logger.UPPLogger) *queueHandler {
	return &queueHandler{
		sc:              sc,
		messageProducer: messageProducer,
		log:             log,
	}
}

func (h *queueHandler) queueConsume(m kafka.FTMessage) {
	if m.Headers["Origin-System-Id"] != nextVideoOrigin {
		h.log.Infof("Ignoring message with different Origin-System-Id: %v", m.Headers["Origin-System-Id"])
		return
	}
	vm := videoMapper{sc: h.sc, strContent: m.Body, tid: m.Headers["X-Request-Id"], log: h.log}
	marshalledEvent, videoUUID, err := h.mapNextVideoAnnotationsMessage(&vm)
	if err != nil {
		h.log.WithMonitoringEvent(mapEvent, vm.tid, contentType).
			WithValidFlag(false).
			WithUUID(videoUUID).
			WithError(err).
			Warnf("Error mapping the message from queue")
		return
	}

	headers := createHeader(m.Headers)
	msgToSend := string(marshalledEvent)
	err = h.messageProducer.SendMessage(kafka.FTMessage{Headers: headers, Body: msgToSend})
	if err != nil {
		h.log.WithMonitoringEvent(mapEvent, vm.tid, contentType).
			WithValidFlag(true).
			WithUUID(videoUUID).
			WithError(err).
			Warnf("Error sending transformed message to queue")
		return
	}

	h.log.WithMonitoringEvent(mapEvent, vm.tid, contentType).
		WithValidFlag(true).
		WithUUID(videoUUID).
		Info("Mapped and sent.")
}

func (h *queueHandler) mapNextVideoAnnotationsMessage(vm *videoMapper) ([]byte, string, error) {
	h.log.WithTransactionID(vm.tid).
		Info("Start mapping next video message.")

	if err := json.Unmarshal([]byte(vm.strContent), &vm.unmarshalled); err != nil {
		return nil, "", fmt.Errorf("video JSON from Next couldn't be unmarshalled: %v. Skipping invalid JSON with tid: %s", err, vm.tid)
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
		"Message-Id":        uuid.New().String(),
		"Message-Type":      generatedMsgType,
		"Content-Type":      "application/json",
		"Origin-System-Id":  origMsgHeaders["Origin-System-Id"],
	}
}
