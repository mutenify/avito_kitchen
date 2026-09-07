# Авито.Кухня — Backend MVP

MVP масштабируемого бэкенд-сервиса для интеграции заведений общественного питания и обработки заказов пользователей в рамках экосистемы Авито.

---

## 1. Архитектура системы (C4 Model: Level 2 — Container Diagram)

Система построена на принципах микросервисной архитектуры с изоляцией контекстов и данных:

```mermaid
C4Container
    title Container Diagram (C4 Level 2) — Сервис Авито.Кухня

    Person(customer, "Клиент", "Пользователь веб-версии Авито")
    Person_Ext(restaurant_staff, "Персонал ресторана", "Сотрудники подключенного заведения")

    System_Boundary(c1, "Периметр Авито.Кухня") {
        Container(kitchen_api, "Avito.Kitchen Core Service", "Go, net/http / Chi", "Предоставляет REST API для клиентов и заведений, управляет заказами и каталогом")
        ContainerDb(postgres_db, "PostgreSQL Database", "PostgreSQL 16", "Хранит каталог заведений, меню, историю заказов и транзакции")
    }

    System_Ext(mock_restaurant, "Mock Restaurant Service", "Go", "Сервис-эмулятор заведения: опрашивает API и обновляет статусы заказов")

    Rel(customer, kitchen_api, "Просмотр каталога, оформление заказов", "HTTPS / JSON REST")
    Rel(mock_restaurant, kitchen_api, "Опрос заказов, смена статусов (COOKING, COMPLETED)", "HTTP / JSON REST")
    Rel(kitchen_api, postgres_db, "Чтение/запись данных, транзакции", "TCP / pgx (Port 5432)")
```

---

## 2. Описание схемы базы данных и индексов

### ER-диаграмма сущностей:

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

### Архитектурные решения по БД:

1. **Стратегия первичных ключей:**
   * Для таблиц `restaurants`, `menu_items` и `order_items` выбран `BIGSERIAL (int64)`. Это публичные справочники с частыми операциями чтения.
   * Для таблицы `orders` выбран `UUIDv4`. Заказы создаются клиентами и передаются по публичным ссылкам. UUID защищает от атак перебора и утечки объема продаж компании конкурентам.

2. **Индексы и производительность:**
   * `idx_menu_items_restaurant_id`: B-Tree индекс на внешний ключ для мгновенной фильтрации меню конкретного ресторана.
   * `idx_orders_restaurant_id` и `idx_orders_status`: Композитный доступ для быстрого пуллинга ресторанами заказов в статусе `CREATED`.
   * `idx_order_items_order_id`: Ускоряет выборку состава чека при детальном запросе заказа.

3. **Снимок цен (`price_at_order`):**
   * В таблице `order_items` сохраняется цена блюда на момент оформления. Это гарантирует финансовую неизменяемость чека при последующем изменении цен рестораном в каталоге `menu_items`.

4. **Целостность данных (Integrity & Constraints):**
   * Ограничения `CHECK (price >= 0)` и `CHECK (quantity > 0)` на уровне БД блокируют запись некорректных финансовых значений.
   * Статусы заказов типизированы через строгий `ENUM` (`order_status`), исключающий запись невалидного состояния.

---

## 3. Пользовательский путь (CJM)

```mermaid
sequenceDiagram
    autonumber
    actor Client as Клиент (Web)
    participant API as Авито.Кухня (Core API)
    participant DB as PostgreSQL
    participant Rest as Mock Restaurant

    Note over Client, API: Сценарий выбора и заказа
    Client->>API: GET /api/v1/restaurants
    API->>DB: SELECT * FROM restaurants WHERE is_active = true
    DB-->>API: Заведения
    API-->>Client: 200 OK [список ресторанов]

    Client->>API: GET /api/v1/restaurants/{id}/menu
    API->>DB: SELECT * FROM menu_items WHERE restaurant_id = {id}
    DB-->>API: Позиции меню
    API-->>Client: 200 OK [меню с ценами и доступностью]

    Client->>API: POST /api/v1/orders (restaurant_id, items)
    API->>DB: BEGIN Transaction -> Проверка остатков -> INSERT order -> COMMIT
    DB-->>API: Order ID (UUID)
    API-->>Client: 201 Created {"order_id": "...", "status": "CREATED"}

    Note over API, Rest: Сценарий обработки рестораном
    loop Каждые 5 секунд (Polling)
        Rest->>API: GET /api/v1/restaurants/{id}/orders?status=CREATED
        API-->>Rest: 200 OK [новые заказы]
    end

    Rest->>API: PATCH /api/v1/orders/{id}/status {"status": "COOKING"}
    API->>DB: UPDATE orders SET status = 'COOKING'
    API-->>Rest: 200 OK

    Rest->>API: PATCH /api/v1/orders/{id}/status {"status": "DELIVERING"}
    Rest->>API: PATCH /api/v1/orders/{id}/status {"status": "COMPLETED"}

    Note over Client, API: Клиент отслеживает доставку
    Client->>API: GET /api/v1/orders/{id}
    API-->>Client: 200 OK {"status": "COMPLETED"}
```

---

## 4. Быстрый запуск

### Требования:
- Docker и Docker Compose

```bash
# Клонировать репозиторий
git clone git@github.com:talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-mutenify-d47ac84f.git
cd backend-trainee-assignment-autumn-2026-flow-2-mutenify-d47ac84f

# Запустить проект и базу данных одной командой
docker compose up --build
```

Сервисы будут доступны:
- **Авито.Кухня API:** `http://localhost:8080`
- **Mock Restaurant:** `http://localhost:8081`
- **PostgreSQL:** `localhost:5432`

---

## 5. Планы по масштабированию (Target Architecture)

Для масштабирования решения за пределы MVP предусмотрены следующие шаги:
1. **Event-Driven Architecture (Apache Kafka / RabbitMQ):** Переход от опроса к асинхронной публикации событий `OrderCreated`, `OrderStatusChanged` для надежной доставки уведомлений ресторанам.
2. **gRPC для внутреннего взаимодействия:** Межсервисный обмен между сервисом кухни и партнерскими системами заведений переводится на бинарный gRPC.
3. **Кэширование каталога (Redis):** Меню ресторанов кэшируются в RAM с инвалидацией по событию обновления меню ресторатором.
4. **Масштабирование БД:** Подключение пулера соединений **PgBouncer**.
