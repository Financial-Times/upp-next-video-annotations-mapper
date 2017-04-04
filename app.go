package main

import (
	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const serviceDescription = "A RESTful API for mapping Next video editor annotations to UPP annotations"

var timeout = 10 * time.Second
var client = &http.Client{Timeout: timeout}
var logger *appLogger

type serviceConfig struct {
	serviceName               string
	appPort                   string
	cacheControlPolicy        string
	publicThingsURI           string
	publicThingsAppName       string
	publicThingsAppHealthURI  string
	publicThingsAppPanicGuide string
	envAPIHost                string
	graphiteTCPAddress        string
	graphitePrefix            string
}

func main() {
	app := cli.App("upp-next-video-annotations-mapper", serviceDescription)
	serviceName := app.StringOpt("app-name", "next-video-annotations-mapper", "The name of this service")
	appPort := app.String(cli.StringOpt{
		Name:   "app-port",
		Value:  "8084",
		Desc:   "Default port for Next Video Annotations Mapper",
		EnvVar: "APP_PORT",
	})
	cacheControlPolicy := app.String(cli.StringOpt{
		Name:   "cache-control-policy",
		Value:  "no-store",
		Desc:   "Cache control policy header",
		EnvVar: "CACHE_CONTROL_POLICY",
	})
	publicThingsURI := app.String(cli.StringOpt{
		Name:   "public-things-uri",
		Value:  "http://localhost:8080/__public-things-api/things/",
		Desc:   "Public Things URI",
		EnvVar: "PUBLIC_THINGS_URI",
	})
	publicThingsAppName := app.String(cli.StringOpt{
		Name:   "public-things-app-name",
		Value:  "Public Things Service",
		Desc:   "Service name of the public things application",
		EnvVar: "PUBLIC_THINGS_APP_NAME",
	})
	publicThingsAppHealthURI := app.String(cli.StringOpt{
		Name:   "public-things-app-health-uri",
		Value:  "http://localhost:8080/__public-things-api/__health",
		Desc:   "URI of the Public Things Application health endpoint",
		EnvVar: "PUBLIC_THINGS_APP_HEALTH_URI",
	})
	// TODO provide correct panic guide (on service file also)
	publicThingsAppPanicGuide := app.String(cli.StringOpt{
		Name:   "public-things-app-panic-guide",
		Value:  "https://sites.google.com/a/ft.com/dynamic-publishing-team/public-things-panic-guide",
		Desc:   "Public things appllication panic guide url for healthcheck. Default panic guide is for public things.",
		EnvVar: "PUBLIC_THINGS_APP_PANIC_GUIDE",
	})
	envAPIHost := app.String(cli.StringOpt{
		Name:   "env-api-host",
		Value:  "api.ft.com",
		Desc:   "API host to use for URLs in responses",
		EnvVar: "ENV_API_HOST",
	})
	graphiteTCPAddress := app.String(cli.StringOpt{
		Name:   "graphite-tcp-address",
		Value:  "",
		Desc:   "Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally)",
		EnvVar: "GRAPHITE_TCP_ADDRESS",
	})
	graphitePrefix := app.String(cli.StringOpt{
		Name:   "graphite-prefix",
		Value:  "coco.services.$ENV.content-preview.0",
		Desc:   "Prefix to use. Should start with content, include the environment, and the host name. e.g. coco.pre-prod.sections-rw-neo4j.1",
		EnvVar: "GRAPHITE_PREFIX",
	})
	addresses := app.Strings(cli.StringsOpt{
		Name:   "queue-addresses",
		Value:  []string{"http://localhost:8080"},
		Desc:   "Addresses to connect to the queue (hostnames).",
		EnvVar: "Q_ADDR",
	})
	group := app.String(cli.StringOpt{
		Name:   "group",
		Value:  "videoAnnotationsMapper",
		Desc:   "Group used to read the messages from the queue.",
		EnvVar: "Q_GROUP",
	})
	readTopic := app.String(cli.StringOpt{
		Name:   "read-topic",
		Value:  "NativeCmsPublicationEvents",
		Desc:   "The topic to read the meassages from.",
		EnvVar: "Q_READ_TOPIC",
	})
	readQueue := app.String(cli.StringOpt{
		Name:   "read-queue",
		Value:  "kafka",
		Desc:   "The queue to read the meassages from.",
		EnvVar: "Q_READ_QUEUE",
	})
	writeTopic := app.String(cli.StringOpt{
		Name:   "write-topic",
		Value:  "V1ConceptAnnotations",
		Desc:   "The topic to write the meassages to.",
		EnvVar: "Q_WRITE_TOPIC",
	})
	writeQueue := app.String(cli.StringOpt{
		Name:   "write-queue",
		Value:  "kafka",
		Desc:   "The queue to write the meassages to.",
		EnvVar: "Q_WRITE_QUEUE",
	})

	app.Action = func() {
		logger = newAppLogger(*serviceName)
		if len(*addresses) == 0 {
			logger.log.Info("No queue address provided. Quitting...")
			cli.Exit(1)
		}
		sc := serviceConfig{
			serviceName:               *serviceName,
			appPort:                   *appPort,
			cacheControlPolicy:        *cacheControlPolicy,
			publicThingsURI:           *publicThingsURI,
			publicThingsAppName:       *publicThingsAppName,
			publicThingsAppHealthURI:  *publicThingsAppHealthURI,
			publicThingsAppPanicGuide: *publicThingsAppPanicGuide,
			graphiteTCPAddress:        *graphiteTCPAddress,
			graphitePrefix:            *graphitePrefix,
			envAPIHost:                *envAPIHost,
		}
		consumerConfig := consumer.QueueConfig{
			Addrs:                *addresses,
			Group:                *group,
			Topic:                *readTopic,
			Queue:                *readQueue,
			ConcurrentProcessing: false,
			AutoCommitEnable:     true,
		}
		producerConfig := producer.MessageProducerConfig{
			Addr:  (*addresses)[0],
			Topic: *writeTopic,
			Queue: *writeQueue,
		}

		annMapper := annotationMapper{sc: sc, consumerConfig: consumerConfig, producerConfig: producerConfig}
		annMapper.init()

		go listen(sc)

		consumeUntilSigterm(annMapper.messageConsumer, consumerConfig)
	}
	err := app.Run(os.Args)
	if err != nil {
		println(err)
	}
}

func listen(sc serviceConfig) {
	r := mux.NewRouter()
	r.Path(httphandlers.BuildInfoPath).HandlerFunc(httphandlers.BuildInfoHandler)
	r.Path(httphandlers.PingPath).HandlerFunc(httphandlers.PingHandler)
	r.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(fthealth.Handler(sc.serviceName, serviceDescription, sc.publicThingsAppCheck()))})

	logger.serviceStartedEvent(sc.asMap())

	err := http.ListenAndServe(":"+sc.appPort, r)
	if err != nil {
		logrus.Fatalf("Unable to start server: %v", err)
	}
}

func consumeUntilSigterm(messageConsumer *consumer.MessageConsumer, config consumer.QueueConfig) {
	logger.messageEvent(config.Topic, "Starting queue consumer")

	var consumerWaitGroup sync.WaitGroup
	consumerWaitGroup.Add(1)
	go func() {
		(*messageConsumer).Start()
		consumerWaitGroup.Done()
	}()
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	(*messageConsumer).Stop()
	consumerWaitGroup.Wait()
}

func (sc serviceConfig) asMap() map[string]interface{} {
	return map[string]interface{}{
		"service-name":                 sc.serviceName,
		"service-port":                 sc.appPort,
		"cache-control-policy":         sc.cacheControlPolicy,
		"public-things-app":            sc.publicThingsAppName,
		"public-things-uri":            sc.publicThingsURI,
		"public-things-app-health-uri": sc.publicThingsAppHealthURI,
		"public-things-app-guide":      sc.publicThingsAppPanicGuide,
		"env-api-host":                 sc.envAPIHost,
		"graphite-tcp-address":         sc.graphiteTCPAddress,
		"graphite-prefix":              sc.graphitePrefix,
	}
}
