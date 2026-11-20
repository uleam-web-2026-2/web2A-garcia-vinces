# Auditoría de posibles inyecciones SQL

La búsqueda real con `git grep` produjo 32 coincidencias candidatas. Se revisaron los usos de SQL, las llamadas GORM, las consultas generadas por sqlc y los falsos positivos. No se encontró concatenación de texto recibido del cliente para construir SQL.

| Coincidencias | ¿Arma SQL? | ¿Llega texto del cliente? | ¿Va aparte como parámetro? | Evaluación |
|---|---|---|---|---|
| `servidor-pareja/internal/latent/eventos_repositorio.go:33,42,55,61` | Sí, mediante GORM. | IDs y datos asociados al evento; los ID de ruta se validan. | Sí: `Where` usa `?` y `First` recibe un ID numérico. | Seguras; los valores no se concatenan al SQL. |
| `servidor-pareja/internal/latent/filas/consultas.sql.go:20,50,84,119,147` | Sí. | Algunos valores provienen de peticiones. | Sí; sqlc usa consultas fijas con `$1`, `$2`, etc. | Seguras; son consultas generadas y parametrizadas. |
| `servidor-pareja/internal/latent/manejadores.go:116,124,136,195,244,263` | Sí, mediante GORM. | Sí, filtros del cuerpo y valores asociados al evento. | Sí: filtros con `?`, ID validados; `Order("id")` es fijo. | Seguras; no se arma SQL con texto del cliente. |
| `servidor-pareja/internal/latent/reservas.go:22,43,65` | Sí. | Los ID proceden de valores del caso de uso. | Sí; `Exec` y `Raw` usan `?` y argumentos separados. | Seguras; las sentencias son fijas y parametrizadas. |
| `servidor-pareja/internal/sesion/usuarios.go:69,98` | Sí, mediante GORM. | Sí; el nombre llega al login o se toma de configuración. | Sí; `Where` usa `?`. | Seguras; el nombre no se concatena al SQL. |
| `servidor-pareja/main.go:89,94` | Sí, mediante GORM. | No directamente; consulta cuentas y usa los ID obtenidos. | Sí; GORM envía los valores por separado. | Seguras; no se concatenan entradas del cliente. |
| `servidor-pareja/internal/latent/eventos_servicio.go:163,168` | No. | No; se genera un código de evento. | No aplica. | Falsos positivos: `Sprintf` y `+` construyen el código, no una consulta. |
| `servidor-pareja/internal/latent/filas/db.go:13,15,16`; `servidor-pareja/internal/latent/rutas.go:23` | No ejecutan consultas. | No. | No aplica. | Falsos positivos: son interfaces de base de datos y agrupación de rutas. |
| `servidor-pareja/internal/sesion/guardia.go:60`; `servidor-pareja/main.go:117,129,130` | No. | No se usa para SQL. | No aplica. | Falsos positivos: se construyen un mensaje, una respuesta o la dirección del servidor. |

La prueba automatizada `TestIDConTextoSQLResponde400` comprueba que un ID con texto tipo `OR 1=1` se rechaza con HTTP 400. La petición manual con el servidor en ejecución y su resultado deben registrarse por separado en `inyeccion.txt`.
