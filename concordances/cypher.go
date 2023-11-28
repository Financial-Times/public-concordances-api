package concordances

import (
	"errors"
	"fmt"
	"net/url"

	ontology "github.com/Financial-Times/cm-graph-ontology/v2"
	cmneo4j "github.com/Financial-Times/cm-neo4j-driver"
)

const thingURL = "http://api.ft.com/things/"

// Driver interface
type Driver interface {
	ReadByConceptID(ids []string) (concordances Concordances, found bool, err error)
	ReadByAuthority(authority string, ids []string) (concordances Concordances, found bool, err error)
	CheckConnectivity() error
}

// CypherDriver struct
type CypherDriver struct {
	driver       *cmneo4j.Driver
	publicAPIURL string
}

// NewCypherDriver instantiate driver
func NewCypherDriver(driver *cmneo4j.Driver, publicAPIURL string) (CypherDriver, error) {
	_, err := url.ParseRequestURI(publicAPIURL)
	if err != nil {
		return CypherDriver{}, err
	}

	return CypherDriver{driver, publicAPIURL}, nil
}

// CheckConnectivity tests neo4j by running a simple cypher query
func (cd CypherDriver) CheckConnectivity() error {
	return cd.driver.VerifyConnectivity()
}

func (cd CypherDriver) ReadByConceptID(identifiers []string) (concordances Concordances, found bool, err error) {
	var results []neoReadStruct
	query := &cmneo4j.Query{
		Cypher: `
		MATCH (p:Thing)
		WHERE p.uuid in $identifiers
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		MATCH (canonical)<-[:EQUIVALENT_TO]-(leafNode:Thing)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, leafNode.authority as authority, leafNode.authorityValue as authorityValue
		UNION ALL

		MATCH (p:Thing)
		WHERE p.uuid in $identifiers
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		WHERE exists(canonical.leiCode)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, 'LEI' as authority, canonical.leiCode as authorityValue
		UNION ALL

		MATCH (p:Location)
		WHERE p.uuid in $identifiers
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		WHERE exists(canonical.iso31661)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, 'ISO-3166-1' as authority, canonical.iso31661 as authorityValue
		UNION ALL

		MATCH (p:NAICSIndustryClassification)
		WHERE p.uuid in $identifiers
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		WHERE exists(canonical.industryIdentifier)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, 'NAICS' as authority, canonical.industryIdentifier as authorityValue
		UNION ALL

		MATCH (p:FTAnIIndustryClassification)
		WHERE p.uuid in $identifiers
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		WHERE exists(canonical.industryIdentifier)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, 'FTAnI' as authority, canonical.industryIdentifier as authorityValue
		UNION ALL

		MATCH (p:Thing)
		WHERE p.uuid in $identifiers
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		MATCH (canonical)<-[:EQUIVALENT_TO]-(leafNode:Thing)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, 'UPP' as authority, leafNode.uuid as authorityValue
        `,
		Params: map[string]interface{}{"identifiers": identifiers},
		Result: &results,
	}

	err = cd.driver.Read(query)

	if errors.Is(err, cmneo4j.ErrNoResultsFound) {
		return Concordances{}, false, nil
	}
	if err != nil {
		return Concordances{}, false, fmt.Errorf("error accessing Concordance datastore for identifier %v: %w", identifiers, err)
	}

	concordances = Concordances{
		Concordance: []Concordance{},
	}

	return processCypherQueryToConcordances(cd, query, results)

}

