package utils

import (
	"encoding/json"
	"errors"
	"wild_project/src/models"
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

// ValidateOrder проверяет корректность данных заказа
// todo эту функцию можно расширить добавив проверку типа данных каждого поля, диапазона сообщения...
func ValidateOrder(order *models.Order) error {
	if order.OrderUID == "" {
		return errors.New("Ошибка в order UID")
	}
	if order.TrackNumber == "" {
		return errors.New("Ошибка в track number")
	}

	return nil
}
