package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"devops-console-backend/internal/common"
)

func EncryptString(plainText string) (string, error) {
	if plainText == "" {
		return "", nil
	}

	block, err := aes.NewCipher(secretKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func DecryptString(cipherText string) (string, error) {
	if cipherText == "" {
		return "", nil
	}

	raw, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(secretKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("invalid encrypted payload")
	}

	nonce, data := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plainText, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func secretKey() []byte {
	seed := "kafka-console-secret"
	if cfg := common.GetGlobalConfig(); cfg != nil && cfg.Jwt != nil && cfg.Jwt.Secret != "" {
		seed = cfg.Jwt.Secret
	}
	result := sha256.Sum256([]byte(seed))
	return result[:]
}
