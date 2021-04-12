package concordances

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"errors"
	"strings"

	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/service-status-go/gtg"
	transactionidutils "github.com/Financial-Times/transactionid-utils-go"
)

type HTTPHandler struct {
	log *logger.UPPLogger
}

const (
	thingURIPrefix = "http://api.ft.com/things/"

	multipleAuthoritiesNotPermitted          = "multiple authorities are not permitted"
	conceptAndAuthorityCannotBeBothPresent   = "if conceptId is present then authority is not a valid parameter"
	authorityIsMandatoryIfConceptIDIsMissing = "if conceptId is absent then authority is mandatory"
	neitherConceptIDNorAuthorityPresent      = "neither conceptId nor authority were present"
	errAccessingConcordanceDatastore         = "error accessing Concordance datastore"
)

// ConcordanceDriver for cypher queries
var ConcordanceDriver Driver
var CacheControlHeader string
var connCheck error

func NewHTTPHandler(log *logger.UPPLogger) *HTTPHandler {
	return &HTTPHandler{log: log}
}

// HealthCheck provides an FT standard timed healthcheck for the /__health endpoint
func HealthCheck() fthealth.TimedHealthCheck {
	return fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{
			SystemCode:  "public-concordances-api",
			Name:        "public-concordances-api",
			Description: "Concords concept identifiers",
			Checks: []fthealth.Check{
				{
					BusinessImpact:   "Unable to respond to Public Concordances API requests",
					Name:             "Check connectivity to Neo4j",
					PanicGuide:       "https://runbooks.ftops.tech/public-concordances-api",
					Severity:         1,
					TechnicalSummary: "Cannot connect to Neo4j a instance with at least one concordance loaded in it",
					Checker:          Checker,
				},
			},
		},
		Timeout: 10 * time.Second,
	}
}

func StartAsyncChecker(checkInterval time.Duration) {
	go func(checkInterval time.Duration) {
		ticker := time.NewTicker(checkInterval)
		for range ticker.C {
			connCheck = ConcordanceDriver.CheckConnectivity()
		}
	}(checkInterval)
}

// Checker does more stuff
func Checker() (string, error) {
	if connCheck == nil {
		return "Connectivity to neo4j is ok", connCheck
	}
	return "Error connecting to neo4j", connCheck
}

// GTG lightly checks the application and conforms to the FT standard GTG format
func GTG() gtg.Status {
	if _, err := Checker(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}

// GetConcordances is the public API
func (hh *HTTPHandler) GetConcordances(w http.ResponseWriter, r *http.Request) {
	hh.log.Debugf("Concordance request: %s", r.URL.RawQuery)
	m, _ := url.ParseQuery(r.URL.RawQuery)

	_, conceptIDExist := m["conceptId"]
	_, authorityExist := m["authority"]
	tid := transactionidutils.GetTransactionIDFromRequest(r)
	logEntry := hh.log.WithTransactionID(tid)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if conceptIDExist && authorityExist {
		err := writeErrorResponse(w, http.StatusBadRequest, conceptAndAuthorityCannotBeBothPresent)
		if err != nil {
			logEntry.WithError(err).Errorf("cannot write response message: %s", conceptAndAuthorityCannotBeBothPresent)
		}
		return
	}

	if !conceptIDExist && !authorityExist {
		err := writeErrorResponse(w, http.StatusBadRequest, authorityIsMandatoryIfConceptIDIsMissing)
		if err != nil {
			logEntry.WithError(err).Errorf("cannot write response message: %s", authorityIsMandatoryIfConceptIDIsMissing)
		}
		return
	}

	if len(m["authority"]) > 1 {
		err := writeErrorResponse(w, http.StatusBadRequest, multipleAuthoritiesNotPermitted)
		if err != nil {
			logEntry.WithError(err).Errorf("cannot write response message: %s", multipleAuthoritiesNotPermitted)
		}
		return
	}

	concordance, _, err := processParams(conceptIDExist, authorityExist, m)
	if err != nil {
		logEntry.WithError(err).Errorf("error looking up Concordances")
		err := writeErrorResponse(w, http.StatusInternalServerError, errAccessingConcordanceDatastore)
		if err != nil {
			logEntry.WithError(err).Errorf("cannot write response message: %s", errAccessingConcordanceDatastore)
		}
		return
	}

	w.Header().Set("Cache-Control", CacheControlHeader)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(concordance)
}

func processParams(conceptIDExist bool, authorityExist bool, m url.Values) (concordances Concordances, found bool, err error) {
	if conceptIDExist {
		conceptUuids := []string{}

		for _, uri := range m["conceptId"] {
			conceptUuids = append(conceptUuids, strings.TrimPrefix(uri, thingURIPrefix))
		}

		return ConcordanceDriver.ReadByConceptID(conceptUuids)
	}

	if authorityExist {
		return ConcordanceDriver.ReadByAuthority(m.Get("authority"), m["identifierValue"])
	}

	return Concordances{}, false, errors.New(neitherConceptIDNorAuthorityPresent)
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, msg string) error {
	w.WriteHeader(statusCode)

	payload := []byte(`{"message": "` + msg + `"}`)
	if _, err := w.Write(payload); err != nil {
		return fmt.Errorf("error while writing response message: %w", err)
	}

	return nil
}
