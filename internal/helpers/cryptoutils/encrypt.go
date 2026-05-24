package cryptoutils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

const (
	privateKeyType    = "PRIVATE KEY"
	rsaPrivateKeyType = "RSA PRIVATE KEY"
)

type envelope struct {
	Key   []byte `json:"key"`
	Nonce []byte `json:"nonce"`
	Data  []byte `json:"data"`
}

func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("decode public key PEM")
	}

	ttt, errParse := x509.ParseCertificate(block.Bytes)
	if errParse != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return ttt.PublicKey.(*rsa.PublicKey), nil
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("decode private key PEM")
	}

	switch block.Type {
	case privateKeyType:
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}

		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA")
		}

		return rsaKey, nil

	case rsaPrivateKeyType:
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse RSA private key: %w", err)
		}

		return key, nil

	default:
		return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
	}
}

func Encrypt(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("generate AES key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	encryptedData := gcm.Seal(nil, nonce, data, nil)

	encryptedKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		aesKey,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("encrypt AES key: %w", err)
	}

	out, err := json.Marshal(envelope{
		Key:   encryptedKey,
		Nonce: nonce,
		Data:  encryptedData,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal encrypted envelope: %w", err)
	}

	return out, nil
}

func Decrypt(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("unmarshal encrypted envelope: %w", err)
	}

	aesKey, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		env.Key,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("decrypt AES key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	plainData, err := gcm.Open(nil, env.Nonce, env.Data, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt data: %w", err)
	}

	return plainData, nil
}
