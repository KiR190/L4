# Календарь событий (Task Manager)

HTTP-сервис для управления событиями календаря с поддержкой напоминаний.

## Требования

-   Go 1.25+
-   PostgreSQL 15+
-   RabbitMQ 3.x (с плагином `rabbitmq_delayed_message_exchange`)

## Быстрый старт

### С Docker Compose

```bash
docker-compose up -d
```

Сервис будет доступен на `http://localhost:8080`

## Переменные окружения

| Переменная                 | Описание                           | По умолчанию                        |
| -------------------------- | ---------------------------------- | ----------------------------------- |
| `APP_PORT`                 | Порт HTTP-сервера                  | `8080`                              |
| `DATABASE_HOST`            | Хост PostgreSQL                    | `localhost`                         |
| `DATABASE_PORT`            | Порт PostgreSQL                    | `5432`                              |
| `DATABASE_USER`            | Пользователь БД                    | `calendar_user`                     |
| `DATABASE_PASSWORD`        | Пароль БД                          | `mysecretpassword`                  |
| `DATABASE_NAME`            | Имя БД                             | `calendar_db`                       |
| `RABBIT_URL`               | URL RabbitMQ                       | `amqp://guest:guest@rabbitmq:5672/` |
| `REMINDER_QUEUE_CAP`       | Размер очереди напоминаний         | `256`                               |
| `ARCHIVE_INTERVAL_MINUTES` | Интервал архивации (мин)           | `60`                                |
| `ARCHIVE_CUTOFF_DAYS`      | Архивировать события старше N дней | `30`                                |

## API Эндпоинты

### Создание события

```bash
curl -X POST http://localhost:8080/create_event \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "event_date": "2026-01-15T00:00:00Z",
    "title": "Встреча с командой",
    "remind_at": "2026-01-15T09:00:00Z",
    "remind_channel": "native"
  }'
```

### Обновление события

```bash
curl -X POST http://localhost:8080/update_event \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "event_date": "2026-01-16T00:00:00Z",
    "title": "Встреча перенесена"
  }'
```

### Удаление события

```bash
curl -X POST http://localhost:8080/delete_event \
  -H "Content-Type: application/json" \
  -d '{"id": 1}'
```

### События на день

```bash
curl "http://localhost:8080/events_for_day?user_id=1&date=2026-01-15"
```

### События на неделю

```bash
curl "http://localhost:8080/events_for_week?user_id=1&date=2026-01-15"
```

### События на месяц

```bash
curl "http://localhost:8080/events_for_month?user_id=1&date=2026-01-15"
```

## Формат полей

| Поле               | Тип    | Описание                                         |
| ------------------ | ------ | ------------------------------------------------ |
| `user_id`          | int64  | ID пользователя                                  |
| `event_date`       | string | Дата события (YYYY-MM-DD или RFC3339)            |
| `title`            | string | Название события                                 |
| `remind_at`        | string | Время напоминания (RFC3339), опционально         |
| `remind_channel`   | string | Канал: `native`, `email`, `telegram`             |
| `remind_recipient` | string | Email или Telegram username (для email/telegram) |

## Напоминания

Поддерживаются 3 канала отправки:

-   **native** — вывод в stdout/логи сервера
-   **email** — отправка на email (требует настройки SMTP)
-   **telegram** — отправка в Telegram (требует TG_BOT_TOKEN)

## Архивация

Фоновый воркер каждые `ARCHIVE_INTERVAL_MINUTES` минут помечает события старше `ARCHIVE_CUTOFF_DAYS` дней как архивные (`is_archived=true`). Архивные события не возвращаются в запросах.

## Тестирование

```bash
go test ./... -v
```
