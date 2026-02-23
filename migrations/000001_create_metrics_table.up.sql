-- migrations/000001_create_metrics_table.up.sql
-- Создание таблицы метрик
CREATE TABLE IF NOT EXISTS metrics (
                         id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
                         name VARCHAR(255) NOT NULL,
                         metric_type VARCHAR(255) NOT NULL,
                         delta BIGINT,
                         value DOUBLE PRECISION
);

-- Базовый индекс для поиска по названию
CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name);

-- Индекс для поиска по типу
CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(metric_type);

-- уникальность записей
ALTER TABLE metrics ADD CONSTRAINT metrics_name_type_uniq UNIQUE (name, metric_type);
