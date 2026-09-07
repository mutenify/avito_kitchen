package usecase

import (
	"context"

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
	if _, err := u.restaurantRepo.GetByID(ctx, restaurantID); err != nil {
		return nil, err
	}
	return u.restaurantRepo.GetMenuByRestaurantID(ctx, restaurantID)
}
