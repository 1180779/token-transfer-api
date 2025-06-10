package eresolvers

import (
	"errors"
	"fmt"
	"token-transfer-api/internal/address"
)

var InsufficientBalanceError = errors.New("insufficient balance")
var NegativeTransferError = errors.New("transfer amount must be positive")
var NonIntegerTransferError = errors.New("transfer amount must be integer")
var BeginTransactionError = errors.New("failed to begin transaction")
var CommitTransactionError = errors.New("failed to commit transaction")

type AddressNotFoundError struct {
	Address address.Address
}

func (e AddressNotFoundError) Error() string {
	return fmt.Sprintf("address not found: %s", e.Address.Hex())
}

type AddressCreationError struct {
	Address address.Address
}

func (e AddressCreationError) Error() string {
	return fmt.Sprintf("address not created: %s", e.Address.Hex())
}

type AddressRetrievalError struct {
	Address address.Address
}

func (e AddressRetrievalError) Error() string {
	return fmt.Sprintf("address could not be retrieved: %s", e.Address.Hex())
}

type AddressAmountUpdateError struct {
	Address address.Address
}

func (e AddressAmountUpdateError) Error() string {
	return fmt.Sprintf("address amount update error: %s", e.Address.Hex())
}
