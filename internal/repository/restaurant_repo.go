package repository

import (
	"context"
	"database/sql"
	"fmt"

	"avito-kitchen/internal/domain"
)

var _ domain.RestaurantRepository = (*RestaurantRepository)(nil)

type RestaurantRepository struct {
	db *sql.DB
}

func NewRestaurantRepository(db *sql.DB) *RestaurantRepository {
	return &RestaurantRepository{db: db}
}

func (r *RestaurantRepository) GetAll(ctx context.Context) ([]domain.Restaurant, error) {
	rows, err := r.db.QueryContext(ctx, queryGetAllRestaurants)
	if err != nil {
		return nil, mapError(err, nil)
	}
	defer rows.Close()

	restaurants := make([]domain.Restaurant, 0)
	for rows.Next() {
		var rest domain.Restaurant
		if err := rows.Scan(&rest.ID, &rest.Name, &rest.Address, &rest.IsActive, &rest.CreatedAt); err != nil {
			return nil, mapError(err, nil)
		}
		restaurants = append(restaurants, rest)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(err, nil)
	}

	return restaurants, nil
}

func (r *RestaurantRepository) GetByID(ctx context.Context, id int64) (*domain.Restaurant, error) {
	var rest domain.Restaurant
	err := r.db.QueryRowContext(ctx, queryGetRestaurantByID, id).Scan(
		&rest.ID, &rest.Name, &rest.Address, &rest.IsActive, &rest.CreatedAt,
	)
	if err != nil {
		return nil, mapError(err, domain.ErrRestaurantNotFound)
	}

	return &rest, nil
}

func (r *RestaurantRepository) GetMenuByRestaurantID(ctx context.Context, restaurantID int64) ([]domain.MenuItem, error) {
	rows, err := r.db.QueryContext(ctx, queryGetMenuByRestaurantID, restaurantID)
	if err != nil {
		return nil, mapError(err, nil)
	}
	defer rows.Close()

	menu := make([]domain.MenuItem, 0)
	for rows.Next() {
		var item domain.MenuItem
		if err := rows.Scan(&item.ID, &item.RestaurantID, &item.Name, &item.Description, &item.Price, &item.IsAvailable, &item.CreatedAt); err != nil {
			return nil, mapError(err, nil)
		}
		menu = append(menu, item)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(err, nil)
	}

	return menu, nil
}

func (r *RestaurantRepository) GetMenuItemsByIDs(ctx context.Context, restaurantID int64, ids []int64) ([]domain.MenuItem, error) {
	if len(ids) == 0 {
		return []domain.MenuItem{}, nil
	}

	rows, err := r.db.QueryContext(ctx, queryGetMenuItemsByIDs, restaurantID, ids)
	if err != nil {
		return nil, mapError(err, nil)
	}
	defer rows.Close()

	items := make([]domain.MenuItem, 0, len(ids))
	for rows.Next() {
		var item domain.MenuItem
		if err := rows.Scan(&item.ID, &item.RestaurantID, &item.Name, &item.Description, &item.Price, &item.IsAvailable, &item.CreatedAt); err != nil {
			return nil, mapError(err, nil)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(err, nil)
	}

	return items, nil
}

func (r *RestaurantRepository) UpdateMenuItem(ctx context.Context, restaurantID, menuItemID int64, patch domain.UpdateMenuItemRequest) (*domain.MenuItem, error) {
	if patch.IsAvailable == nil && patch.Price == nil {
		return nil, fmt.Errorf("%w: at least one field (is_available or price) must be provided", domain.ErrInvalidInput)
	}

	// decimal.Decimal реализует driver.Valuer со значимым (не указательным) ресивером,
	// поэтому *decimal.Decimal тоже удовлетворяет driver.Valuer — и database/sql вызовет
	// Value() прямо на указателе, не проверив его на nil. Разыменовываем сами и передаём
	// interface{}(nil) явно, если поле не задано.
	var priceArg interface{}
	if patch.Price != nil {
		priceArg = *patch.Price
	}

	var item domain.MenuItem
	err := r.db.QueryRowContext(ctx, queryUpdateMenuItem,
		restaurantID, menuItemID, patch.IsAvailable, priceArg,
	).Scan(&item.ID, &item.RestaurantID, &item.Name, &item.Description, &item.Price, &item.IsAvailable, &item.CreatedAt)
	if err != nil {
		return nil, mapError(err, domain.ErrMenuItemNotFound)
	}

	return &item, nil
}
