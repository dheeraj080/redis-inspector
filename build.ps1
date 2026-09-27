$AppName = "redis-inspector"
$Version = "1.0.0"
$BuildTime = (Get-Date).ToString("yyyy-MM-dd_HH:mm:ss")
$CommitHash = "git-$(Get-Random -Minimum 10000 -Maximum 99999)"

Write-Host "🔨 Building $AppName v$Version ($BuildTime)..." -ForegroundColor Cyan

# Ensure dist directory exists
if (!(Test-Path "dist")) {
    New-Item -ItemType Directory -Path "dist" | Out-Null
}

$Platforms = @(
    @{ OS = "windows"; Arch = "amd64"; Ext = ".exe" },
    @{ OS = "linux";   Arch = "amd64"; Ext = "" },
    @{ OS = "darwin";  Arch = "amd64"; Ext = "" },
    @{ OS = "darwin";  Arch = "arm64"; Ext = "" }
)

foreach ($p in $Platforms) {
    $Env:GOOS = $p.OS
    $Env:GOARCH = $p.Arch
    $OutName = "$AppName-$Version-$($p.OS)-$($p.Arch)$($p.Ext)"
    $OutPath = "dist/$OutName"

    Write-Host "  -> Compiling for $($p.OS) / $($p.Arch)..." -ForegroundColor Yellow

    go build -ldflags="-s -w -X main.Version=$Version -X main.BuildTime=$BuildTime" -o $OutPath ./cmd/redis-inspector

    if ($LASTEXITCODE -eq 0) {
        $size = (Get-Item $OutPath).Length / 1MB
        Write-Host "     Successfully built: $OutPath ($([Math]::Round($size, 2)) MB)" -ForegroundColor Green
    } else {
        Write-Host "     Failed to build for $($p.OS)/$($p.Arch)" -ForegroundColor Red
    }
}

Write-Host "`n✨ All builds completed! Check the 'dist/' folder." -ForegroundColor Green
