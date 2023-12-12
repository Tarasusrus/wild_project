package natsclient

import (
	"errors"
	"fmt"
	"github.com/nats-io/stan.go"
	"sync"
	"wild_project/configs"
)

type NatsClient struct {
	nc   stan.Conn
	subs map[string]stan.Subscription
	mu   sync.Mutex
}

func NewNatsClient(url string, clusterID string, clientID string) (*NatsClient, error) {
	logger := configs.NewLogger()

	nc, err := stan.Connect(clusterID, clientID, stan.NatsURL(url),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			logger.Errorf("Connection lost, reason: %v", reason)
		}),
	)

	if err != nil {
		return nil, err
	}

	// Возвращение нового экземпляра NatsClient
	return &NatsClient{
		nc:   nc,
		subs: make(map[string]stan.Subscription),
	}, nil
}

func (c *NatsClient) Subscribe(topic string, handler stan.MsgHandler) error {
	logger := configs.NewLogger()
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.subs[topic]; exists {
		logger.Warnf("Already subscribed to topic: %s", topic)
		return nil // Или возвращаем ошибку, если подписка уже существует
	}

	sub, err := c.nc.Subscribe(topic, handler, stan.DurableName("my-durable"))
	if err != nil {
		logger.Errorf("Failed to subscribe to topic %s: %v", topic, err)
		return err
	}

	c.subs[topic] = sub
	logger.Infof("Subscribed to topic: %s", topic)
	return nil
}

func (c *NatsClient) Unsubscribe(topic string) error {
	logger := configs.NewLogger()
	c.mu.Lock()
	defer c.mu.Unlock()

	sub, ok := c.subs[topic]
	if !ok {
		errorMessage := fmt.Sprintf("Subscription not found for topic: %s", topic)
		logger.Errorf(errorMessage)
		return errors.New(errorMessage)
	}

	err := sub.Unsubscribe()
	if err != nil {
		logger.Errorf("Failed to unsubscribe from topic %s: %v", topic, err)
		return err
	}

	delete(c.subs, topic)
	logger.Infof("Unsubscribed from topic: %s", topic)
	return nil
}

func (c *NatsClient) PublishMessage(topic string, message []byte) error {
	logger := configs.NewLogger()
	err := c.nc.Publish(topic, message)
	if err != nil {
		logger.Errorf("Failed to publish message to topic %s: %v", topic, err)
		return err
	}
	logger.Infof("Message published to topic: %s", topic)
	return nil
}
