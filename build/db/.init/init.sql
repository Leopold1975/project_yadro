SELECT 'CREATE DATABASE comics' 
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'comics')\gexec

CREATE USER developer WITH ENCRYPTED PASSWORD 'developer';

GRANT ALL PRIVILEGES ON DATABASE comics TO developer;

-- Предоставление прав на схему public
GRANT ALL PRIVILEGES ON SCHEMA public TO developer;

-- Предоставление прав на все таблицы в схеме public
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO developer;

ALTER DATABASE comics OWNER TO developer;

-- Для новых таблиц
--ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO developer;