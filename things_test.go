package main

import (
	"bytes"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const scenarioHappy = "happy"
const scenarioNotHappy = "not_happy"

const thingUUID_1 = "a54fda40-7fe7-339a-9b83-2d7b964ff3a4"
const thingUUID_2 = "71a5efa5-e6e0-3ce1-9190-a7eac8bef325"
const thingUUID_3 = "empty_response"

var publicThingsAPIMock *httptest.Server

func TestRetrieveThingsDetails(t *testing.T) {
	startPublicThingsAPIMock(scenarioHappy)
	defer stopService()

	assert := assert.New(t)
	tests := []struct {
		publicThingsAPIURL string
		thingsUUIDs        []string
		expectedTypes      []string
		expectedLabels     []string
	}{
		{
			publicThingsAPIURLMock(),
			[]string{thingUUID_1, thingUUID_2},
			[]string{"http://www.ft.com/ontology/product/Brand", "http://www.ft.com/ontology/Section"},
			[]string{"Market Minute", "Financials"},
		},
		{
			publicThingsAPIURLMock(),
			[]string{"unknown_id", thingUUID_1},
			[]string{"", "http://www.ft.com/ontology/product/Brand"},
			[]string{"", "Market Minute"},
		},
		{
			publicThingsAPIURLMock(),
			[]string{thingUUID_3},
			[]string{""},
			[]string{""},
		},
	}

	for _, test := range tests {
		h := newThingsHandler(test.publicThingsAPIURL)
		anns := newAnnotationsForThingsEndpoint(test.thingsUUIDs)
		h.retrieveThingsDetails(anns)

		for idx, ann := range anns {
			assert.Equal(test.expectedTypes[idx], ann.thing.directType,
				"Types wrong. Test thing uuids: %v, endpoint url: %s", test.thingsUUIDs, test.publicThingsAPIURL)
			assert.Equal(test.expectedLabels[idx], ann.thing.prefLabel,
				"Labels wrong. Test thing uuids: %v, endpoint url: %s", test.thingsUUIDs, test.publicThingsAPIURL)
		}
	}
}

func TestGetThing(t *testing.T) {
	startPublicThingsAPIMock(scenarioHappy)
	defer stopService()

	assert := assert.New(t)
	tests := []struct {
		publicThingsAPIURL string
		thingUUID          string
		expectedRespEmpty  bool
		expectedStatus     bool
	}{
		{
			publicThingsAPIURLMock(),
			thingUUID_1,
			false,
			true,
		},
		{
			publicThingsAPIURLMock(),
			thingUUID_3,
			true,
			true,
		},
		{
			publicThingsAPIURLMock(),
			"unknown_thing",
			true,
			false,
		},
		{
			"http://invalid_url",
			thingUUID_1,
			true,
			false,
		},
	}

	for _, test := range tests {
		h := newThingsHandler(test.publicThingsAPIURL)
		resp, ok := h.getThing(test.thingUUID)

		assert.Equal(test.expectedStatus, ok, "Return status is wrong. Test thing uuid: %s, endpoint url: %s", test.thingUUID, test.publicThingsAPIURL)
		if ok {
			assert.Equal(test.expectedRespEmpty, isEmpty(resp),
				"Response body emptiness status is wrong. Test thing uuid: %s, endpoint url: %s", test.thingUUID, test.publicThingsAPIURL)
		}
	}
}

func TestGetThingDetails(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		handler        http.HandlerFunc
		expectedType   string
		expectedLabel  string
		expectedStatus bool
	}{
		{
			happyResponseHandler1,
			"http://www.ft.com/ontology/product/Brand",
			"Market Minute",
			true,
		},
		{
			invalidContentMock,
			"",
			"",
			false,
		},
		{
			missingTypeFieldResponseHandler,
			"",
			"",
			false,
		},
		{
			missingLabelFieldResponseHandler,
			"http://www.ft.com/ontology/product/Brand",
			"",
			true,
		},
		{
			emptyTypeFieldResponseHandler,
			"",
			"Market Minute",
			true,
		},
	}

	for _, test := range tests {
		h := newThingsHandler("test_url")
		respRecorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "http://"+h.publicThingsURI, nil)
		test.handler(respRecorder, req)
		resp := respRecorder.Result()
		resp.Request = req

		thingType, thingLabel, ok := h.getThingDetails(resp, "")

		assert.Equal(test.expectedStatus, ok, "Status is wrong.")
		if ok {
			assert.Equal(test.expectedType, thingType, "Thing type is wrong. Input handler: %v", test.handler)
			assert.Equal(test.expectedLabel, thingLabel, "Thing label is wrong. Input handler: %v", test.handler)
		}
	}
}

