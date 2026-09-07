package repository

const (
	// Запросы ресторанов и меню
	queryGetAllRestaurants = `
		SELECT id, name, address, is_active, created_at
		FROM restaurants
		WHERE is_active = true
		ORDER BY id ASC;`

	queryGetRestaurantByID = `
		SELECT id, name, address, is_active, created_at
		FROM restaurants
		WHERE id = $1;`

	queryGetMenuByRestaurantID = `
		SELECT id, restaurant_id, name, description, price, is_available, created_at
		FROM menu_items
		WHERE restaurant_id = $1
		ORDER BY id ASC;`

	queryGetMenuItemsByIDs = `
		SELECT id, restaurant_id, name, description, price, is_available, created_at
		FROM menu_items
		WHERE restaurant_id = $1 AND id = ANY($2);`

	// queryUpdateMenuItem — частичное обновление позиции меню.
	// COALESCE позволяет передавать NULL для полей, которые не нужно менять:
	// $3/$4 равны nil, если соответствующее поле не пришло в запросе.
	queryUpdateMenuItem = `
		UPDATE menu_items
		SET
			is_available = COALESCE($3, is_available),
			price        = COALESCE($4, price)
		WHERE restaurant_id = $1 AND id = $2
		RETURNING id, restaurant_id, name, description, price, is_available, created_at;`

	// Запросы заказов
	queryCreateOrder = `
		INSERT INTO orders (restaurant_id, user_id, status, total_amount)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at;`

	queryCreateOrderItem = `
		INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id;`

	queryGetOrderWithItems = `
		SELECT
			o.id, o.restaurant_id, o.user_id, o.status, o.total_amount, o.created_at, o.updated_at,
			oi.id, oi.menu_item_id, oi.quantity, oi.price_at_order, m.name
		FROM orders o
		LEFT JOIN order_items oi ON oi.order_id = o.id
		LEFT JOIN menu_items m ON m.id = oi.menu_item_id
		WHERE o.id = $1;`

	queryUpdateOrderStatus = `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2;`

	queryGetOrdersByRestaurantAndStatus = `
		SELECT id, restaurant_id, user_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE restaurant_id = $1 AND status = $2
		ORDER BY created_at ASC;`
)
