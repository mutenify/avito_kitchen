package mockrestaurant

import (
	"os"
	"strconv"
	"time"
)

// Config — настройки сервиса-эмулятора заведения, из переменных окружения.
// Один процесс = одно заведение: RestaurantID фиксирован на запуск, что
// соответствует "закрытому списку заведений" из ТЗ (см. README раздел 6) —
// каждое реальное заведение получало бы свой собственный процесс/контейнер
// с собственным RestaurantID.
type Config struct {
	CoreAPIURL   string
	RestaurantID int64
	HTTPPort     string

	// PollInterval — как часто опрашивать Core API на предмет новых заказов.
	PollInterval time.Duration
	// StageDelay — пауза перед каждым шагом конвейера статусов
	// (ACCEPTED -> COOKING -> DELIVERING -> COMPLETED), имитирующая время
	// на реальную готовку и доставку. По умолчанию короткая — чтобы весь
	// цикл заказа было видно за секунды, а не за реальные полчаса.
	StageDelay time.Duration
}

func LoadConfig() Config {
	return Config{
		CoreAPIURL:   getEnv("CORE_API_URL", "http://localhost:8080"),
		RestaurantID: getEnvInt64("RESTAURANT_ID", 1),
		HTTPPort:     getEnv("HTTP_PORT", "8081"),
		PollInterval: getEnvDuration("POLL_INTERVAL", 5*time.Second),
		StageDelay:   getEnvDuration("STAGE_DELAY", 5*time.Second),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	v, err := strconv.ParseInt(os.Getenv(key), 10, 64)
	if err != nil {
		return fallback
	}
	return v
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return d
}
