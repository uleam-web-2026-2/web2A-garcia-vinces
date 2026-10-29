param(
    [string]$Directorio = $PSScriptRoot
)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing

function Draw-Capture([string]$Path, [string]$Title, [string[]]$Lines) {
    $width = 1440
    $lineHeight = 30
    $height = [Math]::Max(220, 118 + ($Lines.Count * $lineHeight))
    $bitmap = [System.Drawing.Bitmap]::new($width, $height)
    $graphics = [System.Drawing.Graphics]::FromImage($bitmap)
    $graphics.Clear([System.Drawing.Color]::FromArgb(15, 23, 42))
    $titleFont = [System.Drawing.Font]::new('Segoe UI', 22, [System.Drawing.FontStyle]::Bold)
    $bodyFont = [System.Drawing.Font]::new('Consolas', 15)
    $mutedFont = [System.Drawing.Font]::new('Segoe UI', 11)
    $white = [System.Drawing.Brushes]::White
    $green = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(134, 239, 172))
    $muted = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(148, 163, 184))
    $graphics.DrawString($Title, $titleFont, $white, 42, 30)
    $graphics.DrawString('Latent · evidencia automatizada desde prueba de integración Go (httptest + SQLite)', $mutedFont, $muted, 44, 76)
    $y = 116
    foreach ($line in $Lines) {
        $remaining = $line
        while ($remaining.Length -gt 112) {
            $graphics.DrawString($remaining.Substring(0, 112), $bodyFont, $green, 44, $y)
            $remaining = '  ' + $remaining.Substring(112)
            $y += $lineHeight
        }
        $graphics.DrawString($remaining, $bodyFont, $green, 44, $y)
        $y += $lineHeight
    }
    $bitmap.Save($Path, [System.Drawing.Imaging.ImageFormat]::Png)
    $muted.Dispose(); $green.Dispose(); $titleFont.Dispose(); $bodyFont.Dispose(); $mutedFont.Dispose()
    $graphics.Dispose(); $bitmap.Dispose()
}

$integrationLog = Join-Path $Directorio 'integracion_hito2.txt'
if (-not (Test-Path -LiteralPath $integrationLog)) { throw "No existe $integrationLog; ejecuta antes la prueba de integración." }
$records = Get-Content -LiteralPath $integrationLog
foreach ($line in $records) {
    if ($line -match 'CAPTURA ([0-6]) \| (.*)') {
        $number = [int]$Matches[1]
        $filename = 'peticion-{0:D2}.png' -f $number
        Draw-Capture (Join-Path $Directorio $filename) "Hito 2 · Petición $number" @($Matches[2])
    }
}

$testLog = Join-Path $Directorio 'go_test.txt'
if (Test-Path -LiteralPath $testLog) {
    $testLines = @(Get-Content -LiteralPath $testLog | Select-Object -Last 14)
    Draw-Capture (Join-Path $Directorio 'go-test.png') 'Hito 2 · go test ./...' $testLines
}
