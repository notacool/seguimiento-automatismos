-- Habilitar extensión pg_cron para limpieza automática
CREATE EXTENSION IF NOT EXISTS pg_cron;

-- Nota: Las tablas se crearán mediante migraciones de golang-migrate
-- Este script solo inicializa extensiones necesarias
