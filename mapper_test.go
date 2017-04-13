package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

var testMap = make(map[string]interface{})

func init() {
	logger = newAppLogger("test")
	testMap["string"] = "value1"
	testMap["nullstring"] = nil
	testMap["bool"] = true

	var objArray = make([]interface{}, 0)
	var obj = make(map[string]interface{}, 0)
	obj["field1"] = "test"
	obj["field2"] = true
	objArray = append(objArray, obj)
	testMap["objArray"] = objArray

	testMap["emptyObjArray"] = make([]interface{}, 0)
}

func TestBuildAnnotations(t *testing.T) {
	assert := assert.New(t)
	vm := videoMapper{}
	tests := []struct {
		nextAnns     []map[string]interface{}
		expectedAnns []annotation
	}{
		{
			[]map[string]interface{}{
				newNextAnnotation("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740", "http://www.ft.com/ontology/classification/isClassifiedBy"),
				newNextAnnotation("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-123456666677", "unknown_predicate_id"),
				newNextAnnotation("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-123456666677", "http://www.ft.com/ontology/annotation/mentions"),
			},
			[]annotation{
				{"http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740", "isClassifiedBy"},
				{"http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-123456666677", "mentions"},
			},
		},
		{
			[]map[string]interface{}{
				newNextAnnotation(nil, "http://www.ft.com/ontology/annotation/mentions"),
			},
			[]annotation{},
		},
		{
			[]map[string]interface{}{
				newNextAnnotation("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740", nil),
			},
			[]annotation{},
		},
	}
	for _, test := range tests {
		anns := vm.retrieveAnnotations(test.nextAnns, "")
		assert.Equal(test.expectedAnns, anns, "Annotations are wrong. Test input: [%v]", test.nextAnns)
	}
}

func TestMapNextVideoAnnotationsHappyFlow(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		fileName          string
		expectedContent   string
		expectedVideoUUID string
		expectedErrStatus bool
	}{
		{
			"next-video-input.json",
			newStringConceptSuggestion(t, "e2290d14-7e80-4db8-a715-949da4de9a07",
				newSuggestion("http://api.ft.com/things/71a5efa5-e6e0-3ce1-9190-a7eac8bef325", "isClassifiedBy"),
			),
			"e2290d14-7e80-4db8-a715-949da4de9a07",
			false,
		},
	}

	for _, test := range tests {
		nextVideo, err := readContent(test.fileName)
		if err != nil {
			assert.Fail(err.Error())
		}
		vm := videoMapper{
			sc:           serviceConfig{},
			unmarshalled: nextVideo,
		}

		marshalledContent, videoUUID, err := vm.mapNextVideoAnnotations()

		assert.Equal(test.expectedContent, string(marshalledContent), "Marshalled content wrong. Input JSON: %s", test.fileName)
		assert.Equal(test.expectedVideoUUID, videoUUID, "Video UUID wrong. Input JSON: %s", test.fileName)
		assert.Equal(test.expectedErrStatus, err != nil, "Error status wrong. Input JSON: %s", test.fileName)
	}
}

func TestMapNextVideoAnnotationsMissingFields(t *testing.T) {
	assert := assert.New(t)
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
		if err != nil {
			assert.Fail(err.Error())
		}
		vm := videoMapper{unmarshalled: nextVideo}
		_, _, err = vm.mapNextVideoAnnotations()
		assert.Equal(test.expectedErrStatus, err != nil, "Error status wrong. Input JSON: %s", test.fileName)
	}
}

func TestMapNextVideoAnnotationsDeleteEvent(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		fileName           string
		expectedContentNil bool
		expectedErrStatus  bool
	}{
		{
			"next-video-delete-input.json",
			true,
			false,
		},
	}

	for _, test := range tests {
		nextVideo, err := readContent(test.fileName)
		if err != nil {
			assert.Fail(err.Error())
		}
		vm := videoMapper{unmarshalled: nextVideo}
		marshalledContent, _, err := vm.mapNextVideoAnnotations()
		assert.Equal(test.expectedContentNil, marshalledContent == nil, "Marshalled content nil status wrong. Input JSON: %s", test.fileName)
		assert.Equal(test.expectedErrStatus, err != nil, "Error status wrong. Input JSON: %s", test.fileName)
	}
}

func TestGetRequiredStringField(t *testing.T) {
	assert := assert.New(t)
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
		assert.Equal(test.expectedValue, result, "Value is wrong. Map key: %s", test.key)
		assert.Equal(test.expectedIsErr, err != nil, "Error status is wrong. Map key: %s", test.key)
	}
}

func TestGetBoolField(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		key           string
		expectedValue bool
		expectedIsErr bool
	}{
		{
			"bool",
			true,
			false,
		},
		{
			"string",
			false,
			true,
		},
		{
			"no_key",
			false,
			true,
		},
	}

	for _, test := range tests {
		result, err := getBoolField(test.key, testMap)
		assert.Equal(test.expectedValue, result, "Value is wrong. Map key: %s", test.key)
		assert.Equal(test.expectedIsErr, err != nil, "Error status is wrong. Map key: %s", test.key)
	}
}

func TestGetObjectsArrayField(t *testing.T) {
	assert := assert.New(t)
	vm := videoMapper{}
	var objArray = make([]map[string]interface{}, 0)
	var objValue = make(map[string]interface{}, 0)
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
			nil,
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
		assert.Equal(test.expectedValue, result, "Value is wrong. Map key: %s", test.key)
		assert.Equal(test.expectedIsErr, err != nil, "Error status is wrong. Map key: %s", test.key)
	}
}

func newNextAnnotation(id interface{}, predicate interface{}) map[string]interface{} {
	var obj = make(map[string]interface{})
	if id != nil {
		obj[annotationIdField] = id
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

func newStringConceptSuggestion(t *testing.T, videoUUID string, s suggestion) string {
	cs := newConceptSuggestion(videoUUID, s)
	marshalledContent, err := json.Marshal(cs)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	return string(marshalledContent)
}
