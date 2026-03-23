package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

const aesGCMCiphertextPrefix = "v1:"

func deriveAES256Key() []byte {
	sum := sha256.Sum256([]byte(global.GVA_CONFIG.JWT.SigningKey))
	return sum[:]
}

func EncryptStringAESGCM(plaintext string) (string, error) {
	key := deriveAES256Key()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	raw := append(nonce, ciphertext...)
	return aesGCMCiphertextPrefix + base64.StdEncoding.EncodeToString(raw), nil
}

func DecryptStringAESGCM(enc string) (string, error) {
	if !strings.HasPrefix(enc, aesGCMCiphertextPrefix) {
		return "", errors.New("unsupported ciphertext format")
	}
	rawB64 := strings.TrimPrefix(enc, aesGCMCiphertextPrefix)
	raw, err := base64.StdEncoding.DecodeString(rawB64)
	if err != nil {
		return "", err
	}
	key := deriveAES256Key()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce := raw[:nonceSize]
	ciphertext := raw[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
