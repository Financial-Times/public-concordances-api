package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/http-handlers-go/v2/httphandlers"
	"github.com/Financial-Times/neo-utils-go/v2/neoutils"

	"github.com/Financial-Times/public-concordances-api/concordances"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	cli "github.com/jawher/mow.cli"
	"github.com/rcrowley/go-metrics"
)

const (
	serviceName               = "public-concordances-api"
	dbConnectionTimeout       = 1 * time.Minute
	maxIdleConnectionsPerHost = 100
)

func main() {
	app := cli.App(serviceName, "A public RESTful API for accessing concordances in neo4j")

	appSystemCode := app.String(cli.StringOpt{
		Name:   "app-system-code",
		Value:  "public-concordance-api",
		Desc:   "System Code of the application",
		EnvVar: "APP_SYSTEM_CODE",
	})
	neoURL := app.String(cli.StringOpt{
		Name:   "neo-url",
		Value:  "http://localhost:7474/db/data",
		Desc:   "neo4j endpoint URL",
		EnvVar: "NEO_URL",
	})
	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})
	env := app.String(cli.StringOpt{
		Name:  "env",
		Value: "local",
		Desc:  "environment this app is running in",
	})
	cacheDuration := app.String(cli.StringOpt{
		Name:   "cache-duration",
		Value:  "30s",
		Desc:   "Duration Get requests should be cached for. e.g. 2h45m would set the max-age value to '7440' seconds",
		EnvVar: "CACHE_DURATION",
	})
	logLevel := app.String(cli.StringOpt{
		Name:   "logLevel",
		Value:  "info",
		Desc:   "Log level of the app",
		EnvVar: "LOG_LEVEL",
	})
	batchSize := app.Int(cli.IntOpt{
		Name:   "batch-size",
		Value:  0,
		Desc:   "Max batch size for Neo4j queries",
		EnvVar: "BATCH_SIZE",
	})

	log := logger.NewUPPLogger(*appSystemCode, *logLevel)
	app.Action = func() {
		cacheControlHeader, err := parseCacheDurationArg(*cacheDuration)
		if err != nil {
			log.WithError(err).Fatalf("Application failed to start")
		}

		conf := neoutils.ConnectionConfig{
			BatchSize:     *batchSize,
			Transactional: false,
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					MaxIdleConnsPerHost: maxIdleConnectionsPerHost,
				},
				Timeout: dbConnectionTimeout,
			},
			BackgroundConnect: true,
		}
		db, err := neoutils.Connect(*neoURL, &conf, log)
		if err != nil {
			log.WithError(err).Fatalf("Application failed to connect to neo4j")
		}
		driver := concordances.NewCypherDriver(db, *env)
		hh := concordances.NewHTTPHandler(log, driver, cacheControlHeader)
		router := registerEndpoints(hh, log)
		srv := newHTTPServer(*port, router)
		go startHTTPServer(srv, log)
		log.Infof("service will listen on port: %s", *port)
		waitForSignal()
		stopHTTPServer(srv, log)
	}

	log.WithFields(map[string]interface{}{
		"CACHE_DURATION": *cacheDuration,
		"NEO_URL":        *neoURL,
		"LOG_LEVEL":      *logLevel,
	}).Info("Starting app with arguments")
	app.Run(os.Args)
}

func registerEndpoints(hh *concordances.HTTPHandler, log *logger.UPPLogger) http.Handler {
	servicesRouter := mux.NewRouter()
	mh := &handlers.MethodHandler{
		"GET": http.HandlerFunc(hh.GetConcordances),
	}
	servicesRouter.Handle("/concordances", mh)

	var monitoringRouter http.Handler = servicesRouter
	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(log, monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	// The top one of these feels more correct, but the lower one matches what we have in Dropwizard,
	// so it's what apps expect currently same as ping, the content of build-info needs more definition
	//using http router here to be able to catch "/"
	router := http.NewServeMux()
	router.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)
	router.HandleFunc(status.BuildInfoPathDW, status.BuildInfoHandler)

	router.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(hh.GTG))
	router.HandleFunc("/__health", fthealth.Handler(hh.HealthCheck(serviceName)))

	router.Handle("/", monitoringRouter)

	return router
}

func newHTTPServer(port string, router http.Handler) *http.Server {
	return &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
}

func startHTTPServer(srv *http.Server, log *logger.UPPLogger) {
	log.Info("starting http server...")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server failed to start: %s", err)
	}
}

func stopHTTPServer(srv *http.Server, log *logger.UPPLogger) {
	log.Info("http server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("failed to gracefully shutdown the server: %v", err)
	}
}

func waitForSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}

func parseCacheDurationArg(cacheDuration string) (string, error) {
	duration, err := time.ParseDuration(cacheDuration)
	if err != nil {
		return "", fmt.Errorf("failed to parse cache duration string, %v", err)
	}

	cacheDurationHeader := fmt.Sprintf("max-age=%s, public", strconv.FormatFloat(duration.Seconds(), 'f', 0, 64))
	return cacheDurationHeader, nil
}
