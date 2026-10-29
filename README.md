# Latent

**Aplicaciones Web II · ULEAM · Web II A · Período 2026-2**

| Integrante | Usuario de GitHub | Paralelo |
| --- | --- | --- |
| Ordeoñez Jostin Dionicio | Jostin12Garcia | Web II A |
| Vinces Chonillo Jean Carlos | Jeanvncs | Web II A |

Latent es una cámara desechable digital para fiestas. El anfitrión crea el evento y comparte un código QR; cada invitado entra con nombre y PIN, tiene una cuota de fotos controlada por el servidor y ve el álbum cuando el anfitrión finaliza la fiesta.

## Desarrollo local

Requisitos: Go 1.27.1 o posterior y PostgreSQL 17 o posterior. Para no tocar una base anterior, crea `LatentDemo` desde pgAdmin y usa esa base descartable para la demostración. Desde PowerShell, en `servidor-pareja`, configura el entorno:

```powershell
$env:PUERTO = "8081"
$env:DATABASE_URL = "postgres://usuario:clave@localhost:5432/LatentDemo?sslmode=disable"
$env:CLAVE_FIRMA = "reemplaza-por-un-secreto-aleatorio-de-32-caracteres"
$env:MODO_DEMO = "true"
go run .
```

La guía del Hito 2 evalúa la API mediante peticiones y pruebas; no pide una interfaz web ni Docker. Comprueba `http://localhost:8081/healthz` y ejecuta la colección de [`docs/hito2_bruno/`](docs/hito2_bruno/) en Bruno. La cuenta anfitrión de demostración es `luis / luis-2026`; el login adicional de ejemplo es `ana / ana-2026`. **No habilites `MODO_DEMO` en una instancia pública.**

La aplicación Go se conecta directamente a PostgreSQL usando `DATABASE_URL`; pgAdmin administra la base y no es necesario que la API use Docker. Asegúrate de que la base indicada en esa URL exista y que la contraseña coincida con tu instalación local. El código ejecuta migraciones al arrancar, pero no borra tablas ni filas. `go run . -reset` sí elimina las tablas; no lo uses con una base que quieras conservar.

### Límite actual de almacenamiento de fotos

Para el prototipo, cada foto puede enviarse como data URL de hasta 4 MiB y se guarda en PostgreSQL junto a sus metadatos. Para abrir el servicio a eventos grandes o tráfico público, el siguiente paso es mover los bytes a almacenamiento de objetos privado (S3 compatible) y guardar en PostgreSQL solo claves y URLs firmadas; el límite por invitado reduce el crecimiento, pero no reemplaza ese almacenamiento.

## API y entregables del Hito 2

- Contrato OpenAPI: [`docs/openapi.yaml`](docs/openapi.yaml)
- Documentación navegable: [`docs/api.html`](docs/api.html)
- Hoja para la demostración: [`docs/hito2_hoja.md`](docs/hito2_hoja.md)
- Colección Bruno y entorno local: [`docs/hito2_bruno/`](docs/hito2_bruno/)
- SQL de datos de demostración descartables: [`docs/hito2_demo_seed.sql`](docs/hito2_demo_seed.sql)
- Evidencias: [`docs/hito2_capturas/`](docs/hito2_capturas/)
- Roles, transición y capas: [`servidor-pareja/docs/decisiones.md`](servidor-pareja/docs/decisiones.md)

Rutas principales: login, unión de invitado, CRUD de eventos, cambio de estado, resumen, galería y registro de fotos. Las rutas privadas usan `Authorization: Bearer <token>` y la API responde errores con una envoltura `{ok:false,error:{codigo,mensaje}}`.

## Pruebas y contrato

Desde `servidor-pareja`:

```powershell
go test ./...
go build ./...
```

Desde la raíz del repositorio:

```powershell
npx @redocly/cli lint docs/openapi.yaml --format=stylish
npx @redocly/cli build-docs docs/openapi.yaml -o docs/api.html
```

Las pruebas usan SQLite temporal; el servidor desplegado usa PostgreSQL. Las credenciales de demo se guardan con bcrypt; nunca subas `.env` ni credenciales reales.

## Estructura

- `servidor-pareja/internal/latent/`: HTTP, reglas de eventos, repositorio y modelo.
- `servidor-pareja/internal/sesion/`: login, bcrypt, JWT y permisos por rol.
- `servidor-pareja/internal/respuesta/`: formato JSON común.
El trabajo anterior está en [el repositorio previo](https://github.com/Jeanvncs/-web2A-Garcia-Vinces).
