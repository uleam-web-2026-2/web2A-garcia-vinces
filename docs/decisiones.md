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

Se tocaron `servidor-pareja/internal/latent/fotos_servicio.go`, `fotos_repositorio.go`, `fotos_repositorio_sqlc.go`, `reservas.go`, `manejadores.go`, `rutas.go`, `main.go`, `latent_test.go`, `postgres_transacciones_test.go`, `servidor-pareja/db/schema.sql`, `db/consultas.sql`, los fuentes generados de `internal/latent/filas`, y `docs/taller10/error_forzado.txt`.

`eventos_servicio.go`, las rutas HTTP y los códigos esperados por las seis peticiones no cambiaron. `AutoMigrate` continúa siendo responsable de crear las tablas; el esquema SQL es descriptivo para SQLC.

## Semana 11 · decisiones de contenedores (D1–D9)

| ID | Decisión | Motivo y efecto |
|---|---|---|
| D1 | Construir la API con una imagen multietapa. | El compilador y las dependencias de construcción no pasan a la imagen final. |
| D2 | Compilar con `CGO_ENABLED=0`. | El binario Go queda estático y puede correr en una imagen mínima. |
| D3 | Usar Distroless Debian 13 como runtime. | La imagen final contiene el binario y certificados necesarios, sin shell ni gestor de paquetes. |
| D4 | Ejecutar como usuario no root. | Reduce los privilegios del proceso dentro del contenedor. |
| D5 | Entregar configuración por variables de entorno. | Las claves y la URL de base no se incluyen en la imagen ni en el código. |
| D6 | Publicar el puerto del servicio con `8081:8081`. | El puerto izquierdo es el de la máquina; el derecho es el que escucha la API dentro del contenedor. |
| D7 | En el reto Compose, conectar la API a PostgreSQL por el nombre de servicio `db`. | Los contenedores comparten una red y no deben usar `localhost` para comunicarse entre sí. |
| D8 | Persistir los datos de PostgreSQL en un volumen nombrado. | Recrear el contenedor de la base no borra sus datos; `down` conserva el volumen. |
| D9 | Esperar a que PostgreSQL pase su healthcheck antes de iniciar la API. | Evita que la API intente conectarse mientras la base todavía está arrancando. |

El reto Compose es local y de demostración. Usa credenciales de ejemplo reemplazables; no debe publicarse con esos valores.

## Semana 12: la seguridad de la API

### D1 Lo que encontramos y cerramos

La búsqueda de patrones devuelve 32 coincidencias candidatas. La cifra incluye llamadas SQL, consultas GORM y falsos positivos como `fmt.Sprintf`, concatenaciones para mensajes/rutas y las interfaces de sqlc; la salida y la clasificación línea por línea se deben guardar y completar en `docs/taller12/auditoria.txt` y `docs/taller12/auditoria.md` al ejecutar la auditoría real.

En la revisión del código, las consultas SQL de `reservas.go` usan `?` con argumentos separados, las consultas sqlc usan `$1`, `$2`, etc., los filtros GORM pasan valores como argumentos y los nombres de orden están fijos en el código. Los ID de ruta pasan por `strconv.Atoi` y se rechazan con `400` si no son enteros positivos. Se añadió una prueba automatizada para la ruta con `1 OR 1=1`; falta ejecutar y guardar la petición manual con el servidor local en `inyeccion.txt`.

### D2 El origen que puede leer nuestras respuestas

Configuramos `http://localhost:5173`, que corresponde al origen local de desarrollo previsto para el front. No usamos `*` para limitar qué sitio puede leer las respuestas. CORS se registra inmediatamente después de crear el router y antes de otros middleware/rutas, de modo que atienda las peticiones previas `OPTIONS` antes de la autenticación. La API acepta los métodos y encabezados que usa el proyecto.

`curl` no aplica la política CORS del navegador: una petición enviada desde `http://malo.example` puede ser procesada por el servidor aunque el navegador no permita que ese sitio lea la respuesta. La verificación real de las cabeceras y del cambio de estado queda pendiente en `cors_permitido.txt`, `cors_no_permitido.txt` y `cors_curl_ejecuta.txt`. Al publicar el front, `AllowedOrigins` debe cambiarse al origen HTTPS exacto del front desplegado.

### D3 El límite del login

`POST /login` admite cinco intentos por minuto y dirección IP. Es una ruta pública que verifica credenciales y puede ser objetivo de intentos repetidos; por eso el límite se aplica solo al login y no bloquea lecturas o acciones protegidas de usuarios legítimos. El sexto intento dentro de la ventana debe recibir `429` y `Retry-After`. La ráfaga real y el valor observado de `Retry-After` se documentarán en `docs/taller12/rafaga.txt`; todavía debe ejecutarla el equipo.

### D4 Detrás de un proxy

El límite toma la IP de la conexión remota mediante `ClientIPFromRemoteAddr`, sin aceptar directamente `X-Forwarded-For`, que el cliente podría falsificar. Detrás de un proxy todas las conexiones pueden aparecer con la IP del proxy. En ese despliegue se debe configurar una cadena de proxies de confianza y validar sus rangos antes de aceptar una cabecera reenviada; confiar en cualquier `X-Forwarded-For` permitiría evadir el límite cambiándola en cada petición.

> **Estado de las evidencias:** las salidas de auditoría, peticiones, ráfaga, logs y prueba de permisos deben generarse desde la ejecución real por los integrantes. No se han inventado ni agregado archivos de evidencia.
