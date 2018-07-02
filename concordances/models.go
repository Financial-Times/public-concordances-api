package concordances

// Concordances is a list of concordances wrapped like this for parity in the JSON currently produced
type Concordances struct {
	Concordance []Concordance `json:"concordances,omitempty"`
}

// Concept is a concept equivilant to a thing
type Concept struct {
	ID     string `json:"id"`
	APIURL string `json:"apiUrl"`
}

// Concordance is the structure used for the people API
type Concordance struct {
	Concept    Concept    `json:"concept,omitempty"`
	Identifier Identifier `json:"identifier,omitempty"`
}

// Identifier identifies the concept with alternative identity
type Identifier struct {
	Authority       string `json:"authority"`
	IdentifierValue string `json:"identifierValue"`
}

type neoReadStruct struct {
	CanonicalUUID  string   `json:"canonicalUUID"`
	UUID           string   `json:"UUID"`
	Types          []string `json:"types"`
	Authority      string   `json:"authority"`
	AuthorityValue string   `json:"authorityValue"`
}
