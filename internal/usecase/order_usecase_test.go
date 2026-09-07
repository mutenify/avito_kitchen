package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"avito-kitchen/internal/domain"
)

// newOrderFixture поднимает usecase с предсказуемыми данными:
// заведение 1 — активно, заведение 2 — отключено; в заведении 1 есть
// доступная "Пепперони" (id 101, 599) и недоступная "Кола" (id 102, 120).
func newOrderFixture() (*fakeRestaurantRepo, *fakeOrderRepo, *OrderUsecase) {
	restRepo := newFakeRestaurantRepo()
	restRepo.restaurants[1] = domain.Restaurant{ID: 1, Name: "Додо Пицца", IsActive: true}
	restRepo.restaurants[2] = domain.Restaurant{ID: 2, Name: "Закрытое заведение", IsActive: false}
	restRepo.menuItems[101] = domain.MenuItem{ID: 101, RestaurantID: 1, Name: "Пепперони", Price: decimal.NewFromInt(599), IsAvailable: true}
	restRepo.menuItems[102] = domain.MenuItem{ID: 102, RestaurantID: 1, Name: "Кола", Price: decimal.NewFromInt(120), IsAvailable: false}

	orderRepo := newFakeOrderRepo()
	return restRepo, orderRepo, NewOrderUsecase(orderRepo, restRepo)
}

func TestOrderUsecase_CreateOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("успешный заказ считает сумму и фиксирует снимок цены", func(t *testing.T) {
		_, _, uc := newOrderFixture()

		order, err := uc.CreateOrder(ctx, domain.CreateOrderRequest{
			RestaurantID: 1,
			UserID:       42,
			Items: []domain.CreateOrderItemItem{
				{MenuItemID: 101, Quantity: 2},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantTotal := decimal.NewFromInt(1198) // 2 * 599
		if !order.TotalAmount.Equal(wantTotal) {
			t.Errorf("total_amount = %s, want %s", order.TotalAmount, wantTotal)
		}
		if order.Status != domain.StatusCreated {
			t.Errorf("status = %s, want %s", order.Status, domain.StatusCreated)
		}
		if len(order.Items) != 1 || order.Items[0].Quantity != 2 {
			t.Errorf("items = %+v, want one item with quantity 2", order.Items)
		}
	})

	t.Run("пустая корзина -> ErrInvalidInput", func(t *testing.T) {
		_, _, uc := newOrderFixture()

		_, err := uc.CreateOrder(ctx, domain.CreateOrderRequest{RestaurantID: 1, UserID: 1})
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("err = %v, want ErrInvalidInput", err)
		}
	})

	t.Run("несуществующее заведение -> ErrRestaurantNotFound", func(t *testing.T) {
		_, _, uc := newOrderFixture()

		_, err := uc.CreateOrder(ctx, domain.CreateOrderRequest{
			RestaurantID: 999,
			UserID:       1,
			Items:        []domain.CreateOrderItemItem{{MenuItemID: 101, Quantity: 1}},
		})
		if !errors.Is(err, domain.ErrRestaurantNotFound) {
			t.Errorf("err = %v, want ErrRestaurantNotFound", err)
		}
	})

	t.Run("отключённое заведение -> ErrRestaurantInactive", func(t *testing.T) {
		_, _, uc := newOrderFixture()

		_, err := uc.CreateOrder(ctx, domain.CreateOrderRequest{
			RestaurantID: 2,
			UserID:       1,
			Items:        []domain.CreateOrderItemItem{{MenuItemID: 101, Quantity: 1}},
		})
		if !errors.Is(err, domain.ErrRestaurantInactive) {
			t.Errorf("err = %v, want ErrRestaurantInactive", err)
		}
	})

	t.Run("недоступное блюдо -> ErrOutOfStock", func(t *testing.T) {
		_, _, uc := newOrderFixture()

		_, err := uc.CreateOrder(ctx, domain.CreateOrderRequest{
			RestaurantID: 1,
			UserID:       1,
			Items:        []domain.CreateOrderItemItem{{MenuItemID: 102, Quantity: 1}},
		})
		if !errors.Is(err, domain.ErrOutOfStock) {
			t.Errorf("err = %v, want ErrOutOfStock", err)
		}
	})

	t.Run("блюдо не из меню этого заведения -> ErrMenuItemNotFound", func(t *testing.T) {
		_, _, uc := newOrderFixture()

		_, err := uc.CreateOrder(ctx, domain.CreateOrderRequest{
			RestaurantID: 1,
			UserID:       1,
			Items:        []domain.CreateOrderItemItem{{MenuItemID: 9999, Quantity: 1}},
		})
		if !errors.Is(err, domain.ErrMenuItemNotFound) {
			t.Errorf("err = %v, want ErrMenuItemNotFound", err)
		}
	})
}

func TestOrderUsecase_UpdateStatus(t *testing.T) {
	ctx := context.Background()

	// setupOrder создаёт заказ восстанавливаемого статуса у заданного заведения
	// и возвращает usecase поверх него.
	setupOrder := func(status domain.OrderStatus, restaurantID int64) (*OrderUsecase, uuid.UUID) {
		orderRepo := newFakeOrderRepo()
		id := uuid.New()
		orderRepo.orders[id] = domain.Order{ID: id, RestaurantID: restaurantID, Status: status}
		return NewOrderUsecase(orderRepo, newFakeRestaurantRepo()), id
	}

	t.Run("допустимый переход обновляет и возвращает заказ", func(t *testing.T) {
		uc, id := setupOrder(domain.StatusCreated, 1)

		order, err := uc.UpdateStatus(ctx, 1, id, domain.StatusAccepted)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.Status != domain.StatusAccepted {
			t.Errorf("status = %s, want %s", order.Status, domain.StatusAccepted)
		}
	})

	t.Run("запрещённый переход -> ErrInvalidStatusTransition", func(t *testing.T) {
		uc, id := setupOrder(domain.StatusCreated, 1)

		_, err := uc.UpdateStatus(ctx, 1, id, domain.StatusCooking) // минуя ACCEPTED
		if !errors.Is(err, domain.ErrInvalidStatusTransition) {
			t.Errorf("err = %v, want ErrInvalidStatusTransition", err)
		}
	})

	t.Run("заказ другого заведения -> ErrOrderNotFound", func(t *testing.T) {
		uc, id := setupOrder(domain.StatusCreated, 1)

		// Заказ существует, но принадлежит заведению 1 — заведение 2 не должно
		// суметь ни изменить его, ни (по коду ошибки) даже узнать, что он есть.
		_, err := uc.UpdateStatus(ctx, 2, id, domain.StatusAccepted)
		if !errors.Is(err, domain.ErrOrderNotFound) {
			t.Errorf("err = %v, want ErrOrderNotFound", err)
		}
	})

	t.Run("несуществующий заказ -> ErrOrderNotFound", func(t *testing.T) {
		uc, _ := setupOrder(domain.StatusCreated, 1)

		_, err := uc.UpdateStatus(ctx, 1, uuid.New(), domain.StatusAccepted)
		if !errors.Is(err, domain.ErrOrderNotFound) {
			t.Errorf("err = %v, want ErrOrderNotFound", err)
		}
	})
}
