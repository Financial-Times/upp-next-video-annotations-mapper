package main

import (
	"encoding/json"
	"errors"
	"fmt"
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"io/ioutil"
	"net/http"
)

type healthCheck struct {
	client       http.Client
	consumerConf consumer.QueueConfig
	producerConf producer.MessageProducerConfig
	panicGuide   string
}

func (h *healthCheck) check() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Annotations from published Next videos will not be created, clients will not see them within content.",
		Name:             "MessageQueueProxyReachable",
		PanicGuide:       h.panicGuide,
		Severity:         1,
		TechnicalSummary: "Message queue proxy is not reachable/healthy",
		Checker:          h.checkAggregateMessageQueueProxiesReachable,
	}
}

func (h *healthCheck) checkAggregateMessageQueueProxiesReachable() (string, error) {
	errMsg := ""

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

func (h *healthCheck) checkMessageQueueProxyReachable(address string, topic string, authKey string, queue string) error {
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
	resp, err := h.client.Do(req)
	if err != nil {
		logger.messageEvent(topic, fmt.Sprintf("Could not connect to proxy: %v", err.Error()))
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Proxy returned status: %d", resp.StatusCode)
		return errors.New(errMsg)
	}
	body, err := ioutil.ReadAll(resp.Body)
	return checkIfTopicIsPresent(body, topic)
}

func checkIfTopicIsPresent(body []byte, searchedTopic string) error {
	var topics []string
	err := json.Unmarshal(body, &topics)
	if err != nil {
		return fmt.Errorf("Error occured and topic could not be found. %v", err.Error())
	}
	for _, topic := range topics {
		if topic == searchedTopic {
			return nil
		}
	}
	return errors.New("Topic was not found")
}
