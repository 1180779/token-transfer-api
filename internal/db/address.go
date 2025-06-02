package db

import (
	"database/sql/driver"
	"github.com/ethereum/go-ethereum/common"
	"reflect"
	errs "token-transfer-api/internal/errors"
)

const AddressLength = common.AddressLength

type Address common.Address

func (a *Address) Scan(src any) error {
	if src == nil {
		return errs.NilError{Name: "src"}
	}

	v, ok := src.([]byte)
	if !ok {
		return errs.TypeError{ExpectedType: reflect.TypeOf([]byte{}), ActualType: reflect.TypeOf(src)}
	}

	if len(v) != common.AddressLength {
		return errs.LengthError{ExpectedLength: common.AddressLength, ActualLength: len(v)}
	}

	copy(a[:], v)
	return nil
}

func (a *Address) Value() (driver.Value, error) {
	return a[:], nil
}

func (Address) DataType() string {
	return "BYTEA"
}

func (a Address) String() string {
	tempAddress := common.Address(a)
	return tempAddress.String()
}

func (a Address) Hex() string {
	return common.Address(a).Hex()
}

func HexToAddress(s string) *Address {
	tempAddress := Address(common.HexToAddress(s))
	return &tempAddress
}
