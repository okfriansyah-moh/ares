package arslib

// Composer translates .ai/ into a provider-specific artifact.
type Composer interface {
	Compose(root string, repo *Repository) error
	Target() string
}

// Importer reads a provider artifact and returns an in-memory Repository.
type Importer interface {
	Import(root string) (*Repository, error)
	Source() string
}

// Validator checks .ai/ structure and returns all findings.
type Validator interface {
	Validate(root string) ([]Finding, error)
}
