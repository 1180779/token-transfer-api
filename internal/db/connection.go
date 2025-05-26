package db

import (
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
		return nil, err
	}

	return db, nil
}
