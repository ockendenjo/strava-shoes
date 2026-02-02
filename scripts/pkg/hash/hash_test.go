package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetBash64FromSHA256Hex(t *testing.T) {
	hexStr := "562f780d4508a4d7455cbf88c8b3aa269ec947227d0702f8bff1b01095b63f94"
	b64Hash := GetBase64FromSHA256Hex(hexStr)

	exp := "Vi94DUUIpNdFXL+IyLOqJp7JRyJ9BwL4v/GwEJW2P5Q="
	assert.Equal(t, exp, b64Hash)
}

func Test_GetSHA256HexFromBase64(t *testing.T) {
	lambdaHash := "Vi94DUUIpNdFXL+IyLOqJp7JRyJ9BwL4v/GwEJW2P5Q="
	hexStr, err := GetSHA256HexFromBase64(lambdaHash)
	require.NoError(t, err)

	exp := "562f780d4508a4d7455cbf88c8b3aa269ec947227d0702f8bff1b01095b63f94"
	assert.Equal(t, exp, hexStr)
}
