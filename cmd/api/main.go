// Command api запускает Avito.Kitchen Core Service: HTTP API поверх
// PostgreSQL, описанный в README и docs/openapi.yaml.
//
// main.go нарочно не содержит бизнес-логики — только сборку зависимостей
// (repository → usecase → httpserver) и управление жизненным циклом
// процесса (запуск сервера, graceful shutdown по сигналу).
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

	"avito-kitchen/internal/config"
	"avito-kitchen/internal/httpserver"
	"avito-kitchen/internal/repository"
	"avito-kitchen/internal/usecase"
)

func main() {
	cfg := config.Load()

	db, err := repository.NewPostgresDB(repository.PostgresConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		// NewPostgresDB сам делает Ping с таймаутом (см. repository/postgres.go),
		// поэтому если мы дошли сюда с ошибкой — Postgres на момент старта
		// либо недоступен, либо не готов принимать соединения. Без базы сервис
		// бесполезен, поэтому падаем сразу, а не поднимаемся в нерабочем виде.
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// --- Сборка слоёв: repository -> usecase -> httpserver ---
	restaurantRepo := repository.NewRestaurantRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	restaurantUC := usecase.NewRestaurantUsecase(restaurantRepo)
	orderUC := usecase.NewOrderUsecase(orderRepo, restaurantRepo)

	router := httpserver.NewRouter(
		httpserver.NewRestaurantHandler(restaurantUC),
		httpserver.NewOrderHandler(orderUC),
		httpserver.NewRestaurantOrderHandler(orderUC),
	)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second, // защита от slowloris-подобных запросов
	}

	// Сервер запускаем в отдельной горутине, чтобы основная могла ждать
	// сигнал остановки и сделать graceful shutdown.
	go func() {
		log.Printf("avito-kitchen api listening on :%s", cfg.HTTPPort)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	waitForShutdown(server)
}

// waitForShutdown блокируется до SIGINT/SIGTERM (в т.ч. `docker compose stop`)
// и затем даёт серверу время доработать уже начатые запросы вместо того,
// чтобы обрывать соединения мгновенно.
func waitForShutdown(server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutdown signal received, draining active requests")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
