package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings" // Butuh strings split
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/eveeze/flash-sale/internal/config" // Import Config
	"github.com/eveeze/flash-sale/internal/database"
	"github.com/eveeze/flash-sale/internal/models"
)

func main() {
	// 1. Load Config Dulu!
	config.LoadConfig()

	fmt.Println("👷 Worker started... Menunggu pesanan...")

	// 2. Konek Database (Menggunakan config dari step 1)
	database.ConnectDB()

	// 3. Setup Kafka Consumer
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Ambil brokers dari Config
	brokers := strings.Split(config.AppConfig.KafkaBrokers, ",")

	// 4. Buat Consumer Group
	consumerGroup, err := sarama.NewConsumerGroup(brokers, "order-group", saramaConfig)
	if err != nil {
		log.Fatal("❌ Gagal membuat consumer group:", err)
	}

	// 5. Jalankan Consumer Loop
	ctx, cancel := context.WithCancel(context.Background())
	handler := &OrderHandler{}

	// Graceful Shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{"orders"}, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-sigterm
	fmt.Println("\n⚠️ Worker shutting down...")
	cancel()
	if err = consumerGroup.Close(); err != nil {
		log.Printf("Error closing client: %v", err)
	}
}

// --- LOGIKA SAMA SEPERTI SEBELUMNYA ---

type OrderHandler struct{}

func (h *OrderHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *OrderHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *OrderHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("📩 Pesan masuk: %s\n", string(msg.Value))

		var order models.Order
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("❌ Error decode JSON: %v", err)
			session.MarkMessage(msg, "")
			continue
		}

		query := "INSERT INTO orders (user_id, product_id, status, created_at) VALUES ($1, $2, 'success', $3)"
		_, err := database.DB.Exec(query, order.UserID, order.ProductID, time.Now())

		if err != nil {
			log.Printf("❌ Gagal insert ke DB: %v", err)
		} else {
			fmt.Printf("✅ Order Berhasil Disimpan! UserID: %d, ProductID: %d\n", order.UserID, order.ProductID)
		}

		session.MarkMessage(msg, "")
	}
	return nil
}