package concordances

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"errors"
	"github.com/Financial-Times/go-fthealth/v1a"
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
		PanicGuide:       "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/public-concordances-api",
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

	if len(m["identifierValue"]) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"message": "` + multipleAuthorityValuesNotSupported + `"}`))
		return
	}

	if len(m["conceptId"]) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"message": "` + multipleConceptIDsNotSupported + `"}`))
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
		w.Write([]byte(`{"message":"` + concordanceNotFound + `"}`))
		return
	}

	//check the first element of concordances - to see if it's canonical and a redirect is required
	if conceptIDExist && len(concordance.Concordance) > 0 {
		canonicalConceptID := concordance.Concordance[0].Concept.ID
		conceptID := m.Get("conceptId")
		if !strings.EqualFold(canonicalConceptID, conceptID) {
			redirectURL := strings.Replace(r.RequestURI, conceptID, canonicalConceptID, 1)
			w.Header().Set("Location", redirectURL)
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}
	}

	Jason, _ := json.Marshal(concordance)
	log.Debugf("Concordance(uuid:%s): %s\n", Jason)
	w.Header().Set("Cache-Control", CacheControlHeader)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(concordance)
}

func processParams(conceptIDExist bool, authorityExist bool, m url.Values) (concordances Concordances, found bool, err error) {
	if conceptIDExist {
		conceptUUID := strings.TrimPrefix(m.Get("conceptId"), thingURIPrefix)
		return ConcordanceDriver.ReadByConceptID(conceptUUID)
	}

	if authorityExist {
		identifierValue := m.Get("identifierValue")
		return ConcordanceDriver.ReadByAuthority(m.Get("authority"), identifierValue)
	}

	return Concordances{}, false, errors.New(neitherConceptIdNorAuthorityPresent)
}

const (
	thingURIPrefix = "http://api.ft.com/things/"

	multipleAuthoritiesNotPermitted          = "Multiple authorities are not permitted"
	multipleAuthorityValuesNotSupported      = "Multiple authority values (identifiers) are not supported"
	multipleConceptIDsNotSupported           = "Multiple conceptIDs are not supported"
	conceptAndAuthorityCannotBeBothPresent   = "If conceptId is present then authority is not a valid parameter"
	authorityIsMandatoryIfConceptIdIsMissing = "If conceptId is absent then authority is mandatory"
	neitherConceptIdNorAuthorityPresent      = "Neither conceptId nor authority were present"
	concordanceNotFound                      = "Concordance not found."
)
