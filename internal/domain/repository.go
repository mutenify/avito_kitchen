package domain

import (
	"context"

	"github.com/google/uuid"
)

// RestaurantRepository — контракт хранилища ресторанов и меню.
type RestaurantRepository interface {
	GetAll(ctx context.Context) ([]Restaurant, error)
	GetByID(ctx context.Context, id int64) (*Restaurant, error)
	GetMenuByRestaurantID(ctx context.Context, restaurantID int64) ([]MenuItem, error)
	GetMenuItemsByIDs(ctx context.Context, restaurantID int64, ids []int64) ([]MenuItem, error)
	// UpdateMenuItem — частичное обновление позиции меню (доступность и/или цена).
	// Используется заведением, чтобы снять блюдо с продажи или скорректировать цену,
	// не трогая остальные поля.
	UpdateMenuItem(ctx context.Context, restaurantID, menuItemID int64, patch UpdateMenuItemRequest) (*MenuItem, error)
}

// OrderRepository — контракт хранилища заказов.
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *Order) error
	GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status OrderStatus) error
	GetOrdersByRestaurantAndStatus(ctx context.Context, restaurantID int64, status OrderStatus) ([]Order, error)
}
