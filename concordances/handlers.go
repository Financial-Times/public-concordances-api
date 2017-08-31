package concordances

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"errors"
	"strings"

	"time"

	"github.com/Financial-Times/go-fthealth/v1a"
	log "github.com/sirupsen/logrus"
)

// ConcordanceDriver for cypher queries
var ConcordanceDriver Driver
var CacheControlHeader string
var connCheck error

// HealthCheck does something
func HealthCheck() v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Unable to respond to Public Concordances api requests",
		Name:             "Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service",
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/public-concordances-api",
		Severity:         1,
		TechnicalSummary: "Cannot connect to Neo4j a instance with at least one concordance loaded in it",
		Checker:          Checker,
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

//GoodToGo returns a 503 if the healthcheck fails - suitable for use from varnish to check availability of a node
func GoodToGo(writer http.ResponseWriter, req *http.Request) {
	if _, err := Checker(); err != nil {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}
}

// BuildInfoHandler - This is a stop gap and will be added to when we can define what we should display here
func BuildInfoHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "build-info")
}

// GetConcordances is the public API
func GetConcordances(w http.ResponseWriter, r *http.Request) {

	log.Debugf("Concordance request: %s", r.URL.RawQuery)
	m, _ := url.ParseQuery(r.URL.RawQuery)

	_, conceptIDExist := m["conceptId"]
	_, authorityExist := m["authority"]

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if conceptIDExist && authorityExist {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"message": "` + conceptAndAuthorityCannotBeBothPresent + `"}`))
		return
	}

	if !conceptIDExist && !authorityExist {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"message": "` + authorityIsMandatoryIfConceptIdIsMissing + `"}`))
		return
	}

	if len(m["authority"]) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"message": "` + multipleAuthoritiesNotPermitted + `"}`))
		return
	}

	concordance, _, err := processParams(conceptIDExist, authorityExist, m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}

	Jason, _ := json.Marshal(concordance)
	log.Debugf("Concordance(uuid:%s): %s\n", Jason)
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

	return Concordances{}, false, errors.New(neitherConceptIdNorAuthorityPresent)
}

const (
	thingURIPrefix = "http://api.ft.com/things/"

	multipleAuthoritiesNotPermitted          = "Multiple authorities are not permitted"
	conceptAndAuthorityCannotBeBothPresent   = "If conceptId is present then authority is not a valid parameter"
	authorityIsMandatoryIfConceptIdIsMissing = "If conceptId is absent then authority is mandatory"
	neitherConceptIdNorAuthorityPresent      = "Neither conceptId nor authority were present"
)
