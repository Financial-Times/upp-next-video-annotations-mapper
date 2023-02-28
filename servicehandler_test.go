package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
)

func TestMapRequest(t *testing.T) {
	h := newServiceHandler(serviceConfig{}, logger.NewUPPLogger("upp-next-video-annotations-test-logger", "debug"))

	tests := []struct {
		fileName           string
		expectedContent    string
		expectedHTTPStatus int
	}{
		{
			"next-video-input.json",
			newStringConceptAnnotation(t, "e2290d14-7e80-4db8-a715-949da4de9a07",
				[]annotation{{"http://api.ft.com/things/71a5efa5-e6e0-3ce1-9190-a7eac8bef325", "isClassifiedBy", defaultRelevanceScore, defaultConfidenceScore}},
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
			assert.NoError(t, err)
		case test.expectedHTTPStatus != http.StatusOK:
			assert.Equal(t, test.expectedHTTPStatus, http.StatusBadRequest, "HTTP status wrong. Input JSON: %s", test.fileName)
		default:
			assert.Equal(t, test.expectedHTTPStatus, http.StatusOK, "HTTP status wrong. Input JSON: %s", test.fileName)
			assert.Equal(t, test.expectedContent, string(body), "Marshalled content wrong. Input JSON: %s", test.fileName)
		}
	}
}

func getReader(fileName string, t *testing.T) *os.File {
	file, err := os.Open("test-resources/" + fileName)
	if err != nil {
		assert.NoError(t, err)
		return nil
	}

	return file
}
