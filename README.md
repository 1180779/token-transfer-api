# Token Transfer API

This project is a Go-based GraphQL API designed for transferring BTP tokens between wallets. It is inspired by the functionality of ERC20 token transfers on the blockchain, focusing on data integrity, simplicity, and robust handling of concurrent requests.

The API exposes a single `transfer` mutation to move tokens from one address to another, ensuring that wallet balances never go negative and that race conditions are handled gracefully at the database level.

## ✨ Features

*   **GraphQL API**: A single, clear `transfer` mutation for all token movements.
*   **PostgreSQL Backend**: Uses a PostgreSQL database for persistent storage of wallet balances.
*   **Race Condition Safe**: Implements pessimistic locking within database transactions to ensure atomic and consistent updates during concurrent transfers.
*   **Dockerized Environment**: Fully containerized with Docker and Docker Compose for easy setup, development, and testing.
*   **Data Integrity**: Uses a custom `decimal` type to handle large numerical values with precision, preventing floating-point errors.
*   **Comprehensive Tests**: Includes a suite of tests covering core functionality, edge cases, and race condition scenarios against a real test database.

## 🛠️ Tech Stack

*   **Language**: Go
*   **API Framework**: GraphQL (using [gqlgen](https://github.com/99designs/gqlgen))
*   **Database**: PostgreSQL
*   **ORM**: [GORM](https://gorm.io/)
*   **Containerization**: Docker & Docker Compose

## 📋 Prerequisites

Before you begin, ensure you have the following installed on your system:
*   [Docker](https://www.docker.com/get-started)
*   [Docker Compose](https://docs.docker.com/compose/install/)

## 🚀 Getting Started

Follow these steps to get the application running locally.

### 1. Clone the Repository

```bash
git clone https://github.com/1180779/token-transfer-api
cd token-transfer-api
```

### 2. Configure Environment Variables

Create a `.env` file in the root of the project. This file contains the database credentials for both the main and test databases.

```env
# Main Database
DATABASE_URL=postgres://user:password@db:5432/mydatabase?sslmode=disable
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=mydatabase

# Test Database
DATABASE_TEST_URL=postgres://test_user:test_password@test-db:5432/token_transfer_test_db?sslmode=disable
POSTGRES_TEST_USER=test_user
POSTGRES_TEST_PASSWORD=test_password
POSTGRES_TEST_DB=token_transfer_test_db
```

### 3. Run the Application

Build and start the application and database services using Docker Compose.

```bash
docker-compose up -d --build
```

The API server will be running and accessible. Upon the first run, the application will automatically:
1.  Create the `accounts` table in the database.
2.  Create a default wallet with address `0x0000000000000000000000000000000000000000` holding **1,000,000** BTP tokens.

You can access the GraphQL Playground in your browser to interact with the API:
**[http://localhost:8080](http://localhost:8080)**

## ⚙️ Usage / API Examples

The API has a single `transfer` mutation.

#### Arguments
*   `from_address` (Address!): The wallet address to send tokens from.
*   `to_address` (Address!): The wallet address to send tokens to.
*   `amount` (Decimal!): The amount of tokens to transfer (must be a positive integer string).

#### Returns
*   `balance` (Decimal!): The updated balance of the `from_address` wallet.

---

### Example 1: Successful Transfer

This mutation transfers 1000 tokens from the default wallet to a new wallet.

```graphql
mutation TransferTokens {
  transfer(
    input: {
      from_address: "0x0000000000000000000000000000000000000000"
      to_address: "0x1234567890123456789012345678901234567890"
      amount: "1000"
    }
  ) {
    balance
  }
}
```

**Successful Response:**
```json
{
  "data": {
    "transfer": {
      "balance": "999000"
    }
  }
}
```

### Example 2: Failed Transfer (Insufficient Balance)

This mutation attempts to transfer more tokens than the sender possesses.

```graphql
mutation FailTransfer {
  transfer(
    input: {
      from_address: "0x0000000000000000000000000000000000000000"
      to_address: "0x1234567890123456789012345678901234567890"
      amount: "999999999"
    }
  ) {
    balance
  }
}
```

**Error Response:**
```json
{
  "errors": [
    {
      "message": "insufficient balance",
      "path": [
        "transfer"
      ]
    }
  ],
  "data": null
}
```

Of course! I've formatted the section you provided and expanded it slightly to include the best practices and alternative shell commands we discussed. This will make it much clearer and more helpful for all users.

Here is the formatted and enhanced section, ready to be pasted into your `README.md`:

---

### Manual API Usage with `curl`

You can also send requests directly from your command line using a tool like `curl`.

#### Recommended Method: Using a Data File

The most reliable way to send a `curl` request is to save the JSON payload to a file. This avoids issues with how different command-line shells handle quotes and special characters.

1.  Create a file named `payload.json`:

    ```json
    {
      "query": "mutation TransferTokens($input: Transfer!) { transfer(input: $input) { balance } }",
      "variables": {
        "input": {
          "from_address": "0x0000000000000000000000000000000000000000",
          "to_address": "0xabcdef1234567890abcdef1234567890abcdef12",
          "amount": "500"
        }
      }
    }
    ```

2.  Execute the request from your terminal:

    *   **On Linux, macOS or Git Bash:**
        ```bash
        curl -X POST -H "Content-Type: application/json" --data @payload.json http://localhost:8080/query
        ```

    *   **On Windows PowerShell:**
        ```powershell
        $payload = Get-Content -Raw -Path .\payload.json
        Invoke-WebRequest -Uri http://localhost:8080/query -Method POST -ContentType "application/json" -Body $payload
        ```

---

## ✅ Running Tests

To run the entire test suite, including the race condition tests, use the `tester` service defined in `docker-compose.yml`. This ensures the tests run in an isolated environment with a dedicated test database.

```bash
docker-compose run --rm tester go test -v ./...
```

## 🧠 Implementation Notes

### Race Condition Handling

To handle concurrent transfers safely, this API employs pessimistic locking at the database level.
*   When a transfer is initiated, a transaction is started.
*   The rows for both the sender and receiver accounts are locked using `SELECT ... FOR UPDATE`. This prevents any other transaction from modifying these rows until the current transaction is committed or rolled back.
*   To prevent database deadlocks, the wallet addresses involved in the transaction are sorted alphabetically before their corresponding rows are locked. This ensures a consistent lock acquisition order across all concurrent transactions.

### Data Types
*   **`address.Address`**: A custom type that wraps `Address` from `ethereum/go-ethereum/common` for Ethereum-style addresses to ensure format validation and type safety.
*   **`decimal.Decimal`**: A custom type that wraps `shopspring/decimal` to handle monetary values with arbitrary precision, avoiding floating-point inaccuracies. All transfer amounts are validated to be positive integers.