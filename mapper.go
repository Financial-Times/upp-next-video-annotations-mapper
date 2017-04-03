package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/satori/go.uuid"
	"net/http"
	"strings"
	"time"
)

const nextVideoOrigin = "next-video-editor"
const dateFormat = "2006-01-02T03:04:05.000Z0700"
const generatedMsgType = "concept-suggestions"

const videoUUIDField = "id"
const annotationsField = "annotations"
const annotationIdField = "id"
const annotationNameField = "name"
const annotationPrimaryField = "primary"

type annotationMapper struct {
	sc              serviceConfig
	consumerConfig  consumer.QueueConfig
	producerConfig  producer.MessageProducerConfig
	messageConsumer *consumer.Consumer
	messageProducer *producer.MessageProducer
}

type videoMessage struct {
	origMsg      *consumer.Message
	tid          string
	unmarshalled map[string]interface{}
}

type nextAnnotation struct {
	thingID     string
	thingText   string
	primaryFlag bool
	thing       thingInfo
}

func (am *annotationMapper) init() {
	*(am.messageProducer) = producer.NewMessageProducer(am.producerConfig)
	*(am.messageConsumer) = consumer.NewConsumer(am.consumerConfig, am.queueConsume, http.Client{})
}

func (am *annotationMapper) queueConsume(m consumer.Message) {
	vm := videoMessage{origMsg: &m, tid: m.Headers["X-Request-Id"]}
	if vm.origMsg.Headers["Origin-System-Id"] != nextVideoOrigin {
		logger.messageEvent(am.consumerConfig.Topic, fmt.Sprintf("Ignoring message with different Origin-System-Id: %v", vm.origMsg.Headers["Origin-System-Id"]))
		return
	}
	marshalledEvent, videoUUID, err := am.mapMessage(&vm)
	if err != nil {
		logger.warnMessageEvent(queueEvent{am.sc.serviceName, am.consumerConfig.Queue, am.consumerConfig.Topic, vm.tid}, err, "Error mapping the message from queue")
		return
	}
	headers := createHeader(m.Headers)
	err = (*am.messageProducer).SendMessage("", producer.Message{Headers: headers, Body: string(marshalledEvent)})
	if err != nil {
		logger.warnMessageEvent(queueEvent{am.sc.serviceName, am.producerConfig.Queue, am.producerConfig.Topic, vm.tid}, err, "Error sending transformed message to queue")
		return
	}
	logger.messageSentEvent(queueEvent{am.sc.serviceName, am.producerConfig.Queue, am.producerConfig.Topic, vm.tid}, videoUUID, "Mapped and sent for uuid: %v")
}

func (am *annotationMapper) mapMessage(vm *videoMessage) ([]byte, string, error) {
	if err := json.Unmarshal([]byte(vm.origMsg.Body), &vm.unmarshalled); err != nil {
		return nil, "", fmt.Errorf("Video JSON from Next couldn't be unmarshalled: %v. Skipping invalid JSON: %v", err.Error(), vm.origMsg.Body)
	}
	if vm.tid == "" {
		return nil, "", errors.New("X-Request-Id not found in kafka message headers. Skipping message")
	}
	lastModified := vm.origMsg.Headers["Message-Timestamp"]
	if lastModified == "" {
		return nil, "", errors.New("Message-Timestamp not found in kafka message headers. Skipping message")
	}
	return am.mapNextVideoAnnotations(vm, lastModified)
}

