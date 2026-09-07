package usecase

import (
	"context"

	"github.com/google/uuid"

	"avito-kitchen/internal/domain"
)

// fakeRestaurantRepo — реализация domain.RestaurantRepository в памяти.
// Usecase-слой знает только про интерфейс (см. internal/domain/repository.go),
// поэтому для теста бизнес-логики реальный Postgres не нужен вообще —
// это и есть та причина, по которой репозитории спрятаны за интерфейсами.
type fakeRestaurantRepo struct {
	restaurants map[int64]domain.Restaurant
	menuItems   map[int64]domain.MenuItem // menuItemID -> item
}

var _ domain.RestaurantRepository = (*fakeRestaurantRepo)(nil)

func newFakeRestaurantRepo() *fakeRestaurantRepo {
	return &fakeRestaurantRepo{
		restaurants: make(map[int64]domain.Restaurant),
		menuItems:   make(map[int64]domain.MenuItem),
	}
}

func (f *fakeRestaurantRepo) GetAll(_ context.Context) ([]domain.Restaurant, error) {
	out := make([]domain.Restaurant, 0, len(f.restaurants))
	for _, r := range f.restaurants {
		out = append(out, r)
	}
	return out, nil
}

func (f *fakeRestaurantRepo) GetByID(_ context.Context, id int64) (*domain.Restaurant, error) {
	r, ok := f.restaurants[id]
	if !ok {
		return nil, domain.ErrRestaurantNotFound
	}
	return &r, nil
}

func (f *fakeRestaurantRepo) GetMenuByRestaurantID(_ context.Context, restaurantID int64) ([]domain.MenuItem, error) {
	out := make([]domain.MenuItem, 0)
	for _, item := range f.menuItems {
		if item.RestaurantID == restaurantID {
			out = append(out, item)
		}
	}
	return out, nil
}

// GetMenuItemsByIDs повторяет поведение настоящего SQL-запроса
// (queryGetMenuItemsByIDs): позиция, которой нет или которая принадлежит
// другому ресторану, просто не попадает в результат — вызывающий код сам
// сверяет len(result) с len(ids), чтобы отличить "не найдено".
func (f *fakeRestaurantRepo) GetMenuItemsByIDs(_ context.Context, restaurantID int64, ids []int64) ([]domain.MenuItem, error) {
	out := make([]domain.MenuItem, 0, len(ids))
	for _, id := range ids {
		item, ok := f.menuItems[id]
		if !ok || item.RestaurantID != restaurantID {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

func (f *fakeRestaurantRepo) UpdateMenuItem(_ context.Context, restaurantID, menuItemID int64, patch domain.UpdateMenuItemRequest) (*domain.MenuItem, error) {
	item, ok := f.menuItems[menuItemID]
	if !ok || item.RestaurantID != restaurantID {
		return nil, domain.ErrMenuItemNotFound
	}

	if patch.IsAvailable != nil {
		item.IsAvailable = *patch.IsAvailable
	}
	if patch.Price != nil {
		item.Price = *patch.Price
	}

	f.menuItems[menuItemID] = item
	return &item, nil
}

// fakeOrderRepo — реализация domain.OrderRepository в памяти.
type fakeOrderRepo struct {
	orders map[uuid.UUID]domain.Order
}

var _ domain.OrderRepository = (*fakeOrderRepo)(nil)

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{orders: make(map[uuid.UUID]domain.Order)}
}

func (f *fakeOrderRepo) CreateOrder(_ context.Context, order *domain.Order) error {
	order.ID = uuid.New()
	f.orders[order.ID] = *order
	return nil
}

func (f *fakeOrderRepo) GetOrderByID(_ context.Context, id uuid.UUID) (*domain.Order, error) {
	o, ok := f.orders[id]
	if !ok {
		return nil, domain.ErrOrderNotFound
	}
	return &o, nil
}

func (f *fakeOrderRepo) UpdateStatus(_ context.Context, id uuid.UUID, status domain.OrderStatus) error {
	o, ok := f.orders[id]
	if !ok {
		return domain.ErrOrderNotFound
	}
	o.Status = status
	f.orders[id] = o
	return nil
}

func (f *fakeOrderRepo) GetOrdersByRestaurantAndStatus(_ context.Context, restaurantID int64, status domain.OrderStatus) ([]domain.Order, error) {
	out := make([]domain.Order, 0)
	for _, o := range f.orders {
		if o.RestaurantID == restaurantID && o.Status == status {
			out = append(out, o)
		}
	}
	return out, nil
}
