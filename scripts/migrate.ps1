param(
    [string]$DatabaseURL = "postgres://gitops:gitops@localhost:5432/gitops?sslmode=disable"
)

$migrationsDir = Join-Path $PSScriptRoot ".." "migrations" -Resolve

$migrations = Get-ChildItem -Path $migrationsDir -Filter "*.sql" | Sort-Object Name

foreach ($migration in $migrations) {
    Write-Host "Running migration: $($migration.Name)..." -ForegroundColor Cyan
    $sql = Get-Content $migration.FullName -Raw
    $result = psql $DatabaseURL -c $sql 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Migration failed: $($migration.Name)" -ForegroundColor Red
        Write-Host $result
        exit 1
    }
    Write-Host "Migration completed: $($migration.Name)" -ForegroundColor Green
}

Write-Host "All migrations applied successfully!" -ForegroundColor Green
