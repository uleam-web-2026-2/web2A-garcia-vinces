# Hoja de demostración — Hito 2

**Producto:** Latent · **Paralelo:** Web II A · **Entidad principal:** Evento

## 1. Usuarios de ejemplo

| Usuario | Rol | Contraseña de ejemplo |
| --- | --- | --- |
| `ana` | invitada (cuenta de prueba) | `ana-2026` |
| `luis` | anfitrión | `luis-2026` |

El flujo real de una invitada no requiere cuenta: usa `POST /eventos/unirse` con código, nombre y PIN y recibe un token ligado a su registro y evento.

## 2. Preparación y seis peticiones Bruno

Primero levanten la aplicación y carguen la base desechable con `docs/hito2_demo_seed.sql` (instrucciones en `docs/hito2_bruno/README.md`). En Bruno, elijan el entorno `local` y ejecuten en orden:

| # | Petición | Credencial | Resultado esperado |
| :-: | --- | --- | --- |
| 0 | `POST /eventos/unirse` con `EVT-DEMBBB`, Carlos y PIN `0427` | Ninguna | 200 y token ligado al evento 2 |
| 1 | `GET /eventos/2/fotos` | Ninguna | **401** `no_autenticado` |
| 2 | `POST /login` con Luis | Ninguna | **200** y JWT de anfitrión |
| 3 | `GET /eventos/2/fotos` | Token de Carlos | **200** y álbum vacío |
| 4 | `PATCH /eventos/2` a `activo` | Token de Carlos (invitado) | **403** `rol_no_permitido` |
| 5 | `PATCH /eventos/2` a `activo` | Token de Luis (anfitrión) | **200** y estado `activo` |
| 6 | `PATCH /eventos/1` a `activo` | Token de Luis; evento finalizado | **422** `transicion_prohibida` |

## 3. Matriz ruta × rol

| Ruta | Sin sesión | Invitado | Anfitrión propietario |
| --- | :-: | :-: | :-: |
| `POST /login` | Sí | Sí | Sí |
| `POST /eventos/unirse` | Sí | Sí | Sí |
| `GET /eventos` | 401 | 403 | Sí |
| `POST /eventos` | 401 | 403 | Sí |
| `GET /eventos/{id}` | 401 | 403 | Sí |
| `PUT /eventos/{id}` | 401 | 403 | Sí, si está creado |
| `PATCH /eventos/{id}` | 401 | 403 | Sí |
| `DELETE /eventos/{id}` | 401 | 403 | Sí, si está vacío y no iniciado |
| `GET /eventos/{id}/fotos` | 401 | Sí, solo su evento | Sí, solo propio |
| `POST /fotos` | 401 | Sí, propias y dentro de cuota | 403 |

## 4. Cumplimiento

- [x] Los structs de Usuario, Evento, Invitado y Foto se migran a PostgreSQL.
- [x] CRUD y cambio de estado de Evento; transiciones prohibidas responden 422.
- [x] JWT, rutas protegidas, separación de roles y prueba de integración sin token.
- [x] Código de eventos separado en manejador HTTP, servicio y repositorio.
- [x] Seis peticiones Bruno en orden; la petición 0 prepara la sesión invitada.
- [x] `npx @redocly/cli lint docs/openapi.yaml --format=stylish` sin avisos.
- [x] `go test ./...` verde y respaldos en `docs/hito2_capturas/`.

Las capturas en `docs/hito2_capturas/` corresponden a las respuestas en vivo de Bruno y a la ejecución de `go test ./...`. Antes de compartirlas, oculten los valores JWT de las respuestas de preparación y login.
