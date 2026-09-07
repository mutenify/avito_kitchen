package httpserver

import (
	"log"
	"net/http"
	"time"
)

// withMiddleware оборачивает роутер сквозной обработкой, общей для всех
// ручек: логирование запроса и защита от паники в одном хендлере.
// Порядок важен: recoverPanic должен быть снаружи, чтобы успеть перехватить
// панику, случившуюся уже внутри logRequests/самого хендлера.
func withMiddleware(next http.Handler) http.Handler {
	return recoverPanic(logRequests(next))
}

// logRequests пишет в лог метод, путь, итоговый статус и время обработки
// каждого запроса. Для MVP этого достаточно; структурированные логи
// (request id, трейсинг) — уже за рамками задания.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

// statusRecorder перехватывает код ответа: стандартный http.ResponseWriter
// не даёт прочитать его обратно после вызова WriteHeader, поэтому без такой
// обёртки middleware не узнал бы, чем хендлер в итоге ответил.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// recoverPanic не даёт необработанной панике в одном запросе уронить весь
// процесс: перехватывает её, логирует и отвечает клиенту как обычной
// внутренней ошибкой.
func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v", rec)
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
