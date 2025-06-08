package resolvers

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"log"
	"os"
	"sync"
	"testing"
	"token-transfer-api/internal/address"
	"token-transfer-api/internal/db"
	"token-transfer-api/internal/decimal"
	"token-transfer-api/internal/errors/eresolvers"
	"token-transfer-api/internal/graph"
	"token-transfer-api/internal/graph/model"
)

var (
	testDB       *gorm.DB
	testResolver *graph.Resolver
)

func TestMain(m *testing.M) {
	if os.Getenv("DATABASE_URL") == "" {
		log.Fatalf("DATABASE_URL not set")
	}

	var err error
	testDB, err = db.ConnectDb()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	testResolver = &graph.Resolver{Db: testDB}

	err = db.CreateDefaultAccount(testDB)
	if err != nil {
		log.Fatalf("Failed to create default account: %v", err)
	}

	exitCode := m.Run()

	if err := db.CloseDb(testDB); err != nil {
		log.Printf("Failed to close test database connection: %v", err)
	}

	os.Exit(exitCode)
}

// clearDBState truncates all tables and recreates default data for a clean test run.
func clearDBState(t *testing.T) {
	t.Helper()
	err := testDB.Exec("TRUNCATE TABLE accounts RESTART IDENTITY CASCADE").Error
	require.NoError(t, err, setupFailed)

	err = db.CreateDefaultAccount(testDB)
	require.NoError(t, err, setupFailed)
}

// getAccountBalance fetches the balance of a given address.
func getAccountBalance(t *testing.T, addr address.Address) decimal.Decimal {
	t.Helper()
	var account = db.Account{Address: addr}
	err := testDB.Where("address = ?", addr).FirstOrCreate(&account).Error
	require.NoError(t, err, fmt.Sprintf("Failed to get balance for %s", addr.Hex()))
	return account.Amount
}

// TestTransfer_SuccessfulTransfer tests a basic successful transfer.
func TestTransfer_SuccessfulTransfer(t *testing.T) {
	// assemble
	clearDBState(t)

	mutationResolver := testResolver.Mutation()
	ctx := context.Background()

	defaultAddress := address.HexToAddress(db.DefaultAccountHex)
	initialDefaultBalance := getAccountBalance(t, defaultAddress)
	assert.Equal(t, initialDefaultBalance, decimal.NewFromInt64(db.DefaultCurrencyAmount))

	recipientAddress := address.HexToAddress("0x1234567890123456789012345678901234567890")
	recipientInitialBalance := getAccountBalance(t, recipientAddress)
	assert.True(t, recipientInitialBalance.IsZero())

	transferAmount := decimal.NewFromInt64(100)

	input := model.Transfer{
		FromAddress: defaultAddress,
		ToAddress:   recipientAddress,
		Amount:      transferAmount,
	}

	// act
	sender, err := mutationResolver.Transfer(ctx, input)

	// assert
	require.NoError(t, err, transferShouldSucceed)
	require.NotNil(t, sender)

	expectedSenderBalance := initialDefaultBalance.Sub(transferAmount)
	assert.Equal(t, expectedSenderBalance, sender.Balance)

	finalSenderBalance := getAccountBalance(t, defaultAddress)
	assert.Equal(t, expectedSenderBalance, finalSenderBalance)

	finalRecipientBalance := getAccountBalance(t, recipientAddress)
	assert.Equal(t, transferAmount, finalRecipientBalance)
}

