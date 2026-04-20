package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

type Service struct {
	key []byte
}

var (
	globalService *Service
	once          sync.Once
)

func InitService(secretKey string) {
	once.Do(func() {
		hash := sha256.Sum256([]byte(secretKey))
		globalService = &Service{key: hash[:]}
	})
}

func GetService() *Service {
	return globalService
}

func (s *Service) EncryptAPIKey(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	return Encrypt(plaintext, s.key)
}

func (s *Service) DecryptAPIKey(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	return Decrypt(ciphertext, s.key)
}

func EncryptAPIKey(plaintext string) (string, error) {
	if globalService == nil {
		return plaintext, nil
	}
	return globalService.EncryptAPIKey(plaintext)
}

func DecryptAPIKey(ciphertext string) (string, error) {
	if globalService == nil {
		return ciphertext, nil
	}
	return globalService.DecryptAPIKey(ciphertext)
}

func DeriveKey(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}
