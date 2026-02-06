package crypto2

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

// GenerateEncryptKey derives an encryption key using scrypt and SHA-256
func GenerateEncryptKey(data []byte) []byte {
	sk, err := scrypt.Key(data, nil, 32768, 8, 1, 32)
	if err != nil {
		panic(err)
	}

	sum := sha256.Sum256(append(data, sk...))
	return sum[:]
}

// GenerateEncryptKeyPBKDF2 derives an encryption key using PBKDF2 and SHA-256
func GenerateEncryptKeyPBKDF2(password, salt []byte, iterations int) []byte {
	return pbkdf2.Key(password, salt, iterations, 32, sha256.New)
}

// Hash256 computes SHA-256 hash of data
func Hash256(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

// Encrypt encrypts data using AES-CTR mode
func Encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], data)

	return ciphertext, nil
}

// Decrypt decrypts data using AES-CTR mode
func Decrypt(encryptedData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	data := make([]byte, len(encryptedData[aes.BlockSize:]))

	iv := encryptedData[:aes.BlockSize]
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(data, encryptedData[aes.BlockSize:])

	return data, nil
}
