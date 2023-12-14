package main

import (
	"github.com/nats-io/stan.go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
	"wild_project/src/cache"
	"wild_project/src/models"
	natsclient "wild_project/src/nats"
	"wild_project/src/tests"
	"wild_project/src/utils"
)

const (
	clusterID   = "my_cluster"
	clientID    = "client-123"
	natsURL     = "nats://localhost:4222"
	channelName = "tests-channel"
)

func main() {
	// Инициализация логгера
	Natslogger := log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Порог медленных SQL запросов
			LogLevel:                  logger.Warn, // Уровень логирования
			IgnoreRecordNotFoundError: true,        // Игнорировать ошибки 'запись не найдена'
			ParameterizedQueries:      true,        // Параметризация запросов в логах
			Colorful:                  false,       // Отключение цветов в логах
		},
	)

	// Подключение к NATS Streaming
	client, err := natsclient.NewNatsClient(natsURL, clusterID, clientID)
	if err != nil {
		Natslogger.Fatalf("Failed to create NATS client: %v", err)
	}
	defer client.Close()

	// Подключение к базе данных
	dsn := "user=admin password=root dbname=mydatabase sslmode=disable host=localhost port=5433"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormLogger})
	if err != nil {
		Natslogger.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	Natslogger.Println("Успешное подключение к базе данных")

	// Автоматическая миграция моделей
	err = db.AutoMigrate(&models.Order{}, &models.Delivery{}, &models.Payment{}, &models.Items{})
	if err != nil {
		Natslogger.Fatalf("Ошибка миграции: %v", err)
	}
	Natslogger.Println("Миграция успешно завершена")

	// Инициализация кэша
	orderCache := cache.NewOrderCache()

	// Получение количества элементов в кэше
	cacheSizeBefore := orderCache.Count()
	log.Printf("Количество элементов в кэше: %d", cacheSizeBefore)

	// Загрузка данных из БД в кэш
	if err := orderCache.LoadFromDB(db); err != nil {
		Natslogger.Fatalf("Ошибка при загрузке данных из БД в кэш: %v", err)
	}
	Natslogger.Println("Данные успешно загружены из БД в кэш")

	// Получение количества элементов в кэше
	cacheSizeAfter := orderCache.Count()
	log.Printf("Количество элементов в кэше: %d", cacheSizeAfter)

	// Подписка на канал в NATS Streaming
	err = client.Subscribe(channelName, func(m *stan.Msg) {
		Natslogger.Printf("Получено сообщение: %s\n", string(m.Data))

		// Десериализация сообщения
		order, err := utils.DeserializeOrder(string(m.Data))
		if err != nil {
			Natslogger.Printf("Ошибка десериализации заказа: %v", err)
			return
		}

		// Сохранение заказа в базу данных
		if err := db.Create(&order).Error; err != nil {
			Natslogger.Printf("Ошибка при сохранении заказа в БД: %v", err)
		} else {
			Natslogger.Printf("Заказ успешно сохранен в БД: %v", order.OrderUID)
		}
	})
	if err != nil {
		Natslogger.Fatalf("Ошибка при подписке: %v", err)
	}

	// Генерация и отправка тестовых сообщений в NATS Streaming
	messages, err := tests.GenerateTestMessages(1)
	if err != nil {
		Natslogger.Fatalf("Ошибка при генерации тестовых сообщений: %v", err)
	}

	for _, message := range messages {
		err = client.PublishMessage(channelName, []byte(message))
		if err != nil {
			Natslogger.Printf("Ошибка при отправке сообщения: %v", err)
		}
	}
}

// to do
