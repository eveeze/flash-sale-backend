package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eveeze/flash-sale/internal/database"
	"github.com/eveeze/flash-sale/internal/models"
)

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid Json Bdoy", http.StatusBadRequest)
		return
	}

	query := "INSERT INTO products (name, quantity, price) VALUES ($1, $2, $3) RETURNING id"
	err := database.DB.QueryRow(query, p.Name, p.Quantity, p.Price).Scan(&p.ID)
	if err != nil {
		http.Error(w, "Gagal menyimpan product ke database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	redisKey := fmt.Sprintf("product_stock:%d", p.ID)

	err = database.Rdb.Set(context.Background(), redisKey, p.Quantity, 0).Err()
	if err != nil {
		fmt.Printf("⚠️ Gagal simpan ke Redis: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Barang siap untuk Flash Sale!",
		"data":      p,
		"redis_key": redisKey, // Biar kita tau key-nya apa buat dicek nanti
	})
}
