package database

import (
	"fmt"
	"os"
	"os/exec"
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DbName   string
}

// Dump creates a PostgreSQL dump file
func Dump(cfg DBConfig, outFile string) error {
	cmd := exec.Command(
		"pg_dump",
		"-h", cfg.Host,
		"-p", fmt.Sprintf("%d", cfg.Port),
		"-U", cfg.User,
		"-F", "c", // custom format
		"-f", outFile,
		cfg.DbName,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.Password)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
