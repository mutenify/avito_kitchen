package mockrestaurant

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// NewServer — собственный HTTP API сервиса-эмулятора заведения. Это ответ на
// пункт ТЗ "...в случае необходимости предоставлять свой API для обмена
// информацией": здесь у заведения есть свой (пусть и минимальный) интерфейс —
// посмотреть, что происходит с заказами, и поуправлять каталогом, — который
// уже сам сервис транслирует в вызовы Core API.
func NewServer(worker *Worker, client *CoreClient) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /orders", handleListOrders(worker))
	mux.HandleFunc("GET /menu", handleGetMenu(client))
	mux.HandleFunc("PATCH /menu/{itemId}", handleUpdateMenuItem(client))

	return mux
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleListOrders — GET /orders: что этот воркер сам провёл (или ведёт
// прямо сейчас) по конвейеру статусов. Источник — не Core API, а локальная
// память Worker, поэтому ручка отвечает мгновенно и не зависит от него.
func handleListOrders(worker *Worker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, worker.Snapshot())
	}
}

// handleGetMenu — GET /menu: каталог заведения глазами самого заведения
// (проксирует GET /api/v1/restaurants/{id}/menu на Core API).
func handleGetMenu(client *CoreClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		menu, err := client.GetMenu(r.Context())
		if err != nil {
			log.Printf("mock-restaurant: get menu failed: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, menu)
	}
}

// handleUpdateMenuItem — PATCH /menu/{itemId}: ручка персонала заведения —
// снять блюдо с продажи, вернуть в продажу или поменять цену (см. README,
// CJM 3.2, шаг "Актуализация каталога"). Сам сервис-эмулятор только
// проксирует запрос в Core API; у настоящего заведения на этом месте была
// бы его собственная админка/POS-система.
func handleUpdateMenuItem(client *CoreClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		itemID, err := strconv.ParseInt(r.PathValue("itemId"), 10, 64)
		if err != nil || itemID <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "itemId must be a positive integer"})
			return
		}

		var req struct {
			IsAvailable *bool   `json:"is_available"`
			Price       *string `json:"price"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		item, err := client.UpdateMenuItem(r.Context(), itemID, req.IsAvailable, req.Price)
		if err != nil {
			log.Printf("mock-restaurant: update menu item failed: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("mock-restaurant: failed to encode response: %v", err)
	}
}
