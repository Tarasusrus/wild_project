package cache

import (
	"gorm.io/gorm"
	"log"
	"sync"
	"wild_project/src/models"
)

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

// todo надо еще тащить заказ из кеша и залогировать это, еще сделать логи по папкам в одной портянке будет тяжело разобраться
