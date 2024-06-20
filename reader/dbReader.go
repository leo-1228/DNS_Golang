package reader

import (
	"dnscheck/dbs"
)

type DbReader struct {
	// batchSize is the number of lines that will be read at once in every call.
	batchSize int
}

func (r *DbReader) Close() {

}

func (r *DbReader) Batch() ([]string, error) {
	processedDomains, err := dbs.Service.GetNextDomainsToCheck(r.batchSize)
	if err != nil {
		return nil, err
	}

	err = dbs.Service.UpdateTimestampsForDomains(processedDomains)
	if err != nil {
		return nil, err
	}

	domainNames := make([]string, len(processedDomains))
	for i, v := range processedDomains {
		domainNames[i] = v.Domain
	}

	return domainNames, nil
}

func NewDbReader(cfg Config) (*DbReader, error) {

	r := &DbReader{
		batchSize: cfg.BatchSize,
	}

	return r, nil
}
