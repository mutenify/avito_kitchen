package mockrestaurant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// CoreClient — тонкий HTTP-клиент к ресторанному API основного сервиса
// (см. docs/openapi.yaml, ручки /api/v1/restaurants/{restaurantId}/...).
// Всё, что он умеет, — это то же самое, что мог бы сделать curl: собрать
// запрос, проверить код ответа, распарсить JSON.
type CoreClient struct {
	baseURL      string
	restaurantID int64
	httpClient   *http.Client
}

func NewCoreClient(baseURL string, restaurantID int64) *CoreClient {
	return &CoreClient{
		baseURL:      baseURL,
		restaurantID: restaurantID,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// GetOrdersByStatus — GET /api/v1/restaurants/{id}/orders?status=...
func (c *CoreClient) GetOrdersByStatus(ctx context.Context, status string) ([]Order, error) {
	u := fmt.Sprintf("%s/api/v1/restaurants/%d/orders?status=%s", c.baseURL, c.restaurantID, url.QueryEscape(status))

	var orders []Order
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

// UpdateOrderStatus — PATCH /api/v1/restaurants/{id}/orders/{orderId}/status
func (c *CoreClient) UpdateOrderStatus(ctx context.Context, orderID, status string) (*Order, error) {
	u := fmt.Sprintf("%s/api/v1/restaurants/%d/orders/%s/status", c.baseURL, c.restaurantID, orderID)

	var order Order
	if err := c.doJSON(ctx, http.MethodPatch, u, map[string]string{"status": status}, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// GetMenu — GET /api/v1/restaurants/{id}/menu.
// Ручка формально клиентская (без авторизации в MVP доступна всем), но
// заведению тоже удобно посмотреть на свой каталог тем же способом, каким
// его видит покупатель — например, чтобы решить, какую позицию снять с продажи.
func (c *CoreClient) GetMenu(ctx context.Context) ([]MenuItem, error) {
	u := fmt.Sprintf("%s/api/v1/restaurants/%d/menu", c.baseURL, c.restaurantID)

	var menu []MenuItem
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &menu); err != nil {
		return nil, err
	}
	return menu, nil
}

// UpdateMenuItem — PATCH /api/v1/restaurants/{id}/menu/{itemId}.
// isAvailable и price — указатели по той же причине, что и в
// domain.UpdateMenuItemRequest на стороне Core API: nil значит "поле не
// меняем", а не "сбросить в zero value".
func (c *CoreClient) UpdateMenuItem(ctx context.Context, itemID int64, isAvailable *bool, price *string) (*MenuItem, error) {
	u := fmt.Sprintf("%s/api/v1/restaurants/%d/menu/%d", c.baseURL, c.restaurantID, itemID)

	body := struct {
		IsAvailable *bool   `json:"is_available,omitempty"`
		Price       *string `json:"price,omitempty"`
	}{IsAvailable: isAvailable, Price: price}

	var item MenuItem
	if err := c.doJSON(ctx, http.MethodPatch, u, body, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// doJSON — общая механика запроса: сериализовать тело (если есть), отправить,
// на код ответа >= 400 вернуть ошибку с текстом из ErrorResponse Core API
// (см. docs/openapi.yaml), иначе распарсить тело в dst.
func (c *CoreClient) doJSON(ctx context.Context, method, reqURL string, body any, dst any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call core api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		var apiErr errorResponse
		_ = json.NewDecoder(resp.Body).Decode(&apiErr) // best-effort — тело может быть пустым
		return fmt.Errorf("core api %s %s -> %d: %s", method, reqURL, resp.StatusCode, apiErr.Error)
	}

	if dst == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
