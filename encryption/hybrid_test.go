package encryption

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func writeRSAKeyPair(t *testing.T, dir string) (pubPath, privPath string) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}

	privPath = filepath.Join(dir, "private.pem")
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	if err := os.WriteFile(privPath, privPEM, 0600); err != nil {
		t.Fatalf("write private key: %v", err)
	}

	pubPath = filepath.Join(dir, "public.pem")
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	if err := os.WriteFile(pubPath, pubPEM, 0644); err != nil {
		t.Fatalf("write public key: %v", err)
	}
	return pubPath, privPath
}

func TestAESRoundtrip(t *testing.T) {
	dir := t.TempDir()
	plaintext := []byte("this is a fake postgres dump payload, repeated to compress well. " +
		"this is a fake postgres dump payload, repeated to compress well.")

	src := filepath.Join(dir, "dump.sql")
	if err := os.WriteFile(src, plaintext, 0600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	key, err := GenerateRandomKey(32)
	if err != nil {
		t.Fatalf("GenerateRandomKey: %v", err)
	}

	enc := filepath.Join(dir, "dump.enc")
	if err := EncryptFileWithAES(key, src, enc); err != nil {
		t.Fatalf("EncryptFileWithAES: %v", err)
	}

	dec := filepath.Join(dir, "dump.decrypted.sql")
	if err := DecryptFileWithAES(key, enc, dec); err != nil {
		t.Fatalf("DecryptFileWithAES: %v", err)
	}

	got, err := os.ReadFile(dec)
	if err != nil {
		t.Fatalf("read decrypted: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("roundtrip mismatch:\nwant %q\ngot  %q", plaintext, got)
	}
}

func TestRSAKeyWrapRoundtrip(t *testing.T) {
	dir := t.TempDir()
	pubPath, privPath := writeRSAKeyPair(t, dir)

	aesKey, err := GenerateRandomKey(32)
	if err != nil {
		t.Fatalf("GenerateRandomKey: %v", err)
	}

	wrapped, err := EncryptKeyWithPublicRSA(pubPath, aesKey)
	if err != nil {
		t.Fatalf("EncryptKeyWithPublicRSA: %v", err)
	}

	encKeyFile := filepath.Join(dir, "aes.key")
	if err := os.WriteFile(encKeyFile, wrapped, 0600); err != nil {
		t.Fatalf("write wrapped key: %v", err)
	}

	unwrapped, err := DecryptKeyWithPrivateRSA(privPath, encKeyFile)
	if err != nil {
		t.Fatalf("DecryptKeyWithPrivateRSA: %v", err)
	}
	if !bytes.Equal(unwrapped, aesKey) {
		t.Fatalf("AES key mismatch after RSA roundtrip")
	}
}

func TestDecryptWithWrongKeyFails(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "dump.sql")
	if err := os.WriteFile(src, []byte("secret"), 0600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	good, err := GenerateRandomKey(32)
	if err != nil {
		t.Fatalf("GenerateRandomKey good: %v", err)
	}
	bad, err := GenerateRandomKey(32)
	if err != nil {
		t.Fatalf("GenerateRandomKey bad: %v", err)
	}

	enc := filepath.Join(dir, "dump.enc")
	if err := EncryptFileWithAES(good, src, enc); err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	dec := filepath.Join(dir, "dump.dec")
	if err := DecryptFileWithAES(bad, enc, dec); err == nil {
		t.Fatalf("expected decrypt with wrong key to fail")
	}
}
