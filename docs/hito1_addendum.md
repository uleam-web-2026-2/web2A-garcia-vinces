# Hito 1 · Addendum técnico

**Pareja:** [Nombre] Garcia · Vinces
**Paralelo:** Aplicaciones Web II A

## A. Estructura del proyecto

```
web2A-garcia-vinces/             raíz del repositorio
├── .github/                     flujos de GitHub Actions
├── .gitignore                   archivos que no se suben (por ejemplo, .env)
├── README.md                    cómo levantar el servidor desde cero
├── docs/                        ficha, addendum y capturas del Hito 1
└── servidor-pareja/             proyecto en Go
    ├── main.go                  punto de entrada: levanta el servidor en el puerto 8081
    ├── go.mod, go.sum           módulo y dependencias de Go
    ├── .env.example             variables de entorno de ejemplo, sin datos reales
    ├── docs/
    │   └── decisiones.md        decisiones de diseño del proyecto
    └── internal/
        ├── latent/              el negocio: eventos, invitados y fotos
        │   ├── modelos.go       structs de Usuario, Evento, Invitado y Foto
        │   ├── manejadores.go   un manejador por endpoint, con sus validaciones
        │   ├── rutas.go         rutas y verbos de la API
        │   └── latent_test.go   pruebas de las reglas de negocio
        ├── middleware/          middleware.go: funciones que se ejecutan antes de los manejadores
        └── respuesta/           formato común de las respuestas JSON
```

## B. Configuración y secretos

| Variable | Para qué sirve | Ejemplo (sin datos reales) |
|----------|----------------|----------------------------|
| PUERTO | Puerto en el que escucha el servidor | 8081 |
| DATABASE_URL | Cadena de conexión a PostgreSQL | postgres://usuario:clave@localhost:5432/latent |

**Archivo de ejemplo:** `servidor-pareja/.env.example`. El repositorio no tiene credenciales reales: `.env` está en `.gitignore`.

## C. Pruebas

Las pruebas están en `servidor-pareja/internal/latent/latent_test.go`, una por cada regla de negocio del Hito 1.

| Prueba | Qué caso cubre |
|--------|----------------|
| Evento inválido | `POST /eventos` sin `nombre` responde 400 |
| Código inexistente | `POST /eventos/unirse` con un código que no existe responde 404 |
| Cuota agotada | Con 3 fotos tomadas, `POST /fotos` responde 403 |
| Estado desconocido | `PATCH /eventos/{id}/estado` con un estado que no existe responde 400 |
| Transición prohibida | `PATCH /eventos/{id}/estado` de `finalizado` a `activo` responde 409 y no cambia el estado |

Se corren con `go test ./...` desde la carpeta `servidor-pareja`.

**Captura de `go test ./...`:** ![Pruebas](hito1_pruebas.png)

## D. Boceto de la pantalla principal

Pantalla: la galería del álbum, que consume `GET /eventos/{id}/fotos`. En el boceto se ven la foto, el nombre del invitado que la tomó, el estado de sincronización y, arriba, el estado del evento.

![Boceto](hito1_boceto.png)

## E. Diagrama de secuencia del caso de uso principal

Caso: un invitado toma una foto sin señal y esta se sube sola al recuperarla.

```mermaid
sequenceDiagram
    Invitado->>Cámara: toma una foto
    Cámara->>Cámara: guarda la foto en el celular con un uuid (sincronizado = no)
    Note over Cámara: el salón no tiene señal
    Cámara->>API: POST /fotos (al recuperar la señal)
    API->>BD: valida que fotos_tomadas sea menor que disparo_por_invitado
    alt el invitado aún tiene fotos
        BD-->>API: guarda la foto y suma 1 a fotos_tomadas
        API-->>Cámara: 201 con la foto guardada
        Cámara->>Cámara: marca la foto como sincronizada
    else el invitado ya tomó 3 fotos
        BD-->>API: límite alcanzado
        API-->>Cámara: 403
    end
```

## F. Capturas de respuestas

**Caso correcto:** `POST /eventos` con la cabecera `X-Usuario-ID: 1` responde 201 Created en `http://localhost:8081/eventos`.

![Respuesta correcta](hito1_respuesta_ok.png)

**Caso con error de validación:** `POST /eventos` con `nombre` vacío responde 400 Bad Request, con `"ok": false` y el mensaje de que nombre, disparo_por_invitado y max_invitado son obligatorios.

![Respuesta con error](hito1_respuesta_error.png)