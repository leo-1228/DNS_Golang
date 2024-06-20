package saver

import (
	"dnscheck/dbs"
	"dnscheck/logger"
	"dnscheck/report"
	"dnscheck/tools"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

type MixValid struct {
	dbs.DomainInfo
	reportC chan report.ReportInfo
}
type MixError struct {
	err     string
	reportC chan report.ReportInfo
}
type Saver struct {
	workspace          string
	mainC              chan MixValid
	mainErr            chan MixError
	closer             sync.WaitGroup
	sessionManager     sync.WaitGroup
	errFile            *os.File
	invalidMainFile    *os.File
	invalidVariantFile *os.File
	validVariantFile   *os.File
	validMainFile      *os.File
	reporter           *report.Report
	saveInvalid        bool
}

func (s *Saver) Run(dblocks []dbs.ResultInfo, reporter *report.Report) {

	/*s.reporter = reporter

	s.mainC = make(chan MixValid)
	s.mainErr = make(chan MixError)

	go s.saveValid()
	go s.saveError()

	s.closer.Add(len(dblocks) * 2)

	for i := 0; i < len(dblocks); i++ {
		domainBlock := dblocks[i]

		reportInfo := make(chan report.ReportInfo)

		s.reporter.Log(domainBlock.Domain, domainBlock.All, reportInfo)

		go func(ldomainBlock dbs.ResultInfo, lreportInfo chan report.ReportInfo) {
			for _, d := range ldomainBlock.Infos {
				s.mainC <- MixValid{
					reportC:    lreportInfo,
					DomainInfo: d,
				}
			}
			s.closer.Done()
		}(domainBlock, reportInfo)

		go func(ldomainBlock dbs.ResultInfo, lreportInfo chan report.ReportInfo) {
			for _, d := range ldomainBlock.Errors {
				s.mainErr <- MixError{
					err:     d,
					reportC: lreportInfo,
				}
			}
			s.closer.Done()
		}(domainBlock, reportInfo)
	}

	go func() {
		s.closer.Wait()
		close(s.mainC)
		close(s.mainErr)
	}()

	s.sessionManager.Add(2)
	s.sessionManager.Wait()
	*/
}
func (s *Saver) Close() {

}
func New(workspace string, saveInvalid bool) (*Saver, error) {

	created, validMainDomain, err := tools.CreateOrOpenFile(path.Join(workspace, "valid_main_domains.csv"))
	if err != nil {
		return nil, err
	}
	if created {
		validMainDomain.WriteString("main domain , A , name servers\n")
	}
	created, inValidMainDomain, err := tools.CreateOrOpenFile(path.Join(workspace, "invalid_main_domains.txt"))
	if err != nil {
		return nil, err
	}
	if created {
		inValidMainDomain.WriteString("main domain\n")
	}
	created, inValidVariantDomain, err := tools.CreateOrOpenFile(path.Join(workspace, "invalid_variant_domains.csv"))
	if err != nil {
		return nil, err
	}
	if created {
		inValidVariantDomain.WriteString("variant domain , action , main domain\n")
	}
	created, validVariantDomain, err := tools.CreateOrOpenFile(path.Join(workspace, "valid_variant_domains.csv"))
	if err != nil {
		return nil, err
	}
	if created {
		validVariantDomain.WriteString("variant domain , A , name servers , action , main domain\n")
	}
	_, errFile, err := tools.CreateOrOpenFile(path.Join(workspace, "errors.txt"))
	if err != nil {
		return nil, err
	}

	srv := &Saver{
		workspace:          workspace,
		mainC:              make(chan MixValid, 1000),
		mainErr:            make(chan MixError, 1000),
		sessionManager:     sync.WaitGroup{},
		closer:             sync.WaitGroup{},
		invalidMainFile:    inValidMainDomain,
		invalidVariantFile: inValidVariantDomain,
		validVariantFile:   validVariantDomain,
		validMainFile:      validMainDomain,
		errFile:            errFile,
		saveInvalid:        saveInvalid,
	}

	return srv, nil
}

func (s *Saver) saveValid() {

	validMainDomainData := strings.Builder{}
	validVariantDomainData := strings.Builder{}

	inValidMainDomainData := strings.Builder{}
	inValidVariantDomainData := strings.Builder{}

	for info := range s.mainC {
		if !info.IsValid {

			if info.Domain.How == "Main" {

				if s.saveInvalid {
					_, err := inValidMainDomainData.WriteString(info.Domain.Variant + "\n")
					if err != nil {
						log.Fatal("writing to invalid file error: ", err)
					}
				}
				info.reportC <- report.ReportInfo{
					Domain: info.MainDomain,
					Valid:  false,
				}
				continue
			}

			if s.saveInvalid {
				// store variant domain
				// save to invalid valid files
				_, err := inValidVariantDomainData.WriteString(
					strings.Join([]string{
						info.Domain.Variant,
						info.Domain.How,
						info.MainDomain}, ",") + "\n")
				if err != nil {
					log.Fatal("writing to invalid file error: ", err)
				}
			}
			info.reportC <- report.ReportInfo{
				Domain: info.MainDomain,
				Valid:  false,
			}

			continue
		}

		if info.Domain.How == "Main" {

			_, err := validMainDomainData.WriteString(strings.Join([]string{
				info.Domain.Variant,
				"[" + strings.Join(info.Host, "|") + "]",
				"[" + strings.Join(info.IP, "|") + "]",
			}, ",") + "\n")
			if err != nil {
				log.Fatal("writing to invalid file error: ", err)
			}
			if err != nil {
				log.Fatal("writing to file error: ", err)
			}

			info.reportC <- report.ReportInfo{
				Domain: info.MainDomain,
				Valid:  true,
			}

			continue
		}

		// save to valid files
		_, err := validVariantDomainData.WriteString(strings.Join([]string{
			info.Domain.Variant,
			"[" + strings.Join(info.Host, "|") + "]",
			"[" + strings.Join(info.IP, "|") + "]",
			info.Domain.How,
			info.MainDomain,
		}, ",") + "\n")

		if err != nil {
			log.Fatal("writing to file error: ", err)
		}

		info.reportC <- report.ReportInfo{
			Domain: info.MainDomain,
			Valid:  true,
		}

	}

	invalidMain := inValidMainDomainData.String()
	invalidVariant := inValidVariantDomainData.String()
	validMain := validMainDomainData.String()
	validVariant := validVariantDomainData.String()

	if len(invalidMain) > 0 {
		logger.Info("saving invalid main domains to the invalid_main_domains.txt ...")

		_, err := s.invalidMainFile.WriteString(invalidMain)
		if err != nil {
			log.Fatal("writing to invalid file error: ", err)
		}
	}
	if len(invalidVariant) > 0 {
		logger.Info("saving invalid variant domains to the invalid_variant_domains.csv ...")

		_, err := s.invalidVariantFile.WriteString(invalidVariant)
		if err != nil {
			log.Fatal("writing to invalid file error: ", err)
		}
	}
	if len(validMain) > 0 {
		logger.Info("saving valid main domains to the valid_main_domains.csv ...")

		_, err := s.validMainFile.WriteString(validMain)
		if err != nil {
			log.Fatal("writing to invalid file error: ", err)
		}

	}
	if len(validVariant) > 0 {
		logger.Info("saving valid variant domains to the valid_variant_domains.csv ...")

		_, err := s.validVariantFile.WriteString(validVariant)
		if err != nil {
			log.Fatal("writing to invalid file error: ", err)
		}
	}

	err := s.invalidMainFile.Sync()
	if err != nil {
		log.Fatal("writing to file error: ", err)
	}
	logger.Info("saving invalid main domains done")

	err = s.invalidVariantFile.Sync()
	if err != nil {
		log.Fatal("writing to file error: ", err)
	}
	logger.Info("saving invalid variant domains done")

	err = s.validMainFile.Sync()
	if err != nil {
		log.Fatal("writing to file error: ", err)
	}
	logger.Info("saving valid main domains done")

	err = s.validVariantFile.Sync()
	if err != nil {
		log.Fatal("writing to file error: ", err)
	}
	logger.Info("saving valid variant domains done")

	s.sessionManager.Done()
}

func (s *Saver) saveError() {

	tempData := strings.Builder{}

	for err := range s.mainErr {
		// save to valid files

		_, ef := tempData.WriteString(err.err + "\n")
		if ef != nil {
			log.Fatal("writing to error file error: ", ef)
		}

		err.reportC <- report.ReportInfo{
			Domain: err.err,
			Valid:  true,
		}

	}

	logger.Info("saving errors to the errors.txt..")

	_, ef := s.errFile.WriteString(tempData.String())
	if ef != nil {
		log.Fatal("writing to error file error: ", ef)
	}

	ef = s.errFile.Sync()
	if ef != nil {
		log.Fatal("writing to error file error: ", ef)
	}

	logger.Info("saving errors done.")

	s.sessionManager.Done()

}
