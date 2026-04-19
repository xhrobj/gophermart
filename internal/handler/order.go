package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/xhrobj/gophermart/internal/middleware"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/service"
)

type getOrdersResponseItem struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

/*

#### **Загрузка номера заказа**

Хендлер: `POST /api/user/orders`.

Хендлер доступен только аутентифицированным пользователям. Номером заказа является последовательность цифр произвольной длины.

Номер заказа может быть проверен на корректность ввода с помощью [алгоритма Луна](https://ru.wikipedia.org/wiki/Алгоритм_Луна){target="_blank"}.

Формат запроса:

```
POST /api/user/orders HTTP/1.1
Content-Type: text/plain
...

12345678903
```

Возможные коды ответа:

- `200` — номер заказа уже был загружен этим пользователем;
- `202` — новый номер заказа принят в обработку;
- `400` — неверный формат запроса;
- `401` — пользователь не аутентифицирован;
- `409` — номер заказа уже был загружен другим пользователем;
- `422` — неверный формат номера заказа;
- `500` — внутренняя ошибка сервера.

*/

func UploadOrder(orderService service.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		if rq.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		userID, ok := middleware.UserIDFromContext(rq.Context())
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(rq.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result, err := orderService.UploadOrder(rq.Context(), userID, string(body))
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidOrderInput):
				w.WriteHeader(http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidOrderNumber):
				w.WriteHeader(http.StatusUnprocessableEntity)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		switch result.Status {
		case model.UploadOrderAccepted:
			w.WriteHeader(http.StatusAccepted)
		case model.UploadOrderDuplicate:
			w.WriteHeader(http.StatusOK)
		case model.UploadOrderConflict:
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

/*

#### **Получение списка загруженных номеров заказов**

Хендлер: `GET /api/user/orders`.

Хендлер доступен только авторизованному пользователю. Номера заказа в выдаче должны быть отсортированы по времени загрузки от самых новых к самым старым. Формат даты — RFC3339.

Доступные статусы обработки расчётов:

- `NEW` — заказ загружен в систему, но не попал в обработку;
- `PROCESSING` — вознаграждение за заказ рассчитывается;
- `INVALID` — система расчёта вознаграждений отказала в расчёте;
- `PROCESSED` — данные по заказу проверены и информация о расчёте успешно получена.

Формат запроса:

```
GET /api/user/orders HTTP/1.1
Content-Length: 0
```

Возможные коды ответа:

- `200` — успешная обработка запроса.

  Формат ответа:

    ```
    200 OK HTTP/1.1
    Content-Type: application/json
    ...

    [
    	{
            "number": "9278923470",
            "status": "PROCESSED",
            "accrual": 500,
            "uploaded_at": "2020-12-10T15:15:45+03:00"
        },
        {
            "number": "12345678903",
            "status": "PROCESSING",
            "uploaded_at": "2020-12-10T15:12:01+03:00"
        },
        {
            "number": "346436439",
            "status": "INVALID",
            "uploaded_at": "2020-12-09T16:09:53+03:00"
        }
    ]
    ```

- `204` — нет данных для ответа.
- `401` — пользователь не авторизован.
- `500` — внутренняя ошибка сервера.

*/

/*

curl -i -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"login": "admin","password": "god"}'

curl -i http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer jwt"

*/

/*

curl -i http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTc3NjY4NTQ4NSwiaWF0IjoxNzc2NTk5MDg1fQ.wiyicfgtpgEEhA3xQisZyvy9ov_DSBZBpuh_Ssgqy6A"

HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 84

[{"number":"12345678903","status":"NEW","uploaded_at":"2026-04-18T20:27:42+03:00"}]

*/

func GetOrders(orderService service.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		if rq.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		userID, ok := middleware.UserIDFromContext(rq.Context())
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		orders, err := orderService.ListOrders(rq.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		result := make([]getOrdersResponseItem, 0, len(orders))
		for _, order := range orders {
			result = append(result, buildGetOrdersResponseItem(order))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err = json.NewEncoder(w).Encode(result); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func buildGetOrdersResponseItem(order model.Order) getOrdersResponseItem {
	item := getOrdersResponseItem{
		Number:     order.Number,
		Status:     string(order.Status),
		UploadedAt: order.UploadedAt.Format(time.RFC3339),
	}

	if order.Status == model.OrderStatusProcessed {
		accrual := hundredthsToAmount(order.Accrual)
		item.Accrual = &accrual
	}

	return item
}
