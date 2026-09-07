package httpserver

import (
	"fmt"
	"net/http"
	"testing"

	"avito-kitchen/internal/domain"
)

// TestStatusFor проверяет таблицу маппинга из README (раздел 4): каждая
// доменная ошибка — в том числе обёрнутая через fmt.Errorf("%w: ...") —
// должна разбираться через errors.Is в правильный HTTP-статус.
func TestStatusFor(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"invalid input", domain.ErrInvalidInput, http.StatusBadRequest},
		{"обёрнутый invalid input", fmt.Errorf("%w: quantity must be > 0", domain.ErrInvalidInput), http.StatusBadRequest},

		{"restaurant not found", domain.ErrRestaurantNotFound, http.StatusNotFound},
		{"menu item not found", domain.ErrMenuItemNotFound, http.StatusNotFound},
		{"order not found", domain.ErrOrderNotFound, http.StatusNotFound},
		{"generic not found", domain.ErrNotFound, http.StatusNotFound},

		{"restaurant inactive", domain.ErrRestaurantInactive, http.StatusConflict},
		{"обёрнутый out of stock", fmt.Errorf("%w: Кола", domain.ErrOutOfStock), http.StatusConflict},
		{"invalid status transition", domain.ErrInvalidStatusTransition, http.StatusConflict},
		{"already exists", domain.ErrAlreadyExists, http.StatusConflict},

		{"неизвестная ошибка", fmt.Errorf("boom"), http.StatusInternalServerError},
		{"явный internal", domain.ErrInternal, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusFor(tt.err); got != tt.want {
				t.Errorf("statusFor(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}
