package main

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testMap = make(map[string]interface{})

func init() {
	testMap["string"] = "value1"
	testMap["nullstring"] = nil
	testMap["bool"] = true

	var objArray = make([]interface{}, 0)
	var obj = make(map[string]interface{})
	obj["field1"] = "test"
	obj["field2"] = true
	objArray = append(objArray, obj)
	testMap["objArray"] = objArray

	testMap["emptyObjArray"] = make([]interface{}, 0)
}

func getLogger() *logger.UPPLogger {
	return logger.NewUPPLogger("video-annotations-mapper", "Debug")
}

func TestBuildAnnotations(t *testing.T) {
	vm := videoMapper{
		log: getLogger(),
	}
	tests := []struct {
		nextAnns     []map[string]interface{}
		expectedAnns []tag
	}{
		{
			[]map[string]interface{}{
				newNextAnnotation("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740", "http://www.ft.com/ontology/classification/isClassifiedBy"),
				newNextAnnotation("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-123456666677", "unknown_predicate_id"),
				newNextAnnotation("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-123456666677", "http://www.ft.com/ontology/annotation/mentions"),
			},
			[]tag{
				{"http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740", "isClassifiedBy"},
				{"http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-123456666677", "mentions"},
			},
		},
		{
			[]map[string]interface{}{
				newNextAnnotation(nil, "http://www.ft.com/ontology/annotation/mentions"),
			},
			[]tag{},
		},
		{
			[]map[string]interface{}{
				newNextAnnotation("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740", nil),
			},
			[]tag{},
		},
	}
	for _, test := range tests {
		anns := vm.retrieveAnnotations(test.nextAnns, "")
		assert.Equal(t, test.expectedAnns, anns, "Annotations are wrong. Test input: [%v]", test.nextAnns)
	}
}

func TestMapNextVideoAnnotationsHappyFlow(t *testing.T) {
	tests := []struct {
		fileName          string
		expectedContent   string
		expectedVideoUUID string
		expectedErrStatus bool
	}{
		{
			"next-video-input.json",
			newStringConceptAnnotation(t, "e2290d14-7e80-4db8-a715-949da4de9a07",
				[]annotation{newTestAnnotation("http://api.ft.com/things/71a5efa5-e6e0-3ce1-9190-a7eac8bef325", "isClassifiedBy")},
			),
			"e2290d14-7e80-4db8-a715-949da4de9a07",
			false,
		},
	}

	for _, test := range tests {
		nextVideo, err := readContent(test.fileName)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		vm := videoMapper{
			sc:           serviceConfig{},
			unmarshalled: nextVideo,
			log:          getLogger(),
		}

		marshalledContent, videoUUID, err := vm.mapNextVideoAnnotations()

		assert.Equal(t, test.expectedContent, string(marshalledContent), "Marshalled content wrong. Input JSON: %s", test.fileName)
		assert.Equal(t, test.expectedVideoUUID, videoUUID, "Video UUID wrong. Input JSON: %s", test.fileName)
		assert.Equal(t, test.expectedErrStatus, err != nil, "Error status wrong. Input JSON: %s", test.fileName)
	}
}

func TestMapNextVideoAnnotationsMissingFields(t *testing.T) {
	tests := []struct {
		fileName          string
		expectedErrStatus bool
	}{
		{
			"next-video-no-anns-input.json",
			false,
		},
		{
			"next-video-empty-anns-input.json",
			false,
		},
		{
			"next-video-invalid-anns-input.json",
			true,
		},
		{
			"next-video-no-videouuid-input.json",
			true,
		},
	}

	for _, test := range tests {
		nextVideo, err := readContent(test.fileName)
		require.NoError(t, err)
		vm := videoMapper{
			unmarshalled: nextVideo,
			log:          getLogger(),
		}
		_, _, err = vm.mapNextVideoAnnotations()
		assert.Equal(t, test.expectedErrStatus, err != nil, "Error status wrong. Input JSON: %s", test.fileName)
	}
}

func TestMapNextVideoAnnotationsDeleteEvent(t *testing.T) {
	tests := []struct {
		fileName          string
		expectedContent   ConceptAnnotation
		expectedErrStatus bool
	}{
		{
			"next-video-delete-input.json",
			ConceptAnnotation{"e2290d14-7e80-4db8-a715-949da4de9a07", []annotation{}},
			false,
		},
	}

	for _, test := range tests {
		nextVideo, err := readContent(test.fileName)
		require.NoError(t, err)
		vm := videoMapper{
			unmarshalled: nextVideo,
			log:          getLogger(),
		}
		marshalledContent, _, err := vm.mapNextVideoAnnotations()
		var concept ConceptAnnotation
		err = json.Unmarshal(marshalledContent, &concept)
		assert.Nil(t, err)
		assert.Equal(t, test.expectedContent, concept, "Marshalled content differs from expected. Input JSON: %s", test.fileName)
		assert.Equal(t, test.expectedErrStatus, err != nil, "Error status wrong. Input JSON: %s", test.fileName)
	}
}

func TestGetRequiredStringField(t *testing.T) {
	tests := []struct {
		key           string
		expectedValue interface{}
		expectedIsErr bool
	}{
		{
			"string",
			"value1",
			false,
		},
		{
			"nullstring",
			"",
			true,
		},
		{
			"bool",
			"",
			true,
		},
		{
			"no_key",
			"",
			true,
		},
	}

	for _, test := range tests {
		result, err := getRequiredStringField(test.key, testMap)
		assert.Equal(t, test.expectedValue, result, "Value is wrong. Map key: %s", test.key)
		assert.Equal(t, test.expectedIsErr, err != nil, "Error status is wrong. Map key: %s", test.key)
	}
}

func TestGetObjectsArrayField(t *testing.T) {
	vm := videoMapper{
		log: getLogger(),
	}
	var objArray = make([]map[string]interface{}, 0)
	var objValue = make(map[string]interface{})
	objValue["field1"] = "test"
	objValue["field2"] = true
	objArray = append(objArray, objValue)
	tests := []struct {
		key           string
		expectedValue []map[string]interface{}
		expectedIsErr bool
	}{
		{
			"objArray",
			objArray,
			false,
		},
		{
			"string",
			nil,
			true,
		},
		{
			"no_key",
			make([]map[string]interface{}, 0),
			false,
		},
		{
			"emptyObjArray",
			make([]map[string]interface{}, 0),
			false,
		},
	}

	for _, test := range tests {
		result, err := getObjectsArrayField(test.key, testMap, "videoUUID", &vm)
		assert.Equal(t, test.expectedValue, result, "Value is wrong. Map key: %s", test.key)
		assert.Equal(t, test.expectedIsErr, err != nil, "Error status is wrong. Map key: %s", test.key)
	}
}

func newNextAnnotation(id interface{}, predicate interface{}) map[string]interface{} {
	var obj = make(map[string]interface{})
	if id != nil {
		obj[annotationIDField] = id
	}
	if predicate != nil {
		obj[annotationPredicateField] = predicate
	}
	return obj
}

func readContent(fileName string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile("test-resources/" + fileName)
	if err != nil {
		return nil, err
	}

	var result = make(map[string]interface{})
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func newStringConceptAnnotation(t *testing.T, videoUUID string, s []annotation) string {
	var annotations = make([]annotation, 0)
	if s != nil {
		annotations = append(annotations, s...)
	}
	cs := newConceptAnnotation(videoUUID, annotations...)
	marshalledContent, err := json.Marshal(cs)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	return string(marshalledContent)
}
