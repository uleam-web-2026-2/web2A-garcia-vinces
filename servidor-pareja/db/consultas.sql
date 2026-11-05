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


-- name: EstadoCuotaFoto :one
SELECT i.evento_id, i.fotos_tomadas, e.disparo_por_invitado, e.estado
FROM invitados AS i
JOIN eventos AS e ON e.id = i.evento_id
WHERE i.id = $1;

-- name: IncrementarCuotaFoto :execrows
UPDATE invitados
SET fotos_tomadas = fotos_tomadas + 1
WHERE invitados.id = $1
  AND invitados.fotos_tomadas < (
      SELECT e.disparo_por_invitado
      FROM eventos AS e
      WHERE e.id = invitados.evento_id AND e.estado = 'activo'
  );

-- name: CrearFoto :one
INSERT INTO fotos (uuid, url_archivo, sincronizado, evento_id, invitado_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, uuid, url_archivo, sincronizado, evento_id, invitado_id;
