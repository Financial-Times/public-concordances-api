package concordances

import (
	"fmt"

	log "github.com/Financial-Times/go-logger"
	"github.com/Financial-Times/neo-model-utils-go/mapper"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
)

// Driver interface
type Driver interface {
	ReadByConceptID(ids []string) (concordances Concordances, found bool, err error)
	ReadByAuthority(authority string, ids []string) (concordances Concordances, found bool, err error)
	CheckConnectivity() error
}

// CypherDriver struct
type CypherDriver struct {
	conn neoutils.NeoConnection
	env  string
}

//NewCypherDriver instantiate driver
func NewCypherDriver(conn neoutils.NeoConnection, env string) CypherDriver {
	return CypherDriver{conn, env}
}

// CheckConnectivity tests neo4j by running a simple cypher query
func (pcw CypherDriver) CheckConnectivity() error {
	return neoutils.Check(pcw.conn)
}

func (pcw CypherDriver) ReadByConceptID(identifiers []string) (concordances Concordances, found bool, err error) {
	var results []neoReadStruct
	query := &neoism.CypherQuery{
		Statement: `
		MATCH (p:Thing)
		WHERE p.uuid in {identifiers}
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		MATCH (canonical)<-[:EQUIVALENT_TO]-(leafNode:Thing)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, leafNode.authority as authority, leafNode.authorityValue as authorityValue
		UNION ALL

		MATCH (p:Thing)
		WHERE p.uuid in {identifiers}
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		WHERE exists(canonical.leiCode)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, 'LEI' as authority, canonical.leiCode as authorityValue
		UNION ALL
		MATCH (p:Thing)
		WHERE p.uuid in {identifiers}
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		WHERE exists(canonical.iso31661)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, 'ISO-3166-1' as authority, canonical.iso31661 as authorityValue
		UNION ALL
		MATCH (p:Thing)
		WHERE p.uuid in {identifiers}
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		MATCH (canonical)<-[:EQUIVALENT_TO]-(leafNode:Thing)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, 'UPP' as authority, leafNode.uuid as authorityValue
        `,
		Parameters: neoism.Props{"identifiers": identifiers},
		Result:     &results,
	}

	err = pcw.conn.CypherBatch([]*neoism.CypherQuery{query})
	if err != nil {
		log.Errorf("Error looking up Concordances with query %s from neoism: %+v\n", query.Statement, err)
		return Concordances{}, false, fmt.Errorf("error accessing Concordance datastore for identifier: %v", identifiers)
	}

	if (len(results)) == 0 {
		return Concordances{}, false, nil
	}

	concordances = Concordances{
		Concordance: []Concordance{},
	}

	return processCypherQueryToConcordances(pcw, query, results)

}

func (pcw CypherDriver) ReadByAuthority(authority string, identifierValues []string) (concordances Concordances, found bool, err error) {
	var results []neoReadStruct

	authorityProperty, found := AuthorityFromURI(authority)
	if !found {
		return Concordances{}, false, nil
	}

	var query *neoism.CypherQuery

	if authorityProperty == "UPP" {
		// We need to treat the UPP authority slightly different as it's stored elsewhere.
		query = &neoism.CypherQuery{
			Statement: `
		MATCH (p:Thing)
		WHERE p.uuid IN {authorityValue}
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, p.uuid as UUID, 'UPP' as authority, p.uuid as authorityValue`,

			Parameters: neoism.Props{
				"authorityValue": identifierValues,
			},
			Result: &results,
		}
	} else if authorityProperty == "LEI" {
		// We've gotta treat LEI special like as well.
		query = &neoism.CypherQuery{
			Statement: `
		MATCH (p:Concept)
		WHERE p.leiCode IN {authorityValue}
		AND exists(p.prefUUID)
		RETURN DISTINCT p.prefUUID AS canonicalUUID, labels(p) AS types, p.uuid as UUID, 'LEI' as authority, p.leiCode as authorityValue`,

			Parameters: neoism.Props{
				"authorityValue": identifierValues,
			},
			Result: &results,
		}
	} else if authorityProperty == "ISO-3166-1" {
		query = &neoism.CypherQuery{
			Statement: `
		MATCH (p:Concept)
		WHERE exists(p.iso31661)
		AND p.iso31661 IN {authorityValue}
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, p.uuid as UUID, 'ISO-3166-1' as authority, p.iso31661 as authorityValue
			`,
			Parameters: neoism.Props{
				"authorityValue": identifierValues,
			},
			Result: &results,
		}
	} else {
		query = &neoism.CypherQuery{
			Statement: `
		MATCH (p:Thing)
		WHERE p.authority = {authority} AND p.authorityValue IN {authorityValue}
		MATCH (p)-[:EQUIVALENT_TO]->(canonical:Concept)
		RETURN DISTINCT canonical.prefUUID AS canonicalUUID, labels(canonical) AS types, p.uuid as UUID, p.authority as authority, p.authorityValue as authorityValue`,

			Parameters: neoism.Props{
				"authorityValue": identifierValues,
				"authority":      authorityProperty,
			},
			Result: &results,
		}
	}

	err = pcw.conn.CypherBatch([]*neoism.CypherQuery{query})
	if err != nil {
		log.Errorf("Error looking up Concordances with query %s from neoism: %+v\n", query.Statement, err)
		return Concordances{}, false, fmt.Errorf("error accessing Concordance datastore for identifier: %v", identifierValues)
	}

	if (len(results)) == 0 {
		return Concordances{}, false, nil
	}

	concordances = Concordances{
		Concordance: []Concordance{},
	}

	return processCypherQueryToConcordances(pcw, query, results)
}

func processCypherQueryToConcordances(pcw CypherDriver, q *neoism.CypherQuery, results []neoReadStruct) (concordances Concordances, found bool, err error) {
	err = pcw.conn.CypherBatch([]*neoism.CypherQuery{q})
	if err != nil {
		log.Errorf("Error looking up Concordances with query %s from neoism: %+v\n", q.Statement, err)
		return Concordances{}, false, fmt.Errorf("Error accessing Concordance datastore")
	}

	concordances = neoReadStructToConcordances(results, pcw.env)

	if (len(concordances.Concordance)) == 0 {
		return Concordances{}, false, nil
	}
	return concordances, true, nil
}

func neoReadStructToConcordances(neo []neoReadStruct, env string) (concordances Concordances) {
	concordances = Concordances{
		Concordance: []Concordance{},
	}
	for _, neoCon := range neo {
		var con = Concordance{}
		var concept = Concept{}

		concept.ID = mapper.IDURL(neoCon.CanonicalUUID)
		concept.APIURL = mapper.APIURL(neoCon.CanonicalUUID, neoCon.Types, env)
		authorityURI, found := AuthorityToURI(neoCon.Authority)
		if !found {
			log.Debugf("Unsupported authority: %s", neoCon.Authority)
			continue
		}
		con.Identifier = Identifier{Authority: authorityURI, IdentifierValue: neoCon.AuthorityValue}

		con.Concept = concept
		concordances.Concordance = append(concordances.Concordance, con)
	}
	return concordances
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
