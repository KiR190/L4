-- Удаляем индексы
DROP INDEX IF EXISTS idx_events_user_date;
DROP INDEX IF EXISTS idx_events_date;
DROP INDEX IF EXISTS idx_events_user_id;
DROP INDEX IF EXISTS idx_events_is_archived;
DROP INDEX IF EXISTS idx_events_user_active;

-- Удаляем триггер
DROP TRIGGER IF EXISTS set_updated_at ON events;

-- Удаляем функцию
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаляем представления
DROP VIEW IF EXISTS events_active;
DROP VIEW IF EXISTS events_archived;

-- Удаляем таблицу
DROP TABLE IF EXISTS events;
