package dbs

import (
	"time"

	"gorm.io/gorm"
)

type DNSDataResult struct {
	DomainName  string `gorm:"column:Domain_Name" json:"domainName"`
	ARecords    string `gorm:"column:A_Records" json:"aRecords"`
	Nameservers string `gorm:"column:Nameservers" json:"nameServers"`
}

// TableName overrides the table name
func (DNSDataResult) TableName() string {
	return "tblDomainDNSData"
}

type AvailableDomain struct {
	Domain string `gorm:"column:Domain"`
	Date   string `gorm:"column:Date"`
}

// TableName overrides the table name
func (AvailableDomain) TableName() string {
	return "tblAvailableDomains"
}

type ProcessedDomain struct {
	Domain     string    `gorm:"column:Domain;primaryKey"`
	Variations string    `gorm:"column:Variations"`
	StartDate  time.Time `gorm:"column:Startdate"`
	NewlyAdded bool      `gorm:"column:NewlyAdded"`
	RunId      string    `gorm:"column:RunId"`
}

// TableName overrides the table name
func (ProcessedDomain) TableName() string {
	return "tblProcessedDomains"
}

type DbVariationDomain struct {
	Domain     string `gorm:"column:Domain;primaryKey"`
	MainDomain string `gorm:"column:MainDomain"`
	Processed  bool   `gorm:"column:Processed"`
}

// TableName overrides the table name
func (DbVariationDomain) TableName() string {
	return "tblVariationDomains"
}

type DbVariants struct {
	gorm.Model
	Domain     string `gorm:"column:domain"`
	MainDomain string `gorm:"column:mainDomain"`
	BlockID    uint
}

type DbBlock struct {
	gorm.Model
	Concurrency     int
	CheckMainDomain bool
	Domain          string
	Dns             string
	Id              string
	Variants        []DbVariants
}

/*type VariationRecord struct {
	Variant    string
	How        string
	MainDomain string
}*/
