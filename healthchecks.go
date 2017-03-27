package main

import (
	"errors"
	"fmt"
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"net/http"
)

// TODO: add appropriate parameters for checking public-things-api app
func (sc *serviceConfig) publicThingsAppCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Videos cannot be published",
		Name:             "",
		PanicGuide:       "",
		Severity:         1,
		TechnicalSummary: "Checks that " + "" + " Service is reachable. Internal Content Service requests content from " + "" + " service.",
		Checker: func() (string, error) {
			return checkServiceAvailability("", "")
		},
	}
}

func checkServiceAvailability(serviceName string, healthURI string) (string, error) {
	req, err := http.NewRequest("GET", healthURI, nil)
	resp, err := client.Do(req)
	if err != nil {
		msg := fmt.Sprintf("%s service is unreachable: %v", serviceName, err)
		return msg, errors.New(msg)
	}
	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("%s service is not responding with OK. status=%d", serviceName, resp.StatusCode)
		return msg, errors.New(msg)
	}
	return "Ok", nil
}
