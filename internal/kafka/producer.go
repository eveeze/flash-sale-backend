package kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// Variabel global agar bisa dipanggil di mana saja
var Producer sarama.SyncProducer

// InitProducer: Menyiapkan koneksi ke Kafka
func InitProducer() {
	// Setup konfigurasi Kafka
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll // Tunggu sampai data benar-benar tersimpan
	config.Producer.Retry.Max = 5

	// Koneksi ke Kafka (Port 9092 sesuai Docker)
	var err error
	Producer, err = sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal("❌ Gagal connect ke Kafka:", err)
	}
	fmt.Println("✅ Sukses connect ke Kafka Producer")
}

// PublishOrder: Fungsi untuk mengirim pesan ke topik Kafka
func PublishOrder(topic string, message interface{}) error {
	// 1. Ubah data (struct) jadi JSON (byte)
	val, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// 2. Siapkan pesan
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(val),
	}

	// 3. Kirim pesan
	_, _, err = Producer.SendMessage(msg)
	return err
}
