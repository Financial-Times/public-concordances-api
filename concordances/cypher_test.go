//go:build integration
// +build integration

package concordances

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"

	ontology "github.com/Financial-Times/cm-graph-ontology/v2"
	"github.com/Financial-Times/cm-graph-ontology/v2/neo4j"
	cmneo4j "github.com/Financial-Times/cm-neo4j-driver"
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

var expectedConcordanceSVProvision = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/1808c3fc-04bb-589b-a457-640bffa8f6c6",
				APIURL: "http://api.ft.com/concepts/1808c3fc-04bb-589b-a457-640bffa8f6c6"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "1808c3fc-04bb-589b-a457-640bffa8f6c6"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/1808c3fc-04bb-589b-a457-640bffa8f6c6",
				APIURL: "http://api.ft.com/concepts/1808c3fc-04bb-589b-a457-640bffa8f6c6"},
			Identifier{
				Authority:       "http://api.ft.com/system/8e6c705e-1132-42a2-8db0-c295e29e8658",
				IdentifierValue: "65d735ebad5f88460e919a42"},
		},
	},
}

var expectedConcordanceSVProvisionByAuthority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/1808c3fc-04bb-589b-a457-640bffa8f6c6",
				APIURL: "http://api.ft.com/concepts/1808c3fc-04bb-589b-a457-640bffa8f6c6"},
			Identifier{
				Authority:       "http://api.ft.com/system/8e6c705e-1132-42a2-8db0-c295e29e8658",
				IdentifierValue: "65d735ebad5f88460e919a42"},
		},
	},
}

var expectedConcordanceFTPCGenre = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/e02fb4c0-1fe5-476b-b791-e921db5b99f2",
				APIURL: "http://api.ft.com/things/e02fb4c0-1fe5-476b-b791-e921db5b99f2"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "e02fb4c0-1fe5-476b-b791-e921db5b99f2"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/e02fb4c0-1fe5-476b-b791-e921db5b99f2",
				APIURL: "http://api.ft.com/things/e02fb4c0-1fe5-476b-b791-e921db5b99f2"},
			Identifier{
				Authority:       "http://api.ft.com/system/724b5e36-6d45-4cf1-b1c2-3f676b21f21b",
				IdentifierValue: "e02fb4c0-1fe5-476b-b791-e921db5b99f2"},
		},
	},
}

var expectedConcordanceFTPCGenreByAuthority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/e02fb4c0-1fe5-476b-b791-e921db5b99f2",
				APIURL: "http://api.ft.com/things/e02fb4c0-1fe5-476b-b791-e921db5b99f2"},
			Identifier{
				Authority:       "http://api.ft.com/system/724b5e36-6d45-4cf1-b1c2-3f676b21f21b",
				IdentifierValue: "e02fb4c0-1fe5-476b-b791-e921db5b99f2"},
		},
	},
}

var expectedConcordanceFTPCSource = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/8a852776-38b4-47fc-bb5e-e496801a28bf",
				APIURL: "http://api.ft.com/things/8a852776-38b4-47fc-bb5e-e496801a28bf"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "8a852776-38b4-47fc-bb5e-e496801a28bf"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/8a852776-38b4-47fc-bb5e-e496801a28bf",
				APIURL: "http://api.ft.com/things/8a852776-38b4-47fc-bb5e-e496801a28bf"},
			Identifier{
				Authority:       "http://api.ft.com/system/724b5e36-6d45-4cf1-b1c2-3f676b21f21b",
				IdentifierValue: "8a852776-38b4-47fc-bb5e-e496801a28bf"},
		},
	},
}

var expectedConcordanceFTPCSourceByAuthority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/8a852776-38b4-47fc-bb5e-e496801a28bf",
				APIURL: "http://api.ft.com/things/8a852776-38b4-47fc-bb5e-e496801a28bf"},
			Identifier{
				Authority:       "http://api.ft.com/system/724b5e36-6d45-4cf1-b1c2-3f676b21f21b",
				IdentifierValue: "8a852776-38b4-47fc-bb5e-e496801a28bf"},
		},
	},
}

