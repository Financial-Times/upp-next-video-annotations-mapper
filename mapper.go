package main

import (
	"encoding/json"
	"fmt"
)

const videoUUIDField = "id"
const annotationsField = "annotations"
const annotationIdField = "id"
const annotationPredicateField = "predicate"
const deletedField = "deleted"

type videoMapper struct {
	sc           serviceConfig
	strContent   string
	tid          string
	unmarshalled map[string]interface{}
}

type annotation struct {
	thingID   string
	predicate string
}

func (vm *videoMapper) mapNextVideoAnnotations() ([]byte, string, error) {
	videoUUID, err := getRequiredStringField(videoUUIDField, vm.unmarshalled)
	if err != nil {
		return nil, "", err
	}

	if vm.isDeleteEvent() {
		logger.videoMapEvent(vm.tid, videoUUID, fmt.Sprint("Ignoring delete Next video message."))
		return nil, videoUUID, nil
	}

	nextAnnsArray, err := getObjectsArrayField(annotationsField, vm.unmarshalled, videoUUID, vm)
	switch {
	case err != nil:
		return nil, videoUUID, err
	case nextAnnsArray == nil:
		return nil, videoUUID, nil
	}

	annotations := vm.retrieveAnnotations(nextAnnsArray, videoUUID)

	if len(annotations) == 0 {
		logger.videoMapEvent(vm.tid, videoUUID, fmt.Sprintf("No annotation could be retrieved for Next video: [%s]", vm.strContent))
		return nil, videoUUID, nil
	}

	annHandler := annHandler{videoUUID: videoUUID, transactionID: vm.tid}
	conceptSuggestion := annHandler.createAnnotations(annotations)

	marshalledPubEvent, err := json.Marshal(conceptSuggestion)
	if err != nil {
		logger.videoEvent(vm.tid, videoUUID, "Error marshalling processed annotations")
		return nil, videoUUID, err
	}

	return marshalledPubEvent, videoUUID, nil
}

func (vm *videoMapper) retrieveAnnotations(nextAnnsArray []map[string]interface{}, videoUUID string) []annotation {
	var annotations = make([]annotation, 0)
	for _, ann := range nextAnnsArray {
		thingID, err := getRequiredStringField(annotationIdField, ann)
		if err != nil {
			logger.videoErrorEvent(vm.tid, videoUUID, err, "Cannot extract concept id from annotation field")
			continue
		}

		predicate, err := getRequiredStringField(annotationPredicateField, ann)
		if err != nil {
			logger.videoErrorEvent(vm.tid, videoUUID, err, "Cannot extract predicate from annotation field")
			continue
		}

		ann := annotation{
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
	objArrayI, ok := obj[key]
	if !ok {
		logger.videoMapEvent(vm.tid, videoUUID, nullFieldError(key).Error())
		return nil, nil
	}

	var objArray []interface{}
	objArray, ok = objArrayI.([]interface{})
	if !ok {
		return nil, wrongFieldTypeError("object array", key, objArrayI)
	}

	var result = make([]map[string]interface{}, 0)
	for _, objI := range objArray {
		obj, ok = objI.(map[string]interface{})
		if !ok {
			return nil, wrongFieldTypeError("object array", key, objArrayI)
		}
		result = append(result, obj)
	}
	return result, nil
}

func getBoolField(key string, obj map[string]interface{}) (bool, error) {
	valueI, ok := obj[key]
	if !ok {
		return false, nullFieldError(key)
	}

	val, ok := valueI.(bool)
	if !ok {
		return false, wrongFieldTypeError("bool", key, valueI)
	}
	return val, nil
}

func nullFieldError(fieldKey string) error {
	return fmt.Errorf("[%s] field of native Next video JSON is missing or is null", fieldKey)
}

func wrongFieldTypeError(expectedType, fieldKey string, value interface{}) error {
	return fmt.Errorf("[%s] field of native Next video JSON is not of type %s: [%v]", fieldKey, expectedType, value)
}

func (vm *videoMapper) isDeleteEvent() bool {
	if _, present := vm.unmarshalled[deletedField]; present {
		return true
	}
	return false
}
