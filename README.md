# Авито.Кухня — Backend MVP

MVP бэкенд-сервиса для интеграции заведений общественного питания с площадкой Авито и обработки заказов пользователей на доставку еды.

> Упрощения, на которые сознательно пошли в MVP, помечены пометкой **[MVP]** — почему именно так и что изменится при масштабировании, см. в разделе 6.

---

## 1. Архитектура системы (C4 Model: Level 2 — Container Diagram)

```mermaid
C4Container
    title Container Diagram (C4 Level 2) — Сервис Авито.Кухня

    Person(customer, "Клиент", "Пользователь веб-версии Авито")
    Person_Ext(restaurant_staff, "Персонал заведения", "Сотрудники подключенного к платформе заведения")

    System_Boundary(c1, "Периметр Авито.Кухня") {
        Container(kitchen_api, "Avito.Kitchen Core Service", "Go, net/http", "REST API для клиентов и заведений: каталог, заказы, управление меню")
        ContainerDb(postgres_db, "PostgreSQL Database", "PostgreSQL 16", "Каталог заведений, меню, заказы и их состав")
    }

    System_Ext(mock_restaurant, "Mock Restaurant Service", "Go", "Эмулятор заведения: управляет доступностью блюд, опрашивает новые заказы и меняет их статус")

    Rel(customer, kitchen_api, "Просмотр каталога, оформление и отслеживание заказов", "HTTPS / JSON REST")
    Rel(mock_restaurant, kitchen_api, "Управление меню, опрос и обработка заказов", "HTTP / JSON REST")
    Rel(kitchen_api, postgres_db, "Чтение/запись, транзакции", "TCP / pgx (5432)")
```

**Почему один Core Service, а не отдельные сервисы каталога/заказов [MVP]:** домены каталога и заказов тесно связаны (снимок цены при заказе, проверка доступности), а нагрузка и команда на старте — одна. Границы внутри сервиса всё равно проведены на уровне кода (`domain` / `usecase` / `repository` по сущностям), поэтому вынести заказы в отдельный сервис в будущем — это в первую очередь вопрос деплоя, а не переписывания бизнес-логики.

---

## 2. Схема базы данных

```mermaid
erDiagram
    RESTAURANTS ||--o{ MENU_ITEMS : "содержит позиции"
    RESTAURANTS ||--o{ ORDERS : "принимает"
    ORDERS ||--|{ ORDER_ITEMS : "состоит из"
    MENU_ITEMS ||--o{ ORDER_ITEMS : "входит в"

    RESTAURANTS {
        bigserial id PK
        varchar name
        text address
        boolean is_active
        timestamp created_at
    }

    MENU_ITEMS {
        bigserial id PK
        bigint restaurant_id FK
        varchar name
        text description
        numeric price
        boolean is_available
        timestamp created_at
    }

    ORDERS {
        uuid id PK
        bigint restaurant_id FK
        bigint user_id
        order_status status
        numeric total_amount
        timestamp created_at
        timestamp updated_at
    }

    ORDER_ITEMS {
        bigserial id PK
        uuid order_id FK
        bigint menu_item_id FK
        int quantity
        numeric price_at_order
    }
```

### Архитектурные решения

1. **Стратегия первичных ключей.** `restaurants`, `menu_items`, `order_items` — `BIGSERIAL`: внутренние справочные сущности с частым чтением по индексу. `orders` — `UUID`: заказ передаётся клиенту как публичная ссылка (`GET /orders/{id}` без авторизации, см. раздел 6), и предсказуемый инкрементный id раскрывал бы объём заказов конкурентам и позволял перебором читать чужие заказы.

2. **Индексы.**
   - `idx_menu_items_restaurant_id` — фильтрация меню конкретного заведения.
   - `idx_orders_restaurant_id`, `idx_orders_status` — под опрос заведением своих заказов по статусу (`GET /restaurants/{id}/orders?status=CREATED`), самый частый запрос со стороны заведения.
   - `idx_order_items_order_id` — сборка состава заказа при чтении.

3. **Снимок цены (`order_items.price_at_order`).** Цена блюда на момент заказа фиксируется отдельно от `menu_items.price` — последующее изменение цены заведением не меняет сумму уже оформленных заказов.

4. **Целостность на уровне БД.** `CHECK (price >= 0)`, `CHECK (quantity > 0)`, `CHECK (total_amount >= 0)` — база не даёт записать финансово некорректное состояние, даже если в коде выше будет баг. Статус заказа — строгий `ENUM (order_status)`.

---

## 3. Пользовательские сценарии (CJM)

Заведение и клиент — разные роли с разными API-неймспейсами (см. раздел 4), поэтому и сценарии разведены на два независимых CJM.

### 3.1. CJM клиента: находит блюдо → делает заказ → отслеживает статус

