# Semana 10 · línea de base de peticiones

La línea de base se obtuvo antes de los cambios con la prueba HTTP `TestDemoHito2RespuestasDeRolesYTransicion` (`httptest` + SQLite en memoria). No se pudo ejecutar la colección Bruno contra PostgreSQL en esta máquina: no hay servidor disponible en `localhost:5432` ni API activa en `localhost:8081`. Estos resultados son la comprobación automatizada disponible, no una corrida Bruno.

| Petición | Antes | Después |
|---|---:|---:|
| 1 sin token | 401 | Pendiente de ejecutar Bruno después de los cambios |
| 2 login | 200 | Pendiente de ejecutar Bruno después de los cambios |
| 3 con token | 200 | Pendiente de ejecutar Bruno después de los cambios |
| 4 rol sin permiso | 403 | Pendiente de ejecutar Bruno después de los cambios |
| 5 cambio permitido | 200 | Pendiente de ejecutar Bruno después de los cambios |
| 6 transición prohibida | 422 | Pendiente de ejecutar Bruno después de los cambios |

La prueba de referencia está en `servidor-pareja/internal/latent/latent_test.go`; puede repetirse con:

```powershell
cd servidor-pareja
go test -count=1 -run '^TestDemoHito2RespuestasDeRolesYTransicion$' -v ./internal/latent
```

Para completar la evidencia exigida por la guía, ejecutar la colección `docs/hito2_bruno` contra PostgreSQL nativo y completar «Después» tras desplegar la rama.

## Fase 1 · interfaz del repositorio

Este punto ya estaba implementado al iniciar el taller: `ServicioEventos` recibe `RepositorioEventos`, definido en `internal/latent/eventos_repositorio.go`, y `eventos_servicio.go` no importa GORM. `main.go` construye el repositorio GORM y se lo entrega al servicio. Se conserva este diseño.

## Fase 3 · consultas sqlc

Se añadió `servidor-pareja/sqlc.yaml` con sqlc v1.31.1, un esquema descriptivo de `eventos` y las consultas `BuscarEvento` y `GuardarEvento`. El adaptador SQLC usa la misma conexión `database/sql` que GORM, incrusta el repositorio GORM y reemplaza solo esos dos métodos. `sql.ErrNoRows` y un `UPDATE` con cero filas se convierten en `ErrEventoNoEncontrado`. `AutoMigrate` permanece activo.

Comprobaciones locales: `sqlc generate`, `go build ./...` y `go test ./...` desde `servidor-pareja`.

## Fase 4 · pruebas automatizadas

`go test ./...` pasó; su salida quedó guardada en `docs/taller10/go_test.txt`. Las pruebas HTTP del repositorio usan SQLite en memoria y el repositorio GORM, por lo que no sustituyen la validación del adaptador SQLC ni la repetición de las seis peticiones en Bruno. Esas dos comprobaciones requieren que PostgreSQL nativo y la API estén activos.

## Fase 5 · contrato y configuración

El contrato documenta las cuatro entidades del paquete `latent`: Evento, Invitado, Foto y Usuario. Las peticiones POST y PUT de eventos usan `EventoNuevo`, sin `id` ni `estado`; la respuesta Evento documenta también `invitados`. `redocly.yaml` conserva la regla de validación recomendada, excepto `no-unused-components`, porque la guía pide incluir Usuario aunque la API no lo devuelve en ninguna ruta. `docs/taller10/redocly.txt` registra una validación limpia con Redocly CLI 2.54.3.

`servidor-pareja/.env.example` ya enumera todas las variables leídas por el proceso (`PUERTO`, `DATABASE_URL`, `CLAVE_FIRMA`, `MODO_DEMO`, `ANFITRION_USUARIO` y `ANFITRION_CLAVE`) con valores de ejemplo o vacíos; no contiene credenciales reales.
