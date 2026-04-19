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

// UploadOrder обрабатывает загрузку номера заказа текущим пользователем.
//
// Хендлер принимает номер заказа в теле запроса в формате text/plain,
// вызывает сервис загрузки заказа и возвращает статус в зависимости
// от результата обработки.
//
// Возможные коды ответа:
//   - 200 -> номер уже был загружен этим пользователем
//   - 202 -> новый номер принят в обработку
//   - 400 -> тело запроса некорректно
//   - 401 -> пользователь не аутентифицирован
//   - 409 -> номер уже загружен другим пользователем
//   - 422 -> номер заказа не проходит валидацию
//   - 500 при внутренней ошибке сервера
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

// GetOrders возвращает список заказов, загруженных текущим пользователем.
//
// Хендлер получает userID из контекста, запрашивает список заказов в сервисе
// и отдает их в формате JSON. Заказы должны быть отсортированы от новых к старым,
// а время загрузки сериализуется в формате RFC3339.
//
// Возможные коды ответа:
//   - 200 -> список заказов успешно получен
//   - 204 -> у пользователя нет загруженных заказов
//   - 401 -> пользователь не аутентифицирован
//   - 500 -> при внутренней ошибке сервера
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
