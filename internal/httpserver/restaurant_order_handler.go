package httpserver

import (
	"fmt"
	"net/http"

	"avito-kitchen/internal/domain"
	"avito-kitchen/internal/usecase"
)

// RestaurantOrderHandler — API заведения над заказами: опрос новых заказов
// (polling, см. README раздел 3.2) и смена их статуса.
//
// Обе ручки принимают restaurantId в пути запроса — этим заведение опознаёт
// себя. Полноценной аутентификации в MVP нет (см. README раздел 6), но
// usecase.OrderUsecase.UpdateStatus дополнительно сверяет, что заказ с
// переданным orderId реально принадлежит этому restaurantId — иначе заказ
// одного заведения можно было бы изменить, зная только его UUID.
type RestaurantOrderHandler struct {
	uc *usecase.OrderUsecase
}

func NewRestaurantOrderHandler(uc *usecase.OrderUsecase) *RestaurantOrderHandler {
	return &RestaurantOrderHandler{uc: uc}
}

// ListOrders — GET /api/v1/restaurants/{restaurantId}/orders?status=CREATED.
func (h *RestaurantOrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := parsePathInt64(r, "restaurantId")
	if err != nil {
		writeError(w, err)
		return
	}

	rawStatus := r.URL.Query().Get("status")
	if rawStatus == "" {
		writeError(w, fmt.Errorf("%w: query parameter status is required", domain.ErrInvalidInput))
		return
	}

	status := domain.OrderStatus(rawStatus)
	if !status.IsValid() {
		writeError(w, fmt.Errorf("%w: unknown status %q", domain.ErrInvalidInput, rawStatus))
		return
	}

	orders, err := h.uc.GetOrdersForRestaurant(r.Context(), restaurantID, status)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, orders)
}

// UpdateStatus — PATCH /api/v1/restaurants/{restaurantId}/orders/{orderId}/status.
// Допустимые переходы статуса описаны в domain.OrderStatus.CanTransitionTo;
// здесь только валидируется, что status вообще является одним из известных
// значений (иначе это 400, а не 409 — см. writeError/statusFor).
func (h *RestaurantOrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := parsePathInt64(r, "restaurantId")
	if err != nil {
		writeError(w, err)
		return
	}

	orderID, err := parsePathUUID(r, "orderId")
	if err != nil {
		writeError(w, err)
		return
	}

	var req domain.UpdateOrderStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	if !req.Status.IsValid() {
		writeError(w, fmt.Errorf("%w: unknown status %q", domain.ErrInvalidInput, req.Status))
		return
	}

	order, err := h.uc.UpdateStatus(r.Context(), restaurantID, orderID, req.Status)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, order)
}
