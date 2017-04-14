package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func init() {
	logger = newAppLogger("test")
}

func TestMapRequest(t *testing.T) {
	h := serviceHandler{
		sc: serviceConfig{},
	}

	assert := assert.New(t)
	tests := []struct {
		fileName           string
		expectedContent    string
		expectedHTTPStatus int
	}{
		{
			"next-video-input.json",
			newStringConceptSuggestion(t, "e2290d14-7e80-4db8-a715-949da4de9a07",
				newSuggestion("http://api.ft.com/things/71a5efa5-e6e0-3ce1-9190-a7eac8bef325", "isClassifiedBy"),
			),
			http.StatusOK,
		},
		{
			"next-video-invalid-anns-input.json",
			"",
			http.StatusBadRequest,
		},
		{
			"invalid-format.json",
			"",
			http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		fileReader := getReader(test.fileName, t)
		req, _ := http.NewRequest("POST", "http://next-video-annotaitons-mapper.ft.com/map", fileReader)
		w := httptest.NewRecorder()

		h.mapRequest(w, req)

		body, err := ioutil.ReadAll(w.Body)

		switch {
		case err != nil:
			assert.Fail(err.Error())
		case test.expectedHTTPStatus != http.StatusOK:
			assert.Equal(test.expectedHTTPStatus, http.StatusBadRequest, "HTTP status wrong. Input JSON: %s", test.fileName)
		default:
			assert.Equal(test.expectedHTTPStatus, http.StatusOK, "HTTP status wrong. Input JSON: %s", test.fileName)
			assert.Equal(test.expectedContent, string(body), "Marshalled content wrong. Input JSON: %s", test.fileName)
		}
	}
}

func getReader(fileName string, t *testing.T) *os.File {
	file, err := os.Open("test-resources/" + fileName)
	if err != nil {
		assert.Fail(t, err.Error())
		return nil
	}

	return file
}
