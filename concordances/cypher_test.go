package concordances

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"reflect"

	"sort"

	"github.com/Financial-Times/concepts-rw-neo4j/concepts"
	"github.com/Financial-Times/go-logger"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.InitDefaultLogger("TestPublicConcordancesAPI")
}

var concordedBrandSmartlogic = Concordance{
	Concept{
		ID:     "http://api.ft.com/things/b20801ac-5a76-43cf-b816-8c3b2f7133ad",
		APIURL: "http://api.ft.com/brands/b20801ac-5a76-43cf-b816-8c3b2f7133ad"},
	Identifier{
		Authority:       "http://api.ft.com/system/SMARTLOGIC",
		IdentifierValue: "b20801ac-5a76-43cf-b816-8c3b2f7133ad"},
}

var concordedBrandSmartlogicUPP = Concordance{
	Concept{
		ID:     "http://api.ft.com/things/b20801ac-5a76-43cf-b816-8c3b2f7133ad",
		APIURL: "http://api.ft.com/brands/b20801ac-5a76-43cf-b816-8c3b2f7133ad"},
	Identifier{
		Authority:       "http://api.ft.com/system/UPP",
		IdentifierValue: "b20801ac-5a76-43cf-b816-8c3b2f7133ad"},
}

var concordedBrandTME = Concordance{
	Concept{
		ID:     "http://api.ft.com/things/b20801ac-5a76-43cf-b816-8c3b2f7133ad",
		APIURL: "http://api.ft.com/brands/b20801ac-5a76-43cf-b816-8c3b2f7133ad"},
	Identifier{
		Authority:       "http://api.ft.com/system/FT-TME",
		IdentifierValue: "VGhlIFJvbWFu-QnJhbmRz"},
}

var concordedBrandTMEUPP = Concordance{
	Concept{
		ID:     "http://api.ft.com/things/b20801ac-5a76-43cf-b816-8c3b2f7133ad",
		APIURL: "http://api.ft.com/brands/b20801ac-5a76-43cf-b816-8c3b2f7133ad"},
	Identifier{
		Authority:       "http://api.ft.com/system/UPP",
		IdentifierValue: "70f4732b-7f7d-30a1-9c29-0cceec23760e"},
}

var expectedConcordanceBankOfTest = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115",
				APIURL: "http://api.ft.com/organisations/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "2cdeb859-70df-3a0e-b125-f958366bea44"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115",
				APIURL: "http://api.ft.com/organisations/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
			Identifier{
				Authority:       "http://api.ft.com/system/FACTSET",
				IdentifierValue: "7IV872-E"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115",
				APIURL: "http://api.ft.com/organisations/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
			Identifier{
				Authority:       "http://api.ft.com/system/FT-TME",
				IdentifierValue: "QmFuayBvZiBUZXN0-T04="},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115",
				APIURL: "http://api.ft.com/organisations/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
			Identifier{
				Authority:       "http://api.ft.com/system/LEI",
				IdentifierValue: "VNF516RB4DFV5NQ22UF0"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115",
				APIURL: "http://api.ft.com/organisations/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
			Identifier{
				Authority:       "http://api.ft.com/system/SMARTLOGIC",
				IdentifierValue: "cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115",
				APIURL: "http://api.ft.com/organisations/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115",
				APIURL: "http://api.ft.com/organisations/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "d56e7388-25cb-343e-aea9-8b512e28476e"},
		},
	},
}

var expectedConcordanceBankOfTestByAuthority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115",
				APIURL: "http://api.ft.com/organisations/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
			Identifier{
				Authority:       "http://api.ft.com/system/FACTSET",
				IdentifierValue: "7IV872-E"},
		},
	},
}

var expectedConcordanceBankOfTestByUPPAuthority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115",
				APIURL: "http://api.ft.com/organisations/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "d56e7388-25cb-343e-aea9-8b512e28476e"},
		},
	},
}

var expectedConcordanceBankOfTestByLEIAuthority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115",
				APIURL: "http://api.ft.com/organisations/cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
			Identifier{
				Authority:       "http://api.ft.com/system/LEI",
				IdentifierValue: "VNF516RB4DFV5NQ22UF0"},
		},
	},
}

var unconcordedBrandTME = Concordance{
	Concept{
		ID:     "http://api.ft.com/things/ad56856a-7d38-48e2-a131-7d104f17e8f6",
		APIURL: "http://api.ft.com/brands/ad56856a-7d38-48e2-a131-7d104f17e8f6"},
	Identifier{
		Authority:       "http://api.ft.com/system/FT-TME",
		IdentifierValue: "UGFydHkgcGVvcGxl-QnJhbmRz"},
}

