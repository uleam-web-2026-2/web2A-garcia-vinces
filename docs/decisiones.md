# Decisiones de diseño — Latent

## Semana 7 D1 · Nivel de acceso de cada ruta

| Ruta | Nivel | Por qué |
| --- | --- | --- |
| `POST /login` | Pública | Una cuenta anfitrión obtiene su sesión aquí. Las credenciales se verifican con bcrypt y el servidor firma un JWT. |
| `POST /eventos/unirse` | Pública | El invitado llega desde un QR/código, sin crear cuenta. El servidor valida nombre y PIN, registra la plaza y emite un token ligado a ese invitado y a ese evento. |
| `GET /eventos` | Con sesión, solo anfitrión | Solo el anfitrión puede consultar la lista de los eventos que administra. Un invitado no recibe códigos de acceso de otras fiestas. |
| `GET /eventos/{id}` | Con sesión y propietario, solo anfitrión | El detalle incluye el código de acceso y la configuración privada del evento. |
| `POST /eventos` | Con sesión, solo anfitrión | Crear y configurar una fiesta es una acción del anfitrión. |
| `PUT /eventos/{id}` | Con sesión y propietario, solo anfitrión | Solo se ajusta una fiesta propia que aún no empezó. |
| `PATCH /eventos/{id}` | Con sesión y propietario, solo anfitrión | El anfitrión controla el inicio, cierre y cancelación. Invitados no pueden alterar el ciclo de vida. |
| `DELETE /eventos/{id}` | Con sesión y propietario, solo anfitrión | Solo se elimina un evento propio sin invitados ni fotos y que no esté activo o revelado; esto protege el álbum y su auditoría. |
| `GET /eventos/{id}/resumen` y `/fotos` | Con sesión y pertenencia | El anfitrión dueño o una persona con token de invitado de esa fiesta puede consultar sus datos. El invitado ve las fotos después del revelado. |
| `POST /fotos` | Con sesión, solo invitado | El JWT incluye el ID del invitado; el servidor no confía en un ID que el cliente pueda cambiar para gastar la cuota de otra persona. |

## Semana 7 D2 · La ruta de un solo rol

- **Decisión:** `PATCH /eventos/{id}` y las operaciones de escritura del evento son exclusivas del anfitrión propietario; `POST /fotos` es exclusivo de un invitado con sesión.
- **Por qué:** el anfitrión responde por el inicio y final de la fiesta; cada invitado solo administra sus propias capturas. Si cualquier invitado pudiera cambiar el estado, podría revelar o cancelar el álbum antes de tiempo. Si pudiera presentar el ID de otra persona, podría gastar su cuota de fotos.
- **Qué descartamos:** un único rol genérico para todos. Haría que el cliente pudiera intentar acciones ajenas al negocio y trasladaría la autorización a la interfaz, que se puede modificar.
- **Cómo se aplica:** primero `tokens.Autenticar` valida la firma y expiración; luego `sesion.SoloRol` restringe rol y los manejadores comprueban propiedad o pertenencia al evento.

## Semana 7 D3 · La transición prohibida

- **Transición:** `finalizado → activo`.
- **Por qué el negocio no la permite:** al finalizar, el anfitrión revela un álbum cerrado. Volver a abrir la cámara permitiría añadir fotos después del revelado y rompería la cuota y el orden de la fiesta.
- **Qué responde y dónde se valida:** `422 Unprocessable Entity`, código `transicion_prohibida`. `ServicioEventos.CambiarEstado` consulta el mapa `transiciones` de `modelos.go` antes de guardar. La ruta protegida está en `internal/latent/rutas.go`.

## Semana 8 · Estructura en capas

El flujo del cambio de estado atraviesa tres responsabilidades:

