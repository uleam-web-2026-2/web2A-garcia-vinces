# Taller 12 — Seguridad de la API

La rama de trabajo es `taller12`. Este directorio contiene la lista de patrones para la auditoría y esta guía de ejecución. Las capturas y salidas se deben generar con la API, Docker y Bruno ejecutándose en el entorno real del equipo.

## Preparar el Compose aislado

Desde la raíz del repositorio, en PowerShell:

```powershell
Copy-Item .env.taller12.example .env.taller12
notepad .env.taller12
```

Reemplaza todas las contraseñas y `CLAVE_FIRMA` por valores propios antes de arrancar. Para `MIGRATION_DB_PASSWORD` y `API_DB_PASSWORD`, usa letras, números, guion o guion bajo para que puedan viajar en el URL de PostgreSQL sin escapar caracteres. No uses los valores de ejemplo en un servidor público. Este Compose usa el proyecto `latent-taller12`, con contenedores y volumen separados del taller 11. Si el taller 11 ya usa `8081`, deja `PUERTO=8082` para que ambos puedan estar activos.

```powershell
docker compose -p latent-taller12 -f compose.taller12.yaml --env-file .env.taller12 up --build -d
docker compose -p latent-taller12 -f compose.taller12.yaml --env-file .env.taller12 ps
Invoke-RestMethod http://localhost:8082/healthz
```

Si configuraste `PUERTO=8081` porque está libre, consulta `http://localhost:8081/healthz`.

La primera ejecución crea los roles y el esquema en un volumen limpio, otorga los permisos mínimos y luego inicia la API. Para detener el entorno sin borrar los datos:

```powershell
docker compose -p latent-taller12 -f compose.taller12.yaml --env-file .env.taller12 down
```

No ejecutes `down -v` si quieres conservar la base de demostración. El volumen se declara como `latent_taller12_data` dentro del proyecto `latent-taller12`.

## Evidencias que debes generar y subir

No se incluyen salidas de ejemplo como si fueran ejecuciones reales. Luego de ejecutar cada paso en tu máquina, guarda las evidencias reales en este directorio:

- `peticiones.md`: resultados de las seis peticiones Bruno antes de defensas, con defensas y con el usuario restringido.
- `auditoria.txt`, `auditoria.md` y, si cambiaste código, `auditoria_despues.txt`.
- `inyeccion.txt`: petición normal y petición manipulada; la segunda no debe ampliar los resultados.
- `cors_permitido.txt`, `cors_no_permitido.txt`, `cambio_estado.json` y `cors_curl_ejecuta.txt`.
- `login_malo.json`, `rafaga.txt` y `logs_seis.txt`.
- Reto: `reto_drop.txt` y `reto_usuarios.txt`.
- Si la prueba exige capturas, adjunta capturas directas de las herramientas mostrando la ejecución real; no las fabriques ni edites.

Guarda `.txt` en UTF-8. La guía de comandos está en `S12/taller_semana_12_seguridad_de_la_api.md`; ejecuta las peticiones contra tu propio servicio local. Las salidas deben corresponder a la versión de código que subiste.

## Comandos de auditoría

Desde la raíz del repositorio, Git Bash:

```bash
git grep -n -E -f docs/taller12/patrones.txt -- '*.go' ':(exclude)*_test.go' > docs/taller12/auditoria.txt
```

Revisa cada coincidencia, incluyendo falsos positivos, y completa una fila por coincidencia en `auditoria.md`. Si modificas código, repite la búsqueda en `auditoria_despues.txt`.
