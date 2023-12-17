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
