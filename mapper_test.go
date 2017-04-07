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
				newNextAnnotation("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740", "text1", false),
			},
			[]annotation{
				newAnnWithoutThingDetail("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740", "text1", false, "d969d76e-f8f4-34ae-bc38-95cfd0884740"),
			},
		},
		{
			[]map[string]interface{}{
				newNextAnnotation("", "text1", false),
			},
			[]annotation{},
		},
		{
			[]map[string]interface{}{
				newNextAnnotation(nil, "text1", false),
			},
			[]annotation{},
		},
		{
			[]map[string]interface{}{
				newNextAnnotation("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740", "", false),
			},
			[]annotation{
				newAnnWithoutThingDetail("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740", "", false, "d969d76e-f8f4-34ae-bc38-95cfd0884740"),
			},
		},
		{
			[]map[string]interface{}{
				newNextAnnotation("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740", nil, false),
			},
			[]annotation{
				newAnnWithoutThingDetail("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740", "", false, "d969d76e-f8f4-34ae-bc38-95cfd0884740"),
			},
		},
		{
			[]map[string]interface{}{
				newNextAnnotation("http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740", "text1", nil),
			},
			[]annotation{},
		},
	}
	for _, test := range tests {
		Anns := vm.buildAnnotations(test.nextAnns, "")
		assert.Equal(test.expectedAnns, Anns, "Annotations are wrong. Test input: [%v]", test.nextAnns)
	}
}

func TestMapNextVideoAnnotationsMissingFields(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		fileName  string
		errStatus bool
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
		assert.Equal(test.errStatus, err != nil, "Error status wrong. Input JSON: %s", test.fileName)
	}
}

func TestMapNextVideoAnnotationsDeleteEvent(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		fileName  string
		expectedContentNil bool
		expectedErrStatus bool
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

func TestGetStringField(t *testing.T) {
	assert := assert.New(t)
	vm := videoMapper{}
	tests := []struct {
		key           string
		expectedValue string
		expectedIsErr bool
	}{
		{
			"string",
			"value1",
			false,
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
		result, err := getStringField(test.key, testMap, &vm)
		assert.Equal(test.expectedValue, result, "Value is wrong. Map key: %s", test.key)
		assert.Equal(test.expectedIsErr, err != nil, "Error status is wrong. Map key: %s", test.key)
	}
}

func TestGetBoolField(t *testing.T) {
	assert := assert.New(t)
	vm := videoMapper{}
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
		result, err := getBoolField(test.key, testMap, &vm)
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

func TestParseThingUUID(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		id             string
		expectedUUID   string
		expectedStatus bool
	}{
		{
			"http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740",
			"d969d76e-f8f4-34ae-bc38-95cfd0884740",
			true,
		},
		{
			"d969d76e-f8f4-34ae-bc38-95cfd0884740",
			"d969d76e-f8f4-34ae-bc38-95cfd0884740",
			true,
		},
		{
			"http://api.ft.com/things/",
			"",
			false,
		},
	}

	for _, test := range tests {
		result, status := parseThingUUID(test.id)
		assert.Equal(test.expectedUUID, result, "Parsed value is wrong. Input id: %s", test.id)
		assert.Equal(test.expectedStatus, status, "Error status is wrong. Input id: %s", test.id)
	}
}

func newNextAnnotation(id interface{}, name interface{}, primaryFlag interface{}) map[string]interface{} {
	var obj = make(map[string]interface{})
	if id != nil {
		obj[annotationIdField] = id
	}
	if name != nil {
		obj[annotationNameField] = name
	}
	if primaryFlag != nil {
		obj[annotationPrimaryField] = primaryFlag
	}
	return obj
}

func newAnnWithoutThingDetail(id string, name string, primaryFlag bool, thingUUID string) annotation {
	return annotation{
		thingID:     id,
		thingText:   name,
		primaryFlag: primaryFlag,
		thing: &thingInfo{
			uuid: thingUUID,
		},
	}
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
