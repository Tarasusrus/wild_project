package main

import (
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

func loadAndCheckCache(orderCache *cache.OrderCache, db *gorm.DB) {
	// Проверка кеша до подключения к БД и после с сообщением о успешной загрузке кеша из БД
	cacheSizeBefore := orderCache.Count()
	mainLog.Printf("В кеше до загрузки даты : %d", cacheSizeBefore)
	// Из БД в кеш
	if err := orderCache.LoadFromDB(db); err != nil {
		mainLog.Fatalf("Ошибка в загрузке даты из БД: %v", err)
	}
	mainLog.Println("Дата успешно загрузилась")
	// Смотрим количество запискей в кеше
	cacheSizeAfter := orderCache.Count()
	mainLog.Printf("В кеше после загрузки даты : %d", cacheSizeAfter)
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
		mainLog.Fatalf("Ошибка в создании клиента NAts: %v", err)
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

	// Инициализация кэша и копирование из бд
	orderCache := cache.NewOrderCache()
	loadAndCheckCache(orderCache, db)

	// Подключение к NATS Streaming и подписка на канал
	err = client.Subscribe(channelName, func(m *stan.Msg) {
		mainLog.Printf("Получено новое сообщение: %s\n", string(m.Data))

		// Обработка сообщения
		utils.ProcessNatsMessage(orderCache, db, m)
	})
	if err != nil {
		mainLog.Fatalf("Ошибка при подписке на канал NATS: %v", err)
	}

	// Генерация и отправка тестовых сообщений в NATS Streaming
	messages, err := tests.GenerateTestMessages(1)
	if err != nil {
		mainLog.Fatalf("Ошибка при генерации тестовых сообщений: %v", err)
	}

	for _, message := range messages {
		if err := client.PublishMessage(channelName, []byte(message)); err != nil {
			mainLog.Printf("Ошибка при отправке сообщения: %v", err)
		}
	}

	// Генерация и отправка тестовых сообщений в NATS Streaming
	go func() {
		messages, err := tests.GenerateTestMessages(1)
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
		log.Fatalf("Ошибка во время запуска HTTP серваака: %v", err)
	}
	select {}

}
