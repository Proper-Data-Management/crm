package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
)

type KafkaReader struct {
	reader *kafka.Reader
	cancel context.CancelFunc
}
type KafkaReaderCallback func(offset int64, data string)

var _kafkaReaders map[string]*KafkaReader
var _onceInit sync.Once

func kafkaReadersSingletone() map[string]*KafkaReader {
	_onceInit.Do(func() {
		_kafkaReaders = make(map[string]*KafkaReader)
	})
	return _kafkaReaders
}

func KafkaReaderCreate(readerId, host, topic string, offset int64, cb KafkaReaderCallback) {
	KafkaReaderClose(readerId)
	krs := kafkaReadersSingletone()
	kr := KafkaReader{}
	krs[readerId] = &kr
	ctx, cancel := context.WithCancel(context.Background())
	kr.cancel = cancel
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{host}, //"localhost:9092"
		Topic:     topic,          // "testtopic1"
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})
	defer func() {
		r.Close()
		delete(krs, readerId)
		fmt.Printf("kafka reader closed:%s\n", readerId)
	}()
	r.SetOffset(offset)
	fmt.Printf("kafka reader started:%s offset:%d\n", readerId, offset)
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			break
		}
		fmt.Printf("kafka message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
		cb(m.Offset, string(m.Value))
	}
}

func KafkaReaderClose(readerId string) {
	krs := kafkaReadersSingletone()
	if kr, ok := krs[readerId]; ok {
		kr.cancel()
		delete(krs, readerId)
		fmt.Printf("kafka reader force closed:%s\n", readerId)
	}
}
func KafkaReaderList() []string {
	krs := kafkaReadersSingletone()
	keys := make([]string, 0, len(krs))
	for k := range krs {
		keys = append(keys, k)
	}
	return keys
}
