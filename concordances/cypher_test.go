package concordances

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"

	cmneo4j "github.com/Financial-Times/cm-neo4j-driver"
	"github.com/Financial-Times/concepts-rw-neo4j/concepts"
	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
)

var concordedBrandSmartlogic = Concordance{
	Concept{
		ID:     "http://api.ft.com/things/b20801ac-5a76-43cf-b816-8c3b2f7133ad",
		APIURL: "http://api.ft.com/brands/b20801ac-5a76-43cf-b816-8c3b2f7133ad"},
	Identifier{
		Authority:       "http://api.ft.com/system/SMARTLOGIC",
		IdentifierValue: "b20801ac-5a76-43cf-b816-8c3b2f7133ad"},
}

var concordedManagedLocationByConceptId = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44",
				APIURL: "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
			Identifier{
				Authority:       "http://api.ft.com/system/WIKIDATA",
				IdentifierValue: "http://www.wikidata.org/entity/Q218"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44",
				APIURL: "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
			Identifier{
				Authority:       "http://api.ft.com/system/FT-TME",
				IdentifierValue: "TnN0ZWluX0dMX1JP-R0w="},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44",
				APIURL: "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
			Identifier{
				Authority:       "http://api.ft.com/system/MANAGEDLOCATION",
				IdentifierValue: "5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44",
				APIURL: "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
			Identifier{
				Authority:       "http://api.ft.com/system/ISO-3166-1",
				IdentifierValue: "RO"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44",
				APIURL: "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "4534282c-d3ee-3595-9957-81a9293200f3"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44",
				APIURL: "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "4411b761-e632-30e7-855c-06aeca76c48d"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44",
				APIURL: "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
		},
	},
}

var concordedManagedLocationByAuthority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44",
				APIURL: "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
			Identifier{
				Authority:       "http://api.ft.com/system/MANAGEDLOCATION",
				IdentifierValue: "5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
		},
	},
}

var concordedManagedLocationByISO31661Authority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44",
				APIURL: "http://api.ft.com/things/5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
			Identifier{
				Authority:       "http://api.ft.com/system/ISO-3166-1",
				IdentifierValue: "RO"},
		},
	},
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

var expectedConcordanceNAICSIndustryClassification = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/38ee195d-ebdd-48a9-af4b-c8a322e7b04d",
				APIURL: "http://api.ft.com/things/38ee195d-ebdd-48a9-af4b-c8a322e7b04d"},
			Identifier{
				Authority:       "http://api.ft.com/system/SMARTLOGIC",
				IdentifierValue: "38ee195d-ebdd-48a9-af4b-c8a322e7b04d"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/38ee195d-ebdd-48a9-af4b-c8a322e7b04d",
				APIURL: "http://api.ft.com/things/38ee195d-ebdd-48a9-af4b-c8a322e7b04d"},
			Identifier{
				Authority:       "http://api.ft.com/system/NAICS",
				IdentifierValue: "5111"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/38ee195d-ebdd-48a9-af4b-c8a322e7b04d",
				APIURL: "http://api.ft.com/things/38ee195d-ebdd-48a9-af4b-c8a322e7b04d"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "38ee195d-ebdd-48a9-af4b-c8a322e7b04d"},
		},
	},
}

var expectedConcordanceNAICSIndustryClassificationByAuthority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/38ee195d-ebdd-48a9-af4b-c8a322e7b04d",
				APIURL: "http://api.ft.com/things/38ee195d-ebdd-48a9-af4b-c8a322e7b04d"},
			Identifier{
				Authority:       "http://api.ft.com/system/NAICS",
				IdentifierValue: "5111"},
		},
	},
}

