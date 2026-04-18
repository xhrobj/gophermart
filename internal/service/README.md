# Сервисные интерфейсы

## `AuthService` -> регистрация и логин

```go
type AuthService interface {
    Register(ctx context.Context, login, password string) (model.AuthResult, error)
    Login(ctx context.Context, login, password string) (model.AuthResult, error)
}
```

Используется ручками:

- `POST /api/user/register`
  - вызывает: `AuthService.Register`

- `POST /api/user/login`
  - вызывает: `AuthService.Login`

---

## `OrderService` -> загрузка и список заказов пользователя

```go
type OrderService interface {
    UploadOrder(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error)
    ListOrders(ctx context.Context, userID int64) ([]model.Order, error)
}
```

Используется ручками:

- `POST /api/user/orders`
  - вызывает: `OrderService.UploadOrder`

- `GET /api/user/orders`
  - вызывает: `OrderService.ListOrders`

---

## `BalanceService` -> баланс и списания

```go
type BalanceService interface {
    GetBalance(ctx context.Context, userID int64) (model.Balance, error)
    Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error
    ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}
```

Используется ручками:

- `GET /api/user/balance`
  - вызывает: `BalanceService.GetBalance`

- `POST /api/user/balance/withdraw`
  - вызывает: `BalanceService.Withdraw`

- `GET /api/user/withdrawals`
  - вызывает: `BalanceService.ListWithdrawals`

---

## `AccrualService` -> фоновая обработка заказов через accrual

```go
type AccrualService interface {
    ProcessPendingOrders(ctx context.Context) error
}
```

Используется:

- фоновым обработчиком заказов, который периодически проверяет заказы во внешнем сервисе accrual.

Межсервисный запрос к accrual:

- `GET /api/orders/{number}`
