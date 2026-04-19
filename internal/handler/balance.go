package handler

import (
	"encoding/json"
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
