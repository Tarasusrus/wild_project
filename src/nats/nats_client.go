package natsclient

import (
	"errors"
	"fmt"
	"github.com/nats-io/stan.go"
	"log"
	"os"
	"sync"
)

// NatsClient хранит экземпляр соединения, карту подписок и мьютекс для синхронизации
type NatsClient struct {
	nc     stan.Conn
	subs   map[string]stan.Subscription
	mu     sync.Mutex
	logger *log.Logger
}

// NewNatsClient устанавливает новое соединение с сервером NATS Streaming и возвращает новый NatsClient
func NewNatsClient(url string, clusterID string, clientID string) (*NatsClient, error) {
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	nc, err := stan.Connect(clusterID, clientID, stan.NatsURL(url),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			logger.Println("Соединение потеряно, причина: %v", reason)
		}),
	)

	if err != nil {
		return nil, err
	}

	// Возвращение нового экземпляра NatsClient
	return &NatsClient{
		nc:     nc,
		subs:   make(map[string]stan.Subscription),
		logger: logger,
	}, nil
}

// Subscribe подписывается на тему
func (c *NatsClient) Subscribe(topic string, handler stan.MsgHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.subs[topic]; exists {
		c.logger.Printf("Уже подписаны на тему: %s", topic)
		return nil // Или возвращаем ошибку, если подписка уже существует
	}

	sub, err := c.nc.Subscribe(topic, handler, stan.DurableName("my-durable"))
	if err != nil {
		c.logger.Printf("Не удалось подписаться на тему %s: %v", topic, err)
		return err
	}

	c.subs[topic] = sub
	c.logger.Printf("Subscribed to topic: %s", topic)
	return nil
}

// Unsubscribe отписывается от темы
func (c *NatsClient) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sub, ok := c.subs[topic]
	if !ok {
		errorMessage := fmt.Sprintf("Подписка не найдена для темы: %s", topic)
		c.logger.Println(errorMessage)
		return errors.New(errorMessage)
	}

	err := sub.Unsubscribe()
	if err != nil {
		c.logger.Printf("Ошибка при отписке %s: %v", topic, err)
		return err
	}

	delete(c.subs, topic)
	c.logger.Printf("Unsubscribed from topic: %s", topic)
	return nil
}

// PublishMessage публикует новое сообщение на тему
func (c *NatsClient) PublishMessage(topic string, message []byte) error {
	err := c.nc.Publish(topic, message)
	if err != nil {
		c.logger.Printf("Ошибка публикации сообщения %s: %v", topic, err)
		return err
	}
	c.logger.Printf("Сообщение опубликовано в тему: %s", message, topic)
	return nil
}

// Close закрывает соединение с сервером NATS Streaming
func (c *NatsClient) Close() error {
	err := c.nc.Close()
	if err != nil {
		return err
	}
	c.logger.Println("Соединение успешно закрыто")
	return nil
}
