package crypto2

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/scrypt"
)

// 加密参数常量
const (
	// Scrypt 参数 (N=2^17, r=8, p=1) - 高安全性配置
	ScryptN      = 1 << 17 // 131072
	ScryptR      = 8
	ScryptP      = 1
	ScryptKeyLen = 32

	// Argon2id 参数
	Argon2Time    = 3
	Argon2Memory  = 64 * 1024 // 64 MB
	Argon2Threads = 4
	Argon2KeyLen  = 32
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrDecryptionFailed  = errors.New("decryption failed: authentication error")
)

// GenerateEncryptKey derives an encryption key using Scrypt + Argon2id
// 双重密钥派生：Scrypt 抗 ASIC，Argon2id 抗 GPU
func GenerateEncryptKey(password, salt []byte) ([]byte, error) {
	// 第一层：Scrypt 派生
	scryptKey, err := scrypt.Key(password, salt, ScryptN, ScryptR, ScryptP, ScryptKeyLen)
	if err != nil {
		return nil, err
	}
	// 第二层：Argon2id 派生
	return argon2.IDKey(scryptKey, salt, Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLen), nil
}

// Hash256 computes SHA-256 hash of data
func Hash256(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

// EncryptGCM encrypts data using AES-256-GCM (authenticated encryption)
// Returns: nonce (12 bytes) + ciphertext + tag (16 bytes)
func EncryptGCM(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptGCM decrypts data using AES-256-GCM (authenticated encryption)
func DecryptGCM(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, ErrInvalidCiphertext
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}