// TestTransfer_InsufficientBalance tests the case where the sender has insufficient funds.
func TestTransfer_InsufficientBalance(t *testing.T) {
	// assemble
	clearDBState(t)

	mutationResolver := testResolver.Mutation()
	ctx := context.Background()

	defaultAddress := address.HexToAddress(db.DefaultAccountHex)
	initialDefaultBalance := getAccountBalance(t, defaultAddress)

	// Transfer an amount greater than the initial balance
	transferAmount := initialDefaultBalance.Add(decimal.NewFromInt64(1))

	input := model.Transfer{
		FromAddress: defaultAddress,
		ToAddress:   address.HexToAddress("0x1234567890123456789012345678901234567890"),
		Amount:      transferAmount,
	}

	// act
	sender, err := mutationResolver.Transfer(ctx, input)

	// assert
	assert.Error(t, err, transferShouldFail)
	assert.Nil(t, sender)
	assert.IsType(t, eresolvers.InsufficientBalanceError, err)

	finalSenderBalance := getAccountBalance(t, defaultAddress)
	assert.True(t, finalSenderBalance.Equal(initialDefaultBalance))
}

// TestTransfer_NegativeAmount tests transferring a negative amount.
func TestTransfer_NegativeAmount(t *testing.T) {
	// assemble
	clearDBState(t)

	mutationResolver := testResolver.Mutation()
	ctx := context.Background()
	fromAddr := address.HexToAddress(db.DefaultAccountHex)
	toAddr := address.HexToAddress("0x1111111111111111111111111111111111111111")

	testCases := []struct {
		name   string
		amount decimal.Decimal
	}{
		{
			name:   "Negative Amount 1",
			amount: decimal.NewFromInt64(-50),
		},
		{
			name:   "Negative Amount 2",
			amount: decimal.NewFromFloat64(-100_000.88),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := model.Transfer{
				FromAddress: fromAddr,
				ToAddress:   toAddr,
				Amount:      tc.amount,
			}
			// act
			sender, err := mutationResolver.Transfer(ctx, input)

			// assert
			assert.Error(t, err)
			assert.IsType(t, err, eresolvers.NegativeTransferError)
			assert.Equal(t, *sender, model.Sender{Balance: decimal.Zero}, "")
		})
	}
}

// TestTransfer_NonInteger tests transferring a non integer amount.
func TestTransfer_NonInteger(t *testing.T) {
	// assemble
	clearDBState(t)

	mutationResolver := testResolver.Mutation()
	ctx := context.Background()

	testAddress := address.HexToAddress("0x0123456789012345678901234567890123456789")
	transferAmount := decimal.NewFromFloat64(150.5)

	input := model.Transfer{
		FromAddress: address.HexToAddress(db.DefaultAccountHex),
		ToAddress:   testAddress,
		Amount:      transferAmount,
	}

	// act
	sender, err := mutationResolver.Transfer(ctx, input)

	// assert
	require.Error(t, err, transferShouldFail)
	require.IsType(t, err, eresolvers.NonIntegerTransferError)
	require.Equal(t, model.Sender{Balance: decimal.Zero}, *sender)
}

// TestTransfer_SelfTransfer tests transferring to the same address.
func TestTransfer_SelfTransfer(t *testing.T) {
	// assemble
	clearDBState(t)

	mutationResolver := testResolver.Mutation()
	ctx := context.Background()

	testAddress := address.HexToAddress(db.DefaultAccountHex)
	initialBalance := getAccountBalance(t, testAddress)
	transferAmount := decimal.NewFromInt64(100)

	input := model.Transfer{
		FromAddress: testAddress,
		ToAddress:   testAddress,
		Amount:      transferAmount,
	}

	// act
	sender, err := mutationResolver.Transfer(ctx, input)

	// assert
	require.NoError(t, err, transferShouldSucceed)
	require.NotNil(t, sender)

	assert.True(t, sender.Balance.Equal(initialBalance))

	finalBalance := getAccountBalance(t, testAddress)
	assert.True(t, finalBalance.Equal(initialBalance))
}

