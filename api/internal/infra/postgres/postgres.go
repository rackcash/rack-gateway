package postgres

import (
	"fmt"
	"infra/api/internal/config"
	"infra/api/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init(config *config.Config) *gorm.DB {
	dbConfig := config.Postgres
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.Db_name, dbConfig.Port, dbConfig.Ssl_mode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("Gorm error: " + err.Error())
	}

	err = db.AutoMigrate(&domain.Wallets{}, &domain.Merchants{}, &domain.Invoices{}, &domain.Balances{})
	if err != nil {
		panic("Auto migrate error: " + err.Error())
	}

	if err := db.AutoMigrate(&domain.Wallets{}, &domain.Merchants{}, &domain.Invoices{}, &domain.Balances{}, &domain.Events{}); err != nil {
		panic("Auto migrate error: " + err.Error())
	}

	return db
}

type TestConfig struct {
	Host     string
	User     string
	Password string
	DbName   string
	Port     uint16
}

var TEST_CONFIG = TestConfig{
	Host:     "localhost",
	User:     "postgres",
	Password: "lol",
	DbName:   "test",
	Port:     5432,
}

func InitTest(dbConfig TestConfig) *gorm.DB {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.DbName, dbConfig.Port, "disable")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("Gorm error: " + err.Error())
	}

	err = db.AutoMigrate(&domain.Wallets{}, &domain.Merchants{}, &domain.Invoices{}, &domain.Balances{})
	if err != nil {
		panic("Auto migrate error: " + err.Error())
	}

	if err := db.AutoMigrate(&domain.Wallets{}, &domain.Merchants{}, &domain.Invoices{}, &domain.Balances{}, &domain.Events{}); err != nil {
		panic("Auto migrate error: " + err.Error())
	}

	return db
}

func DropTables(db *gorm.DB) error {
	return db.Migrator().DropTable(&domain.Wallets{}, &domain.Merchants{}, &domain.Invoices{}, &domain.Balances{}, &domain.Events{})
}
