package cache

import (
	"gorm.io/gorm"
	"log"
	"os"
	"sync"
	"wild_project/src/models"
)

var logger *log.Logger

func init() {
	logDir := "logs"
	logFile := "cache.log"
	fullPath := logDir + "/" + logFile

	// Проверка на существование папки для логов
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.Mkdir(logDir, 0755)
		if err != nil {
			log.Fatalln("Failed to create log directory:", err)
		}
	}

	// Открытие файла логов с созданием, если он не существует
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
	}

	// Инициализация глобального логгера для этого пакета
	logger = log.New(file, "CACHE: ", log.Ldate|log.Ltime|log.Lshortfile)
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
	oc.mu.Lock()
	defer oc.mu.Unlock()
	oc.orders[order.OrderUID] = order
	logger.Println("Order added to cache:", order.OrderUID)
}

// Get извлекает заказ из кеша по его уникальному идентификатору
func (oc *OrderCache) Get(orderUID string) (models.Order, bool) {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	order, exists := oc.orders[orderUID]
	return order, exists
}

// SaveToDB сохраняет заказ в базу данных и добавляет его в кеш
func (oc *OrderCache) SaveToDB(db *gorm.DB, order models.Order, logger *log.Logger) error {
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
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	return len(oc.orders)
}

// todo надо еще тащить заказ из кеша и залогировать это, еще сделать логи по папкам в одной портянке будет тяжело разобраться
