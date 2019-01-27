// +build integration

package test

import (
	"fmt"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"natschat/config"
	"sync"
)

var (
	dbOnce sync.Once
	db     *gorm.DB
)

// Ensure that the testdb is a singleton
func GetTestDB() *gorm.DB {
	cfg := config.GetTestConfig()
	dbOnce.Do(func() {
		var err error
		db, err = gorm.Open("postgres", fmt.Sprintf("host=localhost port=5432 user=ben password=password dbname=%s sslmode=disable", cfg.DB.Name))
		if err != nil {
			log.Fatalf("failed to connect to testdb: %v", err)
		}
	})
	return db
}

//// Initial setup to create db and run migrations
//func SetupTestDB(cfg *config.Config) *gorm.DB {
//	db := getPostgresDB()
//
//	var res struct{ Count int }
//	if err := db.Raw("select count(*) from pg_database where datname = ?", cfg.DB.Name).Scan(&res).Error; err != nil {
//		log.Fatalf("failed to get count of dbs: %v", err)
//	}
//
//	if res.Count > 0 {
//		if _, err := db.DB().Exec("drop database " + cfg.DB.Name); err != nil {
//			log.Fatalf("failed to drop old testdb: %v", err)
//		}
//	}
//
//	if _, err := db.DB().Exec("create database " + cfg.DB.Name); err != nil {
//		log.Fatalf("failed to create new test db: %v", err)
//	}
//
//	err := db.Close()
//	if err != nil {
//		log.Fatalf("failed to close postgres: %v", err)
//	}
//
//	db = GetTestDB()
//	driver, err := postgres.WithInstance(db.DB(), &postgres.Config{DatabaseName: cfg.DB.Name})
//	if err != nil {
//		log.Fatalf("Err while getting postgres driver: %v", err)
//	}
//
//	m, err := migrate.NewWithDatabaseInstance("file://resources/db/migrations", cfg.DB.Name, driver)
//	if err != nil {
//		log.Fatalf("got err while migrating testdb: %v", err)
//	}
//	if err := m.Up(); err != nil {
//		log.Fatalf("err while migrating testdb: %v", err)
//	}
//	return db
//}
//
//func getPostgresDB() *gorm.DB {
//	cfg := config.GetTestConfig()
//	db, err := gorm.Open("postgres",
//		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
//			cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.SSLMode))
//	if err != nil {
//		log.Fatalf("failed to connect to postgres: %v", err)
//	}
//	return db
//}
//
//func TearDownDB(db *gorm.DB) {
//	if err := db.Close(); err != nil {
//		log.Fatalf("failed to close test db: %v", err)
//	}
//
//	pdb := getPostgresDB()
//	cfg := config.GetTestConfig()
//	// Force all connections to close
//	q := fmt.Sprintf(`
//		SELECT pg_terminate_backend(pg_stat_activity.pid)
//		FROM pg_stat_activity
//		WHERE pg_stat_activity.datname = '%s'
//		  AND pid <> pg_backend_pid();
//	`, cfg.DB.Name)
//	if err := pdb.Exec(q).Error; err != nil {
//		log.Errorf("failed to drop pg connections: %v", err)
//	}
//	if _, err := pdb.DB().Exec("drop database " + cfg.DB.Name); err != nil {
//		log.Fatalf("failed to drop old testdb: %v", err)
//	}
//	if err := pdb.Close(); err != nil {
//		log.Fatalf("failed to close test db: %v", err)
//	}
//}
