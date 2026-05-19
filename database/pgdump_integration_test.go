//go:build integration

package database_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"arcana-db-backup/database"
	"arcana-db-backup/encryption"
)

// dbConfigFromEnv reads connection settings from env vars set by CI / the
// developer's local environment. Required: PGHOST, PGPORT, PGUSER, PGPASSWORD,
// PGDATABASE.
func dbConfigFromEnv(t *testing.T) database.DBConfig {
	t.Helper()
	host := os.Getenv("PGHOST")
	portStr := os.Getenv("PGPORT")
	user := os.Getenv("PGUSER")
	pass := os.Getenv("PGPASSWORD")
	name := os.Getenv("PGDATABASE")
	if host == "" || portStr == "" || user == "" || name == "" {
		t.Skip("integration test requires PGHOST/PGPORT/PGUSER/PGPASSWORD/PGDATABASE env vars")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("invalid PGPORT: %v", err)
	}
	return database.DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: pass,
		DbName:   name,
	}
}

// TestBackupEncryptDecryptRoundtrip exercises Step 1: pg_dump → AES encrypt →
// AES decrypt and verifies the decrypted bytes match the original dump exactly.
func TestBackupEncryptDecryptRoundtrip(t *testing.T) {
	cfg := dbConfigFromEnv(t)
	dir := t.TempDir()

	dumpFile := filepath.Join(dir, "dump.sql")
	if err := database.Dump(cfg, dumpFile); err != nil {
		t.Fatalf("pg_dump failed: %v", err)
	}

	original, err := os.ReadFile(dumpFile)
	if err != nil {
		t.Fatalf("read dump: %v", err)
	}
	if len(original) == 0 {
		t.Fatalf("dump file is empty")
	}

	aesKey, err := encryption.GenerateRandomKey(32)
	if err != nil {
		t.Fatalf("GenerateRandomKey: %v", err)
	}

	encFile := dumpFile + ".enc"
	if err := encryption.EncryptFileWithAES(aesKey, dumpFile, encFile); err != nil {
		t.Fatalf("EncryptFileWithAES: %v", err)
	}

	decFile := filepath.Join(dir, "dump.decrypted.sql")
	if err := encryption.DecryptFileWithAES(aesKey, encFile, decFile); err != nil {
		t.Fatalf("DecryptFileWithAES: %v", err)
	}

	got, err := os.ReadFile(decFile)
	if err != nil {
		t.Fatalf("read decrypted: %v", err)
	}
	if !bytes.Equal(original, got) {
		t.Fatalf("decrypted dump does not match original (orig=%d bytes, got=%d bytes)",
			len(original), len(got))
	}
}
