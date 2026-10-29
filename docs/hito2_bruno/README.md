# Colección Bruno — Hito 2

## Requisitos

- Servidor Latent disponible en `http://localhost:8081`.
- PostgreSQL local administrado desde pgAdmin.
- Base de datos desechable `LatentDemo`. No usar el seed en la base anterior ni en una base con datos reales.
- Seleccionar el entorno `local` de esta colección.

## Preparar la demostración

1. En pgAdmin, crea `LatentDemo`. No borres ni reutilices una base que tenga información que quieras conservar.
2. Desde PowerShell, dentro de `servidor-pareja`, configura `DATABASE_URL` para apuntar a `LatentDemo`, `CLAVE_FIRMA` y `MODO_DEMO=true`, y ejecuta `go run .`. El servidor aplica migraciones y crea las cuentas de ejemplo.
3. En pgAdmin, selecciona `LatentDemo`, abre Query Tool, carga `docs/hito2_demo_seed.sql` desde el repositorio y ejecútalo allí. El seed solo añade los dos eventos de demostración.

4. Ejecuta la petición `0 Preparación · registrar invitado`. El script guarda `tokenInvitado` en el entorno.
5. Ejecuta la petición 1, luego 2 a 6 en orden. La petición 2 guarda `tokenAnfitrion`.

La petición 0 es preparación y no cuenta entre las seis demostraciones evaluadas. El seed asigna IDs 1 y 2 para que la demo sea repetible en una base desechable. El archivo no reconstruye ni borra toda la base.
