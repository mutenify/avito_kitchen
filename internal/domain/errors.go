package domain

import "errors"

// Базовые ошибки
var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrInvalidInput  = errors.New("invalid input data")
	ErrInternal      = errors.New("internal server error")
)

// Специфичные бизнес-ошибки
var (
	// Рестораны и меню
	ErrRestaurantNotFound = errors.New("restaurant not found")
	ErrRestaurantInactive = errors.New("restaurant is currently inactive")
	ErrMenuItemNotFound   = errors.New("menu item not found")
	ErrOutOfStock         = errors.New("menu item is out of stock or unavailable")

	// Заказы
	ErrOrderNotFound           = errors.New("order not found")
	ErrInvalidStatusTransition = errors.New("invalid order status transition")
)
