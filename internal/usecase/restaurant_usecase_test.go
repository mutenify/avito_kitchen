package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"

	"avito-kitchen/internal/domain"
)

func TestRestaurantUsecase_GetRestaurantByID(t *testing.T) {
	repo := newFakeRestaurantRepo()
	repo.restaurants[1] = domain.Restaurant{ID: 1, IsActive: true}
	repo.restaurants[2] = domain.Restaurant{ID: 2, IsActive: false}
	uc := NewRestaurantUsecase(repo)
	ctx := context.Background()

	if _, err := uc.GetRestaurantByID(ctx, 1); err != nil {
		t.Errorf("активное заведение: unexpected error %v", err)
	}

	if _, err := uc.GetRestaurantByID(ctx, 2); !errors.Is(err, domain.ErrRestaurantInactive) {
		t.Errorf("отключённое заведение: err = %v, want ErrRestaurantInactive", err)
	}

	if _, err := uc.GetRestaurantByID(ctx, 999); !errors.Is(err, domain.ErrRestaurantNotFound) {
		t.Errorf("несуществующее заведение: err = %v, want ErrRestaurantNotFound", err)
	}
}

// TestRestaurantUsecase_GetMenu_RespectsIsActive — регрессионный тест: изначально
// GetMenu ходил в repo.GetByID напрямую и не видел is_active, из-за чего меню
// отключённого заведения спокойно отдавалось клиенту (см. историю правок).
func TestRestaurantUsecase_GetMenu_RespectsIsActive(t *testing.T) {
	repo := newFakeRestaurantRepo()
	repo.restaurants[1] = domain.Restaurant{ID: 1, IsActive: false}
	repo.menuItems[101] = domain.MenuItem{ID: 101, RestaurantID: 1}
	uc := NewRestaurantUsecase(repo)

	_, err := uc.GetMenu(context.Background(), 1)
	if !errors.Is(err, domain.ErrRestaurantInactive) {
		t.Errorf("err = %v, want ErrRestaurantInactive", err)
	}
}

func TestRestaurantUsecase_UpdateMenuItem(t *testing.T) {
	ctx := context.Background()

	newRepo := func() *fakeRestaurantRepo {
		repo := newFakeRestaurantRepo()
		repo.restaurants[1] = domain.Restaurant{ID: 1, IsActive: true}
		repo.menuItems[101] = domain.MenuItem{ID: 101, RestaurantID: 1, IsAvailable: true, Price: decimal.NewFromInt(599)}
		return repo
	}

	t.Run("без единого поля -> ErrInvalidInput", func(t *testing.T) {
		uc := NewRestaurantUsecase(newRepo())

		_, err := uc.UpdateMenuItem(ctx, 1, 101, domain.UpdateMenuItemRequest{})
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("err = %v, want ErrInvalidInput", err)
		}
	})

	t.Run("отрицательная цена -> ErrInvalidInput", func(t *testing.T) {
		uc := NewRestaurantUsecase(newRepo())
		price := decimal.NewFromInt(-1)

		_, err := uc.UpdateMenuItem(ctx, 1, 101, domain.UpdateMenuItemRequest{Price: &price})
		if !errors.Is(err, domain.ErrInvalidInput) {
			t.Errorf("err = %v, want ErrInvalidInput", err)
		}
	})

	t.Run("снятие с продажи применяется, остальные поля не трогаются", func(t *testing.T) {
		uc := NewRestaurantUsecase(newRepo())
		notAvailable := false

		item, err := uc.UpdateMenuItem(ctx, 1, 101, domain.UpdateMenuItemRequest{IsAvailable: &notAvailable})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.IsAvailable {
			t.Error("is_available = true, want false")
		}
		if !item.Price.Equal(decimal.NewFromInt(599)) {
			t.Errorf("price = %s, want unchanged 599", item.Price)
		}
	})

	t.Run("чужая позиция меню -> ErrMenuItemNotFound", func(t *testing.T) {
		repo := newRepo()
		repo.restaurants[2] = domain.Restaurant{ID: 2, IsActive: true} // заведение существует...
		uc := NewRestaurantUsecase(repo)
		notAvailable := false

		// ...но itemID=101 принадлежит заведению 1, а не 2.
		_, err := uc.UpdateMenuItem(ctx, 2, 101, domain.UpdateMenuItemRequest{IsAvailable: &notAvailable})
		if !errors.Is(err, domain.ErrMenuItemNotFound) {
			t.Errorf("err = %v, want ErrMenuItemNotFound", err)
		}
	})
}
