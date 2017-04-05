package main

import (
	"io/ioutil"
	"net/http"
	"encoding/json"
	"fmt"
)

type serviceHandler struct {
	sc serviceConfig
}

func (h serviceHandler) mapRequest(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writerBadRequest(w, err, "")
	}
	tid := r.Header.Get("X-Request-Id")

	vm := videoMapper{sc: h.sc, strContent: string(body), tid: tid}

	mappedVideoBytes, _, err := h.mapNextVideoAnnotationsRequest(&vm)
	if err != nil {
		writerBadRequest(w, err, tid)
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(mappedVideoBytes)
	if err != nil {
		logger.serviceEvent(tid, err, "Writing response error.")
	}
}

func (h serviceHandler) mapNextVideoAnnotationsRequest(vm *videoMapper) ([]byte, string, error) {
	if err := json.Unmarshal([]byte(vm.strContent), &vm.unmarshalled); err != nil {
		return nil, "", fmt.Errorf("Video JSON from Next couldn't be unmarshalled: %v. Skipping invalid JSON: %v", err.Error(), vm.strContent)
	}
	return vm.mapNextVideoAnnotations()
}

func writerBadRequest(w http.ResponseWriter, err error, tid string) {
	w.WriteHeader(http.StatusBadRequest)
	_, err2 := w.Write([]byte(err.Error()))
	if err2 != nil {
		logger.serviceEvent(tid, err, "Couldn't write Bad Request response.")
	}
	return
}
