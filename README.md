# Latent

**Hilo Servidor · ULEAM · Período 2026-2**
Aplicaciones Web II (TDI-610) · Aplicación para el Servidor Web (IS-503)

## Integrantes

| Integrante | Usuario de GitHub | Paralelo |
| --- | --- | --- |
| Ordeoñez Jostin Dionicio | Jostin12Garcia | Web II A |
| Vinces Chonillo Jean Carlos | Jeanvncs | Web II A |

## El producto

Latent es una cámara desechable digital para eventos sociales (bodas, cumpleaños y fiestas). Vende a los anfitriones un evento con código de acceso en el que cada invitado toma un número limitado de fotos (3), sin instalar apps ni crear una cuenta, y el anfitrión recibe el álbum completo. El negocio de referencia es ONCE.

La ficha del negocio y el addendum técnico del Hito 1 están en [`docs/hito1_ficha_del_negocio.md`](docs/hito1_ficha_del_negocio.md) y [`docs/hito1_addendum.md`](docs/hito1_addendum.md).

## Cómo levantar el servidor

Requisitos: Go (versión estable) y PostgreSQL.

1. Crea una base de datos en PostgreSQL llamada `Latent`.
2. El servidor no lee archivos `.env`: toma `PUERTO` y `DATABASE_URL` del entorno. Los valores de ejemplo están en [`servidor-pareja/.env.example`](servidor-pareja/.env.example).
3. Desde la carpeta `servidor-pareja`, define las variables y arranca el servidor.

En PowerShell (Windows):

```powershell
cd servidor-pareja
$env:PUERTO = "8081"
$env:DATABASE_URL = "postgres://usuario:clave@localhost:5432/Latent?sslmode=disable"
go run .
```

En bash (Linux o macOS):

```bash
cd servidor-pareja
export PUERTO=8081
export DATABASE_URL="postgres://usuario:clave@localhost:5432/Latent?sslmode=disable"
go run .
```

El servidor responde en `http://localhost:8081`.

Todas las peticiones necesitan la cabecera `X-Usuario-ID` (por ejemplo, `X-Usuario-ID: 1`). Sin ella, el servidor responde 401. Para crear un evento, envía `POST /eventos` con este cuerpo JSON:

```json
{
  "nombre": "Boda de prueba",
  "disparo_por_invitado": 3,
  "max_invitado": 50
}
```

La respuesta correcta es 201 Created. Con el nombre vacío, responde 400.

## Base de datos

- Opción usada por la pareja: PostgreSQL nativo
- Base de datos del proyecto: `Latent`

## Cómo correr las pruebas

```bash
cd servidor-pareja
go test ./...
```

Son 5 pruebas, en `servidor-pareja/internal/latent/latent_test.go`, una por cada regla de negocio del Hito 1. Las pruebas corren también en la integración continua (pestaña Actions). Desde la semana 4, un entregable cuyas pruebas no pasan en la integración no se recibe.

## Convenciones del repositorio

- Un commit de cada integrante como mínimo por taller; el commit de cierre se hace en clase.
- Mensajes de commit: qué cambió y por qué, entendibles sin el autor presente.
- Uso de IA declarado en el cuerpo del commit: una línea con qué herramienta y para qué parte.
- Ningún secreto en el código ni en el historial: la configuración se externaliza (semana 4).

## Estructura

```
docs/ficha_negocio.md              → ficha preliminar del negocio (semana 3)
docs/hito1_ficha_del_negocio.md    → ficha del Hito 1
docs/hito1_addendum.md             → addendum técnico del Hito 1
docs/hito1_*.png                   → capturas de pruebas y de respuestas de la API
servidor-pareja/main.go            → punto de entrada del servidor
servidor-pareja/internal/latent/   → modelos, manejadores, rutas y pruebas
servidor-pareja/internal/middleware/ → identificación del usuario (X-Usuario-ID)
servidor-pareja/internal/respuesta/  → formato común de las respuestas JSON
servidor-pareja/.env.example       → variables de entorno de ejemplo
```

## Historial previo

El trabajo anterior a este repositorio, con sus commits, está en https://github.com/Jeanvncs/-web2A-Garcia-Vinces
