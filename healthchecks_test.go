package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	"net/http/httptest"
	"net/http"
)

var publicThingsAPIMock *httptest.Server

func TestCheckPublicThingsAPIServiceAvailability(t *testing.T) {
	startPublicThingsAPIServiceMock("happy")
	defer stopServices()

	result, err := checkServiceAvailability("publicThingsAPI", publicThingsAPIMock.URL + "/__health")
	assert.Equal(t, result, "Ok")
	assert.Nil(t, err)
}

func TestCheckPublicThingsAPIServiceNonAvailability(t *testing.T) {
	startPublicThingsAPIServiceMock("notHappy")
	defer stopServices()

	_, err := checkServiceAvailability("publicThingsAPI", publicThingsAPIMock.URL + "/__health")
	assert.Error(t, err)
}

func startPublicThingsAPIServiceMock(handlerType string) {
	router := mux.NewRouter()
	var health http.HandlerFunc

	switch handlerType {
	case "happy" :
		health = happyHandler
	case "notHappy" :
		health = notHappyHandler
	}

	router.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(health)})

	publicThingsAPIMock = httptest.NewServer(router)
}

func happyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func notHappyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func stopServices() {
	publicThingsAPIMock.Close()
}
