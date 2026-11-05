-- Esquema descriptivo para sqlc. GORM AutoMigrate sigue siendo responsable
-- de crear y actualizar estas tablas en ejecución.
CREATE TABLE eventos (
    id bigserial PRIMARY KEY,
    nombre text NOT NULL,
    codigo_acceso text NOT NULL,
    disparo_por_invitado bigint NOT NULL,
    max_invitado bigint NOT NULL,
    estado text NOT NULL,
    fecha_revelado timestamptz,
    usuario_id bigint NOT NULL
);
