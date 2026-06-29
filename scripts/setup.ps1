param(
    [switch]$InitTerraform = $false
)

Write-Host "=== GitOps Lite Platform Setup ===" -ForegroundColor Cyan

Write-Host "`n[1/3] Verificando dependências..." -ForegroundColor Yellow
$missing = @()

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    $missing += "Go (https://go.dev/dl/)"
}
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    $missing += "Docker (https://docs.docker.com/get-docker/)"
}
if (-not (Get-Command terraform -ErrorAction SilentlyContinue)) {
    $missing += "Terraform (https://developer.hashicorp.com/terraform/downloads)"
}

if ($missing.Count -gt 0) {
    Write-Host "Ferramentas faltando:" -ForegroundColor Red
    foreach ($m in $missing) { Write-Host "  - $m" }
    exit 1
}
Write-Host "Todas as dependências estão instaladas." -ForegroundColor Green

Write-Host "`n[2/3] Iniciando serviços Docker..." -ForegroundColor Yellow
Push-Location (Join-Path $PSScriptRoot ".." "docker")
try {
    docker compose down -v 2>&1 | Out-Null
    docker compose up -d --build
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Erro ao iniciar serviços Docker" -ForegroundColor Red
        exit 1
    }
    Write-Host "Serviços Docker iniciados." -ForegroundColor Green

    Write-Host "`nAguardando serviços ficarem prontos..." -ForegroundColor Yellow
    Start-Sleep -Seconds 10
}
finally {
    Pop-Location
}

if ($InitTerraform) {
    Write-Host "`n[3/3] Inicializando Terraform..." -ForegroundColor Yellow
    $tfDir = Join-Path $PSScriptRoot ".." "terraform" "app"
    Push-Location $tfDir
    try {
        terraform init
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Erro ao inicializar Terraform" -ForegroundColor Red
            exit 1
        }
        Write-Host "Terraform inicializado." -ForegroundColor Green
    }
    finally {
        Pop-Location
    }
}

Write-Host "`n=== Setup concluído! ===" -ForegroundColor Cyan
Write-Host "API: http://localhost:8080" -ForegroundColor Green
Write-Host "RabbitMQ UI: http://localhost:15672 (guest/guest)" -ForegroundColor Green

"service_healthy"
