package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"strings" // Butuh ini untuk split string brokers

	"github.com/IBM/sarama"
	"github.com/eveeze/flash-sale/internal/config" // Import Config kita
)

var Producer sarama.SyncProducer

// InitProducer: Menyiapkan koneksi ke Kafka
func InitProducer() {
	// Setup konfigurasi Kafka
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 5

	// Ambil alamat broker dari Config Viper
	// Di .env formatnya "host1:9092,host2:9092", Sarama butuh []string
	brokers := strings.Split(config.AppConfig.KafkaBrokers, ",")

	var err error
	Producer, err = sarama.NewSyncProducer(brokers, kafkaConfig)
	if err != nil {
		log.Fatal("❌ Gagal connect ke Kafka Producer:", err)
	}
	fmt.Println("✅ Sukses connect ke Kafka Producer")
}

// PublishOrder: Mengirim pesan order ke Topic Kafka
func PublishOrder(topic string, message interface{}) error {
	val, err := json.Marshal(message)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(val),
	}

	_, _, err = Producer.SendMessage(msg)
	return err
}