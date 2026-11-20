# Peticiones de verificación — Semana 12

Los resultados de S12 se registraron a partir de las respuestas reales de Bruno. La ejecución se hizo contra Compose con las defensas y el rol restringido `latent_api` activos a la vez; por eso los mismos códigos aparecen en ambas columnas de S12, no son dos ejecuciones independientes. La línea base de S11 queda pendiente porque no se adjuntó su evidencia en esta tanda. La petición **0 Preparación** deja listo el invitado; las peticiones 1–6 son las seis verificaciones del Hito 2.

| N.º | Método y ruta | Autenticación / rol | Qué comprueba | Esperado | Antes (S11) | Con defensas (S12) | Con usuario API restringido | Evidencia recibida |
|---:|---|---|---|---:|---|---|---|---|
| 0 Preparación | `POST /eventos/unirse` | Código de acceso | Registra al invitado de prueba. | 200 | Pendiente | 200 | 200 | Omitida: la captura dejaba visible el token. |
| 1 | `GET /eventos/{id}/fotos` | Sin token | El middleware rechaza una ruta protegida. | 401 | Pendiente | 401 | 401 | [Bruno 401](capturas/bruno-01-sin-token-401.png) |
| 2 | `POST /login` | Credenciales de anfitrión | Inicia sesión y obtiene el token del anfitrión. | 200 | Pendiente | 200 | 200 | [Bruno login 200](capturas/bruno-02-token-anfitrion.png) |
| 3 | `GET /eventos/{id}/fotos` | Token de invitado | Permite consultar las fotos del evento. | 200 | Pendiente | 200 | 200 | [Bruno 200](capturas/bruno-03-acceso-invitado-200.png) |
| 4 | `PATCH /eventos/{id}` | Token de invitado | Rechaza una operación reservada al anfitrión. | 403 | Pendiente | 403 | 403 | [Bruno 403](capturas/bruno-04-rol-invitado-403.png) |
| 5 | `PATCH /eventos/{id}` | Token de anfitrión propietario | Acepta un cambio de estado válido. | 200 | Pendiente | 200 | 200 | [Bruno 200](capturas/bruno-05-rol-anfitrion-200.png) |
| 6 | `PATCH /eventos/{id}` | Token de anfitrión | Rechaza `finalizado → activo`. | 422 | Pendiente | 422 | 422 | [Bruno 422](capturas/bruno-06-transicion-prohibida-422.png) |

> Para completar el entregable falta adjuntar la línea base real de S11. La captura del login corresponde al entorno local de demostración.
