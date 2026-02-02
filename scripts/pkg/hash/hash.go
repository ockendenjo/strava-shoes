package hash

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func GetBase64FromSHA256Hex(hexStr string) string {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

func GetBinarySHA256Hex(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func GetSHA256HexFromBase64(lambdaHash string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(lambdaHash)
	if err != nil {
		return "", err
	}
	hexStr := hex.EncodeToString(b)
	return hexStr, nil
}
