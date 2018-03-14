package concordances

import (
	"encoding/json"
	"net/http"
	"net/url"

	"errors"
	"strings"

	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/service-status-go/gtg"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

// ConcordanceDriver for cypher queries
var ConcordanceDriver Driver
var CacheControlHeader string
var connCheck error

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
					PanicGuide:       "https://dewey.in.ft.com/view/system/public-concordances-api",
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
func GetConcordances(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Concordance request: %s", r.URL.RawQuery)
	m, _ := url.ParseQuery(r.URL.RawQuery)

	conceptID := m["conceptId"]
	authority := m.Get("authority")
	identifierValue := m["identifierValue"]

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if len(conceptID) != 0 && authority != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"message": "` + conceptAndAuthorityCannotBeBothPresent + `"}`))
		return
	}

	if len(conceptID) == 0 && authority == "" {
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

	concordance, _, err := processParams(conceptID, authority, identifierValue)
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

// PostConcordances is the public API
func PostConcordances(w http.ResponseWriter, r *http.Request){
	log.Debugf("Concordance request: %s", r.Body)

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"message": "` + invalidBody + `"}`))
		return
	}

	ai := authorityAndConcepts{}
	err = json.Unmarshal(body, &ai)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"message": "` + connotUnmarshallJSON + `"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if ai.Authority == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"message": "` + authorityIsEmpty + `"}`))
		return
	}

	if len(ai.IdentifierValue) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"message": "` + identifierValuesIsMissing + `"}`))
		return
	}

	concordance, _, err := processParams(ai.ConceptID, ai.Authority, ai.IdentifierValue)
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

func processParams(conceptID []string, authority string, identifierValue []string) (concordances Concordances, found bool, err error) {
	if len(conceptID)!=0 {
		conceptUuids := []string{}

		for _, uri := range conceptID {
			conceptUuids = append(conceptUuids, strings.TrimPrefix(uri, thingURIPrefix))
		}

		return ConcordanceDriver.ReadByConceptID(conceptUuids)
	}

	if authority != "" {
		return ConcordanceDriver.ReadByAuthority(authority, identifierValue)
	}

	return Concordances{}, false, errors.New(neitherConceptIdNorAuthorityPresent)
}


const (
	thingURIPrefix = "http://api.ft.com/things/"

	multipleAuthoritiesNotPermitted          = "Multiple authorities are not permitted"
	conceptAndAuthorityCannotBeBothPresent   = "If conceptId is present then authority is not a valid parameter"
	authorityIsMandatoryIfConceptIdIsMissing = "If conceptId is absent then authority is mandatory"
	neitherConceptIdNorAuthorityPresent      = "Neither conceptId nor authority were present"
	invalidBody                              = "The body cannot be processed"
	connotUnmarshallJSON                     = "JSON cannot be processed"
	authorityIsEmpty		                 = "Authority cannot be empty"
	identifierValuesIsMissing				 = "IdentifierValue is missing"

)

type authorityAndConcepts struct {
	ConceptID   []string `json:"conceptId,omitempty"`
	Authority   string `json:"authority,omitempty"`
	IdentifierValue []string `json:"identifierValue,omitempty"`
}
