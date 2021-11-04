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

	cmneo4j "github.com/Financial-Times/cm-neo4j-driver"
	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/http-handlers-go/v2/httphandlers"

	"github.com/Financial-Times/api-endpoint"
	"github.com/Financial-Times/public-concordances-api/concordances"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	cli "github.com/jawher/mow.cli"
	"github.com/rcrowley/go-metrics"
)

const (
	serviceName = "public-concordances-api"
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
		Value:  "bolt://localhost:7687",
		Desc:   "neoURL must point to a leader node or use neo4j:// scheme, otherwise writes will fail",
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
	apiYml := app.String(cli.StringOpt{
		Name:   "api-yml",
		Value:  "./api.yml",
		Desc:   "Location of the API Swagger YML file.",
		EnvVar: "API_YML",
	})

	log := logger.NewUPPLogger(*appSystemCode, *logLevel)
	log.WithFields(map[string]interface{}{
		"CACHE_DURATION": *cacheDuration,
		"NEO_URL":        *neoURL,
		"LOG_LEVEL":      *logLevel,
	}).Info("Starting app with arguments")

	app.Action = func() {
		cacheControlHeader, err := parseCacheDurationArg(*cacheDuration)
		if err != nil {
			log.WithError(err).Fatalf("Application failed to start")
		}

		driver, err := cmneo4j.NewDefaultDriver(*neoURL, log)
		if err != nil {
			log.WithError(err).Fatal("Unable to create a new cmneo4j driver")
		}
		defer driver.Close()

		concordancesDriver := concordances.NewCypherDriver(driver, *env)
		hh := concordances.NewHTTPHandler(log, concordancesDriver, cacheControlHeader)
		router := registerEndpoints(hh, log, apiYml)
		srv := newHTTPServer(*port, router)
		go startHTTPServer(srv, log)
		log.Infof("service will listen on port: %s", *port)
		waitForSignal()
		stopHTTPServer(srv, log)
	}

	app.Run(os.Args)
}

func registerEndpoints(hh *concordances.HTTPHandler, log *logger.UPPLogger, apiYml *string) http.Handler {
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
	if apiYml != nil {
		endpoint, err := api.NewAPIEndpointForFile(*apiYml)
		if err != nil {
			log.WithError(err).WithField("file", apiYml).Warn("Failed to serve the API Endpoint for this service. Please validate the Swagger YML and the file location.")
		} else {
			servicesRouter.HandleFunc(api.DefaultPath, endpoint.ServeHTTP).Methods("GET")
		}
	}

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
