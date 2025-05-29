package db

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
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

	if len(v) == 0 {
		var zeroAdd Address
		*a = zeroAdd
		return nil
	}

	var z Address
	if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&z); err != nil {
		return fmt.Errorf("can't decode Address from []byte: %w", err)
	}
	*a = z
	return nil
}

func (a *Address) Value() (driver.Value, error) {
	underlyingAddress := Address(*a)

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&underlyingAddress); err != nil {
		return nil, fmt.Errorf("can't encode Address to []byte: %w", err)
	}
	return buf.Bytes(), nil
}

func (*Address) DataType() string {
	return "BYTEA"
}

func HexToAddress(s string) *Address {
	tempAddress := Address(common.HexToAddress(s))
	return &tempAddress
}
