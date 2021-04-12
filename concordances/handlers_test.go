package concordances

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	logger "github.com/Financial-Times/go-logger/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var (
	server          *httptest.Server
	concordanceURL  string
	isFound         bool
	conceptIds      []string
	authorityValues []string
	actualAuthority string
)

type mockConcordanceDriver struct{}

func (driver mockConcordanceDriver) ReadByConceptID(ids []string) (concordances Concordances, found bool, err error) {
	conceptIds = ids
	return Concordances{}, isFound, nil
}
func (driver mockConcordanceDriver) ReadByAuthority(authority string, ids []string) (concordances Concordances, found bool, err error) {
	authorityValues = ids
	actualAuthority = authority
	return Concordances{}, isFound, nil
}

func (driver mockConcordanceDriver) CheckConnectivity() error {
	return nil
}

func init() {
	log := logger.NewUPPLogger("public-concordances-api", "info")
	hh := NewHTTPHandler(log)
	ConcordanceDriver = mockConcordanceDriver{}
	r := mux.NewRouter()
	r.HandleFunc("/concordances", hh.GetConcordances).Methods("GET")
	server = httptest.NewServer(r)
	concordanceURL = fmt.Sprintf("%s/concordances", server.URL) //Grab the address for the API endpoint
	isFound = true
}

func TestCanGetOneAuthority(t *testing.T) {
	assert := assert.New(t)
	isFound = true
	req, _ := http.NewRequest("GET", concordanceURL+"?authority=some-authority&identifierValue=some-value", nil)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(err)
	assert.EqualValues(200, res.StatusCode)
	assert.EqualValues(actualAuthority, "some-authority")
	assert.Len(authorityValues, 1)
	assert.Contains(authorityValues, "some-value")
}

func TestCanGetMultipleIdentifiersByAuthority(t *testing.T) {
	assert := assert.New(t)
	isFound = true
	req, _ := http.NewRequest("GET", concordanceURL+"?authority=some-authority&identifierValue=some-value&identifierValue=some-value2", nil)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(err)
	assert.EqualValues(200, res.StatusCode)
	assert.Len(authorityValues, 2)
	assert.Contains(authorityValues, "some-value")
	assert.Contains(authorityValues, "some-value2")
}

func TestReturnBadRequestGivenMoreThanOneAuthority(t *testing.T) {
	assert := assert.New(t)
	isFound = true
	req, _ := http.NewRequest("GET", concordanceURL+"?authority=some-authority&identifierValue=some-value&authority=some-authority-yet-again", nil)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(err)
	assert.EqualValues(400, res.StatusCode)
	msg, err := ioutil.ReadAll(res.Body)
	assert.NoError(err)
	assert.Contains(string(msg), multipleAuthoritiesNotPermitted)
}

func TestCanGetOneConcept(t *testing.T) {
	assert := assert.New(t)
	isFound = true
	req, _ := http.NewRequest("GET", concordanceURL+"?conceptId=bob", nil)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(err)
	assert.EqualValues(200, res.StatusCode)
	assert.Len(conceptIds, 1)
	assert.Contains(conceptIds, "bob")
}

func TestCanGetMultipleConcepts(t *testing.T) {
	assert := assert.New(t)
	isFound = true
	req, _ := http.NewRequest("GET", concordanceURL+"?conceptId=bob&conceptId=carlos", nil)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(err)
	assert.EqualValues(200, res.StatusCode)
	assert.Len(conceptIds, 2)
	assert.Contains(conceptIds, "bob")
	assert.Contains(conceptIds, "carlos")
}

func TestCanParseConceptURI(t *testing.T) {
	assert := assert.New(t)
	isFound = true
	req, _ := http.NewRequest("GET", concordanceURL+"?conceptId=http://api.ft.com/things/8138ca3f-b80d-3ef8-ad59-6a9b6ea5f15e", nil)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(err)
	assert.EqualValues(200, res.StatusCode)
	assert.Len(conceptIds, 1)
	assert.Contains(conceptIds, "8138ca3f-b80d-3ef8-ad59-6a9b6ea5f15e")
}

func TestCanNotRequestAuthorityAndConceptId(t *testing.T) {
	assert := assert.New(t)
	isFound = true
	req, _ := http.NewRequest("GET", concordanceURL+"?conceptId=bob&authority=high-and-mighty", nil)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(err)
	assert.EqualValues(400, res.StatusCode)
}

func TestCanNotRequestWithoutAuthorityOrConceptId(t *testing.T) {
	assert := assert.New(t)
	isFound = true
	req, _ := http.NewRequest("GET", concordanceURL+"?randomRequestParam=bob", nil)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(err)
	assert.EqualValues(400, res.StatusCode)
}

func TestSetCorrectCacheControlHeaders(t *testing.T) {
	assert := assert.New(t)
	duration, err := time.ParseDuration("30s")
	assert.NoError(err)
	CacheControlHeader = fmt.Sprintf("max-age=%s, public", strconv.FormatFloat(duration.Seconds(), 'f', 0, 64))
	isFound = true
	req, _ := http.NewRequest("GET", concordanceURL+"?authority=some-authority&identifierValue=some-value", nil)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(err)
	defer res.Body.Close()
	assert.EqualValues(200, res.StatusCode)
	assert.EqualValues(res.Header.Get("Cache-Control"), "max-age=30, public")
}
