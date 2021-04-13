package concordances

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"errors"
	"strings"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/service-status-go/gtg"
	transactionidutils "github.com/Financial-Times/transactionid-utils-go"
)

type HTTPHandler struct {
	log                *logger.UPPLogger
	concordanceDriver  Driver
	cacheControlHeader string
}

const (
	healthCheckTimeout = 10 * time.Second

	thingURIPrefix = "http://api.ft.com/things/"

	multipleAuthoritiesNotPermitted          = "multiple authorities are not permitted"
	conceptAndAuthorityCannotBeBothPresent   = "if conceptId is present then authority is not a valid parameter"
	authorityIsMandatoryIfConceptIDIsMissing = "if conceptId is absent then authority is mandatory"
	neitherConceptIDNorAuthorityPresent      = "neither conceptId nor authority were present"
	errAccessingConcordanceDatastore         = "error accessing Concordance datastore"
)

func NewHTTPHandler(log *logger.UPPLogger, driver Driver, cacheControlHeader string) *HTTPHandler {
	return &HTTPHandler{
		log:                log,
		concordanceDriver:  driver,
		cacheControlHeader: cacheControlHeader,
	}
}

// HealthCheck provides an FT standard timed healthcheck for the /__health endpoint
func (hh *HTTPHandler) HealthCheck(serviceName string) fthealth.TimedHealthCheck {
	return fthealth.TimedHealthCheck{
		HealthCheck: fthealth.HealthCheck{
			SystemCode:  serviceName,
			Name:        serviceName,
			Description: "Concords concept identifiers",
			Checks: []fthealth.Check{
				{
					BusinessImpact:   "Unable to respond to Public Concordances API requests",
					Name:             "Check connectivity to Neo4j",
					PanicGuide:       "https://runbooks.ftops.tech/public-concordances-api",
					Severity:         1,
					TechnicalSummary: "Cannot connect to Neo4j a instance with at least one concordance loaded in it",
					Checker:          hh.databaseConnectivityChecker,
				},
			},
		},
		Timeout: healthCheckTimeout,
	}
}

func (hh *HTTPHandler) databaseConnectivityChecker() (string, error) {
	connCheck := hh.concordanceDriver.CheckConnectivity()
	if connCheck == nil {
		return "Connectivity to neo4j is ok", connCheck
	}
	return "Error connecting to neo4j", connCheck
}

// GTG lightly checks the application and conforms to the FT standard GTG format
func (hh *HTTPHandler) GTG() gtg.Status {
	if _, err := hh.databaseConnectivityChecker(); err != nil {
		return gtg.Status{GoodToGo: false, Message: err.Error()}
	}
	return gtg.Status{GoodToGo: true}
}

// GetConcordances is the public API
func (hh *HTTPHandler) GetConcordances(w http.ResponseWriter, r *http.Request) {
	tid := transactionidutils.GetTransactionIDFromRequest(r)
	logEntry := hh.log.WithTransactionID(tid)
	logEntry.Debugf("Concordance request: %s", r.URL.RawQuery)
	m, _ := url.ParseQuery(r.URL.RawQuery)

	_, conceptIDExist := m["conceptId"]
	_, authorityExist := m["authority"]

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

	concordance, _, err := hh.processParams(conceptIDExist, authorityExist, m)
	if err != nil {
		logEntry.WithError(err).Errorf("error looking up Concordances")
		err := writeErrorResponse(w, http.StatusInternalServerError, errAccessingConcordanceDatastore)
		if err != nil {
			logEntry.WithError(err).Errorf("cannot write response message: %s", errAccessingConcordanceDatastore)
		}
		return
	}

	w.Header().Set("Cache-Control", hh.cacheControlHeader)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(concordance)
}

func (hh *HTTPHandler) processParams(conceptIDExist bool, authorityExist bool, m url.Values) (concordances Concordances, found bool, err error) {
	if conceptIDExist {
		conceptUuids := []string{}

		for _, uri := range m["conceptId"] {
			conceptUuids = append(conceptUuids, strings.TrimPrefix(uri, thingURIPrefix))
		}

		return hh.concordanceDriver.ReadByConceptID(conceptUuids)
	}

	if authorityExist {
		return hh.concordanceDriver.ReadByAuthority(m.Get("authority"), m["identifierValue"])
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
