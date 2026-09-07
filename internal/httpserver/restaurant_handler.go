package httpserver

import (
	"net/http"

	"avito-kitchen/internal/domain"
	"avito-kitchen/internal/usecase"
)

// RestaurantHandler закрывает две ручки клиентского API (каталог заведений
// и меню) и одну ручку API заведения (правка меню). Все три лежат на одном
// usecase.RestaurantUsecase и работают с одними и теми же доменными типами
// (Restaurant, MenuItem), поэтому отдельного хендлера под каждую заводить
// не стали — граница "клиент vs заведение" видна на уровне HTTP-маршрутов
// в router.go, а не структуры пакета.
type RestaurantHandler struct {
	uc *usecase.RestaurantUsecase
}

func NewRestaurantHandler(uc *usecase.RestaurantUsecase) *RestaurantHandler {
	return &RestaurantHandler{uc: uc}
}

// ListRestaurants — GET /api/v1/restaurants (клиентский API).
func (h *RestaurantHandler) ListRestaurants(w http.ResponseWriter, r *http.Request) {
	restaurants, err := h.uc.GetAllRestaurants(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, restaurants)
}

// GetMenu — GET /api/v1/restaurants/{restaurantId}/menu (клиентский API).
func (h *RestaurantHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := parsePathInt64(r, "restaurantId")
	if err != nil {
		writeError(w, err)
		return
	}

	menu, err := h.uc.GetMenu(r.Context(), restaurantID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, menu)
}

// UpdateMenuItem — PATCH /api/v1/restaurants/{restaurantId}/menu/{itemId}
// (API заведения). Частичное обновление: заведение присылает только то
// поле, которое реально меняется (is_available и/или price) — остальные
// поля запроса в domain.UpdateMenuItemRequest остаются nil и usecase их
// не трогает.
func (h *RestaurantHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	restaurantID, err := parsePathInt64(r, "restaurantId")
	if err != nil {
		writeError(w, err)
		return
	}

	itemID, err := parsePathInt64(r, "itemId")
	if err != nil {
		writeError(w, err)
		return
	}

	var req domain.UpdateMenuItemRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	item, err := h.uc.UpdateMenuItem(r.Context(), restaurantID, itemID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}
