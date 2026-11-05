# Semana 10 · línea de base de peticiones

En la fase 0, antes de los cambios, PostgreSQL y la API todavía no estaban activos; por eso la columna «Antes» se verificó con la prueba HTTP `TestDemoHito2RespuestasDeRolesYTransicion` (`httptest` + SQLite en memoria), no con una corrida Bruno. La columna «Después» sí contiene los resultados de Bruno contra la API conectada a PostgreSQL.

| Petición | Antes | Después |
|---|---:|---:|
| 1 sin token | 401 | 401 (Bruno) |
| 2 login | 200 | 200 (Bruno) |
| 3 con token | 200 | 200 (Bruno) |
| 4 rol sin permiso | 403 | 403 (Bruno) |
| 5 cambio permitido | 200 | 200 (Bruno) |
| 6 transición prohibida | 422 | 422 (Bruno) |

La prueba de referencia está en `servidor-pareja/internal/latent/latent_test.go`; puede repetirse con:

```powershell
cd servidor-pareja
go test -count=1 -run '^TestDemoHito2RespuestasDeRolesYTransicion$' -v ./internal/latent
```

La corrida posterior real con la extensión Bruno de VS Code quedó registrada en `docs/taller10/peticiones_despues.txt`.

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


## Fase 6 · transacciones

`ServicioFotos` depende de `RepositorioFotos`; GORM y SQLC realizan el incremento de cuota y la inserción de foto dentro de una sola transacción. Las pruebas fuerzan un UUID duplicado y comprueban que no se incrementa el contador. `error_forzado.txt` guarda resultados de ambos adaptadores en SQLite aislado y una prueba optativa posterior contra PostgreSQL nativo, con registros temporales eliminados al terminar.