func (am *annotationMapper) mapNextVideoAnnotations(vm *videoMessage, lastModified string) ([]byte, string, error) {
	videoUUID, err := getStringField(videoUUIDField, vm.unmarshalled, vm)
	if err != nil {
		return nil, videoUUID, err
	}

	nextAnnsArray, err := getObjectsArrayField(annotationsField, vm.unmarshalled, vm)
	if err != nil {
		return nil, videoUUID, err
	}

	var nextAnnotations = make([]nextAnnotation, len(nextAnnsArray))
	for _, ann := range nextAnnsArray {
		thingID, err := getStringField(annotationIdField, ann, vm)
		if err != nil {
			logger.videoErrorEvent(vm.tid, videoUUID, err, "Annotation could not be processed")
			continue
		}

		thingUUID, ok := parseThingUUID(thingID)
		if !ok {
			logger.warnMessageEvent(queueEvent{am.sc.serviceName, am.consumerConfig.Queue, am.consumerConfig.Topic, vm.tid}, nil,
				fmt.Sprintf("Cannot extract thing UUID from annotation id field: %s", thingID))
			continue
		}

		thingName, _ := getStringField(annotationNameField, ann, vm) // ignore the error as name field is not used
		thingPrimaryFlag, err := getBoolField(annotationPrimaryField, ann, vm)
		if err != nil {
			logger.videoErrorEvent(vm.tid, videoUUID, err, "Cannot extract primary flag from annotation field")
			continue
		}

		ann := nextAnnotation{
			thingID:     thingID,
			thingText:   thingName,
			primaryFlag: thingPrimaryFlag,
			thing:       thingInfo{uuid: thingUUID},
		}
		nextAnnotations = append(nextAnnotations, ann)
	}

	thingHandler := thingHandler{am.sc, vm.tid, videoUUID}
	thingHandler.retrieveThingsDetails(nextAnnotations)

	annHandler := annHandler{videoUUID, vm.tid}
	conceptSuggestion := annHandler.createAnnotations(nextAnnotations)

	marshalledPubEvent, err := json.Marshal(*conceptSuggestion)
	if err != nil {
		logger.videoEvent(vm.tid, videoUUID, "Error marshalling processed annotations")
		return nil, videoUUID, err
	}

	return marshalledPubEvent, videoUUID, nil
}

func createHeader(origMsgHeaders map[string]string) map[string]string {
	return map[string]string{
		"X-Request-Id":      origMsgHeaders["X-Request-Id"],
		"Message-Timestamp": time.Now().Format(dateFormat),
		"Message-Id":        uuid.NewV4().String(),
		"Message-Type":      generatedMsgType,
		"Content-Type":      "application/json",
		"Origin-System-Id":  origMsgHeaders["Origin-System-Id"], // TODO here is ok ?
	}
}

func getStringField(key string, obj map[string]interface{}, vm *videoMessage) (string, error) {
	valueI, ok := obj[key]
	if !ok {
		return "", nullFieldError(key, vm)
	}

	val, ok := valueI.(string)
	if !ok {
		return "", wrongFieldTypeError("string", key, vm)
	}
	return val, nil
}

func getObjectsArrayField(key string, obj map[string]interface{}, vm *videoMessage) ([]map[string]interface{}, error) {
	objArrayI, ok := obj[key]
	if !ok {
		return nil, nullFieldError(key, vm)
	}

	objArray, ok := objArrayI.([]map[string]interface{})
	if !ok {
		return nil, wrongFieldTypeError("object array", key, vm)
	}
	return objArray, nil
}

func getBoolField(key string, obj map[string]interface{}, vm *videoMessage) (bool, error) {
	valueI, ok := obj[key]
	if !ok {
		return false, nullFieldError(key, vm)
	}

	val, ok := valueI.(bool)
	if !ok {
		return false, wrongFieldTypeError("bool", key, vm)
	}
	return val, nil
}

func parseThingUUID(thingID string) (string, bool) {
	uuidIdx := strings.LastIndex(thingID, "/") + 1
	uuid := thingID[uuidIdx:]
	if uuid != "" {
		return uuid, true
	}
	return "", false
}

func nullFieldError(fieldKey string, vm *videoMessage) error {
	return fmt.Errorf("[%s] field of native Next video JSON is missing: [%s]", fieldKey, vm.origMsg.Body)
}

func wrongFieldTypeError(expectedType, fieldKey string, vm *videoMessage) error {
	return fmt.Errorf("[%s] field of native Next video JSON is not of type %s: [%v]", fieldKey, expectedType, vm.origMsg.Body)
}
