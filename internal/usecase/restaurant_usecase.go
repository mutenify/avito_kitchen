package usecase

import (
	"context"
	"fmt"

	"avito-kitchen/internal/domain"
)

type RestaurantUsecase struct {
	restaurantRepo domain.RestaurantRepository
}

func NewRestaurantUsecase(restaurantRepo domain.RestaurantRepository) *RestaurantUsecase {
	return &RestaurantUsecase{restaurantRepo: restaurantRepo}
}

func (u *RestaurantUsecase) GetAllRestaurants(ctx context.Context) ([]domain.Restaurant, error) {
	return u.restaurantRepo.GetAll(ctx)
}

func (u *RestaurantUsecase) GetRestaurantByID(ctx context.Context, id int64) (*domain.Restaurant, error) {
	rest, err := u.restaurantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !rest.IsActive {
		return nil, domain.ErrRestaurantInactive
	}
	return rest, nil
}

func (u *RestaurantUsecase) GetMenu(ctx context.Context, restaurantID int64) ([]domain.MenuItem, error) {
	// Переиспользуем GetRestaurantByID, а не сырой repo.GetByID, чтобы проверка
	// is_active была одинаковой везде, где мы отдаём данные заведения клиенту.
	if _, err := u.GetRestaurantByID(ctx, restaurantID); err != nil {
		return nil, err
	}
	return u.restaurantRepo.GetMenuByRestaurantID(ctx, restaurantID)
}

// UpdateMenuItem — точка входа для заведения: снять блюдо с продажи, вернуть
// в продажу или скорректировать цену. Не проверяет is_active заведения намеренно:
// даже временно отключённое заведение должно иметь возможность актуализировать
// свой каталог до включения обратно.
func (u *RestaurantUsecase) UpdateMenuItem(ctx context.Context, restaurantID, menuItemID int64, req domain.UpdateMenuItemRequest) (*domain.MenuItem, error) {
	if req.IsAvailable == nil && req.Price == nil {
		return nil, fmt.Errorf("%w: at least one field (is_available or price) must be provided", domain.ErrInvalidInput)
	}
	if req.Price != nil && req.Price.IsNegative() {
		return nil, fmt.Errorf("%w: price must not be negative", domain.ErrInvalidInput)
	}

	if _, err := u.restaurantRepo.GetByID(ctx, restaurantID); err != nil {
		return nil, err
	}

	return u.restaurantRepo.UpdateMenuItem(ctx, restaurantID, menuItemID, req)
}
