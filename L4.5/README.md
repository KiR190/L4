# Оптимизация HTTP API

Этот проект демонстрирует, как оптимизировать простой HTTP API на Go с использованием инструментов **pprof**, **trace** и **benchstat**. Он наглядно показывает разницу между референсной реализацией и оптимизированным подходом с точки зрения производительности и аллокаций памяти.

---

## Структура проекта

-   **cmd/server** — HTTP‑сервер.

    -   `/sum-bad` — референсная реализация (использует `encoding/json`, `fmt.Sprintf`, лишние аллокации).
    -   `/sum` — оптимизированная реализация (использует `strconv`, прямую запись в `[]byte`, zero‑allocation подход).

-   **cmd/load** — утилита для генерации нагрузки.

-   **benchmarks** — Go‑бенчмарки (`go test -bench`).

-   **profiles** — сохранённые профили pprof:

    -   `bad/` — профили для референской реализации.
    -   `good/` — профили для оптимизированной реализации.

---

## Как запустить проект

### 1. Запуск сервера

```bash
go run cmd/server/main.go
```

-   HTTP‑сервер: `:8080`
-   pprof: `:6060`

---

### 2. Генерация нагрузки

```bash
# Нагрузка на референсный endpoint
go run cmd/load/main.go -url "http://localhost:8080/sum-bad" -c 10 -d 30s

# Нагрузка на оптимизированный endpoint
go run cmd/load/main.go -url "http://localhost:8080/sum" -c 10 -d 30s
```

Где:

-   `-c` — количество конкурентных клиентов
-   `-d` — длительность теста

---

### 3. Сбор профилей производительности

#### CPU‑профиль

```bash
curl "http://localhost:6060/debug/pprof/profile?seconds=10" > cpu.out
```

#### Heap‑профиль

```bash
curl "http://localhost:6060/debug/pprof/heap" > heap.out
```

#### Trace‑профиль

```bash
curl "http://localhost:6060/debug/pprof/trace?seconds=5" > trace.out
```

## Анализ производительности

### CPU‑профиль

Для визуального анализа:

```bash
go tool pprof -http=:8081 profiles/bad/cpu.out
```

В **референсной реализации** основное время CPU уходит на:

-   `encoding/json.Marshal`
-   `fmt.Sprintf`
-   отражение (reflection) и создание временных объектов

Для сравнения откройте:

```bash
go tool pprof -http=:8081 profiles/good/cpu.out
```

В **оптимизированной версии** доминируют:

-   `strconv.Atoi`
-   прямая запись ответа в `http.ResponseWriter`

Накладные расходы значительно ниже, а горячие точки хорошо локализованы.

---

### Бенчмарки

Бенчмарки проверяют чистую бизнес‑логику без HTTP‑оверҳеда:

```bash
go test -bench=. ./benchmarks > bench.txt
benchstat bench.txt
```

### Результаты

(Актуальные цифры см. в `bench.txt`)

Оптимизированная версия:

-   избегает `encoding/json` и reflection
-   минимизирует количество аллокаций
-   работает заметно быстрее и стабильнее под нагрузкой
