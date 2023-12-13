package natsclient

import (
	"errors"
	"fmt"
	"github.com/nats-io/stan.go"
	"log"
	"os"
	"sync"
)

// todo надо добавить комменты
type NatsClient struct {
	nc   stan.Conn
	subs map[string]stan.Subscription
	mu   sync.Mutex
}

func NewNatsClient(url string, clusterID string, clientID string) (*NatsClient, error) {
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	nc, err := stan.Connect(clusterID, clientID, stan.NatsURL(url),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			logger.Println("Connection lost, reason: %v", reason)
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
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.subs[topic]; exists {
		logger.Println("Already subscribed to topic: %s", topic)
		return nil // Или возвращаем ошибку, если подписка уже существует
	}

	sub, err := c.nc.Subscribe(topic, handler, stan.DurableName("my-durable"))
	if err != nil {
		logger.Println("Failed to subscribe to topic %s: %v", topic, err)
		return err
	}

	c.subs[topic] = sub
	logger.Println("Subscribed to topic: %s", topic)
	return nil
}

func (c *NatsClient) Unsubscribe(topic string) error {
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	c.mu.Lock()
	defer c.mu.Unlock()

	sub, ok := c.subs[topic]
	if !ok {
		errorMessage := fmt.Sprintf("Subscription not found for topic: %s", topic)
		logger.Println(errorMessage)
		return errors.New(errorMessage)
	}

	err := sub.Unsubscribe()
	if err != nil {
		logger.Println("Failed to unsubscribe from topic %s: %v", topic, err)
		return err
	}

	delete(c.subs, topic)
	logger.Println("Unsubscribed from topic: %s", topic)
	return nil
}

func (c *NatsClient) PublishMessage(topic string, message []byte) error {
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	err := c.nc.Publish(topic, message)
	if err != nil {
		logger.Println("Failed to publish message to topic %s: %v", topic, err)
		return err
	}
	logger.Println("Message published to topic: %s", topic)
	return nil
}

func (c *NatsClient) Close() {
	c.nc.Close() // Закрытие соединения с NATS Streaming
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println("The connection sucsesfull closed")
}