```mermaid
sequenceDiagram
    autonumber
    actor Client as Клиент (Web)
    participant API as Avito.Kitchen Core API
    participant DB as PostgreSQL

    Client->>API: GET /api/v1/restaurants
    API->>DB: SELECT * FROM restaurants WHERE is_active = true
    API-->>Client: 200 OK [список активных заведений]

    Client->>API: GET /api/v1/restaurants/{id}/menu
    API->>DB: проверка is_active + SELECT меню
    alt заведение отключено
        API-->>Client: 409 Conflict (заведение временно не принимает заказы)
    else заведение не существует
        API-->>Client: 404 Not Found
    else
        API-->>Client: 200 OK [меню с ценами и доступностью]
    end

    Note over Client,API: Клиент собрал корзину, часть блюд может быть недоступна
    Client->>API: POST /api/v1/orders {restaurant_id, user_id, items[]}
    API->>DB: BEGIN → проверка ресторана и доступности блюд → INSERT order + items → COMMIT
    alt блюдо недоступно / нет в меню заведения
        API-->>Client: 409 / 404 + ErrorResponse (какое блюдо и почему)
    else заведение не найдено или отключено
        API-->>Client: 404 / 409 + ErrorResponse
    else
        API-->>Client: 201 Created {order_id, status: "CREATED", total_amount}
    end

    loop клиент отслеживает доставку
        Client->>API: GET /api/v1/orders/{id}
        API-->>Client: 200 OK {status: "COOKING" | "DELIVERING" | ...}
    end
```

**Трудности, которые сценарий явно обрабатывает:**
- блюдо пропало из наличия между открытием меню и оформлением заказа — заказ атомарно отклоняется целиком, с указанием конкретного блюда (`ErrOutOfStock`), а не создаётся частично;
- заведение отключилось (закрылось) — отдельно от "не существует", чтобы клиент понимал разницу между 404 и «сейчас недоступно»;
- заказ создаётся одной транзакцией (проверка наличия + запись заказа и его позиций) — исключает запись заказа без позиций или с несуществующим блюдом.

### 3.2. CJM заведения: управляет меню → принимает и готовит заказы

```mermaid
sequenceDiagram
    autonumber
    actor Staff as Персонал заведения
    participant Rest as Restaurant Service (mock)
    participant API as Avito.Kitchen Core API
    participant DB as PostgreSQL

    Note over Staff,Rest: Актуализация каталога
    Staff->>Rest: блюдо закончилось / изменилась цена
    Rest->>API: PATCH /api/v1/restaurants/{id}/menu/{itemId} {is_available: false}
    API->>DB: UPDATE menu_items SET is_available = false WHERE restaurant_id = {id} AND id = {itemId}
    API-->>Rest: 200 OK {обновлённая позиция меню}

    Note over Rest,API: Приём новых заказов (polling)
    loop каждые N секунд
        Rest->>API: GET /api/v1/restaurants/{id}/orders?status=CREATED
        API-->>Rest: 200 OK [новые заказы этого заведения]
    end

    Rest->>API: PATCH /api/v1/restaurants/{id}/orders/{orderId}/status {status: "ACCEPTED"}
    API->>DB: проверка CanTransitionTo(CREATED → ACCEPTED) + UPDATE
    API-->>Rest: 200 OK

    Rest->>API: PATCH .../status {status: "COOKING"}
    Rest->>API: PATCH .../status {status: "DELIVERING"}
    Rest->>API: PATCH .../status {status: "COMPLETED"}
    API-->>Rest: 200 OK (на каждый шаг)

    Note over Rest,API: Альтернативный сценарий — отмена
    Rest->>API: PATCH .../status {status: "CANCELLED"}
    alt заказ уже в доставке
        API-->>Rest: 409 Conflict (недопустимый переход статуса)
    else
        API-->>Rest: 200 OK
    end
```

**Почему статус-машина именно такая:**

```
CREATED ──▶ ACCEPTED ──▶ COOKING ──▶ DELIVERING ──▶ COMPLETED
   │            │            │
   └──────┴────────────┘
             ▼
         CANCELLED
```

- `ACCEPTED` выделен в отдельный шаг: заведение должно явно подтвердить, что взяло заказ в работу, до начала готовки — иначе клиент не отличит «никто не видел заказ» от «уже готовят».
- Отменить можно только до момента, когда заказ выехал на доставку (`DELIVERING`) — далее возможен только `COMPLETED`. Это доменное правило проверяется в `domain.OrderStatus.CanTransitionTo` до обращения к БД.
- Каждая ручка смены статуса и опроса заказов принимает `restaurant_id` в пути (`/restaurants/{id}/orders/...`). Явной авторизации в MVP нет (см. раздел 6), но usecase-слой сверяет `order.RestaurantID` с `{id}` из пути: попытка изменить чужой заказ возвращает `404`, как будто заказа с таким id для этого заведения не существует — так минимальная изоляция между заведениями есть уже сейчас, без полноценного AuthZ.

