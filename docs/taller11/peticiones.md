# Semana 11 · tabla de peticiones

Base URL local: `http://localhost:8081`. Las rutas de negocio se prueban con la colección Bruno de `docs/hito2_bruno/`. La tabla registra el propósito y los estados esperados observados en las pruebas previas del proyecto; al repetirlas mediante Compose, anota los estados reales en la evidencia de la demostración.

| Orden | Método y ruta | Autenticación | Qué comprueba | HTTP esperado |
|---:|---|---|---|---:|
| 1 | `GET /healthz` | No | La API responde después de iniciar el contenedor. | 200 |
| 2 | `POST /eventos/unirse` | Código de acceso | Registra al invitado y entrega su token. | 200 |
| 3 | `GET /eventos/{id}/fotos` sin token | No | El middleware rechaza una petición protegida. | 401 |
| 4 | `POST /login` | Credenciales de anfitrión | Inicia sesión y entrega el token del anfitrión. | 200 |
| 5 | `GET /eventos/{id}/fotos` como invitado | Token de invitado | Permite consultar la galería del evento correspondiente. | 200 |
| 6 | `PATCH /eventos/{id}` como invitado | Token de invitado | Rechaza una operación reservada al anfitrión. | 403 |
| 7 | `PATCH /eventos/{id}` como anfitrión | Token de anfitrión | Permite cambiar el estado cuando la transición es válida. | 200 |
| 8 | `PATCH /eventos/{id}` de finalizado a activo | Token de anfitrión | Rechaza una transición inválida del negocio. | 422 |

Los IDs y códigos usados por la colección Bruno corresponden a la base de demostración y pueden variar si se cambia el seed. No copies tokens JWT a los archivos de evidencia.
