package handlers

import (
	"encoding/json"
	"net/http"
	"wild_project/src/cache"
)

// StartServer запускает HTTP-сервер
func StartServer(oc *cache.OrderCache, port string) error {
	// Обслуживание статических файлов (например, index.html)
	fs := http.FileServer(http.Dir("src/static"))
	http.Handle("/", fs)

	// Обработчик API для получения информации о заказе
	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		// Извлечение ID заказа из запроса
		orderID := r.URL.Query().Get("id")

		// Получение данных заказа из кэша
		if order, exists := oc.Get(orderID); exists {
			json.NewEncoder(w).Encode(order)
		} else {
			http.Error(w, "Order not found", http.StatusNotFound)
		}
	})

	// Запуск сервера
	return http.ListenAndServe(":"+port, nil)
}
