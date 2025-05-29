package db

import (
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

// ConnectDb returns pointer to gorm.DB which can be used to
// interact with the database. Applies migrations.
func ConnectDb() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Account{})
	if err != nil {
		sqlDB, err2 := db.DB()
		if err2 != nil {
			return nil, err2
		}

		err2 = sqlDB.Close()
		if err2 != nil {
			return nil, err2
		}

		return nil, err
	}

	return db, nil
}

const defaultCurrencyAmount int64 = 10

// CreateDefaultAccount creates the default account if it does not exist.
func CreateDefaultAccount(db *gorm.DB) error {
	defaultAccount := Account{
		Address: &Address{},
		Amount:  decimal.NewFromInt(defaultCurrencyAmount),
	}
	tx := db.FirstOrCreate(&defaultAccount)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}
