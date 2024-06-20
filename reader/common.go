package reader

type DomainReader interface {
	Close()
	Batch() ([]string, error)
}

type Config struct {
	DomainsFileName string
	Workspace       string
	From            int
	To              int

	BatchSize int
}
