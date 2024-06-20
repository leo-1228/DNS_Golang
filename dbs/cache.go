package dbs

import (
	"encoding/json"
	"os"

	//"github.com/glebarez/sqlite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
)

// Cache layer for clients in case connection to server is not available
type ClientCacheService struct {
	db *gorm.DB
}

var ClientCache *ClientCacheService

func SetupClientCache() error {
	dbLocation := os.Getenv("DATABASE_CACHE_PATH")
	if dbLocation == "" {
		dbLocation = "clientcache.db"
	}

	// Create the sqlite file if it's not available
	if _, err := os.Stat(dbLocation); err != nil {
		if _, err = os.Create(dbLocation); err != nil {
			return err
		}
	}

	db, err := gorm.Open(sqlite.Open(dbLocation), &gorm.Config{
		Logger: glog.Default.LogMode(glog.Silent),
	})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&ResultInfo{})
	if err != nil {
		return err
	}

	ClientCache = &ClientCacheService{
		db: db,
	}
	return err
}

func (c *ClientCacheService) AddResult(res *ResultInfo) error {

	// JSON encode to string
	u, err := json.Marshal(res.Infos)
	if err != nil {
		return err
	}
	res.InfoJSON = string(u)

	e, err := json.Marshal(res.Errors)
	if err != nil {
		return err
	}
	res.ErrorsJSON = string(e)

	result := c.db.Create(res)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (c *ClientCacheService) GetCachedResults() ([]ResultInfo, error) {
	var results []ResultInfo
	result := c.db.Find(&results)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, v := range results {
		var infos []DomainInfo
		err := json.Unmarshal([]byte(v.InfoJSON), &infos)
		if err != nil {
			return nil, err
		}
		v.Infos = infos

		var errs []string
		err = json.Unmarshal([]byte(v.ErrorsJSON), &errs)
		if err != nil {
			return nil, err
		}
		v.Errors = errs
	}

	return results, nil
}

func (c *ClientCacheService) RemoveResult(res *ResultInfo) error {
	if res.ID == 0 {
		return nil
	}

	result := c.db.Delete(res)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
