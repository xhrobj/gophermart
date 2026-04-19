package handler

import (
	"encoding/json"
	"net/http"

	"github.com/xhrobj/gophermart/internal/middleware"
	"github.com/xhrobj/gophermart/internal/service"
)

type getBalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

/*

#### **Получение текущего баланса пользователя**

Хендлер: `GET /api/user/balance`.

Хендлер доступен только авторизованному пользователю. В ответе должны содержаться данные о текущей сумме баллов лояльности, а также сумме использованных за весь период регистрации баллов.

Формат запроса:

```
GET /api/user/balance HTTP/1.1
Content-Length: 0
```

Возможные коды ответа:

- `200` — успешная обработка запроса.

  Формат ответа:

    ```
    200 OK HTTP/1.1
    Content-Type: application/json
    ...

    {
    	"current": 500.5,
    	"withdrawn": 42
    }
    ```

- `401` — пользователь не авторизован.
- `500` — внутренняя ошибка сервера.

*/

/*

curl -i http://localhost:8080/api/user/balance \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsImV4cCI6MTc3NjY4NTQ4NSwiaWF0IjoxNzc2NTk5MDg1fQ.wiyicfgtpgEEhA3xQisZyvy9ov_DSBZBpuh_Ssgqy6A"

HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 28

{"current":0,"withdrawn":0}

*/

// GetBalance возвращает текущий баланс пользователя и сумму всех списаний.
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
