package integration

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"token-transfer-api/internal/db"
)

func TestConnection(t *testing.T) {
	_, err := db.ConnectDb()
	assert.Equal(t, err, nil, "")
}
