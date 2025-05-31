package db

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

type Address common.Address

func (a *Address) Scan(src any) error {
	if src == nil {
		return errors.New("scan: src cannot be nil for Address")
	}

	v, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("scan: expected []byte from database, got %T", src)
	}

	if len(v) != common.AddressLength {
		return fmt.Errorf("scan: expected byte slice of length %d, got %d", common.AddressLength, len(v))
	}

	copy(a[:], v)
	return nil
}

func (a *Address) Value() (driver.Value, error) {
	return a[:], nil
}

func (*Address) DataType() string {
	return "BYTEA"
}

func (a *Address) String() string {
	tempAddress := common.Address(*a)
	return tempAddress.String()
}

func HexToAddress(s string) *Address {
	tempAddress := Address(common.HexToAddress(s))
	return &tempAddress
}
