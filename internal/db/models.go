package db

// Account mock account type for the database connection test
type Account struct {
	Address int64 `gorm:"primaryKey"`
	Amount  int64
}
