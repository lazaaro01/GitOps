# GitOps Lite Platform

Plataforma de Deploy Automatizado baseada em GitOps e Infrastructure as Code.

> **Versão atual:** MVP v1.0.0 — Backend

---

## Stack

| Categoria | Tecnologia |
|---|---|
| Linguagem | Go 1.24+ |
| Web Framework | Fiber v2 |
| PostgreSQL Driver | pgx v5 |
| Logger | zerolog |
| Message Broker | RabbitMQ |
| Infraestrutura | Terraform + Docker Provider |
| Containerização | Docker + Docker Compose |

---

## Arquitetura

```mermaid
flowchart TD
    User["Usuário / GitHub Webhook"]
    User -->|POST /deploy| API["API (Go / Fiber)"]
    API -->|Publica Job| RMQ["RabbitMQ"]
    RMQ -->|Consome| Worker["Deploy Worker"]
    Worker --> TF["Terraform Executor"]
    TF --> TFInit["terraform init"]
    TFInit --> TFPlan["terraform plan"]
    TFPlan --> TFApply["terraform apply"]
    TFApply --> Docker["Docker Provider"]
    Docker --> Containers["Containers (App + DB)"]
    Worker -->|Health Check| Containers
    Worker -->|Persiste resultado| DB[("PostgreSQL")]
```

---

## Estrutura do Projeto

```
gitops-lite/
├── apps/
│   ├── api/                    # API HTTP (Fiber)
│   │   ├── cmd/main.go
│   │   └── internal/
│   │       ├── config/         # Configurações via env
│   │       ├── handler/        # Handlers HTTP
│   │       └── queue/          # Producer RabbitMQ
│   └── deploy-worker/          # Worker assíncrono
│       ├── cmd/main.go
│       └── internal/
│           ├── config/
│           ├── consumer/       # Consumer RabbitMQ
│           ├── executor/       # Terraform executor
│           └── health/         # Health check HTTP
├── pkg/                        # Pacotes compartilhados
│   ├── model/                  # Structs/entities
│   └── repository/             # Acesso a banco (pgx)
├── terraform/                  # Módulos Terraform
│   ├── modules/
│   │   ├── network/            # Rede Docker
│   │   ├── container/          # Container Docker
│   │   └── volume/             # Volume Docker
│   └── app/                    # Root module (deploy)
├── migrations/                 # Migrations SQL
├── docker/
│   ├── Dockerfile.api
│   ├── Dockerfile.worker
│   └── docker-compose.yml
├── scripts/
│   ├── setup.ps1               # Setup completo
│   ├── migrate.ps1             # Rodar migrations
│   └── deploy.ps1              # Exemplo de deploy via API
├── go.work                     # Go workspace
└── README.md
```

---

## Pré-requisitos

- **Go 1.24+** — [Download](https://go.dev/dl/)
- **Docker Desktop** — [Download](https://docs.docker.com/get-docker/)
- **Terraform 1.6+** — [Download](https://developer.hashicorp.com/terraform/downloads)
- **RabbitMQ** (via Docker)
- **PostgreSQL 16** (via Docker)

---

## Como Rodar

### 1. Suba os serviços de infraestrutura

```bash
docker compose -f docker/docker-compose.yml up -d postgres rabbitmq
```

Aguarde os serviços ficarem prontos (~10s).

### 2. Execute as migrations

```bash
# Usando PowerShell
.\scripts\migrate.ps1

# Ou manualmente com psql
psql postgres://gitops:gitops@localhost:5432/gitops -f migrations/001_create_deployments.sql
psql postgres://gitops:gitops@localhost:5432/gitops -f migrations/002_create_deployment_logs.sql
psql postgres://gitops:gitops@localhost:5432/gitops -f migrations/003_create_jobs.sql
```

### 3. Inicialize o Terraform

```bash
cd terraform/app
terraform init
```

### 4. Rode a API

```bash
cd apps/api
go run ./cmd/main.go
```

A API estará disponível em `http://localhost:8080`.

### 5. Rode o Worker (em outro terminal)

```bash
cd apps/deploy-worker
go run ./cmd/main.go
```

---

## Usando a API

### Criar um deploy

```bash
curl -X POST http://localhost:8080/api/deploy \
  -H "Content-Type: application/json" \
  -d '{"app_name": "my-app", "image_tag": "nginx:latest"}'
```

**Resposta (202 Accepted):**

```json
{
  "success": true,
  "data": {
    "id": "a1b2c3d4-...",
    "app_name": "my-app",
    "image_tag": "nginx:latest",
    "status": "queued",
    "created_at": "2026-06-29T10:00:00Z",
    "updated_at": "2026-06-29T10:00:00Z"
  }
}
```

### Listar deploys

```bash
curl http://localhost:8080/api/deployments
```

### Ver detalhes de um deploy

```bash
curl http://localhost:8080/api/deployments/<deploy-id>
```

### Cancelar um deploy pendente

```bash
curl -X PUT http://localhost:8080/api/deployments/<deploy-id>/cancel
```

---

## Fluxo de Deploy

1. Cliente envia `POST /api/deploy`
2. API valida os parâmetros e salva no banco (status: `pending`)
3. API publica um job no RabbitMQ e atualiza status para `queued`
4. Worker consome o job da fila
5. Worker atualiza status para `in_progress`
6. Worker executa:
   - `terraform init` — inicializa o provider Docker
   - `terraform plan` — gera o plano de execução
   - `terraform apply` — provisiona o container Docker
7. Worker executa health check HTTP no container
8. Worker atualiza status para `success` ou `failed`
9. Logs de cada etapa são persistidos no PostgreSQL

---

## Serviços

| Serviço | Porta | URL |
|---|---|---|
| API | 8080 | http://localhost:8080 |
| RabbitMQ (AMQP) | 5672 | amqp://guest:guest@localhost:5672 |
| RabbitMQ (Management) | 15672 | http://localhost:15672 |
| PostgreSQL | 5432 | `postgres://gitops:gitops@localhost:5432/gitops` |

---

## Scripts Úteis

### Setup completo (Docker + infra)

```powershell
.\scripts\setup.ps1 -InitTerraform
```

### Deploy de exemplo

```powershell
.\scripts\deploy.ps1 -AppName my-app -ImageTag nginx:latest
```

---

## Roadmap

| Versão | Foco |
|---|---|
| **MVP v1.0.0** | Backend: API + Worker + Terraform + RabbitMQ + PostgreSQL |
| v1.1.0 | Dashboard Web + Rollback |
| v1.2.0 | Observabilidade (Prometheus, Grafana, Loki) |
| v2.0.0 | GitOps completo (Kubernetes, Argo CD, Helm) |
