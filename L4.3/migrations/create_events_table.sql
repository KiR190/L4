-- Создание таблицы events
CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    event_date DATE NOT NULL,
    title TEXT NOT NULL CHECK (char_length(title) > 0),
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON events
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Индексы
CREATE INDEX IF NOT EXISTS idx_events_user_date ON events (user_id, event_date);
CREATE INDEX IF NOT EXISTS idx_events_date ON events (event_date);
CREATE INDEX IF NOT EXISTS idx_events_user_id ON events (user_id);
CREATE INDEX IF NOT EXISTS idx_events_is_archived ON events (is_archived);
-- Составной индекс для выборки активных событий пользователя
CREATE INDEX IF NOT EXISTS idx_events_user_active ON events (user_id, is_archived) WHERE is_archived = FALSE;

-- VIEW для активных событий (is_archived = false)
CREATE OR REPLACE VIEW events_active AS
SELECT id, user_id, event_date, title, is_archived, created_at, updated_at
FROM events
WHERE is_archived = FALSE;

-- VIEW для архивных событий (is_archived = true)
CREATE OR REPLACE VIEW events_archived AS
SELECT id, user_id, event_date, title, is_archived, created_at, updated_at
FROM events
WHERE is_archived = TRUE;

-- Начальные данные
INSERT INTO events (user_id, event_date, title, is_archived)
VALUES
    (1, '2025-10-19', 'Посещение врача', false),
    (1, '2025-10-21', 'Встреча с командой', false),
    (2, '2025-10-22', 'Поход в горы', false),
    (1, '2025-11-01', 'День рождения друга', false);