// Command mock-restaurant — эмулятор одного заведения, подключённого к
// Авито.Кухня (README, раздел "Заведения" и CJM 3.2).
//
// Это отдельный процесс/контейнер, который знает про Core API только через
// его HTTP-контракт (docs/openapi.yaml) — намеренно без единого импорта из
// internal/domain основного сервиса, см. комментарий в
// internal/mockrestaurant/types.go.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"avito-kitchen/internal/mockrestaurant"
)

func main() {
	cfg := mockrestaurant.LoadConfig()

	client := mockrestaurant.NewCoreClient(cfg.CoreAPIURL, cfg.RestaurantID)
	worker := mockrestaurant.NewWorker(client, cfg.PollInterval, cfg.StageDelay)

	// ctx для воркера отменяется при получении сигнала остановки — тогда он
	// перестаёт брать в работу новые заказы и не начинает новых шагов
	// конвейера для уже начатых (см. select на ctx.Done() в processOrder).
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	go worker.Run(workerCtx)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           mockrestaurant.NewServer(worker, client),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf(
			"mock-restaurant: restaurant_id=%d, слушаю :%s, опрашиваю %s каждые %s (шаг конвейера — %s)",
			cfg.RestaurantID, cfg.HTTPPort, cfg.CoreAPIURL, cfg.PollInterval, cfg.StageDelay,
		)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("mock-restaurant: shutdown signal received")
	cancelWorker()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("mock-restaurant: graceful shutdown failed: %v", err)
	}
}
