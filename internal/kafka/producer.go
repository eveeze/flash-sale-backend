package kafka

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/IBM/sarama"
	"github.com/eveeze/flash-sale/internal/config"
	"github.com/eveeze/flash-sale/pkg/logger" // Pakai logger kita
	"go.uber.org/zap"
)

var Producer sarama.SyncProducer

func InitProducer() {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 5

	brokers := strings.Split(config.AppConfig.KafkaBrokers, ",")

	var err error
	Producer, err = sarama.NewSyncProducer(brokers, kafkaConfig)
	if err != nil {
		// Fatal error tetap pakai log.Fatal atau logger.Log.Fatal
		log.Fatal("❌ Gagal connect ke Kafka Producer:", err)
	}
	
	// Gunakan Zap untuk info sukses
	logger.Log.Info("✅ Sukses connect ke Kafka Producer", zap.Strings("brokers", brokers))
}

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