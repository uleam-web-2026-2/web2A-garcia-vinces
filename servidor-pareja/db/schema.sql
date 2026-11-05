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

CREATE TABLE invitados (
    id bigserial PRIMARY KEY,
    nombre text NOT NULL,
    pin varchar(4) NOT NULL,
    fotos_tomadas bigint NOT NULL DEFAULT 0,
    evento_id bigint NOT NULL REFERENCES eventos(id),
    UNIQUE (evento_id, nombre)
);

CREATE TABLE fotos (
    id bigserial PRIMARY KEY,
    uuid text NOT NULL UNIQUE,
    url_archivo text NOT NULL,
    sincronizado boolean NOT NULL DEFAULT false,
    evento_id bigint NOT NULL REFERENCES eventos(id),
    invitado_id bigint NOT NULL REFERENCES invitados(id)
);
