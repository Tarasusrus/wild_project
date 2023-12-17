package utils

import (
	"encoding/json"
	"errors"
	"github.com/nats-io/stan.go"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm"
	"log"
	"wild_project/src/cache"
	"wild_project/src/models"
	natsclient "wild_project/src/nats"
)

// DeserializeOrder преобразует JSON-строку в структуру Order
func DeserializeOrder(jsonOrder string) (models.Order, error) {
	var order models.Order
	err := json.Unmarshal([]byte(jsonOrder), &order)
	if err != nil {
		return models.Order{}, err
	}
	return order, nil
}

// Сообщения с ошибками
const (
	ErrMissingOrderUID    = "отсутствует Order UID"
	ErrMissingTrackNumber = "отсутствует Order track number"
)

// ValidateOrder пример валидации полей
func ValidateOrder(order *models.Order) error {
	if order.OrderUID == "" {
		return errors.New(ErrMissingOrderUID)
	}
	if order.TrackNumber == "" {
		return errors.New(ErrMissingTrackNumber)
	}
	return nil
}

var logger *log.Logger
var filePath = "logs/utils.log"

func init() {
	// Инициализация логгера
	logger = log.New(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    10, // Размер файла в мегабайтах до ротации
		MaxBackups: 3,  // Максимальное количество старых файлов логов
		MaxAge:     28, // Максимальное количество дней для хранения логов
		Compress:   true,
	}, "NATS_HANDLER: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func SubscribeToNats(client *natsclient.NatsClient, orderCache *cache.OrderCache, db *gorm.DB, channelName string) {
	err := client.Subscribe(channelName, func(m *stan.Msg) {
		logger.Printf("Получено новое сообщение: %s\n", string(m.Data))
		ProcessNatsMessage(orderCache, db, m)
	})

	if err != nil {
		logger.Fatalf("Ошибка при подписке на канал NATS: %v", err)
	}
}

func ProcessNatsMessage(orderCache *cache.OrderCache, db *gorm.DB, m *stan.Msg) {
	// Десериализация сообщения
	order, err := DeserializeOrder(string(m.Data))
	if err != nil {
		logger.Printf("Ошибка десериализации заказа: %v", err)
		return
	}

	// Проверка наличия заказа в кэше
	if _, exists := orderCache.Get(order.OrderUID); exists {
		logger.Printf("Заказ уже есть в кэше: %v", order.OrderUID)
		return
	}

	// Проверка наличия заказа в БД
	var dbOrder models.Order
	if err := db.Where("order_uid = ?", order.OrderUID).First(&dbOrder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Заказа нет в БД, сохраняем его
			if err := db.Create(&order).Error; err != nil {
				logger.Printf("Ошибка при сохранении заказа в БД: %v", err)
				return
			}
			orderCache.Add(order)
			logger.Printf("Заказ добавлен в БД и кэш: %v", order.OrderUID)
		} else {
			logger.Printf("Ошибка при запросе к БД: %v", err)
		}
	} else {
		// Заказ уже есть в БД, добавляем в кэш, если он отсутствует
		orderCache.Add(dbOrder)
		logger.Printf("Заказ из БД добавлен в кэш: %v", order.OrderUID)
	}
}
