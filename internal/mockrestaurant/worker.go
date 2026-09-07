package mockrestaurant

import (
	"context"
	"log"
	"sync"
	"time"
)

// Worker — сторона заведения из CJM (README раздел 3.2): раз в PollInterval
// опрашивает Core API на предмет заказов в статусе CREATED и для каждого
// нового заказа сам, без участия человека, проводит его по конвейеру
// ACCEPTED -> COOKING -> DELIVERING -> COMPLETED — так демонстрируется
// именно интеграция двух сервисов, а не просто факт существования обеих ручек.
type Worker struct {
	client *CoreClient

	pollInterval time.Duration
	stageDelay   time.Duration

	mu         sync.Mutex
	inProgress map[string]bool // order id -> уже взят в обработку этой копией воркера
	history    []OrderSnapshot // для собственной ручки GET /orders, см. server.go
}

// OrderSnapshot — то, что Worker помнит о заказе для отображения через
// собственный API. Не путать с Order (DTO ответа Core API): здесь только
// минимум, нужный для демонстрации прогресса конвейера.
type OrderSnapshot struct {
	OrderID     string    `json:"order_id"`
	Status      string    `json:"status"`
	TotalAmount string    `json:"total_amount"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewWorker(client *CoreClient, pollInterval, stageDelay time.Duration) *Worker {
	return &Worker{
		client:       client,
		pollInterval: pollInterval,
		stageDelay:   stageDelay,
		inProgress:   make(map[string]bool),
	}
}

// Run блокируется до отмены ctx — запускается из main в отдельной горутине.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.pollOnce(ctx) // не ждать первый тик — иначе первый заказ подхватится с задержкой PollInterval

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.pollOnce(ctx)
		}
	}
}

// pollOnce — один цикл опроса. Ошибка Core API (например, сервис ещё не
// поднялся при старте docker compose) не останавливает воркер — она просто
// логируется, и следующая попытка будет через pollInterval.
func (w *Worker) pollOnce(ctx context.Context) {
	orders, err := w.client.GetOrdersByStatus(ctx, "CREATED")
	if err != nil {
		log.Printf("mock-restaurant: poll failed: %v", err)
		return
	}

	for _, order := range orders {
		if w.claim(order.ID) {
			go w.processOrder(ctx, order)
		}
	}
}

// claim атомарно проверяет и помечает заказ как взятый в обработку.
// Нужен, чтобы при следующем опросе (или при короткой гонке между двумя
// опросами) один и тот же заказ не запустил вторую параллельную горутину
// обработки — GetOrdersByStatus("CREATED") иначе будет продолжать находить
// его, пока PATCH .../status не переведёт заказ в ACCEPTED.
func (w *Worker) claim(orderID string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.inProgress[orderID] {
		return false
	}
	w.inProgress[orderID] = true
	return true
}

// processOrder ведёт один заказ по всей цепочке статусов. Каждый шаг — тот
// же самый PATCH .../status, которым воспользовался бы сотрудник заведения
// вручную; Worker просто делает это по таймеру вместо человека.
func (w *Worker) processOrder(ctx context.Context, order Order) {
	log.Printf("mock-restaurant: заказ %s взят в работу (CREATED)", order.ID)

	pipeline := []string{"ACCEPTED", "COOKING", "DELIVERING", "COMPLETED"}
	for _, status := range pipeline {
		select {
		case <-ctx.Done():
			return
		case <-time.After(w.stageDelay):
		}

		updated, err := w.client.UpdateOrderStatus(ctx, order.ID, status)
		if err != nil {
			// Например, кто-то параллельно отменил заказ (CANCELLED) — тогда
			// следующий переход в пайплайне станет недопустимым, и Core API
			// ответит 409. Останавливаем обработку именно этого заказа, не
			// весь Worker.
			log.Printf("mock-restaurant: заказ %s -> %s не удался: %v", order.ID, status, err)
			return
		}

		log.Printf("mock-restaurant: заказ %s -> %s", order.ID, status)
		w.remember(*updated)
	}
}

func (w *Worker) remember(order Order) {
	w.mu.Lock()
	defer w.mu.Unlock()

	snap := OrderSnapshot{
		OrderID:     order.ID,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		UpdatedAt:   order.UpdatedAt,
	}

	for i, existing := range w.history {
		if existing.OrderID == snap.OrderID {
			w.history[i] = snap
			return
		}
	}
	w.history = append(w.history, snap)
}

// Snapshot отдаёт копию текущей истории — для GET /orders собственного API.
func (w *Worker) Snapshot() []OrderSnapshot {
	w.mu.Lock()
	defer w.mu.Unlock()

	out := make([]OrderSnapshot, len(w.history))
	copy(out, w.history)
	return out
}
