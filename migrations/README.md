# Gophermart: схема БД и использование полей

1. `users` - пользователи системы
2. `orders` - загруженные пользователями номера заказов и результаты начислений
3. `withdrawals` - операции списания

## Таблица `users`

Таблица хранит зарегистрированных пользователей Gophermart.

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY, --  Внутренний идентификатор пользователя
    login TEXT NOT NULL UNIQUE, -- Логин пользователя
    password_hash TEXT NOT NULL, -- Хеш пароля пользователя
    created_at TIMESTAMPTZ NOT NULL DEFAULT now() -- Дата и время регистрации; в API напрямую не используется
);
```

## Enum `order_status` для статусов обработки заказов

```sql
CREATE TYPE order_status AS ENUM (
    'NEW', -- Заказ загружен в систему, но ещё не отправлен/не попал в обработку
    'PROCESSING', -- Заказ находится в обработке у внешнего сервиса accrual
    'INVALID', -- Внешний сервис отказал в расчёте - начисление по заказу невозможно; финальный статус
    'PROCESSED' -- Расчет завершён - начисление успешно рассчитано; финальный статус
);
```

## Таблица `orders`

Таблица хранит номера заказов, которые пользователи передали в систему, а также статус их обработки и сумму начисленных баллов.

```sql
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY, -- Внутренний идентификатор заказа
    number TEXT NOT NULL UNIQUE, -- Номер заказа
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Владелец заказа
    status order_status NOT NULL DEFAULT 'NEW', -- Статус обработки заказа 
    accrual NUMERIC(19,2) NOT NULL DEFAULT 0, -- Начисленные баллы
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT now(), -- Время загрузки заказа пользователем

    next_poll_at TIMESTAMPTZ NOT NULL DEFAULT now() -- Время следующего запроса во внешний accrual
);
```

Бизнес-правила:

- Один номер заказа не может быть загружен повторно другим пользователем -> `number TEXT NOT NULL UNIQUE`
- Если заказ уже загружал тот же пользователь, нужно вернуть `200` -> Поиск заказа по `number` и сравнение `user_id`
- Если заказ уже загружал другой пользователь, нужно вернуть `409` -> Поиск заказа по `number` и сравнение `user_id`
- Заказы пользователя должны быть отсортированы от новых к старым -> `ORDER BY uploaded_at DESC`

## Индекс `idx_orders_user_uploaded_at`

Оптимизирует получение заказов конкретного пользователя с сортировкой от новых к старым. Используется в `GET /api/user/orders`.

```sql
CREATE INDEX idx_orders_user_uploaded_at
    ON orders (user_id, uploaded_at DESC);
```

## Индекс `idx_orders_polling`

Оптимизирует выбор заказов, которые пора отправить во внешний сервис accrual. Планируется использовать в запросе poller'а.

```sql
CREATE INDEX idx_orders_polling
    ON orders (status, next_poll_at);
```

## Таблица `withdrawals`

Таблица хранит операции списания баллов пользователями.

```sql
CREATE TABLE withdrawals (
    id BIGSERIAL PRIMARY KEY, -- Внутренний идентификатор списания
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Пользователь, выполнивший списание
    order_number TEXT NOT NULL UNIQUE, -- Номер нового заказа, в счет оплаты которого списаны баллы
    amount NUMERIC(19,2) NOT NULL CHECK (amount > 0), -- Сумма списанных баллов
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now() -- Время списания
);
```

Бизнес-правила:

- Сумма списания должна быть положительной -> `CHECK (amount > 0)`
- Номер заказа списания уникален -> `order_number TEXT NOT NULL UNIQUE`
- Списания пользователя выдаются в обратном хронологическом порядке -> `ORDER BY processed_at DESC`

## Индекс `idx_withdrawals_user_processed_at`

Оптимизирует получение списка списаний конкретного пользователя с сортировкой от новых к старым. Используется `GET /api/user/withdrawals`.

```sql
CREATE INDEX idx_withdrawals_user_processed_at
    ON withdrawals (user_id, processed_at DESC);
```

## Расчёт баланса пользователя

В текущей схеме отдельной таблицы `balances` нет. На данный момент предполагается рассчитывать баланс из двух источников:

- начисления по обработанным заказам из `orders.accrual`, где `status = 'PROCESSED'`
- списания из `withdrawals.amount`
