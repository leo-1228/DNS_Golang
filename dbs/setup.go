package dbs

import (
	"dnscheck/logger"
	"fmt"
	"os"

	//"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
)

func setupPostgres(dbSettings map[string]string) (*gorm.DB, error) {
	//dbHost := os.Getenv("DB_HOST")
	//dbPort := os.Getenv("DB_PORT")
	//dbName := os.Getenv("DB_DATABASE")
	//dbUser := os.Getenv("DB_USER")
	//dbPassword := os.Getenv("DB_PASSWORD")

	dbHost := dbSettings["host"]
	dbPort := dbSettings["port"]
	dbName := dbSettings["database"]
	dbUser := dbSettings["user"]
	dbPassword := dbSettings["password"]

	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		dbHost,
		dbPort,
		dbUser,
		dbName,
		dbPassword)

	db, err := gorm.Open(postgres.Open(connectionString))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func setupMysql(dbSettings map[string]string) (*gorm.DB, error) {
	//dbHost := os.Getenv("DB_HOST")
	//dbPort := os.Getenv("DB_PORT")
	//dbName := os.Getenv("DB_DATABASE")
	//dbUser := os.Getenv("DB_USER")
	//dbPassword := os.Getenv("DB_PASSWORD")

	dbHost := dbSettings["host"]
	dbPort := dbSettings["port"]
	dbName := dbSettings["database"]
	dbUser := dbSettings["user"]
	dbPassword := dbSettings["password"]

	// "user:pass@tcp(127.0.0.1:3306)/dbname"
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	db, err := gorm.Open(mysql.Open(connectionString))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func setupSQLite(dbLocation string) (*gorm.DB, error) {
	//dbLocation := os.Getenv("DATABASE_PATH")
	//if dbLocation == "" {
	//	dbLocation = "domaindata.sqlite3"
	//}

	// Create the sqlite file if it's not available
	if _, err := os.Stat(dbLocation); err != nil {
		if _, err = os.Create(dbLocation); err != nil {
			return nil, err
		}

		logger.Warning("SQLite file not found, creating it", dbLocation)
	}

	db, err := gorm.Open(sqlite.Open(dbLocation), &gorm.Config{
		Logger: glog.Default.LogMode(glog.Silent),
	})
	return db, err
}

func InitializeDatabaseService(dbType string, dbSettings map[string]string) error {

	// dbs := os.Getenv("DB")
	// var db *gorm.DB
	// var duplicateDB *gorm.DB
	// var err error

	// switch dbType {
	// case "sqlite":
	// 	db, err = setupSQLite(dbSettings["path"])
	// 	duplicateDB, err = setupSQLite(dbSettings["path"])
	// 	break
	// case "postgres":
	// 	db, err = setupPostgres(dbSettings)
	// 	break
	// case "mysql":
	// 	db, err = setupMysql(dbSettings)
	// 	break
	// default:
	// 	db, err = setupSQLite(dbSettings["path"])
	// 	duplicateDB, err = setupSQLite(dbSettings["path"])
	// 	break
	// }

	// if err != nil {
	// 	return err
	// }

	// err = db.AutoMigrate(&ProcessedDomain{})
	// if err != nil {
	// 	return err
	// }

	// Service = NewDBService(db, duplicateDB)
	// return nil
}

func InitializeVariantService(dbType string, dbSettings map[string]string) error {

	// dbs := os.Getenv("DB")
	var db *gorm.DB
	var err error

	switch dbType {
	case "sqlite":
		db, err = setupSQLite(dbSettings["path"])
		break
	case "postgres":
		db, err = setupPostgres(dbSettings)
		break
	case "mysql":
		db, err = setupMysql(dbSettings)
		break
	default:
		db, err = setupSQLite(dbSettings["path"])
		break
	}

	if err != nil {
		return err
	}

	err = db.AutoMigrate(&ProcessedDomain{})
	if err != nil {
		return err
	}

	VariantService = NewVariantService(db)
	return nil
}

func setupInvalidDomainsSQLite() (*gorm.DB, error) {
	dbLocation := os.Getenv("INVALID_DOMAIN_DATABASE_PATH")
	if dbLocation == "" {
		dbLocation = "invaliddomains.sqlite3"
	}

	// Create the sqlite file if it's not available
	if _, err := os.Stat(dbLocation); err != nil {
		if _, err = os.Create(dbLocation); err != nil {
			return nil, err
		}

		logger.Warning("SQLite file not found, creating it", dbLocation)
	}

	db, err := gorm.Open(sqlite.Open(dbLocation), &gorm.Config{
		Logger: glog.Default.LogMode(glog.Silent),
	})
	return db, err
}

func InitializeInvalidDomainDatabaseService() error {

	var db *gorm.DB
	var err error

	db, err = setupInvalidDomainsSQLite()

	if err != nil {
		return err
	}

	err = db.AutoMigrate(&DbVariationDomain{})
	if err != nil {
		return err
	}

	VariationDomainService = NewVariationDomainService(db)
	return nil
}
