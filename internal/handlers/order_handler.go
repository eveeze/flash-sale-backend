package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eveeze/flash-sale/internal/database"
	"github.com/eveeze/flash-sale/internal/kafka"
	"github.com/eveeze/flash-sale/internal/models"
	"github.com/eveeze/flash-sale/pkg/logger" // Import Logger
	"go.uber.org/zap"
)

type PurchaseRequest struct {
	UserID    int `json:"user_id"`
	ProductID int `json:"product_id"`
}

func PurchaseProduct(w http.ResponseWriter, r *http.Request) {
	// 1. Decode
	var req PurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Warn("Invalid purchase request body", zap.Error(err))
		http.Error(w, "Invalid Format", http.StatusBadRequest)
		return
	}

	// 2. Redis Decr
	redisKey := fmt.Sprintf("product_stock:%d", req.ProductID)
	sisaStok, err := database.Rdb.Decr(r.Context(), redisKey).Result()
	if err != nil {
		logger.Log.Error("Redis connection failed", zap.Error(err))
		http.Error(w, "System Error", http.StatusInternalServerError)
		return
	}

	// 3. Validasi Stok
	if sisaStok < 0 {
		database.Rdb.Incr(r.Context(), redisKey) // Rollback
		
		// Log Info (bukan Error) karena stok habis itu wajar bisnis
		logger.Log.Info("Stok Habis (Sold Out)", 
			zap.Int("user_id", req.UserID),
			zap.Int("product_id", req.ProductID))
			
		http.Error(w, "Yah, Stok Habis!", http.StatusConflict)
		return
	}

	// 4. Publish Kafka
	orderEvent := models.Order{
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	err = kafka.PublishOrder("orders", orderEvent)
	if err != nil {
		database.Rdb.Incr(r.Context(), redisKey) // Kompensasi Redis
		logger.Log.Error("Gagal publish ke Kafka", zap.Error(err))
		http.Error(w, "Order processing failed", http.StatusInternalServerError)
		return
	}

	// Log Sukses (Debug level biar gak spam kalau di production, atau Info kalau butuh audit)
	logger.Log.Info("Order masuk antrian", 
		zap.Int("user_id", req.UserID), 
		zap.Int("product_id", req.ProductID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":            "Pesanan diterima! Sedang diproses.",
		"sisa_stok_estimasi": sisaStok,
		"status":             "queued",
	})
}