package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

const videoUUIDField = "id"
const annotationsField = "annotations"
const annotationIdField = "id"
const annotationNameField = "name"
const annotationPrimaryField = "primary"

type videoMapper struct {
	sc           serviceConfig
	strContent   string
	tid          string
	unmarshalled map[string]interface{}
}

type annotation struct {
	thingID     string
	thingText   string
	primaryFlag bool
	thing       *thingInfo
}

func (vm *videoMapper) mapNextVideoAnnotations() ([]byte, string, error) {
	videoUUID, err := getStringField(videoUUIDField, vm.unmarshalled, vm)
	if err != nil {
		return nil, "", err
	}

	nextAnnsArray, err := getObjectsArrayField(annotationsField, vm.unmarshalled, vm)
	if err != nil {
		return nil, videoUUID, err
	}

	annotations := vm.buildAnnotations(nextAnnsArray, videoUUID)

	if len(annotations) == 0 {
		return nil, videoUUID, fmt.Errorf("No annotation could be retrieved for Next video: [%s]", vm.strContent)
	}

	thingHandler := thingHandler{vm.sc, vm.tid, videoUUID}
	thingHandler.retrieveThingsDetails(annotations)

	annHandler := annHandler{videoUUID, vm.tid}
	conceptSuggestion := annHandler.createAnnotations(annotations)

	marshalledPubEvent, err := json.Marshal(*conceptSuggestion)
	if err != nil {
		logger.videoEvent(vm.tid, videoUUID, "Error marshalling processed annotations")
		return nil, videoUUID, err
	}

	return marshalledPubEvent, videoUUID, nil
}

func (vm *videoMapper) buildAnnotations(nextAnnsArray []map[string]interface{}, videoUUID string) []annotation {
	var annotations = make([]annotation, 0)
	for _, ann := range nextAnnsArray {
		thingID, err := getStringField(annotationIdField, ann, vm)
		if err != nil {
			logger.videoErrorEvent(vm.tid, videoUUID, err, "Annotation could not be processed")
			continue
		}

		thingUUID, ok := parseThingUUID(thingID)
		if !ok {
			logger.videoEvent(vm.tid, videoUUID, fmt.Sprintf("Cannot extract thing UUID from annotation id field: %s", thingID))
			continue
		}

		thingName, _ := getStringField(annotationNameField, ann, vm) // ignore the error as name field is not used
		thingPrimaryFlag, err := getBoolField(annotationPrimaryField, ann, vm)
		if err != nil {
			logger.videoErrorEvent(vm.tid, videoUUID, err, "Cannot extract primary flag from annotation field")
			continue
		}

		ann := annotation{
			thingID:     thingID,
			thingText:   thingName,
			primaryFlag: thingPrimaryFlag,
			thing:       &thingInfo{uuid: thingUUID},
		}
		annotations = append(annotations, ann)
	}
	return annotations
}

func getStringField(key string, obj map[string]interface{}, vm *videoMapper) (string, error) {
	valueI, ok := obj[key]
	if !ok {
		return "", nullFieldError(key, vm)
	}

	val, ok := valueI.(string)
	if !ok {
		return "", wrongFieldTypeError("string", key, valueI)
	}
	return val, nil
}

func getObjectsArrayField(key string, obj map[string]interface{}, vm *videoMapper) ([]map[string]interface{}, error) {
	var objArrayI interface{}
	objArrayI, ok := obj[key]
	if !ok {
		return nil, nullFieldError(key, vm)
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

func getBoolField(key string, obj map[string]interface{}, vm *videoMapper) (bool, error) {
	valueI, ok := obj[key]
	if !ok {
		return false, nullFieldError(key, vm)
	}

	val, ok := valueI.(bool)
	if !ok {
		return false, wrongFieldTypeError("bool", key, valueI)
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

func nullFieldError(fieldKey string, vm *videoMapper) error {
	return fmt.Errorf("[%s] field of native Next video JSON is missing: [%s]", fieldKey, vm.strContent)
}

func wrongFieldTypeError(expectedType, fieldKey string, value interface{}) error {
	return fmt.Errorf("[%s] field of native Next video JSON is not of type %s: [%v]", fieldKey, expectedType, value)
}