var unconcordedBrandTMEUPP = Concordance{
	Concept{
		ID:     "http://api.ft.com/things/ad56856a-7d38-48e2-a131-7d104f17e8f6",
		APIURL: "http://api.ft.com/brands/ad56856a-7d38-48e2-a131-7d104f17e8f6"},
	Identifier{
		Authority:       "http://api.ft.com/system/UPP",
		IdentifierValue: "ad56856a-7d38-48e2-a131-7d104f17e8f6"},
}

func TestNeoReadByConceptID_NewModel_Unconcorded(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)

	conceptRW := concepts.NewConceptService(db)
	assert.NoError(conceptRW.Initialise())

	writeGenericConceptJSONToService(conceptRW, "./fixtures/Brand-Unconcorded-ad56856a-7d38-48e2-a131-7d104f17e8f6.json", assert)
	defer cleanUp(assert, db)

	undertest := NewCypherDriver(db, "prod")
	conc, found, err := undertest.ReadByConceptID([]string{"ad56856a-7d38-48e2-a131-7d104f17e8f6"})
	assert.NoError(err)
	assert.True(found)
	assert.Equal(2, len(conc.Concordance))

	readConceptAndCompare(t, Concordances{[]Concordance{unconcordedBrandTME, unconcordedBrandTMEUPP}}, conc, "TestNeoReadByConceptID_NewModel_Unconcorded")
}

func TestNeoReadByConceptID_NewModel_Concorded(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)
	conceptRW := concepts.NewConceptService(db)
	assert.NoError(conceptRW.Initialise())

	writeGenericConceptJSONToService(conceptRW, "./fixtures/Brand-Concorded-b20801ac-5a76-43cf-b816-8c3b2f7133ad.json", assert)
	defer cleanUp(assert, db)

	undertest := NewCypherDriver(db, "prod")
	conc, found, err := undertest.ReadByConceptID([]string{"b20801ac-5a76-43cf-b816-8c3b2f7133ad"})
	assert.NoError(err)
	assert.True(found)
	assert.Equal(4, len(conc.Concordance))

	readConceptAndCompare(t, Concordances{[]Concordance{concordedBrandSmartlogic, concordedBrandSmartlogicUPP, concordedBrandTME, concordedBrandTMEUPP}}, conc, "TestNeoReadByConceptID_NewModel_Concorded")
}

func TestNeoReadByAuthority_NewModel_Unconcorded(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)
	conceptRW := concepts.NewConceptService(db)
	assert.NoError(conceptRW.Initialise())

	writeGenericConceptJSONToService(conceptRW, "./fixtures/Brand-Unconcorded-ad56856a-7d38-48e2-a131-7d104f17e8f6.json", assert)
	defer cleanUp(assert, db)

	undertest := NewCypherDriver(db, "prod")
	conc, found, err := undertest.ReadByAuthority("http://api.ft.com/system/FT-TME", []string{"UGFydHkgcGVvcGxl-QnJhbmRz"})
	assert.NoError(err)
	assert.True(found)
	assert.Equal(1, len(conc.Concordance))

	sliceConcordances := []Concordance{unconcordedBrandTME}
	readConceptAndCompare(t, Concordances{sliceConcordances}, conc, "TestNeoReadByAuthority_NewModel_Unconcorded")
}

func TestNeoReadByAuthority_NewModel_Concorded(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)
	conceptRW := concepts.NewConceptService(db)
	assert.NoError(conceptRW.Initialise())

	writeGenericConceptJSONToService(conceptRW, "./fixtures/Brand-Concorded-b20801ac-5a76-43cf-b816-8c3b2f7133ad.json", assert)
	defer cleanUp(assert, db)

	undertest := NewCypherDriver(db, "prod")
	conc, found, err := undertest.ReadByAuthority("http://api.ft.com/system/SMARTLOGIC", []string{"b20801ac-5a76-43cf-b816-8c3b2f7133ad"})
	assert.NoError(err)
	assert.True(found)
	assert.Equal(1, len(conc.Concordance))

	readConceptAndCompare(t, Concordances{[]Concordance{concordedBrandSmartlogic}}, conc, "TestNeoReadByAuthority_NewModel_Concorded")
}

func TestNeoReadByConceptIDToConcordancesMandatoryFields(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)
	organisationRW := concepts.NewConceptService(db)
	assert.NoError(organisationRW.Initialise())

	writeGenericConceptJSONToService(organisationRW, "./fixtures/Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json", assert)

	defer cleanUp(assert, db)

	undertest := NewCypherDriver(db, "prod")
	cs, found, err := undertest.ReadByConceptID([]string{"cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"})
	assert.NoError(err)
	assert.True(found)
	assert.NotEmpty(cs.Concordance)

	readConceptAndCompare(t, expectedConcordanceBankOfTest, cs, "TestNeoReadByConceptIDToConcordancesMandatoryFields")
}

