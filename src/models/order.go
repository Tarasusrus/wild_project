package models

import (
	"gorm.io/gorm"
	"time"
)

type Order struct {
	gorm.Model
	OrderUID          string `gorm:"uniqueIndex" json:"OrderUID"` // PK
	TrackNumber       string `json:"TrackNumber"`
	Entry             string `json:"Entry"`
	Delivery          Delivery
	Payment           Payment
	Items             []Items
	Locale            string `json:"Locale"`
	InternalSignature string `json:"InternalSignature"`
	CustomerID        string `json:"CustomerID"`
	DeliveryService   string `json:"DeliveryService"`
	Shardkey          string `json:"Shardkey"`
	SmID              string `json:"SmID"`
	DateCreated       time.Time
	OofShard          string `json:"OofShard"`
}

type Delivery struct {
	gorm.Model
	Name    string `json:"Name"`
	Phone   string `json:"Phone"`
	Zip     string `json:"Zip"`
	City    string `json:"City"`
	Adress  string `json:"Adress"`
	Region  string `json:"Region"`
	Email   string `json:"Email"`
	OrderID uint
}

type Payment struct {
	gorm.Model
	Transaction  string `json:"Transaction"`
	RequestID    string `json:"RequestID"`
	Currency     string `json:"Currency"`
	Provider     string `json:"Provider"`
	Amount       int    `json:"Amount"`
	PaymentDt    int    `json:"PaymentDt"`
	Bank         string `json:"Bank"`
	DeliveryCost int    `json:"DeliveryCost"`
	GoodsTotal   int    `json:"GoodsTotal"`
	CustomFee    int    `json:"CustomFee"`
	OrderID      uint   // Связь с Order
}

type Items struct {
	gorm.Model
	ChrtID      int    `json:"Chrt_id"`
	TrackNumber string `json:"Track_number"`
	Price       int    `json:"Price"`
	RID         string `json:"RID"`
	Name        string `json:"Name"`
	Sale        int    `json:"Sale"`
	Size        string `json:"Size"`
	TotalPrice  int    `json:"TotalPrice"`
	NmID        int    `json:"NmID"`
	Brand       string `json:"Brand"`
	Status      int    `json:"Status"`
	OrderID     uint   // Связь с Order
}
