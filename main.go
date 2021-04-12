package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	_ "github.com/joho/godotenv/autoload"
	"github.com/rcrowley/go-metrics"
)

func main() {
	app := cli.App("public-concordances-api", "A public RESTful API for accessing concordances in neo4j")

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
	healthcheckInterval := app.String(cli.StringOpt{
		Name:   "healthcheck-interval",
		Value:  "30s",
		Desc:   "How often the Neo4j healthcheck is called.",
		EnvVar: "HEALTHCHECK_INTERVAL",
	})
	batchSize := app.Int(cli.IntOpt{
		Name:   "batch-size",
		Value:  0,
		Desc:   "Max batch size for Neo4j queries",
		EnvVar: "BATCH_SIZE",
	})

	log := logger.NewUPPLogger(*appSystemCode, *logLevel)
	app.Action = func() {
		log.Infof("service will listen on port: %s, connecting to: %s", *port, *neoURL)
		runServer(*neoURL, *port, *cacheDuration, *env, *healthcheckInterval, *batchSize, log)
	}

	log.WithFields(map[string]interface{}{
		"HEALTHCHECK_INTERVAL": *healthcheckInterval,
		"CACHE_DURATION":       *cacheDuration,
		"NEO_URL":              *neoURL,
		"LOG_LEVEL":            *logLevel,
	}).Info("Starting app with arguments")
	app.Run(os.Args)
}

func runServer(neoURL string, port string, cacheDuration string, env string, healthcheckInterval string, batchSize int, log *logger.UPPLogger) {

	if duration, durationErr := time.ParseDuration(cacheDuration); durationErr != nil {
		log.Fatalf("Failed to parse cache duration string, %v", durationErr)
	} else {
		concordances.CacheControlHeader = fmt.Sprintf("max-age=%s, public", strconv.FormatFloat(duration.Seconds(), 'f', 0, 64))
	}

	conf := neoutils.ConnectionConfig{
		BatchSize:     batchSize,
		Transactional: false,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 100,
			},
			Timeout: 1 * time.Minute,
		},
		BackgroundConnect: true,
	}
	db, err := neoutils.Connect(neoURL, &conf, log)
	if err != nil {
		log.Fatalf("Error connecting to neo4j %s", err)
	}

	concordances.ConcordanceDriver = concordances.NewCypherDriver(db, env)

	checkInterval, err := time.ParseDuration(healthcheckInterval)
	if err != nil {
		checkInterval = time.Second * 30
	}
	concordances.StartAsyncChecker(checkInterval)

	servicesRouter := mux.NewRouter()

	// Then API specific ones:

	mh := &handlers.MethodHandler{
		"GET": http.HandlerFunc(concordances.GetConcordances),
	}
	servicesRouter.Handle("/concordances", mh)

	var monitoringRouter http.Handler = servicesRouter
	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(log, monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	// The top one of these feels more correct, but the lower one matches what we have in Dropwizard,
	// so it's what apps expect currently same as ping, the content of build-info needs more definition
	//using http router here to be able to catch "/"
	http.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)
	http.HandleFunc(status.BuildInfoPathDW, status.BuildInfoHandler)

	http.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(concordances.GTG))
	http.HandleFunc("/__health", fthealth.Handler(concordances.HealthCheck()))

	http.Handle("/", monitoringRouter)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}

}