var expectedConcordanceFTPCAssetType = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/c5440e5e-a472-4948-ab33-97e0089dd926",
				APIURL: "http://api.ft.com/things/c5440e5e-a472-4948-ab33-97e0089dd926"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "c5440e5e-a472-4948-ab33-97e0089dd926"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/c5440e5e-a472-4948-ab33-97e0089dd926",
				APIURL: "http://api.ft.com/things/c5440e5e-a472-4948-ab33-97e0089dd926"},
			Identifier{
				Authority:       "http://api.ft.com/system/724b5e36-6d45-4cf1-b1c2-3f676b21f21b",
				IdentifierValue: "c5440e5e-a472-4948-ab33-97e0089dd926"},
		},
	},
}

var expectedConcordanceFTPCAssetTypeByAuthority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/c5440e5e-a472-4948-ab33-97e0089dd926",
				APIURL: "http://api.ft.com/things/c5440e5e-a472-4948-ab33-97e0089dd926"},
			Identifier{
				Authority:       "http://api.ft.com/system/724b5e36-6d45-4cf1-b1c2-3f676b21f21b",
				IdentifierValue: "c5440e5e-a472-4948-ab33-97e0089dd926"},
		},
	},
}

var expectedConcordanceFTAOrganisationDetails = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/77701984-3542-4f77-91aa-b5f7bfa43330",
				APIURL: "http://api.ft.com/concepts/77701984-3542-4f77-91aa-b5f7bfa43330"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "77701984-3542-4f77-91aa-b5f7bfa43330"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/77701984-3542-4f77-91aa-b5f7bfa43330",
				APIURL: "http://api.ft.com/concepts/77701984-3542-4f77-91aa-b5f7bfa43330"},
			Identifier{
				Authority:       "http://api.ft.com/system/19d50190-8656-4e91-8d34-82e646ada9c9",
				IdentifierValue: "77701984-3542-4f77-91aa-b5f7bfa43330"},
		},
	},
}

var expectedConcordanceFTAOrganisationDetailsByAuthority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/77701984-3542-4f77-91aa-b5f7bfa43330",
				APIURL: "http://api.ft.com/concepts/77701984-3542-4f77-91aa-b5f7bfa43330"},
			Identifier{
				Authority:       "http://api.ft.com/system/19d50190-8656-4e91-8d34-82e646ada9c9",
				IdentifierValue: "77701984-3542-4f77-91aa-b5f7bfa43330"},
		},
	},
}

var expectedConcordanceFTAPersonDetails = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/a671f5a9-b9a4-4836-a174-fc273166f0db",
				APIURL: "http://api.ft.com/concepts/a671f5a9-b9a4-4836-a174-fc273166f0db"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "a671f5a9-b9a4-4836-a174-fc273166f0db"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/a671f5a9-b9a4-4836-a174-fc273166f0db",
				APIURL: "http://api.ft.com/concepts/a671f5a9-b9a4-4836-a174-fc273166f0db"},
			Identifier{
				Authority:       "http://api.ft.com/system/19d50190-8656-4e91-8d34-82e646ada9c9",
				IdentifierValue: "a671f5a9-b9a4-4836-a174-fc273166f0db"},
		},
	},
}

var expectedConcordanceFTAPersonDetailsByAuthority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/a671f5a9-b9a4-4836-a174-fc273166f0db",
				APIURL: "http://api.ft.com/concepts/a671f5a9-b9a4-4836-a174-fc273166f0db"},
			Identifier{
				Authority:       "http://api.ft.com/system/19d50190-8656-4e91-8d34-82e646ada9c9",
				IdentifierValue: "a671f5a9-b9a4-4836-a174-fc273166f0db"},
		},
	},
}

var expectedConcordanceSVCategory = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/e0fc58d1-8dc5-47c6-90b1-59ccf8217366",
				APIURL: "http://api.ft.com/things/e0fc58d1-8dc5-47c6-90b1-59ccf8217366"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "e0fc58d1-8dc5-47c6-90b1-59ccf8217366"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/e0fc58d1-8dc5-47c6-90b1-59ccf8217366",
				APIURL: "http://api.ft.com/things/e0fc58d1-8dc5-47c6-90b1-59ccf8217366"},
			Identifier{
				Authority:       "http://api.ft.com/system/8e6c705e-1132-42a2-8db0-c295e29e8658",
				IdentifierValue: "e0fc58d1-8dc5-47c6-90b1-59ccf8217366"},
		},
	},
}

