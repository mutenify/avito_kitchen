package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"avito-kitchen/internal/domain"
)

type OrderUsecase struct {
	orderRepo      domain.OrderRepository
	restaurantRepo domain.RestaurantRepository
}

func NewOrderUsecase(orderRepo domain.OrderRepository, restaurantRepo domain.RestaurantRepository) *OrderUsecase {
	return &OrderUsecase{
		orderRepo:      orderRepo,
		restaurantRepo: restaurantRepo,
	}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, req domain.CreateOrderRequest) (*domain.Order, error) {
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("%w: order must contain at least one item", domain.ErrInvalidInput)
	}
	if req.RestaurantID <= 0 {
		return nil, fmt.Errorf("%w: restaurant_id is required", domain.ErrInvalidInput)
	}
	if req.UserID <= 0 {
		return nil, fmt.Errorf("%w: user_id is required", domain.ErrInvalidInput)
	}

	qtyByMenuItem := make(map[int64]int, len(req.Items))
	itemIDs := make([]int64, 0, len(req.Items))

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("%w: quantity must be greater than zero", domain.ErrInvalidInput)
		}
		if item.MenuItemID <= 0 {
			return nil, fmt.Errorf("%w: menu_item_id is required", domain.ErrInvalidInput)
		}
		if _, seen := qtyByMenuItem[item.MenuItemID]; !seen {
			itemIDs = append(itemIDs, item.MenuItemID)
		}
		qtyByMenuItem[item.MenuItemID] += item.Quantity
	}

	menuItems, err := u.restaurantRepo.GetMenuItemsByIDs(ctx, req.RestaurantID, itemIDs)
	if err != nil {
		return nil, err
	}
	if len(menuItems) != len(itemIDs) {
		return nil, domain.ErrMenuItemNotFound
	}

	totalAmount := decimal.Zero
	orderItems := make([]domain.OrderItem, 0, len(menuItems))

	for _, item := range menuItems {
		if !item.IsAvailable {
			return nil, fmt.Errorf("%w: %s", domain.ErrOutOfStock, item.Name)
		}

		qty := qtyByMenuItem[item.ID]
		qtyDec := decimal.NewFromInt(int64(qty))
		totalAmount = totalAmount.Add(item.Price.Mul(qtyDec))

		orderItems = append(orderItems, domain.OrderItem{
			MenuItemID:   item.ID,
			MenuItemName: item.Name,
			Quantity:     qty,
			PriceAtOrder: item.Price,
		})
	}

	order := &domain.Order{
		RestaurantID: req.RestaurantID,
		UserID:       req.UserID,
		Status:       domain.StatusCreated,
		TotalAmount:  totalAmount,
		Items:        orderItems,
	}

	if err := u.orderRepo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (u *OrderUsecase) UpdateStatus(ctx context.Context, orderID uuid.UUID, newStatus domain.OrderStatus) error {
	order, err := u.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	if !order.Status.CanTransitionTo(newStatus) {
		return fmt.Errorf("%w: cannot transition from %s to %s",
			domain.ErrInvalidStatusTransition, order.Status, newStatus)
	}

	return u.orderRepo.UpdateStatus(ctx, orderID, newStatus)
}

func (u *OrderUsecase) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return u.orderRepo.GetOrderByID(ctx, id)
}

func (u *OrderUsecase) GetOrdersForRestaurant(ctx context.Context, restaurantID int64, status domain.OrderStatus) ([]domain.Order, error) {
	return u.orderRepo.GetOrdersByRestaurantAndStatus(ctx, restaurantID, status)
}
