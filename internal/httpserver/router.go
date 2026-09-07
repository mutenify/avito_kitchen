package httpserver

import "net/http"

// NewRouter собирает все ручки на стандартном http.ServeMux.
//
// Начиная с Go 1.22 у ServeMux появился матчинг по методу ("GET /path") и
// параметры пути ("{id}", читаются через r.PathValue) — для API такого
// размера этого достаточно и не требует внешнего роутера (chi, gorilla/mux
// и т.п.), см. README раздел 5 "Статус реализации".
//
// Маршруты явно сгруппированы по тому, кто их вызывает — клиент или
// заведение, — это то же разделение, что описано в README разделе 4.
func NewRouter(
	restaurantHandler *RestaurantHandler,
	orderHandler *OrderHandler,
	restaurantOrderHandler *RestaurantOrderHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", HealthCheck)

	// --- Клиентский API: каталог заведений и свой заказ ---
	mux.HandleFunc("GET /api/v1/restaurants", restaurantHandler.ListRestaurants)
	mux.HandleFunc("GET /api/v1/restaurants/{restaurantId}/menu", restaurantHandler.GetMenu)
	mux.HandleFunc("POST /api/v1/orders", orderHandler.CreateOrder)
	mux.HandleFunc("GET /api/v1/orders/{orderId}", orderHandler.GetOrder)

	// --- API заведения: управление меню и обработка своих заказов ---
	mux.HandleFunc("PATCH /api/v1/restaurants/{restaurantId}/menu/{itemId}", restaurantHandler.UpdateMenuItem)
	mux.HandleFunc("GET /api/v1/restaurants/{restaurantId}/orders", restaurantOrderHandler.ListOrders)
	mux.HandleFunc("PATCH /api/v1/restaurants/{restaurantId}/orders/{orderId}/status", restaurantOrderHandler.UpdateStatus)

	return withMiddleware(mux)
}

// HealthCheck — GET /healthz. Понадобится docker-compose healthcheck'у
// основного сервиса и mock-сервису заведения, чтобы понять, что Core API
// уже поднялся и можно начинать опрашивать заказы.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
