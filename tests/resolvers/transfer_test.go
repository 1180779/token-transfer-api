package resolvers

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

type testSuite struct {
	suite.Suite
	mutationResolver graph.MutationResolver
	ctx              context.Context
}

func (suite *testSuite) SetupTest() {
	clearDBState(suite.T())

	suite.mutationResolver = testResolver.Mutation()
	suite.ctx = context.Background()
}

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
func getAccountBalance(suite *testSuite, addr address.Address) decimal.Decimal {
	suite.T().Helper()
	var account = db.Account{Address: addr}
	err := testDB.Where("address = ?", addr).FirstOrCreate(&account).Error
	require.NoError(suite.T(), err, fmt.Sprintf("Failed to get balance for %s", addr.Hex()))
	return account.Amount
}

// TestTransfer_SuccessfulTransfer tests a basic successful transfer.
func (suite *testSuite) TestTransfer_SuccessfulTransfer() {
	// assemble
	defaultAddress := address.HexToAddress(db.DefaultAccountHex)
	initialDefaultBalance := getAccountBalance(suite, defaultAddress)
	assert.Equal(suite.T(), initialDefaultBalance, decimal.NewFromInt64(db.DefaultCurrencyAmount))

	recipientAddress := address.HexToAddress("0x1234567890123456789012345678901234567890")
	recipientInitialBalance := getAccountBalance(suite, recipientAddress)
	assert.True(suite.T(), recipientInitialBalance.IsZero())

	transferAmount := decimal.NewFromInt64(100)

	input := model.Transfer{
		FromAddress: defaultAddress,
		ToAddress:   recipientAddress,
		Amount:      transferAmount,
	}

	// act
	sender, err := suite.mutationResolver.Transfer(suite.ctx, input)

	// assert
	require.NoError(suite.T(), err, transferShouldSucceed)
	require.NotNil(suite.T(), sender)

	expectedSenderBalance := initialDefaultBalance.Sub(transferAmount)
	assert.Equal(suite.T(), expectedSenderBalance, sender.Balance)

	finalSenderBalance := getAccountBalance(suite, defaultAddress)
	assert.Equal(suite.T(), expectedSenderBalance, finalSenderBalance)

	finalRecipientBalance := getAccountBalance(suite, recipientAddress)
	assert.Equal(suite.T(), transferAmount, finalRecipientBalance)
}

// TestTransfer_InsufficientBalance tests the case where the sender has insufficient funds.
func (suite *testSuite) TestTransfer_InsufficientBalance() {
	// assemble
	defaultAddress := address.HexToAddress(db.DefaultAccountHex)
	initialDefaultBalance := getAccountBalance(suite, defaultAddress)

	// Transfer an amount greater than the initial balance
	transferAmount := initialDefaultBalance.Add(decimal.NewFromInt64(1))

	input := model.Transfer{
		FromAddress: defaultAddress,
		ToAddress:   address.HexToAddress("0x1234567890123456789012345678901234567890"),
		Amount:      transferAmount,
	}

	// act
	sender, err := suite.mutationResolver.Transfer(suite.ctx, input)

	// assert
	assert.Error(suite.T(), err, transferShouldFail)
	assert.Nil(suite.T(), sender)
	assert.IsType(suite.T(), eresolvers.InsufficientBalanceError, err)

	finalSenderBalance := getAccountBalance(suite, defaultAddress)
	assert.True(suite.T(), finalSenderBalance.Equal(initialDefaultBalance))
}

// TestTransfer_NegativeAmount tests transferring a negative amount.
func (suite *testSuite) TestTransfer_NegativeAmount() {
	// assemble
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
		suite.T().Run(tc.name, func(t *testing.T) {
			input := model.Transfer{
				FromAddress: fromAddr,
				ToAddress:   toAddr,
				Amount:      tc.amount,
			}
			// act
			sender, err := suite.mutationResolver.Transfer(suite.ctx, input)

			// assert
			assert.Error(t, err)
			assert.IsType(t, eresolvers.NegativeTransferError, err)
			assert.Nil(t, sender)
		})
	}
}

func TestRunAllSuiteTests(t *testing.T) {
	suite.Run(t, new(testSuite))
}

