# Taller 11 · Contenedores

Esta carpeta reúne la entrega del taller de Semana 11 para la API Latent.

## Contenido

- `compose.yaml` en la raíz: reto adicional que levanta API y PostgreSQL.
- [`peticiones.md`](peticiones.md): tabla de peticiones para comprobar la API.
- [`../decisiones.md`](../decisiones.md#semana-11--decisiones-de-contenedores-d1d9): decisiones D1–D9.
- [`capturar-salidas.ps1`](capturar-salidas.ps1): guarda las salidas reales de imagen y contenedor en esta carpeta.

El Dockerfile multietapa original está en `servidor-pareja/Dockerfile`. El `.env.compose.example` solo contiene valores de demostración; nunca lo uses como configuración pública.

## Reto con Compose

Desde la raíz del repositorio, en PowerShell:

```powershell
Copy-Item .env.compose.example .env.compose
docker compose --env-file .env.compose config
docker compose --env-file .env.compose up --build -d
docker compose --env-file .env.compose ps
Invoke-RestMethod http://localhost:8081/healthz
```

La respuesta esperada del healthcheck HTTP es `ok: True`. Compose espera que PostgreSQL esté saludable antes de iniciar la API. La base usa el volumen nombrado `postgres_data`, así que `docker compose down` detiene y elimina contenedores pero conserva los datos.

Para ver los registros y guardar evidencia:

```powershell
docker compose --env-file .env.compose logs --no-color api
.\docs\taller11\capturar-salidas.ps1
```

El script crea `salidas-imagen.txt` y `salidas-contenedor.txt` con información tomada de Docker en ese momento. Ejecútalo después de `docker compose up`. No se incluyen salidas ficticias.

Para detener los servicios sin borrar los datos:

```powershell
docker compose --env-file .env.compose down
```

Solo para reiniciar desde cero la base descartable de demostración se puede ejecutar `docker compose --env-file .env.compose down -v`; esto borra el volumen `postgres_data` y sus datos.

## Evidencias que se deben conservar

1. Salida de `docker image ls latent-api:compose` o `docker image inspect latent-api:compose`.
2. Salida de `docker compose ps` mostrando API y PostgreSQL en ejecución.
3. Salida de `Invoke-RestMethod http://localhost:8081/healthz`.
4. Registros de `docker compose logs api` donde la API indique que escucha en el puerto 8081.

El entorno de edición usado para preparar esta carpeta no tenía Docker Desktop iniciado, así que estas salidas deben generarse al ejecutar el reto en la máquina local.
