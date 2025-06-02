package db

import "github.com/shopspring/decimal"

// Account represents a user's cryptocurrency account in the database.
// It stores a unique address as its primary key and the associated balance.
type Account struct {
	Address Address         `gorm:"primaryKey;type:bytea"`
	Amount  decimal.Decimal `gorm:"type:numeric(78,0)"`
}
