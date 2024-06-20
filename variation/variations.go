package variation

import (
	"errors"
	"log"
	"sync"
)

type VariationRecord struct {
	Variant    string
	How        string
	MainDomain string
}

type VariationResult struct {
	Domain     string
	Variations []VariationRecord
}

type Config struct {
	MainDomains                []string
	CheckMainDomainDuplication bool
	ValidSpaces                []string
	HomoglyphMethod            bool
	HomoglyphNormal            bool
	AdditionMethod             bool
	BitsquattingMethod         bool
	HyphenationMethod          bool
	InsertionMethod            bool
	OmissionMethod             bool
	RepetitionMethod           bool
	ReplacementMethod          bool
	TranspositionMethod        bool
	VowelSwapMethod            bool
	DictionaryMethod           bool
	DictionaryData             []string
}

type Variation struct {
	mainDomains                map[string]bool
	checkMainDomainDuplication bool
	methods                    []string
	dictionaryData             []string
	homoglyphNormal            bool
	validSpaces                []string

	actions map[string]func(domain string) ([]VariationRecord, error)
}

func (v *Variation) Run(domains []string) ([]VariationResult, error) {

	if len(domains) == 0 {
		return nil, errors.New("variations: no domain found")
	}

	result := make([]VariationResult, 0)
	domainMux := sync.Mutex{}
	domainWg := sync.WaitGroup{}
	domainWg.Add(len(domains))

	for i := 0; i < len(domains); i++ {

		go func(domain string) {
			defer func() {
				domainWg.Done()
			}()

			info, err := DomainInfo(domain, v.validSpaces)
			if err != nil {
				log.Fatal(err)
			}

			allDomains := info.GetInAllSpaces(v.validSpaces)

			iResult := VariationResult{
				Variations: make([]VariationRecord, 0),
				Domain:     domain,
			}

			for j := 0; j < len(allDomains); j++ {

				domain := allDomains[j]

				wg := sync.WaitGroup{}
				wg.Add(len(v.actions))
				mux := sync.Mutex{}

				for index := range v.actions {

					ac := v.actions[index]

					go func(domain string, ac func(string) ([]VariationRecord, error)) {
						defer func() {
							wg.Done()
						}()
						tr, ierr := ac(domain)

						if ierr != nil {
							log.Fatal("variations: error in generaing variations")
						}
						mux.Lock()
						iResult.Variations = append(iResult.Variations, tr...)
						mux.Unlock()

					}(domain, ac)

				}
				wg.Wait()

				iResult.Variations = append(iResult.Variations, VariationRecord{
					Variant: domain,
					How:     "Main",
				})

			}

			domainMux.Lock()
			result = append(result, iResult)
			domainMux.Unlock()

		}(domains[i])
	}

	domainWg.Wait()
	return result, nil
}
func (v *Variation) Run2(domains []string) ([]VariationResult, error) {

	if len(domains) == 0 {
		return nil, errors.New("variations: no domain found")
	}

	// counting number of domain in all spaces
	count := 0
	for i := 0; i < len(domains); i++ {

		info, err := DomainInfo(domains[i], v.validSpaces)
		if err != nil {
			return nil, err
		}

		allDomains := info.GetInAllSpaces(v.validSpaces)
		count += len(allDomains)
	}

	result := make([]VariationResult, 0)
	wg := sync.WaitGroup{}
	wg.Add(count)
	mux := sync.Mutex{}
	var err error

	for i := 0; i < len(domains); i++ {

		info, err := DomainInfo(domains[i], v.validSpaces)
		if err != nil {
			return nil, err
		}

		allDomains := info.GetInAllSpaces(v.validSpaces)

		spaceMux := sync.Mutex{}
		spaceWg := sync.WaitGroup{}
		spaceWg.Add(len(allDomains))

		iR := make(map[string]bool)
		iResult := VariationResult{
			Variations: make([]VariationRecord, 0),
			Domain:     domains[i],
		}

		for j := 0; j < len(allDomains); j++ {

			go func(domain string) {

				defer func() {
					if r := recover(); r != nil {
						mux.Lock()
						err = errors.New("variations: panic in generaing variations")
						mux.Unlock()
					}
					wg.Done()
				}()

				mR := make(map[string]bool)
				vMr := make([]VariationRecord, 0)
				iWg := sync.WaitGroup{}
				iWg.Add(len(v.actions))
				iMux := sync.Mutex{}
				for _, ac := range v.actions {

					go func(ac func(domain string) ([]VariationRecord, error)) {

						defer func() {
							iWg.Done()
						}()

						tr, ierr := ac(domain)

						if ierr != nil {
							mux.Lock()
							err = errors.New("variations: error in generaing variations")
							mux.Unlock()
							return
						}
						iMux.Lock()
						for index, key := range tr {
							if _, ok := mR[key.Variant]; !ok {
								mR[key.Variant] = true
								vMr = append(vMr, tr[index])
							}
						}
						iMux.Unlock()

					}(ac)
				}
				iWg.Wait()

				spaceMux.Lock()
				for index, key := range vMr {
					if _, ok := iR[key.Variant]; !ok {
						iR[key.Variant] = true
						iResult.Variations = append(iResult.Variations, vMr[index])
					}
				}
				spaceMux.Unlock()

				spaceWg.Done()

			}(allDomains[j])

		}
		spaceWg.Wait()
		result = append(result, iResult)
	}

	wg.Wait()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func New(cfg Config) (*Variation, error) {

	v := Variation{
		mainDomains:                make(map[string]bool),
		checkMainDomainDuplication: cfg.CheckMainDomainDuplication,
		methods:                    make([]string, 0),
		actions:                    make(map[string]func(string) ([]VariationRecord, error)),
	}

	if v.checkMainDomainDuplication {
		for _, domain := range cfg.MainDomains {
			v.mainDomains[domain] = true
		}
	}

	v.validSpaces = cfg.ValidSpaces

	if cfg.HomoglyphMethod {
		v.methods = append(v.methods, "Homoglyph")
		v.actions["Homoglyph"] = v.Homoglyph
		v.homoglyphNormal = cfg.HomoglyphNormal
	}
	if cfg.AdditionMethod {
		v.methods = append(v.methods, "Addition")
		v.actions["Addition"] = v.Addition
	}
	if cfg.BitsquattingMethod {
		v.methods = append(v.methods, "Bitsquatting")
	}
	if cfg.HyphenationMethod {
		v.methods = append(v.methods, "Hyphenation")
		v.actions["Hyphenation"] = v.Hyphenation

	}
	if cfg.InsertionMethod {
		v.methods = append(v.methods, "Insertion")
		v.actions["Insertion"] = v.Insertion
	}
	if cfg.OmissionMethod {
		v.methods = append(v.methods, "Omission")
		v.actions["Omission"] = v.Omission
	}
	if cfg.RepetitionMethod {
		v.methods = append(v.methods, "Repetition")
		v.actions["Repetition"] = v.Repetition

	}
	if cfg.ReplacementMethod {
		v.methods = append(v.methods, "Replacement")
		v.actions["Replacement"] = v.Replacement
	}
	if cfg.TranspositionMethod {
		v.methods = append(v.methods, "Transposition")
		v.actions["Transposition"] = v.Transposition
	}
	if cfg.VowelSwapMethod {
		v.methods = append(v.methods, "VowelSwap")
		v.actions["VowelSwap"] = v.VowelSwap
	}
	if cfg.DictionaryMethod {
		v.methods = append(v.methods, "Dictionary")
		v.dictionaryData = cfg.DictionaryData
		v.actions["Dictionary"] = v.Dictionary
	}

	return &v, nil
}
