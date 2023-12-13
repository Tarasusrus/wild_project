package utils

import (
	"encoding/json"
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
