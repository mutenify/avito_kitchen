package domain

import "testing"

func TestOrderStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from, to OrderStatus
		want     bool
	}{
		{StatusCreated, StatusAccepted, true},
		{StatusCreated, StatusCancelled, true},
		{StatusCreated, StatusCooking, false}, // нельзя перепрыгнуть через ACCEPTED
		{StatusAccepted, StatusCooking, true},
		{StatusAccepted, StatusCancelled, true},
		{StatusCooking, StatusDelivering, true},
		{StatusCooking, StatusCancelled, true},
		{StatusDelivering, StatusCompleted, true},
		{StatusDelivering, StatusCancelled, false}, // из доставки отменить нельзя
		{StatusCompleted, StatusCancelled, false},  // терминальный статус
		{StatusCompleted, StatusAccepted, false},   // терминальный статус
		{StatusCancelled, StatusAccepted, false},   // терминальный статус
	}

	for _, tt := range tests {
		got := tt.from.CanTransitionTo(tt.to)
		if got != tt.want {
			t.Errorf("%s.CanTransitionTo(%s) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestOrderStatus_IsValid(t *testing.T) {
	valid := []OrderStatus{
		StatusCreated, StatusAccepted, StatusCooking,
		StatusDelivering, StatusCompleted, StatusCancelled,
	}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("%s.IsValid() = false, want true", s)
		}
	}

	if OrderStatus("UNKNOWN").IsValid() {
		t.Error(`OrderStatus("UNKNOWN").IsValid() = true, want false`)
	}
	if OrderStatus("").IsValid() {
		t.Error(`OrderStatus("").IsValid() = true, want false`)
	}
}
