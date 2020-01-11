package pwd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

type AES struct{}

func NewAES() *AES {
	return &AES{}
}

var (
	key = []byte("8RcY34!7dce3,cdcaeb*faeC3cd9fQfe")
)

func (*AES) Encrypter(password string) (string, error) {
	if len(key) != 32 {
		return "", errors.New("The length of the key must be 32.")
	}

	plaintext := []byte(password)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return hex.EncodeToString(ciphertext), nil
}

func (*AES) Decrypter(code string) (string, error) {
	ciphertext, err := hex.DecodeString(code)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("Ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}
