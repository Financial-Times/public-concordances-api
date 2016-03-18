package concordances

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"errors"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/service-status-go/gtg"
	log "github.com/Sirupsen/logrus"
	"strings"
)

// ConcordanceDriver for cypher queries
var ConcordanceDriver Driver
var CacheControlHeader string

// HealthCheck does something
func HealthCheck() v1a.Check {
	return v1a.Check{
		BusinessImpact:   "Unable to respond to Public Concordances api requests",
		Name:             "Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service",
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/concordance-read-api",
		Severity:         1,
		TechnicalSummary: "Cannot connect to Neo4j a instance with at least one concordance loaded in it",
		Checker:          Checker,
	}
}

// Checker does more stuff
func Checker() (string, error) {
	err := ConcordanceDriver.CheckConnectivity()
	if err == nil {
		return "Connectivity to neo4j is ok", err
	}
	return "Error connecting to neo4j", err
}

// G2GChecker simply calls the Neo4J checker and returns it as a gtg.Status
func G2GChecker() gtg.Status {
	msg, err := Checker()
	if err == nil {
		return gtg.Status{GoodToGo: true}
	}
	return gtg.Status{GoodToGo: false, Message: msg}
}

// Ping says pong
func Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
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

// MethodNotAllowedHandler handles 405
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	return
}

// GetConcordances is the public API
func GetConcordances(w http.ResponseWriter, r *http.Request) {

	m, _ := url.ParseQuery(r.URL.RawQuery)

	_, conceptIDExist := m["conceptId"]
	_, authorityExist := m["authority"]

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if conceptIDExist && authorityExist {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"message": "If conceptId is present then authority is not a valid parameter"}`))
		return
	}

	if !conceptIDExist && !authorityExist {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"message": "If conceptId is absent then authority is mandatory"}`))
		return
	}

	if len(m["authority"]) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"message": "Multiple authorities are not permitted"}`))
		return
	}

	concordance, found, err := processParams(conceptIDExist, authorityExist, m)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Concordance not found."}`))
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

	return Concordances{}, false, errors.New("Niether conceptId nor authority were present")
}

const (
	thingURIPrefix = "http://api.ft.com/things/"
)
