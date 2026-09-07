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

// IsValid проверяет, что значение вообще является одним из шести статусов
// заказа. Нужен HTTP-слою: значение status в query/body приходит строкой
// снаружи, и прежде чем передавать его дальше в usecase, надо отличить
// "неизвестный статус" (400 Bad Request) от "переход запрещён" (409 Conflict) —
// это разные ошибки в domain/errors.go.
func (s OrderStatus) IsValid() bool {
	_, ok := validTransitions[s]
	return ok
}
