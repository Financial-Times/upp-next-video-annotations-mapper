package main

import (
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
)

const serviceDescription = "Gets the Next video content from queue, transforms annotations to an internal representation and puts a new created annotation content to queue."

var logger *appLogger
var timeout = 10 * time.Second
var httpCl = &http.Client{Timeout: timeout}

type serviceConfig struct {
	serviceName string
	appPort     string
}

func main() {
	app := cli.App("next-video-annotations-mapper", serviceDescription)
	serviceName := app.String(cli.StringOpt{
		Name:   "service-name",
		Value:  "next-video-annotations-mapper",
		Desc:   "The name of this service",
		EnvVar: "SERVICE_NAME",
	})
	appName := app.String(cli.StringOpt{
		Name:   "app-name",
		Value:  "Next Video Annotations Mapper",
		Desc:   "The name of the application",
		EnvVar: "APP_NAME",
	})
	systemCode := app.String(cli.StringOpt{
		Name:   "app-system-code",
		Value:  "next-video-annotations-mapper",
		Desc:   "App system code",
		EnvVar: "APP_SYSTEM_CODE",
	})
	appPort := app.String(cli.StringOpt{
		Name:   "app-port",
		Value:  "8084",
		Desc:   "Default port for Next Video Annotations Mapper",
		EnvVar: "APP_PORT",
	})
	panicGuide := app.String(cli.StringOpt{
		Name:   "panic-guide",
		Value:  "https://dewey.ft.com/up-nvam.html",
		Desc:   "Path to panic guide",
		EnvVar: "PANIC_GUIDE",
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
			serviceName: *serviceName,
			appPort:     *appPort,
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

		annMapper := queueHandler{sc: sc, httpCl: httpCl, consumerConfig: consumerConfig, producerConfig: producerConfig}
		annMapper.init()

		sh := serviceHandler{sc}
		hc := newHealthCheck(&producerConfig, &consumerConfig, *appName, *systemCode, *panicGuide)
		go listen(sc, sh, hc)

		consumeUntilSigterm(annMapper.messageConsumer, consumerConfig)
	}
	err := app.Run(os.Args)
	if err != nil {
		println(err)
	}
}

func listen(sc serviceConfig, sh serviceHandler, hc *HealthCheck) {
	r := mux.NewRouter()
	r.Path("/map").Handler(handlers.MethodHandler{"POST": http.HandlerFunc(sh.mapRequest)})
	r.Path(httphandlers.BuildInfoPath).HandlerFunc(httphandlers.BuildInfoHandler)
	r.Path(httphandlers.PingPath).HandlerFunc(httphandlers.PingHandler)
	r.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(hc.Health())})
	r.Path(httphandlers.GTGPath).HandlerFunc(httphandlers.NewGoodToGoHandler(hc.GTG))

	logger.serviceStartedEvent(sc.asMap())

	err := http.ListenAndServe(":"+sc.appPort, r)
	if err != nil {
		logrus.Fatalf("Unable to start server: %v", err)
	}
}

func consumeUntilSigterm(messageConsumer consumer.MessageConsumer, config consumer.QueueConfig) {
	logger.messageEvent(config.Topic, "Starting queue consumer")

	var consumerWaitGroup sync.WaitGroup
	consumerWaitGroup.Add(1)
	go func() {
		messageConsumer.Start()
		consumerWaitGroup.Done()
	}()
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	messageConsumer.Stop()
	consumerWaitGroup.Wait()
}

func (sc serviceConfig) asMap() map[string]interface{} {
	return map[string]interface{}{
		"service-name": sc.serviceName,
		"service-port": sc.appPort,
	}
}
