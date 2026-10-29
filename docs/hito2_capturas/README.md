# Respaldos de la demostración

`peticion-00.png` a `peticion-06.png` muestran las respuestas observadas por `TestDemoHito2RespuestasDeRolesYTransicion`. `go-test.png` resume el resultado de `go test ./...`. Los archivos `integracion_hito2.txt` y `go_test.txt` conservan la salida textual completa.

Son respaldos generados desde una prueba de integración HTTP en memoria, no capturas de Bruno ni evidencia de un servidor desplegado. Los JWT se omiten. Para regenerar las imágenes después de ejecutar las pruebas:

```powershell
cd servidor-pareja
$env:GOCACHE = Join-Path $env:TEMP 'latent-hito2-gocache'
go test -v ./... 2>&1 | Tee-Object -FilePath '..\docs\hito2_capturas\go_test.txt'
go test -run TestDemoHito2RespuestasDeRolesYTransicion -v ./internal/latent 2>&1 | Tee-Object -FilePath '..\docs\hito2_capturas\integracion_hito2.txt'
..\docs\hito2_capturas\generar-capturas.ps1
```
