package handlers

import (
	"encoding/json"
	"errors"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
	"wild_project/src/cache"
	"wild_project/src/models"
	"wild_project/src/my_prometheus"
	natsclient "wild_project/src/nats"
)

var logger *log.Logger
var filePath = "logs/handlers.log"

func init() {
	logger = log.New(&lumberjack.Logger{
		Filename:   filePath, // Указываем путь к файлу лога
		MaxSize:    10,       // Размер файла в мегабайтах до ротации
		MaxBackups: 3,        // Максимальное количество старых файлов логов
		MaxAge:     28,       // Максимальное количество дней для хранения логов
		Compress:   true,     // Включение сжатия для старых файлов логов
	}, "HANDLERS: ", log.Ldate|log.Ltime|log.Lshortfile)
}

var path = "/Users/tarasmalinovskij/my_project/src/static"

// StartServer запускает HTTP-сервер
func StartServer(oc *cache.OrderCache, db *gorm.DB, natsClient *natsclient.NatsClient, port string) error {
	// Обслуживание статических файлов
	fs := http.FileServer(http.Dir(path))
	http.Handle("/", fs)

	// Обработчик API для получения информации о заказе
	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		overallStart := time.Now()

		// Извлечение ID заказа из запроса
		orderID := r.URL.Query().Get("id")
		logger.Printf("Получен запрос на заказ с ID: %s", orderID)

		// Получение данных заказа из кэша
		cacheStart := time.Now()
		if order, exists := oc.Get(orderID); exists {
			cacheDuration := time.Since(cacheStart).Seconds()
			my_prometheus.CacheResponseTime.WithLabelValues("/order").Observe(cacheDuration)

			json.NewEncoder(w).Encode(order)
			logger.Printf("Найден в кеше ID: %s", orderID)

			overallDuration := time.Since(overallStart).Seconds()
			my_prometheus.OverallResponseTime.WithLabelValues("/order").Observe(overallDuration)
			my_prometheus.TotalRequests.WithLabelValues("/order").Inc()
			return
		}
		cacheDuration := time.Since(cacheStart).Seconds()
		my_prometheus.CacheResponseTime.WithLabelValues("/order").Observe(cacheDuration)

		// Если заказ не найден в кэше, идем в базу данных
		dbStart := time.Now()
		var order models.Order
		if err := db.Where("order_uid = ?", orderID).First(&order).Error; err != nil {
			dbDuration := time.Since(dbStart).Seconds()
			my_prometheus.DbResponseTime.WithLabelValues("/order").Observe(dbDuration)

			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Order не найден", http.StatusNotFound)
				logger.Printf("Order не найден ID: %s", orderID)
			} else {
				http.Error(w, "Ошибка в БД", http.StatusInternalServerError)
				logger.Printf("Ошибка в БД: %s", orderID)
			}

			overallDuration := time.Since(overallStart).Seconds()
			my_prometheus.OverallResponseTime.WithLabelValues("/order").Observe(overallDuration)
			my_prometheus.TotalRequests.WithLabelValues("/order").Inc()
			return
		}
		dbDuration := time.Since(dbStart).Seconds()
		my_prometheus.DbResponseTime.WithLabelValues("/order").Observe(dbDuration)

		// Добавление заказа в кэш и отправка данных
		oc.Add(order)
		json.NewEncoder(w).Encode(order)

		overallDuration := time.Since(overallStart).Seconds()
		my_prometheus.OverallResponseTime.WithLabelValues("/order").Observe(overallDuration)
		my_prometheus.TotalRequests.WithLabelValues("/order").Inc()
	})

	http.HandleFunc("/sendToNats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		var message map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&message)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Сериализация сообщения для отправки в NATS
		msgData, err := json.Marshal(message)
		if err != nil {
			http.Error(w, "Error encoding message", http.StatusInternalServerError)
			return
		}

		// Отправка сообщения в NATS
		err = natsClient.PublishMessage("tests-channel", msgData)
		if err != nil {
			http.Error(w, "Error sending message to NATS", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Message sent to NATS successfully"))
	})

	// Запуск сервера
	return http.ListenAndServe(":"+port, nil)
}
