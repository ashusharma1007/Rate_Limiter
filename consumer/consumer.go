package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"rate-limiter/models"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Consumer struct {
	consumer   *kafka.Consumer
	aggregator *Aggregator
}

func NewConsumer(brokerAddress, groupId, topic string, aggregator *Aggregator) (*Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokerAddress,
		"group.id":          groupId,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load new consumer: %v", err)
	}

	err = c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to subscribe the topic: %v", err)
	}

	return &Consumer{consumer: c, aggregator: aggregator}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	defer c.consumer.Close()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		msg, err := c.consumer.ReadMessage(200 * time.Millisecond)
		if err != nil {
			if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
				continue
			}
			log.Printf("consumer read error: %v", err)
			continue
		}
		var messageRateLimit models.RateLimitEvent
		if err := json.Unmarshal(msg.Value, &messageRateLimit); err != nil {
			log.Printf("failed to unmarshal event: %v", err)
			continue
		}

		err = c.aggregator.Process(ctx, messageRateLimit)
		if err != nil {
			log.Printf("failed to process event: %v", err)
		}

	}
}
