package config

import (
	"github.com/joho/godotenv"
	"os"
	"sync"
)

type Config struct {
	KTMineURL    string
	KTMineAPIKey string
	DBPort       string
	DBUsername   string
	DBPassword   string
	DBHost       string
	SSLMode      string
	DBName       string
	ENV          string
}

var (
	config *Config
	once   sync.Once
)

func LoadConfig() *Config {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			panic("failed to load env variables")
		}

		config = &Config{
			KTMineURL:    os.Getenv("KTMINE_URL"),
			KTMineAPIKey: os.Getenv("KTMINE_API_KEY"),
			DBPort:       os.Getenv("DB_PORT"),
			DBUsername:   os.Getenv("DB_USERNAME"),
			DBPassword:   os.Getenv("DB_PASSWORD"),
			DBHost:       os.Getenv("DB_HOST"),
			SSLMode:      os.Getenv("SSL_MODE"),
			ENV:          os.Getenv("ENV"),
			DBName:       os.Getenv("DB_NAME"),
		}
	})
	return config
}
