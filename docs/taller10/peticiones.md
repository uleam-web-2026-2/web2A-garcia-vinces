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
