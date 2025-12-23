package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port         string `mapstructure:"PORT"`
	DBUrl        string `mapstructure:"DB_DSN"`
	RedisAddr    string `mapstructure:"REDIS_ADDR"`
	RedisPass    string `mapstructure:"REDIS_PASSWORD"`
	RedisDB      int    `mapstructure:"REDIS_DB"`
	KafkaBrokers string `mapstructure:"KAFKA_BROKERS"`
}

var AppConfig *Config

func LoadConfig() {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Println("⚠️  Tidak menemukan file .env, menggunakan environment variables sistem")
	}

	err := viper.Unmarshal(&AppConfig)
	if err != nil {
		log.Fatal("❌ Gagal membaca config:", err)
	}
	
	log.Println("✅ Konfigurasi berhasil dimuat")
}