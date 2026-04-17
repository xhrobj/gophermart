package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/xhrobj/gophermart/internal/service"
)

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

/*

#### **Регистрация пользователя**

Хендлер: `POST /api/user/register`.

Регистрация производится по паре логин/пароль. Каждый логин должен быть уникальным.
После успешной регистрации должна происходить автоматическая аутентификация пользователя.

Формат запроса:

```
POST /api/user/register HTTP/1.1
Content-Type: application/json
...

{
	"login": "<login>",
	"password": "<password>"
}
```

Возможные коды ответа:

- `200` — пользователь успешно зарегистрирован и аутентифицирован;
- `400` — неверный формат запроса;
- `409` — логин уже занят;
- `500` — внутренняя ошибка сервера.

*/

/*

curl -i -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{"login": "admin","password": "god"}'

*/

func Register(authService service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req authRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result, err := authService.Register(r.Context(), req.Login, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidAuthInput):
				w.WriteHeader(http.StatusBadRequest)
			case errors.Is(err, service.ErrLoginAlreadyExists):
				w.WriteHeader(http.StatusConflict)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Authorization", "Bearer "+result.Token)
		w.WriteHeader(http.StatusOK)
	}
}

/*

#### **Аутентификация пользователя**

Хендлер: `POST /api/user/login`.

Аутентификация производится по паре логин/пароль.

Формат запроса:

```
POST /api/user/login HTTP/1.1
Content-Type: application/json
...

{
	"login": "<login>",
	"password": "<password>"
}
```

Возможные коды ответа:

- `200` — пользователь успешно аутентифицирован;
- `400` — неверный формат запроса;
- `401` — неверная пара логин/пароль;
- `500` — внутренняя ошибка сервера.

*/

/*

curl -i -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"login": "admin","password": "god"}'

*/

func Login(authService service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req authRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result, err := authService.Login(r.Context(), req.Login, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidAuthInput):
				w.WriteHeader(http.StatusBadRequest)
			case errors.Is(err, service.ErrInvalidCredentials):
				w.WriteHeader(http.StatusUnauthorized)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Authorization", "Bearer "+result.Token)
		w.WriteHeader(http.StatusOK)
	}
}
