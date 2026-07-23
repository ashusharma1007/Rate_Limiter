package kafka

import (
	"encoding/json"
	"log"
	"rate-limiter/models"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Producer struct {
	producer *kafka.Producer
	topic    string
}

func NewProducer(brokerAddress string, topic string) (*Producer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokerAddress,
	})
	if err != nil {
		return nil, err
	}

	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("delivery failed: %v", ev.TopicPartition.Error)
				}
			}
		}
	}()
	return &Producer{producer: producer, topic: topic}, nil
}

func (p *Producer) Publish(event models.RateLimitEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal event: %v", err)
	}
	p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Value: data,
	}, nil)
}

func (p *Producer) Flush(timoutMs int) {
	p.producer.Flush(timoutMs)
}
