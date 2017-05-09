package main

import (
	"errors"
	"fmt"
	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"net/http"
)

type queueHealthCheck struct {
	httpCl        *http.Client
	consumerConf  consumer.QueueConfig
	producerConf  producer.MessageProducerConfig
	appName       string
	appSystemCode string
	panicGuide    string
}

func (h *queueHealthCheck) healthCheck() *fthealth.HealthCheck {
	checks := []fthealth.Check{h.queueCheck()}
	return &fthealth.HealthCheck{SystemCode: h.appSystemCode, Name: h.appName, Description: serviceDescription, Checks: checks}
}

func (h *queueHealthCheck) queueCheck() fthealth.Check {
	return fthealth.Check{
		ID:               "message-queue-proxy-reachable",
		Name:             "Message Queue Proxy Reachable",
		Severity:         1,
		BusinessImpact:   "Annotations from published Next videos will not be created, clients will not see them within content.",
		TechnicalSummary: "Message queue proxy is not reachable/healthy",
		PanicGuide:       h.panicGuide,
		Checker:          h.checkAggregateMessageQueueProxiesReachable,
	}
}

func (h *queueHealthCheck) checkAggregateMessageQueueProxiesReachable() (string, error) {
	var errMsg string

	err := h.checkMessageQueueProxyReachable(h.producerConf.Addr, h.producerConf.Topic, h.producerConf.Authorization, h.producerConf.Queue)
	if err != nil {
		return err.Error(), fmt.Errorf("Health check for queue address %s, topic %s failed. Error: %s", h.producerConf.Addr, h.producerConf.Topic, err.Error())
	}

	for i := 0; i < len(h.consumerConf.Addrs); i++ {
		err := h.checkMessageQueueProxyReachable(h.consumerConf.Addrs[i], h.consumerConf.Topic, h.consumerConf.AuthorizationKey, h.consumerConf.Queue)
		if err == nil {
			return "Ok", nil
		}
		errMsg = errMsg + fmt.Sprintf("Health check for queue address %s, topic %s failed. Error: %s", h.consumerConf.Addrs[i], h.consumerConf.Topic, err.Error())
	}
	return errMsg, errors.New(errMsg)
}

func (h *queueHealthCheck) checkMessageQueueProxyReachable(address string, topic string, authKey string, queue string) error {
	req, err := http.NewRequest("GET", address+"/topics", nil)
	if err != nil {
		logger.messageEvent(topic, fmt.Sprintf("Could not connect to proxy: %v", err.Error()))
		return err
	}
	if len(authKey) > 0 {
		req.Header.Add("Authorization", authKey)
	}
	if len(queue) > 0 {
		req.Host = queue
	}
	resp, err := h.httpCl.Do(req)
	if err != nil {
		logger.messageEvent(topic, fmt.Sprintf("Could not connect to proxy: %v", err.Error()))
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Proxy returned status: %d", resp.StatusCode)
		return errors.New(errMsg)
	}
	return nil
}
