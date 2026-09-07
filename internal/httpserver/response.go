package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"avito-kitchen/internal/domain"
)

// ErrorResponse — единый формат тела ответа при любой ошибке.
// Соответствует схеме ErrorResponse в docs/openapi.yaml.
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeJSON сериализует payload и пишет его в ответ с нужным статусом.
// Общая точка для всех хендлеров, чтобы Content-Type и обработка ошибки
// кодирования не дублировались в каждом из них.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if payload == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// На этом этапе клиенту заголовки и статус уже отправлены,
		// исправить ответ нельзя — остаётся только залогировать.
		log.Printf("httpserver: failed to encode response: %v", err)
	}
}

// decodeJSON парсит тело запроса в dst. DisallowUnknownFields сделан
// намеренно строгим: опечатка в поле запроса (например, "quantiy" вместо
// "quantity") станет явной ошибкой 400, а не тихо проигнорируется.
func decodeJSON(r *http.Request, dst any) error {
	defer func() { _ = r.Body.Close() }()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: invalid request body: %v", domain.ErrInvalidInput, err)
	}

	return nil
}

// writeError переводит доменную ошибку в HTTP-ответ по таблице маппинга
// из README (раздел 4 "Маппинг доменных ошибок в HTTP-статусы").
//
// Работает через errors.Is, а не через сравнение текста ошибки — usecase- и
// repository-слои оборачивают базовые ошибки через fmt.Errorf("%w: ...", ...),
// добавляя контекст (например, "%w: %s" с именем блюда для ErrOutOfStock),
// и errors.Is умеет "пробить" через такую обёртку до базовой ошибки.
func writeError(w http.ResponseWriter, err error) {
	status := statusFor(err)

	if status == http.StatusInternalServerError {
		// Наружу уходит только общая фраза: подробности (текст SQL-ошибки,
		// путь до файла и т.п.) — потенциальная утечка деталей реализации.
		// Сами подробности при этом обязательно остаются в логе сервера.
		log.Printf("httpserver: internal error: %v", err)
		writeJSON(w, status, ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, status, ErrorResponse{Error: err.Error()})
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest

	case errors.Is(err, domain.ErrRestaurantNotFound),
		errors.Is(err, domain.ErrMenuItemNotFound),
		errors.Is(err, domain.ErrOrderNotFound),
		errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound

	case errors.Is(err, domain.ErrRestaurantInactive),
		errors.Is(err, domain.ErrOutOfStock),
		errors.Is(err, domain.ErrInvalidStatusTransition),
		errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict

	default:
		return http.StatusInternalServerError
	}
}
