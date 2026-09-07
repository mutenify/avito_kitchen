package domain

import (
	"encoding/json"
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

// MarshalJSON форматирует price ровно двумя знаками после запятой.
// decimal.Decimal.String() по умолчанию обрезает незначащие нули
// (599.00 → "599", 599.50 → "599.5"), что для цены недопустимо: клиент
// должен получать единообразный денежный формат независимо от того,
// круглая цена или нет.
func (m MenuItem) MarshalJSON() ([]byte, error) {
	type alias MenuItem
	return json.Marshal(struct {
		alias
		Price string `json:"price"`
	}{
		alias: alias(m),
		Price: m.Price.StringFixed(2),
	})
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

// MarshalJSON — см. комментарий у MenuItem.MarshalJSON, та же причина:
// total_amount должен всегда приходить с двумя знаками после запятой.
func (o Order) MarshalJSON() ([]byte, error) {
	type alias Order
	return json.Marshal(struct {
		alias
		TotalAmount string `json:"total_amount"`
	}{
		alias:       alias(o),
		TotalAmount: o.TotalAmount.StringFixed(2),
	})
}

type OrderItem struct {
	ID           int64           `json:"id" db:"id"`
	OrderID      uuid.UUID       `json:"order_id" db:"order_id"`
	MenuItemID   int64           `json:"menu_item_id" db:"menu_item_id"`
	Quantity     int             `json:"quantity" db:"quantity"`
	PriceAtOrder decimal.Decimal `json:"price_at_order" db:"price_at_order"`
	MenuItemName string          `json:"name,omitempty" db:"menu_item_name"`
}

// MarshalJSON — см. комментарий у MenuItem.MarshalJSON, та же причина:
// price_at_order должен всегда приходить с двумя знаками после запятой.
func (i OrderItem) MarshalJSON() ([]byte, error) {
	type alias OrderItem
	return json.Marshal(struct {
		alias
		PriceAtOrder string `json:"price_at_order"`
	}{
		alias:        alias(i),
		PriceAtOrder: i.PriceAtOrder.StringFixed(2),
	})
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

// UpdateMenuItemRequest — частичное обновление позиции меню со стороны заведения.
// Указатели позволяют отличить "поле не передано" от "поле сброшено в zero value":
// заведение может прислать только is_available (распродали блюдо) или только price
// (изменили прайс), не переотправляя остальные поля.
type UpdateMenuItemRequest struct {
	IsAvailable *bool            `json:"is_available,omitempty"`
	Price       *decimal.Decimal `json:"price,omitempty"`
}
