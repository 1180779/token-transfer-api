package db

import (
	"token-transfer-api/internal/address"
	"token-transfer-api/internal/decimal"
)

// Account represents a user's cryptocurrency account in the database.
// It stores a unique address as its primary key and the associated balance.
type Account struct {
	Address address.Address `gorm:"primaryKey;type:string;size:42"`
	Amount  decimal.Decimal `gorm:"type:numeric(78,0)"`
}
