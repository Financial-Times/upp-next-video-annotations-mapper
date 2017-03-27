package main

import (
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"net/http"
	"os"
	"github.com/gorilla/handlers"
	"time"
)

const serviceDescription = "A RESTful API for mapping Next video editor annotations to UPP annotations"

var timeout = 10 * time.Second
var client = &http.Client{Timeout: timeout}

type serviceConfig struct {
	serviceName                           string
	appPort                               string
	handlerPath                           string
}

func main() {
	app := cli.App("upp-next-video-annotations-mapper", serviceDescription)
	serviceName := app.StringOpt("app-name", "internal-content-api", "The name of this service")
	appPort := app.String(cli.StringOpt{
		Name:   "app-port",
		Value:  "8084",
		Desc:   "Default port for Next Video Annotations Mapper",
		EnvVar: "APP_PORT",
	})

	app.Action = func() {
		sc := serviceConfig{
			serviceName: *serviceName,
			appPort: *appPort,
		}
		h := setupServiceHandler(sc)
		err := http.ListenAndServe(":" + *appPort, h)
		if err != nil {
			logrus.Fatalf("Unable to start server: %v", err)
		}
	}
	app.Run(os.Args)
}

func setupServiceHandler(sc serviceConfig) *mux.Router {
	r := mux.NewRouter()
	r.Path(httphandlers.BuildInfoPath).HandlerFunc(httphandlers.BuildInfoHandler)
	r.Path(httphandlers.PingPath).HandlerFunc(httphandlers.PingHandler)
	r.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(fthealth.Handler(sc.serviceName, serviceDescription, sc.publicThingsAppCheck()))})
	return r
}
