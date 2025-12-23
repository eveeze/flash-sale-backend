package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eveeze/flash-sale/internal/database"
	"github.com/eveeze/flash-sale/internal/kafka"
	"github.com/eveeze/flash-sale/internal/models"
)

// Request body dari user saat beli
type PurchaseRequest struct {
	UserID    int `json:"user_id"`
	ProductID int `json:"product_id"`
}

func PurchaseProduct(w http.ResponseWriter, r *http.Request) {
	// 1. Decode Request
	var req PurchaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Format", http.StatusBadRequest)
		return
	}

	// 2. CEK STOK DI REDIS (Atomic Decrement)
	// Pastikan key-nya SAMA PERSIS dengan yang kamu buat di product_handler
	// Tadi di output kamu: "product_stock:1" (pakai c)
	redisKey := fmt.Sprintf("product_stock:%d", req.ProductID)

	// Kurangi stok -1. Ini operasi atomik (aman dari race condition)
	sisaStok, err := database.Rdb.Decr(r.Context(), redisKey).Result()
	if err != nil {
		http.Error(w, "System Error: Redis Down", http.StatusInternalServerError)
		return
	}

	// 3. Validasi Stok
	if sisaStok < 0 {
		// Kalau hasil minus, berarti stok habis.
		// Kembalikan angkanya biar gak minus terus (Incr)
		database.Rdb.Incr(r.Context(), redisKey)
		
		http.Error(w, "Yah, Stok Habis! Kalah cepat :(", http.StatusConflict)
		return
	}

	// 4. Stok Aman? KIRIM KE KAFKA!
	// Kita bikin struct order sementara buat dikirim
	orderEvent := models.Order{
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	// Kirim ke topic "orders"
	err = kafka.PublishOrder("orders", orderEvent)
	if err != nil {
		// Kalau gagal kirim ke Kafka, kita harus balikin stok Redis (Kompensasi)
		database.Rdb.Incr(r.Context(), redisKey)
		http.Error(w, "Gagal memproses order", http.StatusInternalServerError)
		return
	}

	// 5. Beri Respon "Accepted" (Bukan Created/Success)
	// Karena aslinya belum masuk database SQL, cuma masuk antrian.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // Code 202
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Pesanan diterima! Sedang diproses.",
		"sisa_stok_estimasi": sisaStok,
		"status": "queued",
	})
}