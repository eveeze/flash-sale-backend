package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/eveeze/flash-sale/internal/database"
	"github.com/eveeze/flash-sale/internal/models"
)

func main() {
	fmt.Println("👷 Worker started... Menunggu pesanan...")

	// 1. Konek Database (Kita butuh akses SQL buat insert)
	database.ConnectDB()

	// 2. Setup Kafka Consumer
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	// Start dari pesan terlama yang belum diproses (OffsetOldest)
	config.Consumer.Offsets.Initial = sarama.OffsetOldest 

	// 3. Buat Consumer Group
	// Group ID "order-group" penting! 
	// Kalau nanti kita jalankan 5 worker, Kafka akan membagi tugas secara adil ke 5 worker ini.
	consumerGroup, err := sarama.NewConsumerGroup([]string{"127.0.0.1:9092"}, "order-group", config)
	if err != nil {
		log.Fatal("❌ Gagal membuat consumer group:", err)
	}

	// 4. Jalankan Consumer di dalam Loop
	ctx, cancel := context.WithCancel(context.Background())
	handler := &OrderHandler{}

	// Handle Graceful Shutdown (Ctrl+C)
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			// Consume topik "orders"
			if err := consumerGroup.Consume(ctx, []string{"orders"}, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			// Kalau context dibatalkan, stop loop
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-sigterm // Tunggu sinyal stop
	fmt.Println("\n⚠️ Worker shutting down...")
	cancel()
	if err = consumerGroup.Close(); err != nil {
		log.Printf("Error closing client: %v", err)
	}
}

// --- LOGIKA PEMROSESAN PESAN ---

type OrderHandler struct{}

// Setup (Dijalankan saat worker mulai)
func (h *OrderHandler) Setup(sarama.ConsumerGroupSession) error { return nil }

// Cleanup (Dijalankan saat worker mati)
func (h *OrderHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim (Logika Utama: Loop mengambil pesan)
func (h *OrderHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		// 1. Terima Pesan
		fmt.Printf("📩 Pesan masuk: %s\n", string(msg.Value))
		
		var order models.Order
		// 2. Decode JSON
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("❌ Error decode JSON: %v", err)
			session.MarkMessage(msg, "") // Tandai sudah dibaca meski error (biar gak macet)
			continue
		}

		// 3. Proses Simpan ke Database (Lambat)
		// Kita lakukan simulasi 'kerja berat' sedikit kalau mau
		// time.Sleep(100 * time.Millisecond) 
		
		query := "INSERT INTO orders (user_id, product_id, status, created_at) VALUES ($1, $2, 'success', $3)"
		_, err := database.DB.Exec(query, order.UserID, order.ProductID, time.Now())
		
		if err != nil {
			log.Printf("❌ Gagal insert ke DB: %v", err)
		} else {
			fmt.Printf("✅ Order Berhasil Disimpan! UserID: %d, ProductID: %d\n", order.UserID, order.ProductID)
		}

		// 4. Tandai pesan sebagai 'Selesai' (Commit Offset)
		session.MarkMessage(msg, "")
	}
	return nil
}