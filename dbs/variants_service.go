package dbs

import (
	"sync"

	"gorm.io/gorm"
)

type VariantServiceT struct {
	db *gorm.DB
	sync.Mutex
}

var VariantService *VariantServiceT

func NewVariantService(db *gorm.DB) *VariantServiceT {
	return &VariantServiceT{
		db: db,
	}
}

// TODO: Transaction and locking
func (vs *VariantServiceT) InsertVariants(blocks []DbBlock) error {
	vs.Lock()
	defer vs.Unlock()
	for _, v := range blocks {
		res := vs.db.Create(&v)
		if res.Error != nil {
			return res.Error
		}
	}

	return nil
}

func (vs *VariantServiceT) GetVariants(limit int) ([]DbBlock, error) {
	vs.Lock()
	defer vs.Unlock()
	var blocks []DbBlock
	err := vs.db.Model(&DbBlock{}).Preload("Variants").Limit(limit).Find(&blocks).Error
	return blocks, err
}

// TODO: Transaction and locking
func (vs *VariantServiceT) MarkMainDomainProcessed(mainDomain string) error {
	vs.Lock()
	defer vs.Unlock()
	res := vs.db.Where("mainDomain LIKE ?", mainDomain).Delete(&DbVariants{})
	if res.Error != nil {
		return res.Error
	}

	res = vs.db.Where("domain LIKE ?", mainDomain).Delete(&DbBlock{})
	if res.Error != nil {
		return res.Error
	}
	return nil
}
