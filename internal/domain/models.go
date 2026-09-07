package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	StatusCreated    OrderStatus = "CREATED"
	StatusAccepted   OrderStatus = "ACCEPTED"
	StatusCooking    OrderStatus = "COOKING"
	StatusDelivering OrderStatus = "DELIVERING"
	StatusCompleted  OrderStatus = "COMPLETED"
	StatusCancelled  OrderStatus = "CANCELLED"
)

type Restaurant struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Address   string    `json:"address" db:"address"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type MenuItem struct {
	ID           int64           `json:"id" db:"id"`
	RestaurantID int64           `json:"restaurant_id" db:"restaurant_id"`
	Name         string          `json:"name" db:"name"`
	Description  string          `json:"description" db:"description"`
	Price        decimal.Decimal `json:"price" db:"price"`
	IsAvailable  bool            `json:"is_available" db:"is_available"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

type Order struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	RestaurantID int64           `json:"restaurant_id" db:"restaurant_id"`
	UserID       int64           `json:"user_id" db:"user_id"`
	Status       OrderStatus     `json:"status" db:"status"`
	TotalAmount  decimal.Decimal `json:"total_amount" db:"total_amount"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
	Items        []OrderItem     `json:"items,omitempty" db:"-"`
}

type OrderItem struct {
	ID           int64           `json:"id" db:"id"`
	OrderID      uuid.UUID       `json:"order_id" db:"order_id"`
	MenuItemID   int64           `json:"menu_item_id" db:"menu_item_id"`
	Quantity     int             `json:"quantity" db:"quantity"`
	PriceAtOrder decimal.Decimal `json:"price_at_order" db:"price_at_order"`
	MenuItemName string          `json:"name,omitempty" db:"menu_item_name"`
}

type CreateOrderRequest struct {
	RestaurantID int64                 `json:"restaurant_id"`
	UserID       int64                 `json:"user_id"`
	Items        []CreateOrderItemItem `json:"items"`
}

type CreateOrderItemItem struct {
	MenuItemID int64 `json:"menu_item_id"`
	Quantity   int   `json:"quantity"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status"`
}
