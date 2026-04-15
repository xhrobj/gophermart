# Сервисные интерфейсы

## `AuthService`

```go
type AuthService interface {
    Register(ctx context.Context, login, password string) (model.AuthResult, error)
    Login(ctx context.Context, login, password string) (model.AuthResult, error)
}
```

Используется ручками:

- `POST /api/user/register`
  - метод сервиса: `AuthService.Register`
- `POST /api/user/login`
  - метод сервиса: `AuthService.Login`

Используемые таблицы БД:

- `users`
  - создать пользователя;
  - найти пользователя по логину;
  - получить хеш пароля;

---

## `OrderService`

```go
type OrderService interface {
    UploadOrder(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error)
    ListOrders(ctx context.Context, userID int64) ([]model.Order, error)
}
```

Используется ручками:

- `POST /api/user/orders`
  - метод сервиса: `OrderService.UploadOrder`
- `GET /api/user/orders`
  - метод сервиса: `OrderService.ListOrders`

Используемые таблицы БД:

- `orders`
  - создать заказ;
  - проверить владельца заказа;
  - получить список заказов пользователя;

---

## `BalanceService`

```go
type BalanceService interface {
    GetBalance(ctx context.Context, userID int64) (model.Balance, error)
    Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error
    ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}
```

Используется ручками:

- `GET /api/user/balance`
  - метод сервиса: `BalanceService.GetBalance`
- `POST /api/user/balance/withdraw`
  - метод сервиса: `BalanceService.Withdraw`
- `GET /api/user/withdrawals`
  - метод сервиса: `BalanceService.ListWithdrawals`

Используемые таблицы БД:

- `orders`
  - посчитать начисления;
- `withdrawals`
  - посчитать списания;
  - создать новое списание;
  - получить списания пользователя;

---

## `AccrualService`

```go
type AccrualService interface {
    ProcessPendingOrders(ctx context.Context) error
}
```

Используется:

- фоновым обработчиком заказов, который периодически проверяет заказы во внешнем сервисе accrual.

Межсервисный запрос к accrual:

- `GET /api/orders/{number}`

Таблицы БД:

- `orders`
  - выбрать pending-заказы (статусы `NEW` и `PROCESSING`);
  - обновить статус;
  - сохранить начисление;
