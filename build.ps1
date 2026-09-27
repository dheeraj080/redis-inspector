# Cross-Platform Build Script for redis-inspector
$appName = "redis-inspector"
$outputDir = "dist"

# Clean old build artifacts
if (Test-Path $outputDir) {
    Remove-Item -Path $outputDir -Recurse -Force
}
New-Item -ItemType Directory -Path $outputDir | Out-Null

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Ext = ".exe" },
    @{ GOOS = "windows"; GOARCH = "arm64"; Ext = ".exe" },
    @{ GOOS = "linux";   GOARCH = "amd64"; Ext = "" },
    @{ GOOS = "linux";   GOARCH = "arm64"; Ext = "" },
    @{ GOOS = "darwin";  GOARCH = "amd64"; Ext = "" },
    @{ GOOS = "darwin";  GOARCH = "arm64"; Ext = "" }
)

Write-Host "`n🚀 Compiling cross-platform binaries..." -ForegroundColor Cyan

foreach ($target in $targets) {
    $os = $target.GOOS
    $arch = $target.GOARCH
    $ext = $target.Ext
    $outputFile = "$outputDir/${appName}-${os}-${arch}${ext}"

    Write-Host "  -> Building $os/$arch..." -NoNewline
    
    $env:GOOS = $os
    $env:GOARCH = $arch

    go build -ldflags="-s -w" -o $outputFile ./cmd/redis-inspector 2>&1 | Out-Null

    if ($LASTEXITCODE -eq 0) {
        Write-Host " [OK]" -ForegroundColor Green
    } else {
        Write-Host " [FAILED]" -ForegroundColor Red
    }
}

# Clean environment variables
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Host "`n✨ Build complete! Binaries ready in ./${outputDir}" -ForegroundColor Green
