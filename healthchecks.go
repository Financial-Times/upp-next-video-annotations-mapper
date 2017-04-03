package main

import (
	"errors"
	"fmt"
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"net/http"
)

func (sc *serviceConfig) publicThingsAppCheck() fthealth.Check {
	return fthealth.Check{
		BusinessImpact:   "Videos cannot be published",
		Name:             sc.publicThingsAppName + " Availabililty Check",
		PanicGuide:       sc.publicThingsAppPanicGuide,
		Severity:         1,
		TechnicalSummary: "Checks that " + sc.publicThingsAppName + " Service is reachable. Next video annotations mapper requests content from " + sc.publicThingsAppName + " service.",
		Checker: func() (string, error) {
			return checkServiceAvailability(sc.publicThingsAppName, sc.publicThingsAppHealthURI)
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
