package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
	"token-transfer-api/internal/address"
	"token-transfer-api/internal/decimal"
)

const useLogger = true

// ConnectDb returns pointer to gorm.DB which can be used to
// interact with the database. Applies migrations.
func ConnectDb() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")

	var newLogger logger.Interface
	if useLogger {
		newLogger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // IO writer
			logger.Config{
				SlowThreshold:             time.Second, // Slow SQL threshold
				LogLevel:                  logger.Info, // Log level
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound
				Colorful:                  true,        // Disable color
			},
		)
	} else {
		newLogger = logger.Discard
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Account{})
	if err != nil {
		sqlDB, err2 := db.DB()
		if err2 != nil {
			return nil, fmt.Errorf("migration failed (%w), and failed to get underlying *sql.DB (%w)", err, err2)
		}

		err2 = sqlDB.Close()
		if err2 != nil {
			return nil, fmt.Errorf("migration failed (%w), and failed to close database connection (%w)", err, err2)
		}

		return nil, err
	}

	return db, nil
}

func CloseDb(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	if err != nil {
		return err
	}
	return nil
}

const DefaultCurrencyAmount int64 = 1_000_000
const DefaultAccountHex = "0x0000000000000000000000000000000000000000"

// CreateDefaultAccount creates the default account if it does not exist.
func CreateDefaultAccount(db *gorm.DB) error {
	defaultAccount := Account{
		Address: address.HexToAddress(DefaultAccountHex),
		Amount:  decimal.NewFromInt64(DefaultCurrencyAmount),
	}
	err := db.Where("address = ?", defaultAccount.Address).
		FirstOrCreate(&defaultAccount).Error
	if err != nil {
		return err
	}

	return nil
}
