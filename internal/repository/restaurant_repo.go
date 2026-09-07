package repository

import (
	"context"
	"database/sql"

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
