package eresolvers

import (
	"errors"
	"fmt"
	"token-transfer-api/internal/address"
)

var InsufficientBalanceError = errors.New("Insufficient balance")

type SenderNotFoundError struct {
	Address address.Address
}

func (e SenderNotFoundError) Error() string {
	return fmt.Sprintf("sender address not found: %s", e.Address.Hex())
}