var expectedConcordanceSVCategoryByAuthority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/e0fc58d1-8dc5-47c6-90b1-59ccf8217366",
				APIURL: "http://api.ft.com/things/e0fc58d1-8dc5-47c6-90b1-59ccf8217366"},
			Identifier{
				Authority:       "http://api.ft.com/system/8e6c705e-1132-42a2-8db0-c295e29e8658",
				IdentifierValue: "e0fc58d1-8dc5-47c6-90b1-59ccf8217366"},
		},
	},
}

var expectedConcordancePersonGeneric = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/3c4666ef-b403-4313-b648-d639762750e4",
				APIURL: "http://api.ft.com/people/3c4666ef-b403-4313-b648-d639762750e4"},
			Identifier{
				Authority:       "http://api.ft.com/system/UPP",
				IdentifierValue: "3c4666ef-b403-4313-b648-d639762750e4"},
		},
		{
			Concept{
				ID:     "http://api.ft.com/things/3c4666ef-b403-4313-b648-d639762750e4",
				APIURL: "http://api.ft.com/people/3c4666ef-b403-4313-b648-d639762750e4"},
			Identifier{
				Authority:       "http://api.ft.com/system/GENERIC",
				IdentifierValue: "3c4666ef-b403-4313-b648-d639762750e4"},
		},
	},
}

var expectedConcordancePersonGenericByAuthority = Concordances{
	[]Concordance{
		{
			Concept{
				ID:     "http://api.ft.com/things/3c4666ef-b403-4313-b648-d639762750e4",
				APIURL: "http://api.ft.com/people/3c4666ef-b403-4313-b648-d639762750e4"},
			Identifier{
				Authority:       "http://api.ft.com/system/GENERIC",
				IdentifierValue: "3c4666ef-b403-4313-b648-d639762750e4"},
		},
	},
}

