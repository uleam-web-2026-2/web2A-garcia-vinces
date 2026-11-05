-- name: BuscarEvento :one
SELECT id, nombre, codigo_acceso, disparo_por_invitado, max_invitado, estado, fecha_revelado, usuario_id
FROM eventos
WHERE id = $1;

-- name: GuardarEvento :execrows
UPDATE eventos
SET nombre = $2,
    codigo_acceso = $3,
    disparo_por_invitado = $4,
    max_invitado = $5,
    estado = $6,
    fecha_revelado = $7,
    usuario_id = $8
WHERE id = $1;
