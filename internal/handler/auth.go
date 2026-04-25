package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/xhrobj/gophermart/internal/service"
)

// authRequest описывает JSON-запрос с логином и паролем
// для регистрации или аутентификации пользователя.
type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register возвращает HTTP-хендлер для регистрации пользователя.
//
// Хендлер принимает JSON с логином и паролем, вызывает AuthService.Register
// и в случае успеха возвращает JWT-токен в заголовке Authorization.
//
// Возможные коды ответа:
//   - 200 -> пользователь успешно зарегистрирован и аутентифицирован
//   - 400 -> неверный JSON, невалидные регистрационные данные или слишком длинный пароль
//   - 409 -> логин уже занят
//   - 500 -> внутренняя ошибка сервера
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
			case errors.Is(err, service.ErrPasswordTooLong):
				const msg = "Пароль слишком длинный. Попробуйте более короткий пароль."
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(msg))
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

// Login возвращает HTTP-хендлер для аутентификации пользователя.
//
// Хендлер принимает JSON с логином и паролем, вызывает AuthService.Login
// и в случае успеха возвращает JWT-токен в заголовке Authorization.
//
// Возможные коды ответа:
//   - 200 -> пользователь успешно аутентифицирован
//   - 400 -> неверный формат запроса
//   - 401 -> неверная пара логин/пароль
//   - 500 -> внутренняя ошибка сервера
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
