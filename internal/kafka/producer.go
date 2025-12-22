package kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

var Producer sarama.SyncProducer

func InitProducer() {
	config := sarama.NewConfig()
	// Kita butuh kepastian pesan sampai (WaitForAll) agar data tidak hilang
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	// Koneksi ke Kafka Localhost Port 9092
	var err error
	Producer, err = sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal("❌ Gagal connect ke Kafka:", err)
	}
	fmt.Println("✅ Sukses connect ke Kafka Producer")
}

// PublishOrder: Mengirim pesan order ke Topic Kafka
func PublishOrder(topic string, message interface{}) error {
	// Ubah data order jadi JSON byte
	val, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Bungkus jadi pesan Kafka
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(val),
	}

	// Kirim!
	_, _, err = Producer.SendMessage(msg)
	return err
}