// TestTransfer_RaceCondition tests concurrent transfers to simulate race conditions.
func TestTransfer_RaceCondition(t *testing.T) {
	// assemble
	clearDBState(t)

	mutationResolver := testResolver.Mutation()
	ctx := context.Background()

	walletAddress := address.HexToAddress(db.DefaultAccountHex)
	err := testDB.Model(&db.Account{}).
		Where("address = ?", walletAddress).
		Update("amount", decimal.NewFromInt64(10)).Error
	require.NoError(t, err, setupFailed)

	currentBalance := getAccountBalance(t, walletAddress)
	assert.Equal(t, decimal.NewFromInt64(10), currentBalance)

	transfers := []struct {
		amount   int64
		fromAddr string
		toAddr   string
	}{
		{amount: 1, fromAddr: "0x1111111111111111111111111111111111111111", toAddr: db.DefaultAccountHex},
		{amount: 4, fromAddr: db.DefaultAccountHex, toAddr: "0x2222222222222222222222222222222222222222"},
		{amount: 7, fromAddr: db.DefaultAccountHex, toAddr: "0x3333333333333333333333333333333333333333"},
	}

	// create the other accounts
	for _, txData := range transfers {
		var otherAddress string
		if txData.fromAddr != db.DefaultAccountHex {
			otherAddress = txData.fromAddr
		} else {
			otherAddress = txData.toAddr
		}
		err := testDB.FirstOrCreate(&db.Account{
			Address: address.FromHex(otherAddress),
			Amount:  decimal.NewFromInt64(txData.amount),
		}).Error
		assert.NoError(t, err, setupFailed)
	}

	// act
	var wg sync.WaitGroup
	results := make(chan error, len(transfers))
	for i, txData := range transfers {
		wg.Add(1)
		go func(idx int, amount int64, fromAddr string, toAddr string) {
			defer wg.Done()
			input := model.Transfer{
				FromAddress: address.FromHex(fromAddr),
				ToAddress:   address.FromHex(toAddr),
				Amount:      decimal.NewFromInt64(amount),
			}

			_, err = mutationResolver.Transfer(ctx, input)
			results <- err
		}(i, txData.amount, txData.fromAddr, txData.toAddr)
	}

	wg.Wait()
	close(results)

	var errorsList []error
	for err := range results {
		if err != nil {
			errorsList = append(errorsList, err)
		}
	}

	// assert
	finalBalanceInDB := getAccountBalance(t, walletAddress)

	possibleFinalBalanceValues := []string{"0", "4", "7"}

	balanceAchieved := false
	for _, expectedValStr := range possibleFinalBalanceValues {
		expectedDec, err := decimal.NewFromString(expectedValStr)
		require.NoError(t, err)
		if finalBalanceInDB.Equal(expectedDec) {
			balanceAchieved = true
			break
		}
	}

	assert.True(t, balanceAchieved, fmt.Sprintf("Final balance %s not among expected outcomes (0, 4, 7)", finalBalanceInDB.String()))

	insufficientBalanceErrorsCount := 0
	for _, err := range errorsList {
		if errors.Is(err, eresolvers.InsufficientBalanceError) {
			insufficientBalanceErrorsCount++
		} else if err != nil {
			t.Errorf("Unexpected error during race condition test: %v", err)
		}
	}
	assert.Contains(t, []int{0, 1}, insufficientBalanceErrorsCount, "Expected 0 or 1 insufficient balance errorsList")
	t.Logf("race test final balance: %s, Insufficient errorsList: %d", finalBalanceInDB.String(), insufficientBalanceErrorsCount)
}

// TestTransfer_SenderNotFound tests transferring from a non-existent sender.
func TestTransfer_SenderNotFound(t *testing.T) {
	// assemble
	clearDBState(t)

	mutationResolver := testResolver.Mutation()
	ctx := context.Background()

	nonExistentAddress := address.HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	toAddress := address.HexToAddress(db.DefaultAccountHex)

	input := model.Transfer{
		FromAddress: nonExistentAddress,
		ToAddress:   toAddress,
		Amount:      decimal.NewFromInt64(100),
	}

	// act
	sender, err := mutationResolver.Transfer(ctx, input)

	// assert
	assert.Error(t, err, transferShouldFail)
	assert.IsType(t, err, eresolvers.AddressNotFoundError{})
	assert.Equal(t, model.Sender{Balance: decimal.Zero}, *sender)
}
