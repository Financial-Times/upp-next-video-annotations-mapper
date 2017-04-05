package main

import (
	tid "github.com/Financial-Times/transactionid-utils-go"
	"github.com/Sirupsen/logrus"
	"net/http"
)

type queueEvent struct {
	serviceName   string
	queueName     string
	queueTopic    string
	transactionID string
}

type appLogger struct {
	log         *logrus.Logger
	serviceName string
}

func newAppLogger(serviceName string) *appLogger {
	logrus.SetLevel(logrus.InfoLevel)
	log := logrus.New()
	log.Formatter = new(logrus.JSONFormatter)
	return &appLogger{log, serviceName}
}

func (appLogger *appLogger) serviceStartedEvent(serviceConfig map[string]interface{}) {
	serviceConfig["event"] = "service_started"
	appLogger.log.WithFields(serviceConfig).Infof("%s started with configuration", appLogger.serviceName)
}

func (appLogger *appLogger) messageEvent(queueTopic string, message string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":        "consume_queue",
		"service_name": appLogger.serviceName,
		"queue_topic":  queueTopic,
	}).Info(message)
}

func (appLogger *appLogger) warnMessageEvent(event queueEvent, videoUUID string, err error, errMessage string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "error",
		"service_name":   event.serviceName,
		"queue_name":     event.queueName,
		"queue_topic":    event.queueTopic,
		"transaction_id": event.transactionID,
		"uuid":           videoUUID,
		"error":          err,
	}).Warn(errMessage)
}

func (appLogger *appLogger) messageSentEvent(event queueEvent, uuid string, message string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "produce_queue",
		"service_name":   event.serviceName,
		"queue_name":     event.queueName,
		"queue_topic":    event.queueTopic,
		"transaction_id": event.transactionID,
		"uuid":           uuid,
	}).Info(message)
}

func (appLogger *appLogger) thingEvent(transactionID string, thingUUID string, videoUUID string, message string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "error",
		"service_name":   appLogger.serviceName,
		"transaction_id": transactionID,
		"thing_uuid":     thingUUID,
		"uuid":           videoUUID,
	}).Warn(message)
}

func (appLogger *appLogger) videoEvent(transactionID string, videoUUID string, message string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "error",
		"service_name":   appLogger.serviceName,
		"transaction_id": transactionID,
		"uuid":           videoUUID,
	}).Warn(message)
}

func (appLogger *appLogger) videoErrorEvent(transactionID string, videoUUID string, err error, message string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "error",
		"service_name":   appLogger.serviceName,
		"transaction_id": transactionID,
		"error":          err,
		"uuid":           videoUUID,
	}).Warn(message)
}

//func (appLogger *appLogger) TransactionStartedEvent(requestURL string, transactionID string, uuid string) {
//	appLogger.log.WithFields(logrus.Fields{
//		"event":          "transaction_started",
//		"request_url":    requestURL,
//		"transaction_id": transactionID,
//		"uuid":           uuid,
//	}).Info()
//}
//
func (appLogger *appLogger) requestEvent(serviceName string, requestURL string, transactionID string, thingUUID string, videoUUID string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "request",
		"service_name":   serviceName,
		"request_url":    requestURL,
		"transaction_id": transactionID,
		"thing_uuid":     thingUUID,
		"uuid":           videoUUID,
	}).Info()
}

func (appLogger *appLogger) errorEvent(serviceName string, requestURL string, transactionID string, err error, thingUUID string, videoUUID, errMessage string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "error",
		"service_name":   serviceName,
		"request_url":    requestURL,
		"transaction_id": transactionID,
		"error":          err,
		"thing_uuid":     thingUUID,
		"uuid":           videoUUID,
	}).Error(errMessage)

}

func (appLogger *appLogger) requestFailedEvent(serviceName string, requestURL string, resp *http.Response, thingUUID string, videoUUID string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "request_failed",
		"service_name":   serviceName,
		"request_url":    requestURL,
		"transaction_id": resp.Header.Get(tid.TransactionIDHeader),
		"status":         resp.StatusCode,
		"thing_uuid":     thingUUID,
		"uuid":           videoUUID,
	}).Warnf("Request failed. %s responded with %s", serviceName, resp.Status)
}

func (appLogger *appLogger) responseEvent(serviceName string, requestURL string, resp *http.Response, thingUUID string, videoUUID string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "response",
		"service_name":   serviceName,
		"status":         resp.StatusCode,
		"request_url":    requestURL,
		"transaction_id": resp.Header.Get(tid.TransactionIDHeader),
		"thing_uuid":     thingUUID,
		"uuid":           videoUUID,
	}).Info("Response from " + serviceName)
}

func (appLogger *appLogger) serviceEvent(transactionID string, err error, message string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "error",
		"service_name":   appLogger.serviceName,
		"transaction_id": transactionID,
		"error":          err,
	}).Warn(message)
}