func TestNeoReadByAuthorityToConcordancesMandatoryFields(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)

	organisationRW := concepts.NewConceptService(db)
	assert.NoError(organisationRW.Initialise())

	writeGenericConceptJSONToService(organisationRW, "./fixtures/Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json", assert)

	defer cleanUp(assert, db)

	undertest := NewCypherDriver(db, "prod")
	cs, found, err := undertest.ReadByAuthority("http://api.ft.com/system/FACTSET", []string{"7IV872-E"})
	assert.NoError(err)
	assert.True(found)
	assert.NotEmpty(cs.Concordance)

	readConceptAndCompare(t, expectedConcordanceBankOfTestByAuthority, cs, "TestNeoReadByAuthorityToConcordancesMandatoryFields")
}

func TestNeoReadByAuthorityToConcordancesByUPPAuthority(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)

	organisationRW := concepts.NewConceptService(db)
	assert.NoError(organisationRW.Initialise())

	writeGenericConceptJSONToService(organisationRW, "./fixtures/Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json", assert)

	defer cleanUp(assert, db)

	undertest := NewCypherDriver(db, "prod")
	cs, found, err := undertest.ReadByAuthority("http://api.ft.com/system/UPP", []string{"d56e7388-25cb-343e-aea9-8b512e28476e"})
	assert.NoError(err)
	assert.True(found)
	assert.NotEmpty(cs.Concordance)

	readConceptAndCompare(t, expectedConcordanceBankOfTestByUPPAuthority, cs, "TestNeoReadByAuthorityToConcordancesByUPPAuthority")
}

func TestNeoReadByAuthorityToConcordancesByLEIAuthority(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)

	organisationRW := concepts.NewConceptService(db)
	assert.NoError(organisationRW.Initialise())

	writeGenericConceptJSONToService(organisationRW, "./fixtures/Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json", assert)

	defer cleanUp(assert, db)

	undertest := NewCypherDriver(db, "prod")
	cs, found, err := undertest.ReadByAuthority("http://api.ft.com/system/LEI", []string{"VNF516RB4DFV5NQ22UF0"})
	assert.NoError(err)
	assert.True(found)
	assert.NotEmpty(cs.Concordance)

	readConceptAndCompare(t, expectedConcordanceBankOfTestByLEIAuthority, cs, "TestNeoReadByAuthorityToConcordancesByLEIAuthority")
}

func TestNeoReadByAuthorityOnlyOneConcordancePerIdentifierValue(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)

	organisationRW := concepts.NewConceptService(db)
	assert.NoError(organisationRW.Initialise())

	writeGenericConceptJSONToService(organisationRW, "./fixtures/Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json", assert)

	defer cleanUp(assert, db)

	undertest := NewCypherDriver(db, "prod")
	cs, found, err := undertest.ReadByAuthority("http://api.ft.com/system/FACTSET", []string{"7IV872-E"})
	assert.NoError(err)
	assert.True(found)
	assert.NotEmpty(cs.Concordance)
	assert.Equal(len(cs.Concordance), 1)

	readConceptAndCompare(t, expectedConcordanceBankOfTestByAuthority, cs, "TestNeoReadByAuthorityOnlyOneConcordancePerIdentifierValue")
}

func TestNeoReadByConceptIdReturnMultipleConcordancesForMultipleIdentifiers(t *testing.T) {

	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)

	organisationRW := concepts.NewConceptService(db)
	assert.NoError(organisationRW.Initialise())

	writeGenericConceptJSONToService(organisationRW, "./fixtures/Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json", assert)

	defer cleanUp(assert, db)

	undertest := NewCypherDriver(db, "prod")
	cs, found, err := undertest.ReadByConceptID([]string{"cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"})
	assert.NoError(err)
	assert.True(found)
	assert.NotEmpty(cs.Concordance)
	assert.Equal(7, len(cs.Concordance))

	readConceptAndCompare(t, expectedConcordanceBankOfTest, cs, "TestNeoReadByConceptIdReturnMultipleConcordancesForMultipleIdentifiers")
}

func TestNeoReadByAuthorityEmptyConcordancesWhenUnsupportedAuthority(t *testing.T) {
	assert := assert.New(t)
	db := getDatabaseConnection(t, assert)

	organisationRW := concepts.NewConceptService(db)
	assert.NoError(organisationRW.Initialise())

	writeGenericConceptJSONToService(organisationRW, "./fixtures/Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json", assert)

	defer cleanUp(assert, db)

	undertest := NewCypherDriver(db, "prod")
	cs, found, err := undertest.ReadByAuthority("http://api.ft.com/system/UnsupportedAuthority", []string{"DANMUR-1"})
	assert.NoError(err)
	assert.False(found)
	assert.Empty(cs.Concordance)
}

