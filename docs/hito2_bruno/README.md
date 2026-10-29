# Colección Bruno — Hito 2

## Requisitos

- Servidor Latent disponible en `http://localhost:8081`.
- Base de datos local desechable. No usar el seed en una base con datos reales.
- Seleccionar el entorno `local` de esta colección.

## Preparar la demostración

1. Levanta el servidor con una base vacía y espera a que aplique las migraciones y cree las cuentas de ejemplo.
2. En la raíz del repositorio, ejecuta el seed solo en esa base local:

   ```powershell
   Get-Content .\docs\hito2_demo_seed.sql | docker compose exec -T db psql -U latent -d latent
   ```

   Si ejecutas PostgreSQL fuera de Docker, usa `psql` con esa base y el mismo archivo SQL.
3. Ejecuta la petición `0 Preparación · registrar invitado`. El script guarda `tokenInvitado` en el entorno.
4. Ejecuta la petición 1, luego 2 a 6 en orden. La petición 2 guarda `tokenAnfitrion`.

La petición 0 es preparación y no cuenta entre las seis demostraciones evaluadas. El seed asigna IDs 1 y 2 únicamente para que la demo sea repetible en una base descartable.