---

## 4. API

Два независимых неймспейса поверх одного сервиса — это ответ на «в чём разница между API для клиентов и API для заведений» из ТЗ:

| Неймспейс | Кто вызывает | Ручки |
|---|---|---|
| `/api/v1/restaurants`, `/api/v1/orders` | Клиентская веб-часть | просмотр каталога, оформление и просмотр своего заказа |
| `/api/v1/restaurants/{id}/orders`, `/api/v1/restaurants/{id}/menu/{itemId}` | Заведение (mock-сервис) | опрос своих заказов, смена статуса, управление доступностью/ценой позиций меню |

Полная спецификация — [`docs/openapi.yaml`](docs/openapi.yaml).

### Маппинг доменных ошибок в HTTP-статусы

Единая точка правды для будущих хендлеров (`internal/domain/errors.go` → HTTP):

| Доменная ошибка | HTTP | Смысл |
|---|---|---|
| `ErrInvalidInput` | 400 | Некорректный запрос (пустая корзина, нулевое количество и т.п.) |
| `ErrRestaurantNotFound`, `ErrMenuItemNotFound`, `ErrOrderNotFound` | 404 | Ресурс не существует (в т.ч. заказ другого заведения — см. раздел 3.2) |
| `ErrRestaurantInactive`, `ErrOutOfStock`, `ErrInvalidStatusTransition`, `ErrAlreadyExists` | 409 | Ресурс существует, но операция противоречит его текущему состоянию |
| `ErrInternal` / прочее необёрнутое | 500 | Непредвиденная ошибка |

Ошибки оборачиваются через `fmt.Errorf("%w: ...")`, поэтому хендлер разбирает их через `errors.Is` по этой таблице, не парся текст.

---

## 5. Быстрый запуск

### Требования
- Docker и Docker Compose

```bash
git clone <repo-url> avito-kitchen
cd avito-kitchen
docker compose up --build
```

Сервисы:
- **Avito.Kitchen Core API:** `http://localhost:8080`
- **Mock Restaurant Service:** `http://localhost:8081`
- **PostgreSQL:** `localhost:5432`

### Статус реализации

Слои `domain` / `usecase` / `repository`, миграции и HTTP API (`cmd/api`, `internal/httpserver`) готовы и покрывают сценарии из раздела 3 — все запросы из обоих CJM вручную прогнаны через поднятый локально сервис. В работе:

- [ ] Mock Restaurant Service (раздел 3.2, отдельный процесс в `cmd/mock-restaurant`)
- [ ] `docker-compose.yml` и `Dockerfile` для обоих сервисов + прогон миграций при старте
- [ ] Юнит-тесты usecase-слоя (репозитории уже спрятаны за интерфейсами в `domain`, поэтому мокаются без поднятия БД) и интеграционные тесты репозиториев
- [ ] `.golangci.yml`

Локальный запуск без Docker (пока не готов `docker-compose.yml`):

```bash
# Poднять Postgres любым способом и применить миграции из migrations/*.sql,
# затем:
DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres \
DB_NAME=avito_kitchen DB_SSLMODE=disable HTTP_PORT=8080 \
  go run ./cmd/api
```

---

## 6. Допущения MVP и план масштабирования

| Упрощение в MVP | Почему это приемлемо сейчас | Что меняется при масштабировании |
|---|---|---|
| Нет аутентификации/авторизации (по условию задания) | Задание явно исключает её из скоупа | Заведения получают API-ключ/JWT; `restaurant_id` в пути перестаёт быть просто конвенцией и проверяется на уровне middleware, а не только в usecase |
| Опрос заказов заведением через polling | Заведений мало, нагрузка небольшая, не нужна брокер-инфраструктура | Переход на событийную модель (Kafka/RabbitMQ: `OrderCreated`, `OrderStatusChanged`) вместо поллинга |
| Разрешён гипотетический гонки-кейс между проверкой доступности блюда и записью заказа (нет `SELECT … FOR UPDATE`) | Вероятность коллизии на объёме MVP пренебрежимо мала, а `IsAvailable` — не физический остаток, а ручной тумблер заведения | Блокировка строк меню в транзакции создания заказа или полноценный резерв позиций |
| Список заведений вносится вручную через миграцию-сид (закрытый доступ по условию задания) | Соответствует «закрытому списку заведений» из ТЗ на MVP | Полноценный onboarding-флоу заведения через отдельную ручку/личный кабинет |
| Один относительно крупный сервис вместо каталога/заказов/уведомлений по отдельности | Один домен данных, одна команда, нагрузка не требует независимого масштабирования частей | Вынесение заказов в отдельный сервис поверх той же доменной модели, gRPC для межсервисного вызова, PgBouncer/read-реплики под БД, Redis-кэш каталога |
