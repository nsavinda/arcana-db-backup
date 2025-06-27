package encryption

import (
	"compress/gzip"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
)

// GenerateRandomKey generates a random key of specified length.
func GenerateRandomKey(length int) ([]byte, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	return key, err
}

// EncryptKeyWithPublicRSA encrypts the AES key with the provided public RSA key file.
func EncryptKeyWithPublicRSA(pubKeyPath string, key []byte) ([]byte, error) {
	pubKeyBytes, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pubKeyBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return rsa.EncryptOAEP(sha256HashFunc(), rand.Reader, rsaPub, key, nil)
}

// DecryptKeyWithPrivateRSA decrypts the encrypted key file using a private RSA key (PKCS#1 or PKCS#8).
func DecryptKeyWithPrivateRSA(privateKeyPath, encKeyFile string) ([]byte, error) {
	encryptedKey, err := os.ReadFile(encKeyFile)
	if err != nil {
		return nil, err
	}
	privKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(privKeyBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	var privKey *rsa.PrivateKey

	switch block.Type {
	case "RSA PRIVATE KEY": // PKCS#1
		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	case "PRIVATE KEY": // PKCS#8
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		privKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
	default:
		return nil, fmt.Errorf("unsupported key type: %s", block.Type)
	}

	return rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, encryptedKey, nil)
}

// EncryptFileWithAES compresses then encrypts inputPath and writes to outputPath.
func EncryptFileWithAES(aesKey []byte, inputPath, outputPath string) error {
	gzPath := inputPath + ".gz"
	err := CompressFile(inputPath, gzPath)
	if err != nil {
		return err
	}
	defer os.Remove(gzPath)

	plaintext, err := os.ReadFile(gzPath)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)
	return os.WriteFile(outputPath, ciphertext, 0644)
}

// DecryptFileWithAES decrypts and decompresses the file, writing the plaintext to outputPath.
func DecryptFileWithAES(aesKey []byte, inputPath, outputPath string) error {
	tmpGz := outputPath + ".gz"

	ciphertext, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return errors.New("ciphertext too short")
	}
	nonce, ciphertextData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return err
	}
	err = os.WriteFile(tmpGz, plaintext, 0644)
	if err != nil {
		return err
	}
	defer os.Remove(tmpGz)

	return DecompressFile(tmpGz, outputPath)
}

// CompressFile compresses inputPath into outputPath as gzip.
func CompressFile(inputPath, outputPath string) error {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	_, err = io.Copy(gw, inFile)
	return err
}

// DecompressFile decompresses a gzip file at inputPath to outputPath.
func DecompressFile(inputPath, outputPath string) error {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	gr, err := gzip.NewReader(inFile)
	if err != nil {
		return err
	}
	defer gr.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, gr)
	return err
}

// Helper for consistent OAEP hash
func sha256HashFunc() hash.Hash {
	return crypto.SHA256.New()
}
