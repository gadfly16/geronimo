package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"math"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/scrypt"
)

// Atomic counter for messages
var globalIDCounter *int64 = new(int64)

func nextID() int64 {
	return atomic.AddInt64(globalIDCounter, 1)
}

// Math
func fitTo01(val, low, high float64) float64 {
	return (val - low) / (high - low)
}

func fit01(val, low, high float64) float64 {
	return val*(high-low) + low
}

func clamp01(val float64) float64 {
	return math.Min(1, math.Max(0, val))
}

// Encryption
var hash256 = sha256.New()

func HashPassword(pw string) string {
	hash256.Write([]byte(pw))
	return base64.StdEncoding.EncodeToString(hash256.Sum(nil))
}

func EncryptString(password, salt, message string) string {
	plaintext := []byte(message)
	key, err := scrypt.Key([]byte(password), []byte(salt), 32768, 8, 1, 32)
	if err != nil {
		log.Fatal("Couldn't create key from password.")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal("Couldn't create cipher block.")
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Fatal("Couldn't create initial vector.")
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext)
}

func decryptString(password, salt, message string) string {
	ciphertext, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		log.Fatal("Couldn't decode encoded message.")
	}
	key, err := scrypt.Key([]byte(password), []byte(salt), 32768, 8, 1, 32)
	if err != nil {
		log.Fatal("Couldn't create key from password.")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal("Couldn't create cipher block.")
	}
	if len(ciphertext) < aes.BlockSize {
		log.Fatal("Ciphertext too short.")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext)
}

func jsonNumToFloat64(j json.Number) (f float64) {
	f, err := j.Float64()
	if err != nil {
		log.Fatal("Couldn't cost json.Number to float64.")
	}
	return
}
