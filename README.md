# Latent

**Aplicaciones Web II · ULEAM · Web II A · Período 2026-2**

## Integrantes

| Integrante | Usuario de GitHub | Paralelo |
| --- | --- | --- |
| Ordeoñez Jostin Dionicio | Jostin12Garcia | Web II A |
| Vinces Chonillo Jean Carlos | Jeanvncs | Web II A |

Latent es una cámara desechable digital para bodas, cumpleaños y fiestas. El anfitrión crea un evento y comparte su código; los invitados entran desde el celular y cada persona tiene un número limitado de fotos. Al terminar la fiesta, el anfitrión revela el álbum completo.

## Aplicación web

La interfaz está en `servidor-pareja/web/` y se sirve desde el mismo servidor Go:

- Portada de Latent y acceso para anfitriones e invitados.
- Panel del anfitrión para crear eventos, consultar el código, iniciar o cancelar la fiesta y revelar el álbum.
- Cámara móvil para invitados, con vista previa, límite de tamaño y contador de fotos.
- Galería de fotos para anfitrión e invitados.

Las fotos seleccionadas se envían al servidor como imágenes en base64 dentro de `url_archivo`. Para la demostración, cada imagen puede pesar hasta 4 MB.

## Requisitos

- Go (versión estable)
- PostgreSQL

Crea una base de datos llamada `Latent`. Desde PowerShell, en la carpeta `servidor-pareja`, configura el entorno y arranca el servidor:

```powershell
$env:PUERTO = "8081"
$env:DATABASE_URL = "postgres://usuario:clave@localhost:5432/Latent?sslmode=disable"
$env:CLAVE_FIRMA = "una-frase-larga-para-firmar-tokens"
go run .
```

Abre `http://localhost:8081` en el navegador. La página y la API usan el mismo puerto.

## Cuentas de demostración

El servidor crea estas cuentas de ejemplo al iniciar:

| Rol | Usuario | Contraseña |
| --- | --- | --- |
| Anfitrión | `luis` | `luis-2026` |
| Invitada | `ana` | `ana-2026` |

El flujo de invitados solicita además el código de evento, un nombre y un PIN de cuatro dígitos. Estas credenciales solo sirven para demostraciones locales.

## Rutas principales

| Método | Ruta | Acceso |
| --- | --- | --- |
| `POST` | `/login` | Público; devuelve un token |
| `POST` | `/eventos/unirse` | Público; registra al invitado con código, nombre y PIN |
| `GET` | `/eventos` | Anfitrión; lista sus eventos |
| `POST` | `/eventos` | Anfitrión; crea un evento |
| `GET` | `/eventos/{id}` | Anfitrión propietario; consulta un evento |
| `PATCH` | `/eventos/{id}/estado` | Anfitrión propietario; cambia el estado |
| `GET` | `/eventos/{id}/resumen` | Usuario autenticado; devuelve los datos de cámara necesarios |
| `GET` | `/eventos/{id}/fotos` | Usuario autenticado; consulta la galería |
| `POST` | `/fotos` | Invitado autenticado; guarda una foto |

Las rutas protegidas usan `Authorization: Bearer <token>`. `CLAVE_FIRMA`, `DATABASE_URL` y `PUERTO` se leen del entorno; el servidor no carga archivos `.env` automáticamente.

## Repositorio

- `servidor-pareja/main.go`: configuración, migración y arranque del servidor.
- `servidor-pareja/internal/latent/`: eventos, invitados, fotos y rutas de la API.
- `servidor-pareja/internal/sesion/`: autenticación, tokens y roles.
- `servidor-pareja/web/`: interfaz web.
- `docs/`: ficha del negocio, addendum técnico y capturas del Hito 1.

## Colaboración

El repositorio recomienda al menos un commit de cada integrante por taller. Incluyan en el cuerpo del commit una línea con la herramienta de IA usada y para qué parte, y no suban credenciales reales.


## Historial previo

El trabajo anterior está en [el repositorio previo](https://github.com/Jeanvncs/-web2A-Garcia-Vinces).

