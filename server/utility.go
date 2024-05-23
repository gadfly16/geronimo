package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"sync/atomic"

	"golang.org/x/crypto/bcrypt"

	"github.com/dgrijalva/jwt-go"
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
func hashPassword(pw string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pw), 14)
	return string(bytes), err
}

func compareHashPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func generateSecret(l int) ([]byte, error) {
	s := make([]byte, l)
	_, err := rand.Read(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func generateOTP() string {
	otp, _ := generateSecret(16)
	for i, b := range otp {
		otp[i] = b%94 + 33
	}
	return string(otp)
}

func createSecret(path string) error {
	if FileExists(path) {
		return errors.New("key file already exists: " + path)
	}
	secret, err := generateSecret(14)
	if err != nil {
		return err
	}
	return os.WriteFile(path, secret, 0600)
}

func (core *Core) ParseToken(tokenString string) (claims *Claims, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return core.jwtKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, err
	}

	return claims, nil
}

func encryptString(password []byte, salt, message string) (encstr string, err error) {
	plaintext := []byte(message)
	key, err := scrypt.Key(password, []byte(salt), 32768, 8, 1, 32)
	if err != nil {
		return
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Fatal("Couldn't create initial vector.")
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptString(password []byte, salt, message string) (decstr string, err error) {
	ciphertext, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return
	}
	key, err := scrypt.Key(password, []byte(salt), 32768, 8, 1, 32)
	if err != nil {
		return
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	if len(ciphertext) < aes.BlockSize {
		return
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func jsonNumToFloat64(j json.Number) (f float64) {
	f, err := j.Float64()
	if err != nil {
		log.Fatal("Couldn't cost json.Number to float64.")
	}
	return
}

func FileExists(fn string) bool {
	if _, err := os.Stat(fn); err == nil {
		return true
	} else if !errors.Is(err, os.ErrNotExist) {
		log.Fatal("Couldn't stat file: ", fn)
	}
	return false
}
