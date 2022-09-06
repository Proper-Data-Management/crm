package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

func KafkaWriteMessage(broker string, topic string, msg string) error {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer w.Close()

	err := w.WriteMessages(context.Background(),
		kafka.Message{Value: []byte(msg)})
	//	Key:   []byte("Key-A"),

	return err
}
