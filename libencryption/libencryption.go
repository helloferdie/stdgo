package libencryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// Encrypt - encrypt data
func Encrypt(txt string) (string, error) {
	key := []byte(os.Getenv("microservice_secret"))
	bTxt := []byte(txt)
	blk, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	cipherTxt := make([]byte, aes.BlockSize+len(txt))
	iv := cipherTxt[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(blk, iv)
	stream.XORKeyStream(cipherTxt[aes.BlockSize:], bTxt)
	return base64.URLEncoding.EncodeToString(cipherTxt), nil
}

// Decrypt - decrypt data
func Decrypt(txt string) (string, error) {
	key := []byte(os.Getenv("microservice_secret"))
	cipherTxt, _ := base64.URLEncoding.DecodeString(txt)

	blk, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(cipherTxt) < aes.BlockSize {
		return "", fmt.Errorf("%v", "Ciphertext too short")
	}
	iv := cipherTxt[:aes.BlockSize]
	cipherTxt = cipherTxt[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(blk, iv)
	stream.XORKeyStream(cipherTxt, cipherTxt)
	return fmt.Sprintf("%s", cipherTxt), nil
}
