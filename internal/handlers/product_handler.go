package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eveeze/flash-sale/internal/database"
	"github.com/eveeze/flash-sale/internal/models"
	"github.com/eveeze/flash-sale/pkg/logger" // Import Logger
	"go.uber.org/zap"
)

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		logger.Log.Warn("Gagal decode JSON body", zap.Error(err))
		http.Error(w, "Invalid JSON Body", http.StatusBadRequest)
		return
	}

	// Insert DB
	query := "INSERT INTO products (name, quantity, price) VALUES ($1, $2, $3) RETURNING id"
	err := database.DB.QueryRow(query, p.Name, p.Quantity, p.Price).Scan(&p.ID)
	if err != nil {
		logger.Log.Error("Gagal insert product ke DB", zap.Error(err))
		http.Error(w, "Database Error", http.StatusInternalServerError)
		return
	}

	// Insert Redis
	redisKey := fmt.Sprintf("product_stock:%d", p.ID)
	err = database.Rdb.Set(context.Background(), redisKey, p.Quantity, 0).Err()
	if err != nil {
		logger.Log.Error("Gagal cache ke Redis", zap.String("key", redisKey), zap.Error(err))
	} else {
		logger.Log.Info("Produk baru ditambahkan", 
			zap.Int("product_id", p.ID), 
			zap.String("name", p.Name), 
			zap.Int("stock", p.Quantity))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Barang siap untuk Flash Sale!",
		"data":      p,
		"redis_key": redisKey,
	})
}