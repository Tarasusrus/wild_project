package main

import (
	"github.com/nats-io/stan.go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"wild_project/configs"
	"wild_project/src/models"
)

func main() {
	sc, err := stan.Connect("my_cluster", "client-123", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		log.Fatal("Ошибка подключения к серверу NATS", err)
	}
	defer sc.Close()

	dsn := "user=admin password=root dbname=mydatabase sslmode=disable host=localhost port=5433"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: configs.NewLogger(),
	})
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных: %v", err)
	}
	log.Println("Подключение успешно к бд и к NATS")

	err = db.AutoMigrate(&models.Order{}, &models.Delivery{}, &models.Payment{}, &models.Items{})
	if err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}
	log.Println("Миграция успешна")
}
