package db

import "github.com/shopspring/decimal"

type Account struct {
	Address *Address        `gorm:"primaryKey;type:bytea"`
	Amount  decimal.Decimal `gorm:"type:numeric(78,0)"`
}
