package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DbName   string `yaml:"dbname"`
	} `yaml:"database"`
	Backup struct {
		PublicKey   string `yaml:"public_key"`
		Destination string `yaml:"destination"`
		KeepLocal   bool   `yaml:"keep_local"`
	} `yaml:"backup_config"`
	Storage struct {
		Provider  string `yaml:"provider"`
		Bucket    string `yaml:"bucket"`
		Region    string `yaml:"region"`
		AccessKey string `yaml:"access_key"`
		SecretKey string `yaml:"secret_key"`
		Endpoint  string `yaml:"endpoint"`
	} `yaml:"storage"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}
