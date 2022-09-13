package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Financial-Times/go-logger/v2"
)

type serviceHandler struct {
	sc  serviceConfig
	log *logger.UPPLogger
}

func newServiceHandler(sc serviceConfig, log *logger.UPPLogger) *serviceHandler {
	return &serviceHandler{sc: sc, log: log}
}

func (h serviceHandler) mapRequest(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writerBadRequest(w, err, "", h.log)
	}
	tid := r.Header.Get("X-Request-Id")

	vm := videoMapper{sc: h.sc, strContent: string(body), tid: tid, log: h.log}

	mappedVideoBytes, _, err := h.mapNextVideoAnnotationsRequest(&vm)
	if err != nil {
		writerBadRequest(w, err, tid, h.log)
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(mappedVideoBytes)
	if err != nil {
		h.log.WithTransactionID(tid).
			WithValidFlag(true).
			WithError(err).
			Error("Writing response error.")
	}
}

func (h serviceHandler) mapNextVideoAnnotationsRequest(vm *videoMapper) ([]byte, string, error) {
	if err := json.Unmarshal([]byte(vm.strContent), &vm.unmarshalled); err != nil {
		return nil, "", fmt.Errorf("video JSON from Next couldn't be unmarshalled: %v. Skipping invalid JSON with tid: %s", err, vm.tid)
	}
	return vm.mapNextVideoAnnotations()
}

func writerBadRequest(w http.ResponseWriter, err error, tid string, log *logger.UPPLogger) {
	w.WriteHeader(http.StatusBadRequest)
	_, err2 := w.Write([]byte(err.Error()))
	if err2 != nil {
		log.WithTransactionID(tid).
			WithValidFlag(false).
			WithError(err).
			Error("Couldn't write Bad Request response.")
	}
}
