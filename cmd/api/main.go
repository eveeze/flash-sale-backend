package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eveeze/flash-sale/internal/config"
	"github.com/eveeze/flash-sale/internal/database"
	"github.com/eveeze/flash-sale/internal/handlers"
	"github.com/eveeze/flash-sale/internal/kafka"
	"github.com/eveeze/flash-sale/pkg/logger" // Import Logger
	"go.uber.org/zap"
)

func main() {
	// 1. Init Logger & Config
	logger.InitLogger()
	defer logger.Log.Sync() // Flush log sebelum mati

	config.LoadConfig() // Load .env

	logger.Log.Info("🔥 Menyalakan Flash Sale Engine...")

	// 2. Init Infrastructure
	database.ConnectDB()
	database.Migrate()
	kafka.InitProducer()

	// 3. Setup Router
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "OK", "message": "Server siap tempur!"}`))
	})
	mux.HandleFunc("POST /products", handlers.CreateProduct)
	mux.HandleFunc("POST /purchase", handlers.PurchaseProduct)

	// 4. Server Setup
	port := config.AppConfig.Port
	server := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 5. Jalankan Server di Background (Goroutine)
	go func() {
		logger.Log.Info("🚀 Server berjalan", zap.String("address", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("❌ Terjadi error fatal server", zap.Error(err))
		}
	}()

	// --- GRACEFUL SHUTDOWN ---
	// Tunggu sinyal matikan (Ctrl+C atau SIGTERM dari Kubernetes)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	<-quit // Block di sini sampai sinyal diterima
	logger.Log.Warn("⚠️  Sinyal shutdown diterima, menutup server...")

	// Beri waktu 5 detik untuk selesaikan request yang sedang berjalan
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Fatal("❌ Server dipaksa mati", zap.Error(err))
	}

	logger.Log.Info("✅ Server mati dengan aman (Graceful Shutdown)")
}