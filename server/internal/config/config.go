package config

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

var (
	config *Config
	once   sync.Once
)

type Config struct {
	Port        int    `mapstructure:"APP_PORT"`
	Environment string `mapstructure:"APP_ENV"`
	AppName     string `mapstructure:"APP_NAME"`
	Debug       bool   `mapstructure:"DEBUG"`

	DBUsername string `mapstructure:"DB_USERNAME"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBDatabase string `mapstructure:"DB_DATABASE"`

	StorageURL        string `mapstructure:"STORAGE_URL"`
	StorageAccessKey  string `mapstructure:"STORAGE_ACCESS_KEY"`
	StorageSecretKey  string `mapstructure:"STORAGE_SECRET_KEY"`
	StorageBucketName string `mapstructure:"STORAGE_BUCKET_NAME"`

	JWTAccessSecret     string `mapstructure:"JWT_ACCESS_SECRET"`
	JWTRefreshSecret    string `mapstructure:"JWT_REFRESH_SECRET"`
	JWTAccessExpMinutes int16  `mapstructure:"JWT_ACCESS_EXP_MINUTES"`
	JWTRefreshExpHours  int16  `mapstructure:"JWT_REFRESH_EXP_HOURS"`
}

func GetConfig() *Config {
	once.Do(func() {
		viper.SetConfigName(".env")
		viper.SetConfigType("env")
		viper.AddConfigPath(".")
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("error reading config file, %s", err)
		}

		if err := viper.Unmarshal(&config); err != nil {
			log.Fatalf("error unmarshalling config, %s", err)
		}
	})

	return config
}
