package handler

import (
	"fmt"
	"net/http"

	"github.com/xhrobj/gophermart/internal/middleware"
)

/*

curl -i -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"login": "admin","password": "god"}'

curl -i http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer jwt"

*/

func GetOrders() http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		if rq.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := middleware.UserIDFromContext(rq.Context())
		if !ok {
			http.Error(w, "(о_0) user is not authorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte(fmt.Sprintf("userId=%d", userID)))
	}
}
