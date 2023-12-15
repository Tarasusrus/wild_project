package cache

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm"
	"log"
	"sync"
	"time"
	"wild_project/src/models"
)

var logger *log.Logger
var filePath = "logs/cache.log"

func init() {
	logger = log.New(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    10,   // Размер файла в мегабайтах до ротации
		MaxBackups: 3,    // Максимальное количество старых файлов логов
		MaxAge:     28,   // Максимальное количество дней для хранения логов
		Compress:   true, // Включение сжатия для старых файлов логов
	}, "CACHE: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// OrderCache структура для кеширования заказов
type OrderCache struct {
	mu     sync.RWMutex
	orders map[string]models.Order
}

// NewOrderCache создает новый экземпляр OrderCache
func NewOrderCache() *OrderCache {
	return &OrderCache{
		orders: make(map[string]models.Order),
	}
}

// Add добавляет заказ в кеш
func (oc *OrderCache) Add(order models.Order) {
	startTime := time.Now()
	defer func() {
		logger.Printf("Add выполнена за %s", time.Since(startTime))
	}()

	oc.mu.Lock()
	defer oc.mu.Unlock()
	oc.orders[order.OrderUID] = order
	logger.Println("Order added to cache:", order.OrderUID)
}

// Get извлекает заказ из кеша по его уникальному идентификатору
func (oc *OrderCache) Get(orderUID string) (models.Order, bool) {
	startTime := time.Now()
	defer func() {
		logger.Printf("Get выполнена за %s", time.Since(startTime))
	}()

	oc.mu.RLock()
	defer oc.mu.RUnlock()
	order, exists := oc.orders[orderUID]
	return order, exists
}

// SaveToDB сохраняет заказ в базу данных и добавляет его в кеш
func (oc *OrderCache) SaveToDB(db *gorm.DB, order models.Order) error {
	startTime := time.Now()
	defer func() {
		logger.Printf("SaveToDB выполнена за %s", time.Since(startTime))
	}()

	if _, exists := oc.Get(order.OrderUID); !exists {
		if err := db.Create(&order).Error; err != nil {
			logger.Printf("Ошибка при сохранении заказа в БД: %v", err)
			return err
		}
		oc.Add(order)
		logger.Printf("Заказ успешно сохранен в БД: %v", order.OrderUID)
	}
	return nil
}

// LoadFromDB принимает объект базы данных *gorm.DB в качестве параметра и выполняет запрос на получение всех заказов.
func (oc *OrderCache) LoadFromDB(db *gorm.DB) error {
	startTime := time.Now()
	defer func() {
		logger.Printf("LoadFromDB выполнена за %s", time.Since(startTime))
	}()

	var orders []models.Order
	if err := db.Find(&orders).Error; err != nil {
		return err
	}

	for _, order := range orders {
		oc.Add(order)
	}
	return nil
}

// Count возвращает количество заказов в кэше
func (oc *OrderCache) Count() int {
	startTime := time.Now()
	defer func() {
		logger.Printf("Count выполнена за %s", time.Since(startTime))
	}()

	oc.mu.RLock()
	defer oc.mu.RUnlock()
	return len(oc.orders)
}
