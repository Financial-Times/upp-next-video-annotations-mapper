package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/kafka-client-go/v3"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	cli "github.com/jawher/mow.cli"
)

const serviceDescription = "Gets the Next video content from queue, transforms annotations to an internal representation and puts a new created annotation content to queue."

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
		Value:  "https://runbooks.in.ft.com/up-nvam",
		Desc:   "Path to panic guide",
		EnvVar: "PANIC_GUIDE",
	})
	kafkaAddress := app.String(cli.StringOpt{
		Name:   "queue-kafkaAddress",
		Value:  "",
		Desc:   "Addresses to connect to the queue (hostnames).",
		EnvVar: "KAFKA_ADDR",
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
		Desc:   "The topic to read the messages from.",
		EnvVar: "Q_READ_TOPIC",
	})
	writeTopic := app.String(cli.StringOpt{
		Name:   "write-topic",
		Value:  "V1ConceptAnnotations",
		Desc:   "The topic to write the messages to.",
		EnvVar: "Q_WRITE_TOPIC",
	})
	logLevel := app.String(cli.StringOpt{
		Name:   "logLevel",
		Value:  "INFO",
		Desc:   "Logging level (DEBUG, INFO, WARN, ERROR)",
		EnvVar: "LOG_LEVEL",
	})
	consumerLagTolerance := app.Int(cli.IntOpt{
		Name:   "consumerLagTolerance",
		Value:  120,
		Desc:   "Kafka lag tolerance",
		EnvVar: "KAFKA_LAG_TOLERANCE",
	})

	log := logger.NewUPPLogger(*serviceName, *logLevel)

	log.Infof("[Startup] %s is starting ", *serviceName)

	app.Action = func() {
		if *kafkaAddress == "" {
			log.Info("No queue kafkaAddress provided. Quitting...")
			cli.Exit(1)
		}

		producerConfig := kafka.ProducerConfig{
			BrokersConnectionString: *kafkaAddress,
			Topic:                   *writeTopic,
			ConnectionRetryInterval: time.Minute,
		}
		producer := kafka.NewProducer(producerConfig, log)

		sc := serviceConfig{
			serviceName: *serviceName,
			appPort:     *appPort,
		}
		annMapper := newQueueHandler(sc, producer, log)

		consumerConfig := kafka.ConsumerConfig{
			BrokersConnectionString: *kafkaAddress,
			ConsumerGroup:           *group,
			ConnectionRetryInterval: time.Minute,
		}
		topics := []*kafka.Topic{
			kafka.NewTopic(*readTopic, kafka.WithLagTolerance(int64(*consumerLagTolerance))),
		}
		consumer := kafka.NewConsumer(consumerConfig, topics, log)
		go consumer.Start(annMapper.queueConsume)
		defer func(consumer *kafka.Consumer) {
			err := consumer.Close()
			if err != nil {
				log.WithError(err).Error("Consumer could not stop")
			}
		}(consumer)

		sh := newServiceHandler(sc, log)
		hc := NewHealthCheck(producer, consumer, *appName, *systemCode, *panicGuide)
		go listen(sh, hc, log)

		log.Infof("[Shutdown] %s is shutting down", *appName)
		waitForSignal()
	}
	err := app.Run(os.Args)
	if err != nil {
		log.WithError(err).Errorf("%s failed to start", *appName)
	}
}

func listen(sh *serviceHandler, hc *HealthCheck, log *logger.UPPLogger) {
	r := mux.NewRouter()
	r.Path("/map").Handler(handlers.MethodHandler{"POST": http.HandlerFunc(sh.mapRequest)})
	r.Path(httphandlers.BuildInfoPath).HandlerFunc(httphandlers.BuildInfoHandler)
	r.Path(httphandlers.PingPath).HandlerFunc(httphandlers.PingHandler)
	r.Path("/__health").Handler(handlers.MethodHandler{"GET": http.HandlerFunc(hc.Health())})
	r.Path(httphandlers.GTGPath).HandlerFunc(httphandlers.NewGoodToGoHandler(hc.GTG))

	log.WithFields(sh.sc.asMap()).Info("Service started with configuration")

	err := http.ListenAndServe(":"+sh.sc.appPort, r)
	if err != nil {
		log.WithField("message", err).Info("Closing HTTP server")
	}
}

func waitForSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}

func (sc serviceConfig) asMap() map[string]interface{} {
	return map[string]interface{}{
		"service-name": sc.serviceName,
		"service-port": sc.appPort,
	}
}