func startPublicThingsAPIMock(status string) {
	router := mux.NewRouter()
	var getContent1 http.HandlerFunc
	var getContent2 http.HandlerFunc
	var getEmptyContent http.HandlerFunc
	var health http.HandlerFunc

	switch status {
	case scenarioHappy:
		getContent1 = happyResponseHandler1
		getContent2 = happyReponseHandler2
		getEmptyContent = emptyResponsePublicThingsAPIMock
		health = happyHandler
	case scenarioNotHappy:
		getContent1 = internalErrorHandler
		health = internalErrorHandler
	}

	router.Path("/" + thingUUID_1).Handler(handlers.MethodHandler{"GET": http.HandlerFunc(getContent1)})
	router.Path("/" + thingUUID_2).Handler(handlers.MethodHandler{"GET": http.HandlerFunc(getContent2)})
	router.Path("/" + thingUUID_3).Handler(handlers.MethodHandler{"GET": http.HandlerFunc(getEmptyContent)})
	router.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(health)})

	publicThingsAPIMock = httptest.NewServer(router)
}

func happyResponseHandler1(writer http.ResponseWriter, request *http.Request) {
	fillResponseFromFile("test-resources/public-things-api-output_1.json", writer, request)
}

func happyReponseHandler2(writer http.ResponseWriter, request *http.Request) {
	fillResponseFromFile("test-resources/public-things-api-output_2.json", writer, request)
}

func missingTypeFieldResponseHandler(writer http.ResponseWriter, request *http.Request) {
	fillResponseFromFile("test-resources/public-things-api-missing-type-output.json", writer, request)
}

func missingLabelFieldResponseHandler(writer http.ResponseWriter, request *http.Request) {
	fillResponseFromFile("test-resources/public-things-api-missing-label-output.json", writer, request)
}

func emptyTypeFieldResponseHandler(writer http.ResponseWriter, request *http.Request) {
	fillResponseFromFile("test-resources/public-things-api-empty-type-output.json", writer, request)
}

func emptyResponsePublicThingsAPIMock(writer http.ResponseWriter, request *http.Request) {
	// empty response
}

func invalidContentMock(writer http.ResponseWriter, request *http.Request) {
	fillResponseFromFile("public-things-api-output_invalid_format.json", writer, request)
}

func fillResponseFromFile(fileName string, writer http.ResponseWriter, request *http.Request) {
	file, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer file.Close()
	io.Copy(writer, file)
}

func happyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func internalErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func stopService() {
	publicThingsAPIMock.Close()
}

func newThingsHandler(publicThingsURL string) thingHandler {
	return thingHandler{
		serviceConfig{
			publicThingsURI: publicThingsURL + "/",
		},
		"tid",
		"a54fda40-7fe7-339a-9b83-2d7b964ff3a4",
	}
}

func newAnnotationForThingsEndpoint(thingUUID string) annotation {
	return annotation{
		primaryFlag: true,
		thing: &thingInfo{
			uuid: thingUUID,
		},
	}
}

func newAnnotationsForThingsEndpoint(thingsUUIDs []string) []annotation {
	var anns = make([]annotation, 0)
	for _, uuid := range thingsUUIDs {
		anns = append(anns, newAnnotationForThingsEndpoint(uuid))
	}
	return anns
}

func publicThingsAPIURLMock() string {
	return publicThingsAPIMock.URL + "/"
}

func isEmpty(resp *http.Response) bool {
	return getStringFromReader(resp.Body) == ""
}

func getStringFromReader(r io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	return buf.String()
}
