$ErrorActionPreference = 'Stop'

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$outImage = Join-Path $PSScriptRoot 'salidas-imagen.txt'
$outContainer = Join-Path $PSScriptRoot 'salidas-contenedor.txt'

Push-Location $repoRoot
try {
    $imageOutput = docker image inspect latent-api:compose --format 'ID={{.Id}} Tamaño={{.Size}} bytes Creada={{.Created}}'
    if ($LASTEXITCODE -ne 0) { throw 'No se encontró latent-api:compose. Ejecuta docker compose up --build -d primero.' }
    $imageOutput | Set-Content -LiteralPath $outImage -Encoding utf8

    $containerOutput = docker compose --env-file .env.compose ps
    if ($LASTEXITCODE -ne 0) { throw 'No se pudo consultar docker compose ps.' }
    $healthOutput = Invoke-RestMethod -Uri 'http://localhost:8081/healthz'
    $logsOutput = docker compose --env-file .env.compose logs --no-color api
    if ($LASTEXITCODE -ne 0) { throw 'No se pudieron leer los registros de la API.' }

    @(
        '=== docker compose ps ==='
        $containerOutput
        ''
        '=== GET /healthz ==='
        ($healthOutput | ConvertTo-Json -Compress)
        ''
        '=== docker compose logs api ==='
        $logsOutput
    ) | Set-Content -LiteralPath $outContainer -Encoding utf8

    Write-Host "Evidencia guardada:`n$outImage`n$outContainer"
}
finally {
    Pop-Location
}
