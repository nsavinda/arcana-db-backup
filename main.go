package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"arcana-db-backup/config"
	"arcana-db-backup/database"
	"arcana-db-backup/encryption"
	"arcana-db-backup/storage"
	"time"
)

func encryptMode() {
	// 1. Load config
	cfg, err := config.LoadConfig("/etc/arcanadbbackup/config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// 2. Dump the database
	dumpFile := fmt.Sprintf("%s/%s_%s.sql", cfg.Backup.Destination, cfg.Database.DbName, time.Now().Format("20060102_150405"))
	dbCfg := database.DBConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DbName:   cfg.Database.DbName,
	}

	if _, err := os.Stat(cfg.Backup.Destination); os.IsNotExist(err) {
		if err := os.MkdirAll(cfg.Backup.Destination, 0755); err != nil {
			log.Fatalf("Failed to create backup directory: %v", err)
		}
	}

	if err := database.Dump(dbCfg, dumpFile); err != nil {
		log.Fatalf("Database dump failed: %v", err)
	}
	defer os.Remove(dumpFile)

	// 3. Generate random AES key
	aesKey, err := encryption.GenerateRandomKey(32) // 32 bytes for AES-256
	if err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}

	// 4. Encrypt the dump file with the AES key
	encFile := dumpFile + ".enc"
	if err := encryption.EncryptFileWithAES(aesKey, dumpFile, encFile); err != nil {
		log.Fatalf("Failed to encrypt dump file: %v", err)
	}

	// 5. Encrypt the AES key with the public RSA key
	pubKeyPath := cfg.Backup.PublicKey
	if pubKeyPath == "" {
		pubKeyPath = "public.pem" // Default public key file
	}
	if _, err := os.Stat(pubKeyPath); os.IsNotExist(err) {
		log.Fatalf("Public key file not found: %s", pubKeyPath)
	}

	encryptedKey, err := encryption.EncryptKeyWithPublicRSA(pubKeyPath, aesKey)
	if err != nil {
		log.Fatalf("Failed to encrypt AES key: %v", err)
	}

	// 6. Write the encrypted AES key to a file
	keyFile := encFile + ".key"
	if err := os.WriteFile(keyFile, encryptedKey, 0600); err != nil {
		log.Fatalf("Failed to write encrypted key file: %v", err)
	}

	//  upload if storage is configured
	if cfg.Storage.Provider == "" {
		fmt.Println("No storage provider configured, skipping upload.")
		return
	}

	s3Cfg := storage.S3Config{
		Bucket:    cfg.Storage.Bucket,
		Region:    cfg.Storage.Region,
		AccessKey: cfg.Storage.AccessKey,
		SecretKey: cfg.Storage.SecretKey,
		Endpoint:  cfg.Storage.Endpoint,
	}

	if err := storage.Upload(s3Cfg, encFile); err != nil {
		log.Fatalf("Upload of encrypted backup failed: %v", err)
	} else {
		fmt.Printf("Successfully uploaded encrypted backup: %s\n", encFile)
	}
	if err := storage.Upload(s3Cfg, keyFile); err != nil {
		log.Fatalf("Upload of encrypted key failed: %v", err)
	} else {
		fmt.Printf("Successfully uploaded encrypted key: %s\n", keyFile)
	}

	// upload status print

	fmt.Println("Encrypted backup and key generated successfully:")
	fmt.Println("  Encrypted dump:", encFile)
	fmt.Println("  Encrypted key: ", keyFile)
	fmt.Println("To decrypt: ")
	fmt.Printf("  %s decrypt -i <privatekeyfile> %s\n", os.Args[0], encFile)

	if !cfg.Backup.KeepLocal {
		if err := os.Remove(encFile); err != nil {
			log.Printf("Failed to remove local encrypted dump file: %v", err)
		} else {
			fmt.Println("Removed local encrypted dump file:", encFile)
		}
		if err := os.Remove(keyFile); err != nil {
			log.Printf("Failed to remove local encrypted key file: %v", err)
		} else {
			fmt.Println("Removed local encrypted key file:", keyFile)
		}
	} else {
		fmt.Println("Local copies retained as per configuration.")
	}
}

func decryptMode(args []string) {
	// flags for decrypt
	fs := flag.NewFlagSet("decrypt", flag.ExitOnError)
	privateKeyPath := fs.String("i", "", "Private RSA key file for decrypting AES key")
	fs.Parse(args)
	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s decrypt -i <privatekeyfile> <EncryptedFile>\n", os.Args[0])
		os.Exit(1)
	}
	if *privateKeyPath == "" {
		fmt.Fprintln(os.Stderr, "Missing -i <privatekeyfile>")
		os.Exit(1)
	}

	encFile := rest[0]
	encKeyFile := encFile + ".key"
	if _, err := os.Stat(encKeyFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Encrypted key file not found: %s\n", encKeyFile)
		os.Exit(1)
	}

	// 1. Decrypt the AES key
	aesKey, err := encryption.DecryptKeyWithPrivateRSA(*privateKeyPath, encKeyFile)
	if err != nil {
		log.Fatalf("Failed to decrypt AES key: %v", err)
	}

	// 2. Decrypt the encrypted file
	outFile := strings.TrimSuffix(encFile, ".enc") + ".decrypted.sql"
	if err := encryption.DecryptFileWithAES(aesKey, encFile, outFile); err != nil {
		log.Fatalf("Failed to decrypt file: %v", err)
	}

	fmt.Println("Decryption successful.")
	fmt.Println("  Decrypted file:", outFile)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "decrypt" {
		decryptMode(os.Args[2:])
	} else {
		encryptMode()
	}
}
