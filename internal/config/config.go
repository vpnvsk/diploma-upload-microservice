package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"sync"
)

type Config struct {
	KTMineURL           string
	KTMineAPIKey        string
	DBPort              string
	DBUsername          string
	DBPassword          string
	DBHost              string
	SSLMode             string
	DBName              string
	ENV                 string
	BrokerURL           string
	BrokerConsumeQueue  string
	BrokerPublishQueue  string
	BrokerPrefetchCount int
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
		brokerPrefetchCount, err := strconv.Atoi(os.Getenv("BROKER_PREFETCH_COUNT"))
		if err != nil {
			panic("failed to parse config")
		}
		config = &Config{
			KTMineURL:           os.Getenv("KTMINE_URL"),
			KTMineAPIKey:        os.Getenv("KTMINE_API_KEY"),
			DBPort:              os.Getenv("DB_PORT"),
			DBUsername:          os.Getenv("DB_USERNAME"),
			DBPassword:          os.Getenv("DB_PASSWORD"),
			DBHost:              os.Getenv("DB_HOST"),
			SSLMode:             os.Getenv("SSL_MODE"),
			ENV:                 os.Getenv("ENV"),
			DBName:              os.Getenv("DB_NAME"),
			BrokerURL:           os.Getenv("BROKER_URL"),
			BrokerConsumeQueue:  os.Getenv("BROKER_CONSUME_QUEUE"),
			BrokerPublishQueue:  os.Getenv("BROKER_PUBLISH_QUEUE"),
			BrokerPrefetchCount: brokerPrefetchCount,
		}
	})
	return config
}