// TestTransfer_NonInteger tests transferring a non integer amount.
func (suite *testSuite) TestTransfer_NonInteger() {
	// assemble
	testAddress := address.HexToAddress("0x0123456789012345678901234567890123456789")
	transferAmount := decimal.NewFromFloat64(150.5)

	input := model.Transfer{
		FromAddress: address.HexToAddress(db.DefaultAccountHex),
		ToAddress:   testAddress,
		Amount:      transferAmount,
	}

	// act
	sender, err := suite.mutationResolver.Transfer(suite.ctx, input)

	// assert
	require.Error(suite.T(), err, transferShouldFail)
	require.IsType(suite.T(), eresolvers.NonIntegerTransferError, err)
	require.Nil(suite.T(), sender)
}

// TestTransfer_SelfTransfer tests transferring to the same address.
func (suite *testSuite) TestTransfer_SelfTransfer() {
	// assemble
	testAddress := address.HexToAddress(db.DefaultAccountHex)
	initialBalance := getAccountBalance(suite, testAddress)
	transferAmount := decimal.NewFromInt64(100)

	input := model.Transfer{
		FromAddress: testAddress,
		ToAddress:   testAddress,
		Amount:      transferAmount,
	}

	// act
	sender, err := suite.mutationResolver.Transfer(suite.ctx, input)

	// assert
	require.NoError(suite.T(), err, transferShouldSucceed)
	require.NotNil(suite.T(), sender)

	assert.True(suite.T(), sender.Balance.Equal(initialBalance))

	finalBalance := getAccountBalance(suite, testAddress)
	assert.True(suite.T(), finalBalance.Equal(initialBalance))
}

// TestTransfer_RaceCondition tests concurrent transfers to simulate race conditions.
func (suite *testSuite) TestTransfer_RaceCondition() {
	// assemble
	walletAddress := address.HexToAddress(db.DefaultAccountHex)
	err := testDB.Model(&db.Account{}).
		Where("address = ?", walletAddress).
		Update("amount", decimal.NewFromInt64(10)).Error
	require.NoError(suite.T(), err, setupFailed)

	currentBalance := getAccountBalance(suite, walletAddress)
	assert.Equal(suite.T(), decimal.NewFromInt64(10), currentBalance)

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
		assert.NoError(suite.T(), err, setupFailed)
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

			_, err = suite.mutationResolver.Transfer(suite.ctx, input)
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
	finalBalanceInDB := getAccountBalance(suite, walletAddress)

	possibleFinalBalanceValues := []string{"0", "4", "7"}

	balanceAchieved := false
	for _, expectedValStr := range possibleFinalBalanceValues {
		expectedDec, err := decimal.NewFromString(expectedValStr)
		require.NoError(suite.T(), err)
		if finalBalanceInDB.Equal(expectedDec) {
			balanceAchieved = true
			break
		}
	}

	assert.True(suite.T(), balanceAchieved, fmt.Sprintf("Final balance %s not among expected outcomes (0, 4, 7)", finalBalanceInDB.String()))

	insufficientBalanceErrorsCount := 0
	for _, err := range errorsList {
		if errors.Is(err, eresolvers.InsufficientBalanceError) {
			insufficientBalanceErrorsCount++
		} else if err != nil {
			suite.T().Errorf("Unexpected error during race condition test: %v", err)
		}
	}
	assert.Contains(suite.T(), []int{0, 1}, insufficientBalanceErrorsCount, "Expected 0 or 1 insufficient balance errorsList")
	suite.T().Logf("race test final balance: %s, Insufficient errorsList: %d", finalBalanceInDB.String(), insufficientBalanceErrorsCount)
}

// TestTransfer_SenderNotFound tests transferring from a non-existent sender.
func (suite *testSuite) TestTransfer_SenderNotFound() {
	// assemble
	nonExistentAddress := address.HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	toAddress := address.HexToAddress(db.DefaultAccountHex)

	input := model.Transfer{
		FromAddress: nonExistentAddress,
		ToAddress:   toAddress,
		Amount:      decimal.NewFromInt64(100),
	}

	// act
	sender, err := suite.mutationResolver.Transfer(suite.ctx, input)

	// assert
	assert.Error(suite.T(), err, transferShouldFail)
	assert.IsType(suite.T(), err, eresolvers.AddressNotFoundError{})
	assert.Nil(suite.T(), sender)
}