1. **HTTP y autenticación — `manejadores.go`, `rutas.go`, `internal/sesion/`:** Chi valida la sesión y el rol; el manejador interpreta el JSON, convierte errores de dominio en HTTP y serializa la respuesta común.
2. **Reglas del negocio — `eventos_servicio.go`:** crea, edita, lista, consulta, elimina y transiciona eventos. Aquí se valida la cuota de configuración, la propiedad, los estados y las condiciones para borrar. No depende de Chi ni de GORM.
3. **Persistencia — `eventos_repositorio.go`:** define `RepositorioEventos` e implementa sus operaciones con GORM. Traduce `record not found` al error de dominio `ErrEventoNoEncontrado`.

`modelos.go` declara las entidades y la máquina de estados. Para cambiar el motor de datos se reemplaza la implementación del repositorio, manteniendo las reglas de `eventos_servicio.go` y los formatos HTTP. La persistencia usa PostgreSQL en ejecución y SQLite en memoria para pruebas de integración.

## Semana 7 D4 · JSON mal formado y datos de negocio

JSON que no se puede decodificar responde `400 json_invalido`. Un evento con nombre vacío o con cuotas fuera de rango responde `400 evento_invalido`; una transición desconocida o prohibida responde `422`. La aplicación distingue forma de la petición, datos del evento y regla de ciclo de vida para que el cliente reciba el motivo concreto.


## Semana 10 transacciones

### D1 El caso

`POST /fotos` consume una captura de la cuota del invitado y crea una fila en `fotos`. Si se incrementa `invitados.fotos_tomadas` y falla después la inserción de `fotos` (por ejemplo, por un UUID duplicado), el contador queda consumido sin una foto correspondiente. Ambas escrituras forman un único caso de uso.

### D2 Opciones que consideramos

1. Hacer dos llamadas de escritura independientes desde el servicio. Si la segunda falla, la primera queda guardada; no cumple la atomicidad.
2. Abrir la transacción en el servicio usando GORM o `database/sql`. Haría posible el rollback, pero acoplaría el negocio a una tecnología y obligaría a cambiar el servicio al reemplazar el adaptador.
3. Exponer una operación `GuardarFoto` en `RepositorioFotos` y dejar que cada adaptador la ejecute dentro de su propia transacción. Mantiene el servicio independiente y permite que GORM y SQLC resuelvan la atomicidad con su conexión.

### D3 La que elegimos y por qué

Elegimos la tercera. Latent ya necesitaba que el contador de cuota y el álbum coincidieran. `ServicioFotos` recibe un repositorio por interfaz y coordina el caso; `RepositorioFotos.GuardarFoto` es la operación atómica. El adaptador GORM usa `Transaction`; el adaptador SQLC usa `BeginTx`, las consultas generadas y `WithTx`.

### D4 La que descartamos y por qué

Descartamos hacer las dos escrituras por separado: una colisión de UUID o un fallo de almacenamiento dejaría una cuota gastada sin una fila en el álbum. También descartamos abrir la transacción desde el servicio porque introduciría GORM o `database/sql` en la capa de negocio y haría que el cambio de adaptador tocara reglas del caso de uso.

### D5 Con sqlc

SQLC genera `EstadoCuotaFoto`, `IncrementarCuotaFoto` y `CrearFoto`. El adaptador inicia una transacción SQL, usa `q.WithTx(tx)` para ejecutar el incremento y el INSERT en la misma conexión, y confirma solo después de ambas operaciones. Un error del INSERT provoca el rollback. `ServicioFotos` no cambia al escoger GORM o SQLC.

### D6 Archivos

Se tocaron `servidor-pareja/internal/latent/fotos_servicio.go`, `fotos_repositorio.go`, `fotos_repositorio_sqlc.go`, `reservas.go`, `manejadores.go`, `rutas.go`, `main.go`, `latent_test.go`, `servidor-pareja/db/schema.sql`, `db/consultas.sql`, los fuentes generados de `internal/latent/filas`, y `docs/taller10/error_forzado.txt`.

`eventos_servicio.go`, las rutas HTTP y los códigos esperados por las seis peticiones no cambiaron. `AutoMigrate` continúa siendo responsable de crear las tablas; el esquema SQL es descriptivo para SQLC.
