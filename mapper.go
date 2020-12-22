package main

import (
	"encoding/json"
	"fmt"

	"github.com/Financial-Times/go-logger"
)

const (
	videoIDField             = "id"
	videoUUIDField           = "uuid"
	annotationsField         = "annotations"
	annotationIDField        = "id"
	annotationPredicateField = "predicate"
	deletedField             = "deleted"
)

type videoMapper struct {
	sc           serviceConfig
	strContent   string
	tid          string
	unmarshalled map[string]interface{}
}

type tag struct {
	thingID   string
	predicate string
}

func (vm *videoMapper) mapNextVideoAnnotations() ([]byte, string, error) {
	var videoUUID string
	var err error

	if vm.isDeleteEvent() {
		videoUUID, err = getRequiredStringField(videoUUIDField, vm.unmarshalled)
		if err != nil {
			return nil, "", err
		}
	} else {
		videoUUID, err = getRequiredStringField(videoIDField, vm.unmarshalled)
		if err != nil {
			return nil, "", err
		}
	}

	nextAnnsArray, err := getObjectsArrayField(annotationsField, vm.unmarshalled, videoUUID, vm)
	if err != nil {
		return nil, videoUUID, err
	}

	annotations := vm.retrieveAnnotations(nextAnnsArray, videoUUID)

	if len(annotations) == 0 {
		logger.WithTransactionID(vm.tid).
			WithUUID(videoUUID).
			Info("No annotation could be retrieved for Next video")
	}

	conceptAnnotations := createAnnotations(annotations, annsContext{videoUUID: videoUUID, transactionID: vm.tid})

	marshalledPubEvent, err := json.Marshal(conceptAnnotations)
	if err != nil {
		return nil, videoUUID, err
	}

	return marshalledPubEvent, videoUUID, nil
}

func (vm *videoMapper) retrieveAnnotations(nextAnnsArray []map[string]interface{}, videoUUID string) []tag {
	var annotations = make([]tag, 0)
	for _, ann := range nextAnnsArray {
		thingID, err := getRequiredStringField(annotationIDField, ann)
		if err != nil {
			logger.WithTransactionID(vm.tid).
				WithUUID(videoUUID).
				WithError(err).
				Error("Cannot extract concept id from annotation field")
			continue
		}

		nextAnnPredicate, err := getRequiredStringField(annotationPredicateField, ann)
		if err != nil {
			logger.WithTransactionID(vm.tid).
				WithUUID(videoUUID).
				WithError(err).
				Error("Cannot extract predicate from annotation field")
			continue
		}

		predicate, ok := getPredicateShortForm(nextAnnPredicate)
		if !ok {
			logger.WithTransactionID(vm.tid).
				WithUUID(videoUUID).
				Errorf("Next video predicate id is not known: %s", nextAnnPredicate)
			continue
		}

		ann := tag{
			thingID:   thingID,
			predicate: predicate,
		}
		annotations = append(annotations, ann)
	}
	return annotations
}

func getRequiredStringField(key string, obj map[string]interface{}) (string, error) {
	valueI, ok := obj[key]
	if !ok || valueI == nil {
		return "", nullFieldError(key)
	}

	val, ok := valueI.(string)
	if !ok {
		return "", wrongFieldTypeError("string", key, valueI)
	}
	return val, nil
}

func getObjectsArrayField(key string, obj map[string]interface{}, videoUUID string, vm *videoMapper) ([]map[string]interface{}, error) {
	var objArrayI interface{}
	var result = make([]map[string]interface{}, 0)
	objArrayI, ok := obj[key]
	if !ok {
		logger.WithTransactionID(vm.tid).
			WithUUID(videoUUID).
			Info(nullFieldError(key).Error())
		return result, nil
	}

	var objArray []interface{}
	objArray, ok = objArrayI.([]interface{})
	if !ok {
		return nil, wrongFieldTypeError("object array", key, objArrayI)
	}

	for _, objI := range objArray {
		obj, ok = objI.(map[string]interface{})
		if !ok {
			return nil, wrongFieldTypeError("object array", key, objArrayI)
		}
		result = append(result, obj)
	}
	return result, nil
}

func nullFieldError(fieldKey string) error {
	return fmt.Errorf("[%s] field of native Next video JSON is missing or is null", fieldKey)
}

func wrongFieldTypeError(expectedType, fieldKey string, value interface{}) error {
	return fmt.Errorf("[%s] field of native Next video JSON is not of type %s.", fieldKey, expectedType)
}

func (vm *videoMapper) isDeleteEvent() bool {
	if _, present := vm.unmarshalled[deletedField]; present {
		return true
	}
	return false
}
