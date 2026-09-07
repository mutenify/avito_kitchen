// Package mockrestaurant реализует сторону заведения из CJM (README, раздел 3.2):
// сервис-эмулятор, который опрашивает Core API за новыми заказами, проводит их
// по цепочке статусов и умеет управлять доступностью/ценой позиций меню.
//
// Пакет сознательно не импортирует avito-kitchen/internal/domain: в реальной
// жизни сервис заведения — независимый деплой (часто на другом стеке),
// который знает про Core API только по его HTTP-контракту
// (docs/openapi.yaml), а не по внутренним Go-типам соседнего сервиса. Поэтому
// здесь свои DTO, дублирующие форму JSON, а не типы.
package mockrestaurant

import "time"

// Order — зеркало схемы Order из docs/openapi.yaml.
type Order struct {
	ID           string      `json:"id"`
	RestaurantID int64       `json:"restaurant_id"`
	UserID       int64       `json:"user_id"`
	Status       string      `json:"status"`
	TotalAmount  string      `json:"total_amount"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Items        []OrderItem `json:"items,omitempty"`
}

// OrderItem — зеркало схемы OrderItem из docs/openapi.yaml.
type OrderItem struct {
	ID           int64  `json:"id"`
	MenuItemID   int64  `json:"menu_item_id"`
	Name         string `json:"name,omitempty"`
	Quantity     int    `json:"quantity"`
	PriceAtOrder string `json:"price_at_order"`
}

// MenuItem — зеркало схемы MenuItem из docs/openapi.yaml.
type MenuItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Price       string `json:"price"`
	IsAvailable bool   `json:"is_available"`
}

// errorResponse — зеркало схемы ErrorResponse из docs/openapi.yaml.
type errorResponse struct {
	Error string `json:"error"`
}