func TestNeoReadByConceptID(t *testing.T) {
	driver := getNeoDriver(assert.New(t))

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
		{
			name:        "SVProvision",
			fixture:     "SVProvision-Unconcorded-1808c3fc-04bb-589b-a457-640bffa8f6c6.json",
			conceptIDs:  []string{"1808c3fc-04bb-589b-a457-640bffa8f6c6"},
			expectedLen: 2,
			expected:    expectedConcordanceSVProvision,
		},
		{
			name:        "FTPCSource",
			fixture:     "FTPCSource-Unconcorded-8a852776-38b4-47fc-bb5e-e496801a28bf.json",
			conceptIDs:  []string{"8a852776-38b4-47fc-bb5e-e496801a28bf"},
			expectedLen: 2,
			expected:    expectedConcordanceFTPCSource,
		},
		{
			name:        "FTPCGenre",
			fixture:     "FTPCGenre-Unconcorded-e02fb4c0-1fe5-476b-b791-e921db5b99f2.json",
			conceptIDs:  []string{"e02fb4c0-1fe5-476b-b791-e921db5b99f2"},
			expectedLen: 2,
			expected:    expectedConcordanceFTPCGenre,
		},
		{
			name:        "FTPCAssetType",
			fixture:     "FTPCAssetType-Unconcorded-c5440e5e-a472-4948-ab33-97e0089dd926.json",
			conceptIDs:  []string{"c5440e5e-a472-4948-ab33-97e0089dd926"},
			expectedLen: 2,
			expected:    expectedConcordanceFTPCAssetType,
		},
		{
			name:        "FTAOrganisationDetails",
			fixture:     "FTAOrganisationDetails-Unconcorded-77701984-3542-4f77-91aa-b5f7bfa43330.json",
			conceptIDs:  []string{"77701984-3542-4f77-91aa-b5f7bfa43330"},
			expectedLen: 2,
			expected:    expectedConcordanceFTAOrganisationDetails,
		},
		{
			name:        "FTAPersonDetails",
			fixture:     "FTAPersonDetails-Unconcorded-a671f5a9-b9a4-4836-a174-fc273166f0db.json",
			conceptIDs:  []string{"a671f5a9-b9a4-4836-a174-fc273166f0db"},
			expectedLen: 2,
			expected:    expectedConcordanceFTAPersonDetails,
		},
		{
			name:        "PersonGeneric",
			fixture:     "Person-Generic-3c4666ef-b403-4313-b648-d639762750e4.json",
			conceptIDs:  []string{"3c4666ef-b403-4313-b648-d639762750e4"},
			expectedLen: 2,
			expected:    expectedConcordancePersonGeneric,
		},
		{
			name:        "SVCategory",
			fixture:     "SVCategory-Unconcorded-e0fc58d1-8dc5-47c6-90b1-59ccf8217366.json",
			conceptIDs:  []string{"e0fc58d1-8dc5-47c6-90b1-59ccf8217366"},
			expectedLen: 2,
			expected:    expectedConcordanceSVCategory,
		},
		{
			name:        "FTAnIIndustryClassification",
			fixture:     "FTAnIIndustryClassification-97b56e0e-3526-4434-ad29-349b06ead4a3.json",
			conceptIDs:  []string{"97b56e0e-3526-4434-ad29-349b06ead4a3"},
			expectedLen: 3,
			expected: Concordances{
				[]Concordance{
					{
						Concept{
							ID:     "http://api.ft.com/things/97b56e0e-3526-4434-ad29-349b06ead4a3",
							APIURL: "http://api.ft.com/things/97b56e0e-3526-4434-ad29-349b06ead4a3"},
						Identifier{
							Authority:       "http://api.ft.com/system/SMARTLOGIC",
							IdentifierValue: "97b56e0e-3526-4434-ad29-349b06ead4a3"},
					},
					{
						Concept{
							ID:     "http://api.ft.com/things/97b56e0e-3526-4434-ad29-349b06ead4a3",
							APIURL: "http://api.ft.com/things/97b56e0e-3526-4434-ad29-349b06ead4a3"},
						Identifier{
							Authority:       "http://api.ft.com/system/FT-AnI",
							IdentifierValue: "ELE"},
					},
					{
						Concept{
							ID:     "http://api.ft.com/things/97b56e0e-3526-4434-ad29-349b06ead4a3",
							APIURL: "http://api.ft.com/things/97b56e0e-3526-4434-ad29-349b06ead4a3"},
						Identifier{
							Authority:       "http://api.ft.com/system/UPP",
							IdentifierValue: "97b56e0e-3526-4434-ad29-349b06ead4a3"},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			writeConceptFixture(t, driver, "./fixtures/"+test.fixture)
			defer cleanUp(assert.New(t), driver)

			undertest, err := NewCypherDriver(driver, "http://api.ft.com")
			assert.NoError(t, err)
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
			name:             "SVProvision",
			fixture:          "SVProvision-Unconcorded-1808c3fc-04bb-589b-a457-640bffa8f6c6.json",
			authority:        "http://api.ft.com/system/8e6c705e-1132-42a2-8db0-c295e29e8658",
			identifierValues: []string{"65d735ebad5f88460e919a42"},
			expected:         expectedConcordanceSVProvisionByAuthority,
		},
		{
			name:             "FTPCSource",
			fixture:          "FTPCSource-Unconcorded-8a852776-38b4-47fc-bb5e-e496801a28bf.json",
			authority:        "http://api.ft.com/system/724b5e36-6d45-4cf1-b1c2-3f676b21f21b",
			identifierValues: []string{"8a852776-38b4-47fc-bb5e-e496801a28bf"},
			expected:         expectedConcordanceFTPCSourceByAuthority,
		},
		{
			name:             "FTPCGenre",
			fixture:          "FTPCGenre-Unconcorded-e02fb4c0-1fe5-476b-b791-e921db5b99f2.json",
			authority:        "http://api.ft.com/system/724b5e36-6d45-4cf1-b1c2-3f676b21f21b",
			identifierValues: []string{"e02fb4c0-1fe5-476b-b791-e921db5b99f2"},
			expected:         expectedConcordanceFTPCGenreByAuthority,
		},
		{
			name:             "FTPCAssetType",
			fixture:          "FTPCAssetType-Unconcorded-c5440e5e-a472-4948-ab33-97e0089dd926.json",
			authority:        "http://api.ft.com/system/724b5e36-6d45-4cf1-b1c2-3f676b21f21b",
			identifierValues: []string{"c5440e5e-a472-4948-ab33-97e0089dd926"},
			expected:         expectedConcordanceFTPCAssetTypeByAuthority,
		},
		{
			name:             "FTAOrganisationDetails",
			fixture:          "FTAOrganisationDetails-Unconcorded-77701984-3542-4f77-91aa-b5f7bfa43330.json",
			authority:        "http://api.ft.com/system/19d50190-8656-4e91-8d34-82e646ada9c9",
			identifierValues: []string{"77701984-3542-4f77-91aa-b5f7bfa43330"},
			expected:         expectedConcordanceFTAOrganisationDetailsByAuthority,
		},
		{
			name:             "FTAPersonDetails",
			fixture:          "FTAPersonDetails-Unconcorded-a671f5a9-b9a4-4836-a174-fc273166f0db.json",
			authority:        "http://api.ft.com/system/19d50190-8656-4e91-8d34-82e646ada9c9",
			identifierValues: []string{"a671f5a9-b9a4-4836-a174-fc273166f0db"},
			expected:         expectedConcordanceFTAPersonDetailsByAuthority,
		},
		{
			name:             "PersonGeneric",
			fixture:          "Person-Generic-3c4666ef-b403-4313-b648-d639762750e4.json",
			authority:        "http://api.ft.com/system/GENERIC",
			identifierValues: []string{"3c4666ef-b403-4313-b648-d639762750e4"},
			expected:         expectedConcordancePersonGenericByAuthority,
		},
		{
			name:             "SVCategory",
			fixture:          "SVCategory-Unconcorded-e0fc58d1-8dc5-47c6-90b1-59ccf8217366.json",
			authority:        "http://api.ft.com/system/8e6c705e-1132-42a2-8db0-c295e29e8658",
			identifierValues: []string{"e0fc58d1-8dc5-47c6-90b1-59ccf8217366"},
			expected:         expectedConcordanceSVCategoryByAuthority,
		},
		{
			name:             "FTAnIIndustryClassification",
			fixture:          "FTAnIIndustryClassification-97b56e0e-3526-4434-ad29-349b06ead4a3.json",
			authority:        "http://api.ft.com/system/FT-AnI",
			identifierValues: []string{"ELE"},
			expected: Concordances{
				[]Concordance{
					{
						Concept{
							ID:     "http://api.ft.com/things/97b56e0e-3526-4434-ad29-349b06ead4a3",
							APIURL: "http://api.ft.com/things/97b56e0e-3526-4434-ad29-349b06ead4a3"},
						Identifier{
							Authority:       "http://api.ft.com/system/FT-AnI",
							IdentifierValue: "ELE"},
					},
				},
			},
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
			writeConceptFixture(t, driver, "./fixtures/"+test.fixture)
			defer cleanUp(assert.New(t), driver)

			undertest, err := NewCypherDriver(driver, "http://api.ft.com")
			assert.NoError(t, err)
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

func writeConceptFixture(t *testing.T, driver *cmneo4j.Driver, fixture string) {
	f, err := os.Open(fixture)
	if err != nil {
		t.Fatalf("failed to open file '%s': %v", fixture, err)
	}
	concept := ontology.CanonicalConcept{}
	err = json.NewDecoder(f).Decode(&concept)
	if err != nil {
		t.Fatalf("failed to read concept data from '%s': %v", fixture, err)
	}
	query, err := neo4j.WriteCanonicalConceptQueries(concept)
	if err != nil {
		t.Fatalf("failed to construct concept write query for '%s': %v", fixture, err)
	}
	err = driver.Write(query...)
	if err != nil {
		t.Fatalf("failed to write concept in neo4j for '%s': %v", fixture, err)
	}
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
