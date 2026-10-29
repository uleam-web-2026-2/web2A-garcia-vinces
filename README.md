# Latent

**Aplicaciones Web II · ULEAM · Web II A · Período 2026-2**

| Integrante | Usuario de GitHub | Paralelo |
| --- | --- | --- |
| Ordeoñez Jostin Dionicio | Jostin12Garcia | Web II A |
| Vinces Chonillo Jean Carlos | Jeanvncs | Web II A |

Latent es una cámara desechable digital para fiestas. El anfitrión crea el evento y comparte un código QR; cada invitado entra con nombre y PIN, tiene una cuota de fotos controlada por el servidor y ve el álbum cuando el anfitrión finaliza la fiesta.

## Desarrollo local

Requisitos: Go 1.27.1 o posterior y PostgreSQL 17 o posterior. Crea la base `Latent` y, desde PowerShell, en `servidor-pareja`, configura el entorno:

```powershell
$env:PUERTO = "8081"
$env:DATABASE_URL = "postgres://usuario:clave@localhost:5432/Latent?sslmode=disable"
$env:CLAVE_FIRMA = "reemplaza-por-un-secreto-aleatorio-de-32-caracteres"
$env:MODO_DEMO = "true"
go run .
```

Abre `http://localhost:8081`. La aplicación web y la API usan el mismo servidor. La cuenta anfitrión de demostración es `luis / luis-2026`; el login de ejemplo para validar roles es `ana / ana-2026`. **No habilites `MODO_DEMO` en una instancia pública.**

La web cachea su estructura después de la primera visita. Si se pierde la señal, guarda la foto preparada en IndexedDB y la intenta sincronizar cuando vuelve la conexión. El servidor valida siempre la cuota; una foto offline no queda confirmada hasta que la API la acepta.

## Despliegue con Docker Compose

1. Copia `.env.example` como `.env` en la raíz.
2. Cambia `POSTGRES_PASSWORD` y `CLAVE_FIRMA` por valores largos y aleatorios. Mantén `DATABASE_URL` sincronizada con el usuario, contraseña y base configurados para PostgreSQL.
3. Para una instancia de producción, establece `MODO_DEMO=false`, `ANFITRION_USUARIO` y `ANFITRION_CLAVE` (mínimo 12 caracteres). Termina TLS en un proxy HTTPS delante del puerto de la aplicación.
4. Ejecuta `docker compose up --build -d` y verifica `http://localhost:8081/healthz`.

Compose conserva la base en un volumen nombrado y no publica el puerto de PostgreSQL al host. La imagen de la aplicación corre sin root. No uses `go run . -reset` ni `docs/hito2_demo_seed.sql` sobre datos desplegados: `-reset` borra las tablas y el seed altera eventos de IDs fijos.

### Límite actual de almacenamiento de fotos

Para el prototipo, cada foto puede enviarse como data URL de hasta 4 MiB y se guarda en PostgreSQL junto a sus metadatos. El volumen de PostgreSQL conserva esas fotos en el despliegue Compose y las copias de seguridad de la base las incluyen. Para abrir el servicio a eventos grandes o tráfico público, el siguiente paso es mover los bytes a almacenamiento de objetos privado (S3 compatible) y guardar en PostgreSQL solo claves y URLs firmadas; el límite por invitado reduce el crecimiento, pero no reemplaza ese almacenamiento.

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
- `servidor-pareja/web/`: aplicación móvil/web y soporte offline básico.
- `compose.yaml` y `servidor-pareja/Dockerfile`: despliegue local autocontenido.

El trabajo anterior está en [el repositorio previo](https://github.com/Jeanvncs/-web2A-Garcia-Vinces).
