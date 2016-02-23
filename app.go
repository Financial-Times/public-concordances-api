package main

import (
	"net/http"
	"os"

	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/public-concordances-api/concordances"

	"fmt"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/jmcvetta/neoism"
	"github.com/rcrowley/go-metrics"
)

func main() {
	log.Infof("Application starting with args %s", os.Args)
	app := cli.App("public-concordances-api-neo4j", "A public RESTful API for accessing concordances in neo4j")
	neoURL := app.StringOpt("neo-url", "http://localhost:7474/db/data", "neo4j endpoint URL")
	//neoURL := app.StringOpt("neo-url", "http://ftper60304-law1a-eu-t:8080/db/data", "neo4j endpoint URL")
	port := app.StringOpt("port", "8080", "Port to listen on")
	env := app.StringOpt("env", "local", "environment this app is running in")
	graphiteTCPAddress := app.StringOpt("graphiteTCPAddress", "",
		"Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally)")
	graphitePrefix := app.StringOpt("graphitePrefix", "",
		"Prefix to use. Should start with content, include the environment, and the host name. e.g. content.test.public.concordances.api.ftaps59382-law1a-eu-t")
	logMetrics := app.BoolOpt("logMetrics", false, "Whether to log metrics. Set to true if running locally and you want metrics output")
	cacheDuration := app.StringOpt("cache-duration", "1h", "Duration Get requests should be cached for. e.g. 2h45m would set the max-age value to '7440' seconds")

	app.Action = func() {
		baseftrwapp.OutputMetricsIfRequired(*graphiteTCPAddress, *graphitePrefix, *logMetrics)
		if *env != "local" {
			f, err := os.OpenFile("/var/log/apps/public-concordances-api-go-app.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
			if err == nil {
				log.SetOutput(f)

			} else {
				log.Fatalf("Failed to initialise log file, %v", err)
			}
			defer f.Close()
		}

		log.Infof("public-concordances-api will listen on port: %s, connecting to: %s", *port, *neoURL)

		runServer(*neoURL, *port, *cacheDuration, *env)

	}
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.InfoLevel)
	log.Infof("Application started with args %s", os.Args)
	app.Run(os.Args)
}

func runServer(neoURL string, port string, cacheDuration string, env string) {

	if duration, durationErr := time.ParseDuration(cacheDuration); durationErr != nil {
		log.Fatalf("Failed to parse cache duration string, %v", durationErr)
	} else {
		concordances.CacheControlHeader = fmt.Sprintf("max-age=%s, public", strconv.FormatFloat(duration.Seconds(), 'f', 0, 64))
	}

	db, err := neoism.Connect(neoURL)
	db.Session.Client = &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: 100}}
	if err != nil {
		log.Fatalf("Error connecting to neo4j %s", err)
	}

	concordances.ConcordanceDriver = concordances.NewCypherDriver(db, env)

	servicesRouter := mux.NewRouter()

	// Healthchecks and standards first
	servicesRouter.HandleFunc("/__health", v1a.Handler("PublicConcordancesRead Healthchecks",
		"Checks for accessing neo4j", concordances.HealthCheck()))
	servicesRouter.HandleFunc("/ping", concordances.Ping)
	servicesRouter.HandleFunc("/__ping", concordances.Ping)
	servicesRouter.HandleFunc("/__gtg", concordances.GoodToGo)

	// Then API specific ones:
	servicesRouter.HandleFunc("/concordances", concordances.GetConcordances).Methods("GET")

	servicesRouter.HandleFunc("/concordances", concordances.MethodNotAllowedHandler)

	var monitoringRouter http.Handler = servicesRouter
	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	// The top one of these feels more correct, but the lower one matches what we have in Dropwizard,
	// so it's what apps expect currently same as ping, the content of build-info needs more definition
	//using http router here to be able to catch "/"
	http.HandleFunc("/__build-info", concordances.BuildInfoHandler)
	http.HandleFunc("/build-info", concordances.BuildInfoHandler)
	http.Handle("/", monitoringRouter)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}

}
