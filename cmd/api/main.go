package main

import (
	"fmt"
	"net/http"
	"time"

	// Pastikan path ini sesuai dengan nama module di go.mod kamu
	"github.com/eveeze/flash-sale/internal/database"
	"github.com/eveeze/flash-sale/internal/handlers"
)

func main() {
	fmt.Println("🔥 Menyalakan Flash Sale Engine...")

	// 1. Inisialisasi Koneksi (Postgres & Redis)
	database.ConnectDB()

	// 2. Jalankan Auto Migration (Buat tabel otomatis)
	database.Migrate()

	// 3. Setup Router (Menggunakan Standard Library Go 1.22+)
	mux := http.NewServeMux()

	// --- DAFTAR ROUTE ---

	// Route 1: Health Check (Cek server hidup/mati)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "OK", "message": "Server siap tempur!"}`))
	})

	// Route 2: Input Barang (Admin) -> Masuk ke DB & Redis
	mux.HandleFunc("POST /products", handlers.CreateProduct)

	// --------------------

	// 4. Konfigurasi Server (Production Ready)
	// Kita menggunakan struct http.Server manual agar bisa mengatur Timeout.
	// Ini PENTING untuk mencegah serangan "Slowloris" atau koneksi gantung saat high traffic.
	port := ":8080"
	server := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,  // Batas waktu baca request user
		WriteTimeout: 10 * time.Second,  // Batas waktu kirim respon ke user
		IdleTimeout:  120 * time.Second, // Batas waktu koneksi nganggur
	}

	fmt.Println("🚀 Server berjalan di http://localhost" + port)

	// 5. Jalankan Server
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("❌ Terjadi error saat menjalankan server:", err)
	}
}
