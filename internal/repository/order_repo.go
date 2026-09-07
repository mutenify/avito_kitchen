package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"avito-kitchen/internal/domain"
)

var _ domain.OrderRepository = (*OrderRepository)(nil)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return mapError(err, nil)
	}
	// После успешного tx.Commit() ниже Rollback() вернёт sql.ErrTxDone — это
	// штатное поведение паттерна "defer Rollback + explicit Commit", ошибку
	// в этом случае осознанно игнорируем.
	defer func() { _ = tx.Rollback() }()

	err = tx.QueryRowContext(ctx, queryCreateOrder,
		order.RestaurantID, order.UserID, order.Status, order.TotalAmount,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return mapError(err, nil)
	}

	for i := range order.Items {
		order.Items[i].OrderID = order.ID
		if _, err := tx.ExecContext(ctx, queryCreateOrderItem,
			order.ID, order.Items[i].MenuItemID, order.Items[i].Quantity, order.Items[i].PriceAtOrder,
		); err != nil {
			return mapError(err, nil)
		}
	}

	if err := tx.Commit(); err != nil {
		return mapError(err, nil)
	}

	return nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, queryGetOrderWithItems, id)
	if err != nil {
		return nil, mapError(err, nil)
	}
	defer func() { _ = rows.Close() }()

	var order *domain.Order

	for rows.Next() {
		if order == nil {
			order = &domain.Order{Items: make([]domain.OrderItem, 0)}
		}

		var (
			rowItemID       sql.NullInt64
			rowMenuItemID   sql.NullInt64
			rowQuantity     sql.NullInt32
			rowPriceAtOrder sql.NullString
			rowItemName     sql.NullString
		)

		if err := rows.Scan(
			&order.ID, &order.RestaurantID, &order.UserID, &order.Status, &order.TotalAmount, &order.CreatedAt, &order.UpdatedAt,
			&rowItemID, &rowMenuItemID, &rowQuantity, &rowPriceAtOrder, &rowItemName,
		); err != nil {
			return nil, mapError(err, nil)
		}

		if !rowItemID.Valid {
			continue
		}

		price, err := decimal.NewFromString(rowPriceAtOrder.String)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to parse price_at_order: %v", domain.ErrInternal, err)
		}

		order.Items = append(order.Items, domain.OrderItem{
			ID:           rowItemID.Int64,
			OrderID:      order.ID,
			MenuItemID:   rowMenuItemID.Int64,
			MenuItemName: rowItemName.String,
			Quantity:     int(rowQuantity.Int32),
			PriceAtOrder: price,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, mapError(err, nil)
	}

	if order == nil {
		return nil, domain.ErrOrderNotFound
	}

	return order, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	res, err := r.db.ExecContext(ctx, queryUpdateOrderStatus, status, id)
	if err != nil {
		return mapError(err, nil)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return mapError(err, nil)
	}
	if rowsAffected == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

func (r *OrderRepository) GetOrdersByRestaurantAndStatus(ctx context.Context, restaurantID int64, status domain.OrderStatus) ([]domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, queryGetOrdersByRestaurantAndStatus, restaurantID, status)
	if err != nil {
		return nil, mapError(err, nil)
	}
	defer func() { _ = rows.Close() }()

	orders := make([]domain.Order, 0)
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.RestaurantID, &o.UserID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, mapError(err, nil)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(err, nil)
	}

	return orders, nil
}
