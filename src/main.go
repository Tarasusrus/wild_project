package main

import (
	"errors"
	"github.com/nats-io/stan.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"os"
	"time"
	"wild_project/src/cache"
	"wild_project/src/handlers"
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

var mainLog *log.Logger
var filePath = "logs/mainLog.log"

func init() {
	mainLog = log.New(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    10,   // Размер файла в мегабайтах до ротации
		MaxBackups: 3,    // Максимальное количество старых файлов логов
		MaxAge:     28,   // Максимальное количество дней для хранения логов
		Compress:   true, // Включение сжатия для старых файлов логов
	}, "CACHE: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	cwd, _ := os.Getwd()
	log.Println("Текущий рабочий каталог:", cwd)
	// Инициализация логгера
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
	http.Handle("/metrics", promhttp.Handler())
	// Подключение к NATS Streaming
	client, err := natsclient.NewNatsClient(natsURL, clusterID, clientID)
	if err != nil {
		mainLog.Fatalf("Failed to create NATS client: %v", err)
	}
	defer client.Close()

	// Подключение к базе данных
	dsn := "user=admin password=root dbname=mydatabase sslmode=disable host=localhost port=5433"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormLogger})
	if err != nil {
		mainLog.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	mainLog.Println("Успешное подключение к базе данных")

	// Автоматическая миграция моделей
	err = db.AutoMigrate(&models.Order{}, &models.Delivery{}, &models.Payment{}, &models.Items{})
	if err != nil {
		mainLog.Fatalf("Ошибка миграции: %v", err)
	}
	mainLog.Println("Миграция успешно завершена")

	// Инициализация кэша
	orderCache := cache.NewOrderCache()

	// Получение количества элементов в кэше
	cacheSizeBefore := orderCache.Count()
	log.Printf("Количество элементов в кэше: %d", cacheSizeBefore)

	// Загрузка данных из БД в кэш
	if err := orderCache.LoadFromDB(db); err != nil {
		mainLog.Fatalf("Ошибка при загрузке данных из БД в кэш: %v", err)
	}
	mainLog.Println("Данные успешно загружены из БД в кэш")

	// Получение количества элементов в кэше
	cacheSizeAfter := orderCache.Count()
	log.Printf("Количество элементов в кэше: %d", cacheSizeAfter)

	go func() {
		err = client.Subscribe(channelName, func(m *stan.Msg) {
			mainLog.Printf("Получено новое сообщение: %s\n", string(m.Data))

			// Десериализация сообщения
			order, err := utils.DeserializeOrder(string(m.Data))
			if err != nil {
				mainLog.Printf("Ошибка десериализации заказа: %v", err)
				return
			}

			// Валидация заказа
			if err := utils.ValidateOrder(&order); err != nil {
				log.Printf("Ошибка валидации заказа: %v", err)
				mainLog.Printf("Ошибка валидации заказа: %v", err)
				return
			}

			// Проверка наличия заказа в кэше
			if _, exists := orderCache.Get(order.OrderUID); exists {
				mainLog.Printf("Заказ уже есть в кэше: %v", order.OrderUID)
				return // Заказ уже есть в кэше, не сохраняем в БД
			}

			// Проверка наличия заказа в БД
			var dbOrder models.Order
			if err := db.Where("order_uid = ?", order.OrderUID).First(&dbOrder).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// Заказа нет в БД, сохраняем его
					if err := db.Create(&order).Error; err != nil {
						mainLog.Printf("Ошибка при сохранении заказа в БД: %v", err)
						return
					}
					orderCache.Add(order) // Добавляем заказ в кэш
					mainLog.Printf("Заказ добавлен в БД и кэш: %v", order.OrderUID)
				} else {
					mainLog.Printf("Ошибка при запросе к БД: %v", err)
				}
			} else {
				// Заказ уже есть в БД, добавляем в кэш, если он отсутствует
				orderCache.Add(dbOrder)
				mainLog.Printf("Заказ из БД добавлен в кэш: %v", order.OrderUID)
			}
		})
		if err != nil {
			mainLog.Fatalf("Ошибка при подписке: %v", err)
		}
	}()

	// Генерация и отправка тестовых сообщений в NATS Streaming
	go func() {
		messages, err := tests.GenerateTestMessages(1260)
		if err != nil {
			mainLog.Fatalf("Ошибка при генерации тестовых сообщений: %v", err)
		}

		for _, message := range messages {
			err = client.PublishMessage(channelName, []byte(message))
			if err != nil {
				mainLog.Printf("Ошибка при отправке сообщения: %v", err)
			}
		}
	}()
	// Запуск HTTP-сервера
	if err := handlers.StartServer(orderCache, db, client, "8080"); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
	select {}

}
