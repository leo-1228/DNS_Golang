package dbs

import (
	"dnscheck/logger"
	"dnscheck/variation"
	"sync"
	"time"

	"gorm.io/gorm"
)

type DBService struct {
	db          *gorm.DB
	duplicateDB *gorm.DB
	sync.Mutex
}

var Service *DBService

func NewDBService(db, duplicateDB *gorm.DB) *DBService {
	return &DBService{
		db:          db,
		duplicateDB: duplicateDB,
	}
}

func (dbService *DBService) GetDomainsForDuplicateCheck() ([]string, error) {
	var processedDomains []string
	// TODO: We may have duplicate dictionary checks
	result := dbService.duplicateDB.Raw("SELECT Domain from tblProcessedDomains").Scan(&processedDomains)
	if result.Error != nil {
		return nil, result.Error
	}

	return processedDomains, nil
}

/*
	INSERT OR IGNORE INTO tblProcessedDomains SELECT Domain, NULL, NULL,NULL, NULL from (SELECT DISTINCT LOWER(DOMAIN) AS Domain from tblAvailableDomains UNION SELECT DISTINCT LOWER(DOMAIN_NAME) AS Domain from tblDomainDNSData UNION SELECT DISTINCT LOWER(DOMAIN_NAME) AS Domain from tblDomains)
*/

func (dbService *DBService) AddNewEntriesToTableProcessedDomains() (int64, error) {

	// TODO: We may have duplicate dictionary checks
	result := dbService.duplicateDB.Exec("INSERT OR IGNORE INTO tblProcessedDomains SELECT Domain, NULL, NULL,NULL, NULL from (SELECT DISTINCT LOWER(DOMAIN) AS Domain from tblAvailableDomains UNION SELECT DISTINCT LOWER(DOMAIN_NAME) AS Domain from tblDomainDNSData UNION SELECT DISTINCT LOWER(DOMAIN_NAME) AS Domain from tblDomains)")
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (dbService *DBService) DomainDuplicateCheck(whereClause []string) ([]string, error) {
	dbService.Lock()
	var processedDomains []ProcessedDomain

	// TODO: We may have duplicate dictionary checks
	// result := dbService.db.Raw("SELECT Domain from tblProcessedDomains WHERE Domain IN ?", whereClause).Scan(&processedDomains)
	result := dbService.duplicateDB.Where("Domain IN ?", whereClause).Find(&processedDomains)
	dbService.Unlock()
	if result.Error != nil {
		return nil, result.Error
	}

	res := make([]string, len(processedDomains))
	for i, v := range processedDomains {
		res[i] = v.Domain
	}

	return res, nil
}

func (dbService *DBService) GetNextDomainsToCheck(batchSize int) ([]ProcessedDomain, error) {
	dbService.Lock()
	var processedDomains []ProcessedDomain
	// TODO: We may have duplicate dictionary checks
	result := dbService.db.Where("StartDate IS NULL AND (Variations = '' OR Variations IS NULL)").Limit(batchSize).Find(&processedDomains)
	dbService.Unlock()
	if result.Error != nil {
		return nil, result.Error
	}

	return processedDomains, nil
}

/*
"
FROM tblDomainDNSData d
LEFT JOIN tblDomains dd ON UPPER(d.Domain_Name) = dd.Domain_Name
WHERE dd.Domain_Name IS NULL;"
*/

func (dbService *DBService) GetAllNewDomainsFromDNSData() ([]string, error) {
	dbService.Lock()
	var newDomains []string
	result := dbService.db.Raw("SELECT d.Domain_Name FROM tblDomainDNSData d LEFT JOIN tblDomains dd ON UPPER(d.Domain_Name) = dd.Domain_Name WHERE dd.Domain_Name IS NULL").Find(&newDomains)
	dbService.Unlock()
	if result.Error != nil {
		return nil, result.Error
	}

	return newDomains, nil
}

func (dbService *DBService) GetAllNewDomains() ([]ProcessedDomain, error) {
	dbService.Lock()
	var processedDomains []ProcessedDomain
	result := dbService.db.Where("NewlyAdded = TRUE").Find(&processedDomains)
	dbService.Unlock()
	if result.Error != nil {
		return nil, result.Error
	}

	return processedDomains, nil
}

func (dbService *DBService) GetCurentNewDomains(runId string) ([]ProcessedDomain, error) {
	dbService.Lock()
	var processedDomains []ProcessedDomain
	result := dbService.db.Where("NewlyAdded = TRUE AND RunId = ?", runId).Find(&processedDomains)
	dbService.Unlock()
	if result.Error != nil {
		return nil, result.Error
	}

	return processedDomains, nil
}

func (dbService *DBService) UpdateTimestampsForDomains(domains []ProcessedDomain) error {

	dbService.Lock()
	for _, d := range domains {
		d.StartDate = time.Now()
		result := dbService.db.Save(&d)
		if result.Error != nil {
			return result.Error
		}
	}
	dbService.Unlock()
	return nil
}

func (dbService *DBService) MarkDomainProcessed(d *ProcessedDomain) error {

	dbService.Lock()
	result := dbService.db.Save(d)
	dbService.Unlock()
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (dbService *DBService) AddValidDomainBlock(d *DNSDataResult) error {
	dbService.Lock()
	result := dbService.db.Create(d)
	dbService.Unlock()
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (dbService *DBService) ResetUnfinishedDomains() error {
	dbService.Lock()
	result := dbService.db.Exec("UPDATE tblProcessedDomains SET StartDate = NULL WHERE StartDate IS NOT NULL and (Variations is NULL or Variations = '')")
	logger.Info(result)
	dbService.Unlock()
	if result.Error != nil {
		return result.Error
	}

	return nil
}

type DBVariationDomainService struct {
	db *gorm.DB
}

var VariationDomainService *DBVariationDomainService

func NewVariationDomainService(db *gorm.DB) *DBVariationDomainService {
	return &DBVariationDomainService{
		db: db,
	}
}

func (dbInvalidDomainService *DBVariationDomainService) AddVariationDomains(domains []*DomainBlock) (int, []*DomainBlock) {

	numDuplicates := 0
	dbInvalidDomainService.db.Transaction(func(tx *gorm.DB) error {
		for _, dblock := range domains {
			variations := make([]variation.VariationRecord, 0)
			for _, v := range dblock.Variations {
				result := tx.Create(&DbVariationDomain{
					MainDomain: v.MainDomain,
					Domain:     v.Variant,
				})
				// Check for constraint error
				if result.Error != nil {
					// logger.Error(result.Error)
					// Duplicate found, ignore
					numDuplicates++
					continue
				}

				variations = append(variations, v)
			}
			dblock.Variations = variations
		}
		return nil
	})
	// logrus.Error(err)
	return numDuplicates, domains
}

func (dbInvalidDomainService *DBVariationDomainService) AddVariationDomain(d *DbVariationDomain) error {

	result := dbInvalidDomainService.db.Create(d)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (dbInvalidDomainService *DBVariationDomainService) MarkVariationsProcessed(mainDomain string) error {
	result := dbInvalidDomainService.db.Exec("UPDATE tblVariationDomains SET PROCESSED = 1 WHERE MainDomain = ?", mainDomain)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
