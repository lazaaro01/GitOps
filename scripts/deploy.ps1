param(
    [Parameter(Mandatory = $true)]
    [string]$AppName,

    [Parameter(Mandatory = $true)]
    [string]$ImageTag,

    [string]$ApiUrl = "http://localhost:8080"
)

$body = @{
    app_name  = $AppName
    image_tag = $ImageTag
} | ConvertTo-Json

Write-Host "Deploying $AppName with image $ImageTag..." -ForegroundColor Cyan

$response = Invoke-RestMethod -Uri "$ApiUrl/api/deploy" -Method Post -Body $body -ContentType "application/json"

if ($response.success) {
    Write-Host "Deploy criado com sucesso!" -ForegroundColor Green
    Write-Host "ID: $($response.data.id)" -ForegroundColor Green
} else {
    Write-Host "Erro: $($response.error)" -ForegroundColor Red
}
