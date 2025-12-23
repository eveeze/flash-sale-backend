package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eveeze/flash-sale/internal/database"
	"github.com/eveeze/flash-sale/internal/handlers"
	"github.com/eveeze/flash-sale/internal/kafka" // <--- Import baru
)

func main() {
	fmt.Println("🔥 Menyalakan Flash Sale Engine...")

	// 1. Database & Redis
	database.ConnectDB()
	database.Migrate()

	// 2. Kafka Producer (PENTING: Jangan lupa ini!)
	kafka.InitProducer()

	// 3. Router
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "OK", "message": "Server siap tempur!"}`))
	})

	mux.HandleFunc("POST /products", handlers.CreateProduct)
	
	// --- Route Baru: BUY ---
	mux.HandleFunc("POST /purchase", handlers.PurchaseProduct)

	// 4. Server Config
	port := ":8080"
	server := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Println("🚀 Server berjalan di http://localhost" + port)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("❌ Terjadi error saat menjalankan server:", err)
	}
}