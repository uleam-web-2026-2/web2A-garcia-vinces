-- SOLO para la base LatentDemo desechable de demostración local.
-- La aplicación debe haber arrancado una vez: así existen tablas y cuentas.
-- Nunca ejecutes este archivo en una base desplegada con datos reales.
BEGIN;
INSERT INTO usuarios (id, nombre, rol) VALUES (2, 'luis', 'anfitrion')
ON CONFLICT (id) DO UPDATE SET nombre = EXCLUDED.nombre, rol = EXCLUDED.rol;
INSERT INTO eventos (id, nombre, codigo_acceso, disparo_por_invitado, max_invitado, estado, fecha_revelado, usuario_id)
VALUES
  (1, 'Evento demo finalizado', 'EVT-DEMAAA', 3, 50, 'finalizado', CURRENT_TIMESTAMP, 2),
  (2, 'Evento demo en preparación', 'EVT-DEMBBB', 3, 50, 'creado', NULL, 2)
ON CONFLICT (id) DO UPDATE SET
  nombre = EXCLUDED.nombre,
  codigo_acceso = EXCLUDED.codigo_acceso,
  disparo_por_invitado = EXCLUDED.disparo_por_invitado,
  max_invitado = EXCLUDED.max_invitado,
  estado = EXCLUDED.estado,
  fecha_revelado = EXCLUDED.fecha_revelado,
  usuario_id = EXCLUDED.usuario_id;
SELECT setval(pg_get_serial_sequence('eventos', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM eventos), 1));
COMMIT;

