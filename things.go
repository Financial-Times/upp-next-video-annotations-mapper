package main

import (
	"encoding/json"
	"fmt"
	tid "github.com/Financial-Times/transactionid-utils-go"
	"github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
)

const thingTypeField = "directType"
const prefLabelField = "prefLabel"

type thingHandler struct {
	serviceConfig
	transactionID string
	videoUUID     string
}

type thingInfo struct {
	uuid       string
	directType string
	prefLabel  string
}

func (t *thingHandler) retrieveThingsDetails(nextAnns []nextAnnotation) {
	var waitGroup sync.WaitGroup
	for _, nextAnn := range nextAnns {
		waitGroup.Add(1)
		go func(thing *thingInfo) {
			defer waitGroup.Done()

			response, ok := t.getThing(thing.uuid)
			defer cleanupResp(response, logger.log)

			if !ok || response == nil {
				// TODO if thing not found propagate relationship !!! consider also wrong primary flag mapping to continue with false as default???
				logger.thingEvent(t.transactionID, thing.uuid, t.videoUUID, "Thing cannot be retrieved")
				return
			}

			directType, prefLabel, ok := t.getThingDetails(response, thing.uuid)
			if !ok {
				logger.thingEvent(t.transactionID, thing.uuid, t.videoUUID, "Thing type cannot be extracted from response")
				return
			}

			thing.directType = directType
			thing.prefLabel = prefLabel

		}(nextAnn.thing)
	}

	waitGroup.Wait()
}

func (t *thingHandler) getThing(thingUUID string) (*http.Response, bool) {
	requestURL := fmt.Sprintf("%s%s", t.publicThingsURI, thingUUID)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		logger.errorEvent(t.publicThingsAppName, requestURL, t.transactionID, err, thingUUID, t.videoUUID,
			fmt.Sprintf("Cannot reach %s host. Status code: %s", t.publicThingsAppName, http.StatusInternalServerError))
		return nil, false
	}

	req.Header.Set(tid.TransactionIDHeader, t.transactionID)
	req.Header.Set("Content-Type", "application/json")

	logger.requestEvent(t.publicThingsAppName, requestURL, t.transactionID, thingUUID, t.videoUUID)

	resp, err := client.Do(req)
	if err != nil {
		logger.errorEvent(t.publicThingsAppName, req.URL.String(), req.Header.Get(tid.TransactionIDHeader), err, thingUUID, t.videoUUID,
			fmt.Sprintf("Cannot reach %s host. Status code: %s", t.publicThingsAppName, http.StatusServiceUnavailable))
		return nil, false
	}

	if resp.StatusCode != http.StatusOK {
		logger.requestFailedEvent(t.publicThingsAppName, req.URL.String(), resp, thingUUID, t.videoUUID)
		return nil, false
	}

	logger.responseEvent(t.publicThingsAppName, req.URL.String(), resp, thingUUID, t.videoUUID)
	return resp, true
}

func (t *thingHandler) getThingDetails(response *http.Response, thingUUID string) (string, string, bool) {
	var thing map[string]interface{}
	thingBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.errorEvent(t.serviceName, response.Request.URL.String(), t.transactionID, err, thingUUID, t.videoUUID,
			"Error while handling the response body")
		return "", "", false
	}

	err = json.Unmarshal(thingBytes, &thing)
	if err != nil {
		logger.errorEvent(t.serviceName, response.Request.URL.String(), t.transactionID, err, thingUUID, t.videoUUID,
			"Error while parsing the response json")
		return "", "", false
	}

	thingType, ok := t.getThingStringField(thingTypeField, thing, response, thingUUID, thingBytes)
	if !ok {
		return "", "", false
	}

	prefLabel, ok := t.getThingStringField(prefLabelField, thing, response, thingUUID, thingBytes)
	if !ok {
		return thingType, "", true
	}

	return thingType, prefLabel, true
}

func cleanupResp(resp *http.Response, log *logrus.Logger) {
	_, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		log.Warningf("[%v]", err)
	}
	err = resp.Body.Close()
	if err != nil {
		log.Warningf("[%v]", err)
	}
}

func (t *thingHandler) getThingStringField(key string, obj map[string]interface{}, response *http.Response, thingUUID string, thingBytes []byte) (string, bool) {
	resultI, ok := obj[key]
	if !ok {
		logger.thingEvent(t.transactionID, thingUUID, t.videoUUID,
			fmt.Sprintf("%v field not found within %s response. Body: [%v]", key, t.publicThingsAppName, fmt.Sprint(thingBytes)))
		return "", false
	}

	return resultI.(string), true
}