func (cd CypherDriver) ReadByAuthority(authority string, identifierValues []string) (concordances Concordances, found bool, err error) {
	var results []neoReadStruct

	authorityProperty, found := AuthorityFromURI(authority)
	if !found {
		return Concordances{}, false, nil
	}

	var query *cmneo4j.Query

	if authorityProperty == "UPP" {
		// We need to treat the UPP authority slightly different as it's stored elsewhere.
		query = &cmneo4j.Query{
			Cypher: `
		MATCH (p:Thing)
		WHERE p.uuid IN $authorityValue
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, p.uuid as UUID, 'UPP' as authority, p.uuid as authorityValue`,

			Params: map[string]interface{}{
				"authorityValue": identifierValues,
			},
			Result: &results,
		}
	} else if authorityProperty == "LEI" {
		// We've gotta treat LEI special like as well.
		query = &cmneo4j.Query{
			Cypher: `
		MATCH (p:Concept)
		WHERE p.leiCode IN $authorityValue
		AND exists(p.prefUUID)
		RETURN DISTINCT p.prefUUID AS canonicalUUID, labels(p) AS types, p.uuid as UUID, 'LEI' as authority, p.leiCode as authorityValue`,

			Params: map[string]interface{}{
				"authorityValue": identifierValues,
			},
			Result: &results,
		}
	} else if authorityProperty == "ISO-3166-1" {
		query = &cmneo4j.Query{
			Cypher: `
		MATCH (canonical:Location)
		WHERE canonical.iso31661 IN $authorityValue
		AND exists(canonical.prefUUID)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, canonical.uuid as UUID, 'ISO-3166-1' as authority, canonical.iso31661 as authorityValue
			`,
			Params: map[string]interface{}{
				"authorityValue": identifierValues,
			},
			Result: &results,
		}
	} else if authorityProperty == "NAICS" {
		query = &cmneo4j.Query{
			Cypher: `
		MATCH (canonical:NAICSIndustryClassification)
		WHERE canonical.industryIdentifier IN $authorityValue
		AND exists(canonical.prefUUID)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, canonical.uuid as UUID, 'NAICS' as authority, canonical.industryIdentifier as authorityValue
			`,
			Params: map[string]interface{}{
				"authorityValue": identifierValues,
			},
			Result: &results,
		}
	} else if authorityProperty == "FTAnI" {
		query = &cmneo4j.Query{
			Cypher: `
		MATCH (canonical:FTAnIIndustryClassification)
		WHERE canonical.industryIdentifier IN $authorityValue
		AND exists(canonical.prefUUID)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, canonical.uuid as UUID, 'FTAnI' as authority, canonical.industryIdentifier as authorityValue
			`,
			Params: map[string]interface{}{
				"authorityValue": identifierValues,
			},
			Result: &results,
		}
	} else {
		query = &cmneo4j.Query{
			Cypher: `
		MATCH (p:Thing)
		WHERE p.authority = $authority AND p.authorityValue IN $authorityValue
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, p.uuid as UUID, p.authority as authority, p.authorityValue as authorityValue`,

			Params: map[string]interface{}{
				"authorityValue": identifierValues,
				"authority":      authorityProperty,
			},
			Result: &results,
		}
	}

	err = cd.driver.Read(query)
	if errors.Is(err, cmneo4j.ErrNoResultsFound) {
		return Concordances{}, false, nil
	}
	if err != nil {
		return Concordances{}, false, fmt.Errorf("error accessing Concordance datastore for authorityValue %v: %w", identifierValues, err)
	}

	concordances = Concordances{
		Concordance: []Concordance{},
	}

	return processCypherQueryToConcordances(cd, query, results)
}

func processCypherQueryToConcordances(cd CypherDriver, q *cmneo4j.Query, results []neoReadStruct) (concordances Concordances, found bool, err error) {
	err = cd.driver.Read(q)
	if errors.Is(err, cmneo4j.ErrNoResultsFound) {
		return Concordances{}, false, nil
	}

	if err != nil {
		return Concordances{}, false, fmt.Errorf("error accessing Concordance datastore: %w", err)
	}

	concordances, err = neoReadStructToConcordances(results, cd.publicAPIURL)
	if err != nil {
		return Concordances{}, false, fmt.Errorf("transforming result from datastore: %w", err)
	}

	return concordances, true, nil
}

func neoReadStructToConcordances(neo []neoReadStruct, baseURL string) (Concordances, error) {
	concordances := Concordances{
		Concordance: []Concordance{},
	}
	for _, neoCon := range neo {
		var con = Concordance{}
		var concept = Concept{}

		apiURL, err := ontology.APIURL(neoCon.CanonicalUUID, neoCon.Types, baseURL)
		if err != nil {
			return Concordances{}, fmt.Errorf("building APIURL for %q: %w", neoCon.CanonicalUUID, err)
		}

		concept.ID = thingIDURL(neoCon.CanonicalUUID)
		concept.APIURL = apiURL
		authorityURI, found := AuthorityToURI(neoCon.Authority)
		if !found {
			continue
		}
		con.Identifier = Identifier{Authority: authorityURI, IdentifierValue: neoCon.AuthorityValue}

		con.Concept = concept
		concordances.Concordance = append(concordances.Concordance, con)
	}
	return concordances, nil
}

// Map of authority to URI for the supported concordance IDs
var authorityMap = map[string]string{
	"TME":             "http://api.ft.com/system/FT-TME",
	"FACTSET":         "http://api.ft.com/system/FACTSET",
	"UPP":             "http://api.ft.com/system/UPP",
	"LEI":             "http://api.ft.com/system/LEI",
	"Smartlogic":      "http://api.ft.com/system/SMARTLOGIC",
	"ManagedLocation": "http://api.ft.com/system/MANAGEDLOCATION",
	"ISO-3166-1":      "http://api.ft.com/system/ISO-3166-1",
	"Geonames":        "http://api.ft.com/system/GEONAMES",
	"Wikidata":        "http://api.ft.com/system/WIKIDATA",
	"DBPedia":         "http://api.ft.com/system/DBPEDIA",
	"NAICS":           "http://api.ft.com/system/NAICS",
	"FTAnI":           "http://api.ft.com/system/FT-AnI",
}

func AuthorityFromURI(uri string) (string, bool) {
	for a, u := range authorityMap {
		if u == uri {
			return a, true
		}
	}
	return "", false
}

func AuthorityToURI(authority string) (string, bool) {
	authorityURI, found := authorityMap[authority]
	return authorityURI, found
}

func thingIDURL(uuid string) string {
	return thingURL + uuid
}