func TestNeoReadByConceptID(t *testing.T) {
	driver := getNeoDriver(assert.New(t))
	log := logger.NewUPPLogger("public-concordances-api-test", "PANIC")

	conceptRW := concepts.NewConceptService(driver, log)
	assert.NoError(t, conceptRW.Initialise())

	tests := []struct {
		name        string
		fixture     string
		conceptIDs  []string
		expectedLen int
		expected    Concordances
	}{
		{
			name:        "NewModel_Unconcorded",
			fixture:     "Brand-Unconcorded-ad56856a-7d38-48e2-a131-7d104f17e8f6.json",
			conceptIDs:  []string{"ad56856a-7d38-48e2-a131-7d104f17e8f6"},
			expectedLen: 2,
			expected:    Concordances{[]Concordance{unconcordedBrandTME, unconcordedBrandTMEUPP}},
		},
		{
			name:        "NewModel_Concorded",
			fixture:     "Brand-Concorded-b20801ac-5a76-43cf-b816-8c3b2f7133ad.json",
			conceptIDs:  []string{"b20801ac-5a76-43cf-b816-8c3b2f7133ad"},
			expectedLen: 4,
			expected:    Concordances{[]Concordance{concordedBrandSmartlogic, concordedBrandSmartlogicUPP, concordedBrandTME, concordedBrandTMEUPP}},
		},
		{
			name:        "ManagedLocation",
			fixture:     "ManagedLocation-Concorded-5aba454b-3e31-31b9-bdeb-0caf83f62b44.json",
			conceptIDs:  []string{"5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
			expectedLen: 7,
			expected:    concordedManagedLocationByConceptId,
		},
		{
			name:        "ToConcordancesMandatoryFields",
			fixture:     "Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json",
			conceptIDs:  []string{"cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
			expectedLen: 7,
			expected:    expectedConcordanceBankOfTest,
		},
		{
			name:        "ReturnMultipleConcordancesForMultipleIdentifiers",
			fixture:     "Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json",
			conceptIDs:  []string{"cd7e4345-f11f-41f3-a0f0-2cf5c43e0115"},
			expectedLen: 7,
			expected:    expectedConcordanceBankOfTest,
		},
		{
			name:        "NAICSIndustryClassification",
			fixture:     "NAICSIndustryClassification-38ee195d-ebdd-48a9-af4b-c8a322e7b04d.json",
			conceptIDs:  []string{"38ee195d-ebdd-48a9-af4b-c8a322e7b04d"},
			expectedLen: 3,
			expected:    expectedConcordanceNAICSIndustryClassification,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writeGenericConceptJSONToService(conceptRW, "./fixtures/"+test.fixture, assert.New(t))
			defer cleanUp(assert.New(t), driver)

			undertest := NewCypherDriver(driver, "prod")
			conc, found, err := undertest.ReadByConceptID(test.conceptIDs)
			assert.NoError(t, err)
			assert.True(t, found)
			assert.Equal(t, test.expectedLen, len(conc.Concordance))

			readConceptAndCompare(t, test.expected, conc, "TestNeoReadByConceptID_"+test.name)
		})
	}
}

func TestNeoReadByAuthority(t *testing.T) {
	driver := getNeoDriver(assert.New(t))
	log := logger.NewUPPLogger("public-concordances-api-test", "PANIC")

	conceptRW := concepts.NewConceptService(driver, log)
	assert.NoError(t, conceptRW.Initialise())

	tests := []struct {
		name             string
		fixture          string
		authority        string
		identifierValues []string
		expected         Concordances
		expectedErr      bool
	}{
		{
			name:             "NewModel_Concorded",
			fixture:          "Brand-Concorded-b20801ac-5a76-43cf-b816-8c3b2f7133ad.json",
			authority:        "http://api.ft.com/system/SMARTLOGIC",
			identifierValues: []string{"b20801ac-5a76-43cf-b816-8c3b2f7133ad"},
			expected:         Concordances{[]Concordance{concordedBrandSmartlogic}},
		},
		{
			name:             "NewModel_Unconcorded",
			fixture:          "Brand-Unconcorded-ad56856a-7d38-48e2-a131-7d104f17e8f6.json",
			authority:        "http://api.ft.com/system/FT-TME",
			identifierValues: []string{"UGFydHkgcGVvcGxl-QnJhbmRz"},
			expected:         Concordances{[]Concordance{unconcordedBrandTME}},
		},
		{
			name:             "ManagedLocation",
			fixture:          "ManagedLocation-Concorded-5aba454b-3e31-31b9-bdeb-0caf83f62b44.json",
			authority:        "http://api.ft.com/system/MANAGEDLOCATION",
			identifierValues: []string{"5aba454b-3e31-31b9-bdeb-0caf83f62b44"},
			expected:         concordedManagedLocationByAuthority,
		},
		{
			name:             "ISO31661",
			fixture:          "ManagedLocation-Concorded-5aba454b-3e31-31b9-bdeb-0caf83f62b44.json",
			authority:        "http://api.ft.com/system/ISO-3166-1",
			identifierValues: []string{"RO"},
			expected:         concordedManagedLocationByISO31661Authority,
		},
		{
			name:             "ToConcordancesMandatoryField",
			fixture:          "Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json",
			authority:        "http://api.ft.com/system/FACTSET",
			identifierValues: []string{"7IV872-E"},
			expected:         expectedConcordanceBankOfTestByAuthority,
		},
		{
			name:             "ToConcordancesByUPPAuthority",
			fixture:          "Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json",
			authority:        "http://api.ft.com/system/UPP",
			identifierValues: []string{"d56e7388-25cb-343e-aea9-8b512e28476e"},
			expected:         expectedConcordanceBankOfTestByUPPAuthority,
		},
		{
			name:             "ToConcordancesByLEIAuthority",
			fixture:          "Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json",
			authority:        "http://api.ft.com/system/LEI",
			identifierValues: []string{"VNF516RB4DFV5NQ22UF0"},
			expected:         expectedConcordanceBankOfTestByLEIAuthority,
		},
		{
			name:             "OnlyOneConcordancePerIdentifierValue",
			fixture:          "Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json",
			authority:        "http://api.ft.com/system/FACTSET",
			identifierValues: []string{"7IV872-E"},
			expected:         expectedConcordanceBankOfTestByAuthority,
		},
		{
			name:             "NAICSIndustryClassification",
			fixture:          "NAICSIndustryClassification-38ee195d-ebdd-48a9-af4b-c8a322e7b04d.json",
			authority:        "http://api.ft.com/system/NAICS",
			identifierValues: []string{"5111"},
			expected:         expectedConcordanceNAICSIndustryClassificationByAuthority,
		},
		{
			name:             "EmptyConcordancesWhenUnsupportedAuthority",
			fixture:          "Organisation-BankOfTest-cd7e4345-f11f-41f3-a0f0-2cf5c43e0115.json",
			authority:        "http://api.ft.com/system/UnsupportedAuthority",
			identifierValues: []string{"DANMUR-1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writeGenericConceptJSONToService(conceptRW, "./fixtures/"+test.fixture, assert.New(t))
			defer cleanUp(assert.New(t), driver)

			undertest := NewCypherDriver(driver, "prod")
			conc, found, err := undertest.ReadByAuthority(test.authority, test.identifierValues)
			assert.NoError(t, err)

			if len(test.expected.Concordance) > 0 {
				assert.True(t, found)
				assert.Equal(t, 1, len(conc.Concordance))
				readConceptAndCompare(t, test.expected, conc, "TestNeoReadByAuthority_"+test.name)
				return
			}

			assert.False(t, found)
			assert.Empty(t, conc.Concordance)
		})
	}
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

func getNeoDriver(assert *assert.Assertions) *cmneo4j.Driver {
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "bolt://localhost:7687"
	}
	log := logger.NewUPPLogger("public-concordances-api-test", "PANIC")
	driver, err := cmneo4j.NewDefaultDriver(url, log)
	assert.NoError(err, "Failed to connect to Neo4j")
	return driver
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

func cleanUp(assert *assert.Assertions, driver *cmneo4j.Driver) {
	var queries []*cmneo4j.Query

	// Concepts with canonical nodes
	uuids := []string{
		"cd7e4345-f11f-41f3-a0f0-2cf5c43e0115",
		"5aba454b-3e31-31b9-bdeb-0caf83f62b44",
		"b20801ac-5a76-43cf-b816-8c3b2f7133ad",
		"ad56856a-7d38-48e2-a131-7d104f17e8f6",
		"38ee195d-ebdd-48a9-af4b-c8a322e7b04d",
	}
	for _, uuid := range uuids {
		query := &cmneo4j.Query{
			Cypher: `
				MATCH (canonical:Concept{prefUUID:$uuid})--(source)
				OPTIONAL MATCH (source)<-[:IDENTIFIES]-(identifier)
				DETACH DELETE canonical, source, identifier`,
			Params: map[string]interface{}{"uuid": uuid},
		}
		queries = append(queries, query)
	}

	// Things
	uuids = []string{
		"dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54",
	}
	for _, uuid := range uuids {
		query := &cmneo4j.Query{
			Cypher: `
				MATCH (source:Thing{uuid:$uuid})
				OPTIONAL MATCH (source)<-[:IDENTIFIES]-(identifier)
				DETACH DELETE source, identifier`,
			Params: map[string]interface{}{"uuid": uuid},
		}
		queries = append(queries, query)
	}

	err := driver.Write(queries...)
	assert.NoError(err)
}
