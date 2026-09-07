package httpserver

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"avito-kitchen/internal/domain"
)

// parsePathInt64 читает числовой параметр пути (id заведения, id позиции меню).
// Вынесен в отдельную функцию, чтобы в каждом хендлере не дублировать
// strconv.ParseInt и, самое главное, чтобы ошибка парсинга всегда превращалась
// в понятную клиенту ErrInvalidInput (400), а не пробрасывалась как есть.
func parsePathInt64(r *http.Request, name string) (int64, error) {
	raw := r.PathValue(name)

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%w: %s must be a positive integer, got %q", domain.ErrInvalidInput, name, raw)
	}

	return id, nil
}

// parsePathUUID делает то же самое, но для id заказа (UUID в пути).
func parsePathUUID(r *http.Request, name string) (uuid.UUID, error) {
	raw := r.PathValue(name)

	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %s must be a valid UUID, got %q", domain.ErrInvalidInput, name, raw)
	}

	return id, nil
}
