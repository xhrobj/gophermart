package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/xhrobj/gophermart/internal/middleware"
	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/service"
)

type getBalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type getWithdrawalsResponseItem struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

// GetBalance возвращает текущий баланс пользователя и общую сумму списаний.
//
// Хендлер получает userID из контекста, запрашивает баланс в сервисе
// и отдает ответ в формате JSON.
//
// Возможные коды ответа:
//   - 200 -> баланс успешно получен
//   - 401 -> пользователь не аутентифицирован
//   - 500 -> при внутренней ошибке сервера
func GetBalance(balanceService service.BalanceService) http.HandlerFunc {
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

		balance, err := balanceService.GetBalance(rq.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := getBalanceResponse{
			Current:   hundredthsToAmount(balance.Current),
			Withdrawn: hundredthsToAmount(balance.Withdrawn),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(response)
	}
}

// Withdraw списывает баллы с бонусного счета пользователя.
//
// Хендлер принимает JSON с номером заказа и суммой списания,
// вызывает BalanceService.Withdraw и возвращает статус в зависимости
// от результата обработки.
//
// Возможные коды ответа:
// - 200 -> списание успешно выполнено
// - 400 -> неверный формат запроса
// - 401 -> пользователь не аутентифицирован
// - 402 -> недостаточно средств
// - 422 -> номер заказа не проходит валидацию
// - 500 -> внутренняя ошибка сервера
func Withdraw(balanceService service.BalanceService) http.HandlerFunc {
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

		var req withdrawRequest
		if err := json.NewDecoder(rq.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sum, err := amountToHundredths(req.Sum)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = balanceService.Withdraw(rq.Context(), userID, req.Order, sum)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidWithdrawOrderNumber):
				w.WriteHeader(http.StatusUnprocessableEntity)
			case errors.Is(err, service.ErrInvalidWithdrawSum):
				w.WriteHeader(http.StatusBadRequest)
			case errors.Is(err, service.ErrInsufficientFunds):
				w.WriteHeader(http.StatusPaymentRequired)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// GetWithdrawals возвращает историю списаний текущего пользователя.
//
// Хендлер получает userID из контекста, запрашивает список списаний в сервисе
// и отдает их в формате JSON. Списания должны быть отсортированы от новых к старым,
// а время обработки сериализуется в формате RFC3339.
//
// Возможные коды ответа:
//   - 200 -> список списаний успешно получен
//   - 204 -> у пользователя нет списаний
//   - 401 -> пользователь не аутентифицирован
//   - 500 -> при внутренней ошибке сервера
func GetWithdrawals(balanceService service.BalanceService) http.HandlerFunc {
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

		withdrawals, err := balanceService.ListWithdrawals(rq.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		result := make([]getWithdrawalsResponseItem, 0, len(withdrawals))
		for _, withdrawal := range withdrawals {
			result = append(result, buildGetWithdrawalsResponseItem(withdrawal))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err = json.NewEncoder(w).Encode(result); err != nil {
			return
		}
	}
}

func buildGetWithdrawalsResponseItem(withdrawal model.Withdrawal) getWithdrawalsResponseItem {
	return getWithdrawalsResponseItem{
		Order:       withdrawal.OrderNumber,
		Sum:         hundredthsToAmount(withdrawal.Sum),
		ProcessedAt: withdrawal.ProcessedAt.Format(time.RFC3339),
	}
}
