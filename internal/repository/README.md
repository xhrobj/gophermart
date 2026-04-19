# Репо-интерфейсы

## `UserRepository` -> пользователи

```go
type UserRepository interface {
    Create(ctx context.Context, login, passwordHash string) (model.User, error)
    FindByLogin(ctx context.Context, login string) (model.User, error)
}
```

Используется сервисом:

- `AuthService`
  - `Register`
    - вызывает `UserRepository.Create`
  - `Login`
    - вызывает `UserRepository.FindByLogin`

Используемые таблицы БД:

- `users`
  - создать пользователя
  - найти пользователя по логину и получить хеш пароля

---

## `OrderRepository` -> заказы и accrual-обработка

```go
type OrderRepository interface {
    Create(ctx context.Context, userID int64, orderNumber string) (model.Order, error)
    FindByNumber(ctx context.Context, orderNumber string) (model.Order, error)
    ListByUserID(ctx context.Context, userID int64) ([]model.Order, error)

    ListPending(ctx context.Context, limit int) ([]model.Order, error)
    SetAccrualResult(ctx context.Context, orderNumber string, status model.OrderStatus, accrual int64) error
}
```

Используется сервисами:

- `OrderService`
  - `UploadOrder`
    - вызывает `OrderRepository.FindByNumber`
    - вызывает `OrderRepository.Create`

  - `ListOrders`
    - вызывает `OrderRepository.ListByUserID`

- `AccrualService`
  - `ProcessPendingOrders`
    - вызывает `OrderRepository.ListPending`
    - вызывает `OrderRepository.SetAccrualResult`

Используемые таблицы БД:

- `orders`
  - создать заказ
  - найти заказ по номеру и определить его владельца
  - получить список заказов пользователя

  - выбрать pending-заказы (в промежуточных статусах `NEW` и `PROCESSING`) для обработки через accrual
  - обновить статус заказа / сохранить сумму начисления

---

## `BalanceRepository` -> баланс и списания

```go
type BalanceRepository interface {
    GetBalance(ctx context.Context, userID int64) (model.Balance, error)
    Withdraw(ctx context.Context, userID int64, orderNumber string, sum int64) error
    ListWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error)
}
```

Используется сервисом:

- `BalanceService`
  - `GetBalance`
    - вызывает `BalanceRepository.GetBalance`

  - `Withdraw`
    - вызывает `BalanceRepository.Withdraw`

  - `ListWithdrawals`
    - вызывает `BalanceRepository.ListWithdrawals`

Используемые таблицы БД:

- `orders`
  - посчитать сумму начислений пользователя

- `withdrawals`
  - посчитать сумму всех списаний пользователя
  - создать новое списание
  - получить список списаний пользователя
