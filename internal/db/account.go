package db

import "github.com/shopspring/decimal"

type Account struct {
	Address *Address        `gorm:"primary_key,type:bytea"`
	Amount  decimal.Decimal `gorm:"type:numeric(78,0)"`
}
