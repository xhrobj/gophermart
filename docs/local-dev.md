# Локальная разработка

## Локальный запуск без Docker Compose

Этот сценарий поднимает backend и PostgreSQL без контейнеров `client` и `accrual`.

Для полного локального окружения с web-клиентом и accrual использовать запуск через Docker Compose (описан выше).

### 1. Поднять PostgreSQL

```bash
make postgres-up
```

Если контейнер уже существует, можно использовать:

```bash
make postgres-start
```

### 2. Запустить приложение

```bash
make run
```

Подключиться к базе:

```bash
make postgres-connect
```

Остановить или удалить контейнер с базой:

```bash
make postgres-stop
make postgres-rm
```

## Конфигурация

Приложение поддерживает конфигурацию через флаги командной строки и переменные окружения:

- адрес запуска HTTP-сервера: флаг `-a` или переменная `RUN_ADDRESS`
- строка подключения к PostgreSQL: флаг `-d` или переменная `DATABASE_URI`
- адрес внешнего accrual-сервиса: флаг `-r` или переменная `ACCRUAL_SYSTEM_ADDRESS`
- секрет для подписи JWT: переменная `JWT_SECRET`

## Gophermart: примеры curl-запросов

После успешного `register` или `login` взять JWT из заголовка `Authorization` ответа и подставить его вместо `<jwt-token>`.

### Регистрация

```bash
curl -i -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{"login":"admin","password":"secret"}'
```

### Логин

```bash
curl -i -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"login":"admin","password":"secret"}'
```

### Загрузка номера заказа

```bash
curl -i -X POST http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: text/plain" \
  -d '12345678903'
```

### Получение списка загруженных заказов

```bash
curl -i http://localhost:8080/api/user/orders \
  -H "Authorization: Bearer <jwt-token>"
```

### Получение баланса

```bash
curl -i http://localhost:8080/api/user/balance \
  -H "Authorization: Bearer <jwt-token>"
```

### Списание баллов

```bash
curl -i -X POST http://localhost:8080/api/user/balance/withdraw \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt-token>" \
  -d '{
    "order": "2377225624",
    "sum": 5.11
  }'
```

### Получение истории списаний

```bash
curl -i http://localhost:8080/api/user/withdrawals \
  -H "Authorization: Bearer <jwt-token>"
```

## Тесты и проверки

Запустить все тесты:

```bash
make test
```

Запустить тесты с race detector:

```bash
make test-race
```

Запустить интеграционные тесты репозиториев:

```bash
make test-integration
```

Посчитать покрытие:

```bash
make test-coverage
```

Запустить линтер:

```bash
make lint
```

## Makefile-команды

Основные команды для локальной разработки:

- `make run-dev` - запустить полное окружение через Docker Compose
- `make run-dev-down` - остановить compose-окружение
- `make run-dev-logs` - смотреть логи compose-окружения
- `make build` - собрать бинарник Gophermart
- `make run` - собрать и запустить backend

Команды для тестов и проверок:

- `make test` - запустить тесты
- `make test-race` - запустить тесты с race detector
- `make test-integration` - запустить интеграционные тесты репозиториев
- `make test-coverage` - собрать покрытие
- `make test-coverage-integration` - собрать покрытие с учетом интеграционных тестов репозиториев
- `make lint` - запустить golangci-lint

Команды для локального PostgreSQL:

- `make postgres-up` - поднять PostgreSQL-контейнер
- `make postgres-start` - запустить существующий PostgreSQL-контейнер
- `make postgres-stop` - остановить PostgreSQL-контейнер
- `make postgres-rm` - удалить PostgreSQL-контейнер
- `make postgres-connect` - подключиться к PostgreSQL через psql