func readConceptAndCompare(t *testing.T, expected Concordances, actual Concordances, testName string) {

	sortConcordances(expected.Concordance)
	sortConcordances(actual.Concordance)

	assert.True(t, reflect.DeepEqual(expected, actual), fmt.Sprintf("Actual aggregated concept differs from expected: Test: %v \n Expected: %v \n Actual: %v", testName, expected, actual))
}

func sortConcordances(concordanceList []Concordance) {
	sort.SliceStable(concordanceList, func(i, j int) bool {
		return concordanceList[i].Concept.ID < concordanceList[j].Concept.ID
	})
	sort.SliceStable(concordanceList, func(i, j int) bool {
		return concordanceList[i].Identifier.Authority < concordanceList[j].Identifier.Authority
	})
	sort.SliceStable(concordanceList, func(i, j int) bool {
		return concordanceList[i].Identifier.IdentifierValue < concordanceList[j].Identifier.IdentifierValue
	})
}

func getDatabaseConnection(t *testing.T, assert *assert.Assertions) neoutils.NeoConnection {
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}

	conf := neoutils.DefaultConnectionConfig()
	conf.Transactional = false
	db, err := neoutils.Connect(url, conf)
	assert.NoError(err, "Failed to connect to Neo4j")
	return db
}

func writeGenericConceptJSONToService(service concepts.ConceptService, pathToJSONFile string, assert *assert.Assertions) {
	f, err := os.Open(pathToJSONFile)
	assert.NoError(err)
	dec := json.NewDecoder(f)
	inst, _, errr := service.DecodeJSON(dec)
	assert.NoError(errr)
	_, errrr := service.Write(inst, "test_transaction_id")
	assert.NoError(errrr)
}

func cleanUp(assert *assert.Assertions, db neoutils.NeoConnection) {

	qs := []*neoism.CypherQuery{
		{
			Statement: fmt.Sprintf("MATCH (a:Thing {uuid: '%s'})--(i:Identifier) DETACH DELETE i, a", "3e844449-b27f-40d4-b696-2ce9b6137133"),
		}, {
			Statement: fmt.Sprintf("MATCH (a:Thing {uuid: '%s'})--(i:Identifier) DETACH DELETE i, a", "f21a5cc0-d326-4e62-b84a-d840c2209fee"),
		}, {
			Statement: fmt.Sprintf("MATCH (a:Thing {uuid: '%s'})--(i:Identifier) DETACH DELETE i, a", "f9694ba7-eab0-4ce0-8e01-ff64bccb813c"),
		}, {
			Statement: fmt.Sprintf("MATCH (t:Thing {uuid: '%v'})--(i:Identifier) OPTIONAL MATCH (t)-[:EQUIVALENT_TO]-(e:Thing) DETACH DELETE t, i", "70f4732b-7f7d-30a1-9c29-0cceec23760e"),
		}, {
			Statement: fmt.Sprintf("MATCH (t:Thing {uuid: '%v'})--(i:Identifier) OPTIONAL MATCH (t)-[:EQUIVALENT_TO]-(e:Thing) DETACH DELETE t, e, i", "b20801ac-5a76-43cf-b816-8c3b2f7133ad"),
		}, {
			Statement: fmt.Sprintf("MATCH (t:Thing {uuid: '%v'})--(i:Identifier) OPTIONAL MATCH (t)-[:EQUIVALENT_TO]-(e:Thing) DETACH DELETE t, i", "ad56856a-7d38-48e2-a131-7d104f17e8f6"),
		}, {
			Statement: fmt.Sprintf("MATCH (t:Thing {uuid: '%v'})--(i:Identifier) OPTIONAL MATCH (t)-[:EQUIVALENT_TO]-(e:Thing) DETACH DELETE t, i", "dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"),
		}, {
			Statement: fmt.Sprintf("MATCH (t:Thing {prefUUID: '%v'}) DETACH DELETE t", "ad56856a-7d38-48e2-a131-7d104f17e8f6"),
		},
	}

	err := db.CypherBatch(qs)
	assert.NoError(err)

}

func cleanUpParentOrgAndUppIdentifier(db neoutils.NeoConnection, t *testing.T, assert *assert.Assertions) {
	qs := []*neoism.CypherQuery{
		{
			//deletes parent 'org' which only has type Thing
			Statement: fmt.Sprintf("MATCH (j:Thing {uuid: '%v'}) DETACH DELETE j", "3e844449-b27f-40d4-b696-2ce9b6137133"),
		},
		{
			//deletes upp identifier for the above parent 'org'
			Statement: fmt.Sprintf("MATCH (k:Identifier {value: '%v'}) DETACH DELETE k", "3e844449-b27f-40d4-b696-2ce9b6137133"),
		},
	}

	err := db.CypherBatch(qs)
	assert.NoError(err)
}
