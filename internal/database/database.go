package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/eveeze/flash-sale/internal/config"
	_ "github.com/lib/pq" // Driver Postgres wajib di-import begini
	"github.com/redis/go-redis/v9"
)

// Variabel Global supaya bisa diakses dari mana saja
var (
	DB  *sql.DB
	Rdb *redis.Client
)

func ConnectDB() {
	// --- 1. POSTGRES ---
	// Gunakan config dari viper
	var err error
	DB, err = sql.Open("postgres", config.AppConfig.DBUrl)
	if err != nil {
		log.Fatal("❌ Error konfigurasi Postgres:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("❌ Gagal connect ke Postgres:", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)
	fmt.Println("✅ Sukses connect ke PostgreSQL")

	// --- 2. REDIS ---
	Rdb = redis.NewClient(&redis.Options{
		Addr:     config.AppConfig.RedisAddr, // Ambil dari Config
		Password: config.AppConfig.RedisPass,
		DB:       config.AppConfig.RedisDB,
	})

	if err := Rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("❌ Gagal connect ke Redis:", err)
	}
	fmt.Println("✅ Sukses connect ke Redis")
}
func Migrate() {
	fmt.Println("sedang melkaukan migrasi database")

	queryProduct := `
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			quantity INT NOT NULL CHECK (quantity >= 0),
			price DECIMAL(10, 2) NOT NULL
	);
	`

	_, err := DB.Exec(queryProduct)
	if err != nil {
		log.Fatal("Gagal membaut tabel product", err)
	}

	queryOrder := `
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		product_id INT NOT NULL,
		status VARCHAR(20) DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (product_id) REFERENCES products(id)
	);`

	_, err = DB.Exec(queryOrder)
	if err != nil {
		log.Fatal("❌ Gagal membuat tabel orders:", err)
	}

	fmt.Println("✨ Tabel 'products' dan 'orders' siap digunakan!")
}
