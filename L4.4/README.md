# Утилита анализа GC и памяти (Go)

Это HTTP-сервер на Go, который экспортирует метрики runtime (память, GC) в формате Prometheus и предоставляет профилирование pprof.

## Функциональность

-   **Метрики Prometheus**: `/metrics`
    -   `go_mem_alloc_bytes`: Текущая выделенная память.
    -   `go_mem_heap_alloc_bytes`: Память, выделенная в куче.
    -   `go_mem_total_alloc_bytes`: Общий объем выделенной памяти (счетчик).
    -   `go_mem_mallocs_total`: Общее количество количеством аллокаций (счетчик).
    -   `go_mem_frees_total`: Общее количество освобождений памяти (счетчик).
    -   `go_gc_cycles_total`: Общее количество циклов сборки мусора (счетчик).
    -   `go_gc_last_time_seconds`: Время последнего GC (Unix timestamp).
-   **Профилирование**: `/debug/pprof/`
    -   Стандартные профили Go (cpu, heap, goroutine и т.д.).
-   **Управление GC**: Через переменную окружения `GC_PERCENT` при старте.

## Запуск

Для запуска сервера используйте команду:

```bash
go run cmd/server/main.go
```

Сервер запустится на порту `:8080`.

### Конфигурация через переменные окружения

-   `GC_PERCENT`: Устанавливает целевой процент сборки мусора (по умолчанию 100).
-   `BALLAST_MB`: Выделяет "балласт" памяти при старте (в МБ) для эмуляции нагрузки.

Пример запуска с измененным GC percent и балластом:

```bash
GC_PERCENT=50 BALLAST_MB=100 go run cmd/server/main.go
```

## Примеры запросов

### 1. Получение метрик (Prometheus)

```bash
curl http://localhost:8080/metrics
```

В ответе вы увидите метрики, начинающиеся с `go_mem_` и `go_gc_`.

### 2. Проверка состояния (Health check)

```bash
curl http://localhost:8080/health
```

### 3. Профилирование (Pprof)

Открыть в браузере: `http://localhost:8080/debug/pprof/`

Или использовать инструмент `go tool pprof`:

```bash
# Профиль кучи (Heap)
go tool pprof http://localhost:8080/debug/pprof/heap

# Профиль CPU (30 секунд)
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
```
