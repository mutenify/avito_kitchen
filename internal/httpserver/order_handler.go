package httpserver

import (
	"net/http"

	"avito-kitchen/internal/domain"
	"avito-kitchen/internal/usecase"
)

// OrderHandler — клиентская часть API заказов: оформить заказ и посмотреть
// его статус по ссылке. Ручки, которыми пользуется заведение (опрос заказов,
// смена статуса), вынесены в RestaurantOrderHandler — см. его комментарий
// про то, почему это разделение принципиально, а не только для красоты кода.
type OrderHandler struct {
	uc *usecase.OrderUsecase
}

func NewOrderHandler(uc *usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

// CreateOrder — POST /api/v1/orders.
// Вся проверка (существование и активность заведения, доступность и цены
// позиций меню) находится в usecase.OrderUsecase.CreateOrder — хендлер
// только декодирует запрос и переводит результат/ошибку в HTTP-ответ.
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	order, err := h.uc.CreateOrder(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, order)
}

// GetOrder — GET /api/v1/orders/{orderId}.
// Ищет заказ только по UUID, без restaurant_id/user_id — см. README раздел
// 3.1 и описание ручки в docs/openapi.yaml про то, почему это осознанный
// выбор, а не забытая проверка доступа.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID, err := parsePathUUID(r, "orderId")
	if err != nil {
		writeError(w, err)
		return
	}

	order, err := h.uc.GetOrderByID(r.Context(), orderID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, order)
}
