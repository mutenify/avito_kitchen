package domain

// validTransitions описывает допустимые переходы между статусами заказа.
var validTransitions = map[OrderStatus][]OrderStatus{
	StatusCreated:    {StatusAccepted, StatusCancelled},
	StatusAccepted:   {StatusCooking, StatusCancelled},
	StatusCooking:    {StatusDelivering, StatusCancelled},
	StatusDelivering: {StatusCompleted},
	StatusCompleted:  {},
	StatusCancelled:  {},
}

// CanTransitionTo проверяет, разрешён ли переход из текущего статуса в next.
func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	allowed, ok := validTransitions[s]
	if !ok {
		return false
	}
	for _, status := range allowed {
		if status == next {
			return true
		}
	}
	return false
}
