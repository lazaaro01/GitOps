# GitOps Lite Platform

Plataforma de Deploy Automatizado baseada em GitOps e Infrastructure as Code.

> **Versão atual:** v1.1.0 — Dashboard + Rollback
>
> **Desenvolvido por:** Lázaro Vasconcelos

---

## Stack

| Categoria | Tecnologia |
|---|---|
| Backend | Go 1.24+ · Fiber v2 · pgx v5 · zerolog |
| Frontend | React 18 · TypeScript · Vite · Tailwind CSS |
| Gráficos | Recharts |
| Mensageria | RabbitMQ |
| Infraestrutura | Terraform + Docker Provider |
| Streaming | SSE (Server-Sent Events) |
| Banco | PostgreSQL 16 |
| Containerização | Docker + Docker Compose |

---

## Arquitetura

```mermaid
flowchart TD
    User["Usuário"]
    User -->|HTTP| API["API (Go / Fiber)"]
    API -->|Serve SPA| Frontend["Dashboard Web (React + Vite)"]
    Frontend -->|REST + SSE| API

    API -->|Publica Job| RMQ["RabbitMQ"]
    RMQ -->|Consome| Worker["Deploy Worker"]

    Worker --> TF["Terraform Executor"]
    TF --> TFInit["terraform init"]
    TFInit --> TFPlan["terraform plan"]
    TFPlan --> TFApply["terraform apply"]
    TFApply --> Docker["Docker Provider"]
    Docker --> Containers["Containers"]

    Worker -->|Health Check| Containers
    Worker -->|SSE Events| API
    Worker -->|Persiste resultado + state snapshot| DB[("PostgreSQL")]

    API -->|Consulta histórico| DB
    Frontend -->|Logs via SSE| API
```

---

## Estrutura do Projeto

```
gitops-lite/
├── apps/
│   ├── api/                        # API HTTP (Fiber)
│   │   ├── cmd/main.go
│   │   └── internal/
│   │       ├── config/             # Config via env + .env loader
│   │       ├── handler/            # Handlers HTTP (deploy, logs, rollback, events, router)
│   │       └── queue/              # Producer RabbitMQ
│   ├── deploy-worker/              # Worker assíncrono
│   │   ├── cmd/main.go
│   │   └── internal/
│   │       ├── config/
│   │       ├── consumer/           # Consumer RabbitMQ
│   │       ├── events/             # HTTP client para SSE
│   │       ├── executor/           # Terraform executor
│   │       └── health/             # Health check HTTP
│   └── frontend/                   # Dashboard Web (React + Vite)
│       ├── src/
│       │   ├── components/         # Layout, DeployTable, DeployTimeline,
│       │   │                       # LogViewer, RollbackModal, HistoryChart, StatusBadge
│       │   ├── pages/              # Dashboard, DeployDetail, History
│       │   ├── services/           # api.ts (Axios), sse.ts (EventSource)
│       │   ├── hooks/              # useDeployments, useDeployDetail, useSSE
│       │   └── types/              # TypeScript interfaces
│       ├── package.json
│       └── vite.config.ts
├── pkg/                            # Pacotes compartilhados
│   ├── model/                      # Deployment, Job, Log, API response
│   └── repository/                 # Acesso a banco (pgx)
├── terraform/                      # Módulos Terraform
│   ├── modules/
│   │   ├── network/                # Rede Docker
│   │   ├── container/              # Container Docker
│   │   └── volume/                 # Volume Docker
│   └── app/                        # Root module (deploy)
├── migrations/                     # Migrations SQL
├── docker/
│   ├── Dockerfile.api
│   ├── Dockerfile.worker
│   └── docker-compose.yml
├── scripts/
│   ├── setup.ps1
│   ├── migrate.ps1
│   └── deploy.ps1
├── docs/                           # PRDs e especificações
├── go.work                         # Go workspace
└── README.md
```

---

## Pré-requisitos

- **Go 1.24+** — [Download](https://go.dev/dl/)
- **Node.js 20+** — [Download](https://nodejs.org/)
- **Docker Desktop** — [Download](https://docs.docker.com/get-docker/)
- **Terraform 1.6+** — [Download](https://developer.hashicorp.com/terraform/downloads)

---

## Como Rodar

### 1. Suba os serviços de infraestrutura

```bash
docker compose -f docker/docker-compose.yml up -d postgres rabbitmq
```

Aguarde os serviços ficarem prontos (~10s).

> O PostgreSQL estará disponível em `localhost:5433` (porta 5432 do container mapeada para 5433 no host para evitar conflitos com outras instalações locais).

### 2. Execute o backend (API + Worker)

**Terminal 1 — API:**

```bash
cd apps/api
go run ./cmd/main.go
```

A API estará disponível em `http://localhost:8080`.

**Terminal 2 — Worker:**

```bash
cd apps/deploy-worker
go run ./cmd/main.go
```

> O `.env` é carregado automaticamente por ambos os serviços.

### 3. Execute o frontend (desenvolvimento)

```bash
cd apps/frontend
npm install    # apenas na primeira vez
npm run dev
```

O dashboard estará disponível em `http://localhost:5173`.

> Em desenvolvimento, o Vite faz proxy de `/api` para a API em `localhost:8080`.
> Em produção, a própria API serve os arquivos estáticos do frontend (`apps/frontend/dist/`).

---

## Endpoints da API

### v1.0.0 (MVP)

| Método | Rota | Descrição |
|---|---|---|
| `POST` | `/api/deploy` | Criar um novo deploy |
| `GET` | `/api/deployments` | Listar deploys (paginado) |
| `GET` | `/api/deployments/:id` | Detalhes de um deploy |
| `PUT` | `/api/deployments/:id/cancel` | Cancelar um deploy pendente |

### v1.1.0 (Dashboard + Rollback)

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/api/deployments/:id/logs` | Logs de um deploy |
| `GET` | `/api/deployments/:id/logs/download` | Download dos logs (.txt) |
| `POST` | `/api/deployments/:id/rollback` | Solicitar rollback para versão anterior |
| `POST` | `/api/deployments/:id/retry` | Reexecutar deploy falho |
| `GET` | `/api/events?deploy_id=:id` | SSE streaming de eventos em tempo real |
| `GET` | `/health` | Health check da API |

---

## Fluxo de Deploy

```mermaid
sequenceDiagram
    participant U as Usuário
    participant API as API
    participant RMQ as RabbitMQ
    participant W as Worker
    participant TF as Terraform
    participant DB as PostgreSQL

    U->>API: POST /api/deploy {app_name, image_tag}
    API->>DB: Salva deploy (status: pending)
    API->>RMQ: Publica job
    API->>DB: Atualiza status (queued)
    API-->>U: 202 Accepted

    RMQ->>W: Consome job
    W->>DB: Atualiza status (in_progress)
    W->>W: Gera terraform.tfvars
    W->>TF: terraform init
    W->>TF: terraform plan
    W->>TF: terraform apply
    TF->>Docker: Cria container
    W->>W: Health Check
    W->>DB: Atualiza status (success/failed)
    W->>API: SSE event (deploy_completed)
    API->>U: SSE: deploy_completed
```

---

## Funcionalidades do Frontend

### Dashboard
- Listagem de deploys com paginação (20 por página)
- Filtros por status (Todos, Sucesso, Falha, Em Andamento)
- Botão "Novo Deploy" com formulário inline
- Botão "Retry" em deploys falhos diretamente na tabela
- Gráficos: deploys por dia (barras) e distribuição por status (pizza)

### Detalhes do Deploy
- Metadados completos (ID, app, imagem, datas)
- Timeline visual das etapas do pipeline
- Logs com auto-scroll e suporte a SSE em tempo real
- Botões de ação: Rollback (modal), Retry, Cancelar, Download de logs

### Histórico Visual
- Taxa de sucesso, falha e cancelamento
- Gráfico de deploys por dia (últimos 100 deploys)
- Gráfico de distribuição por status

---

## Exemplos de Uso da API

### Criar um deploy

```bash
curl -X POST http://localhost:8080/api/deploy \
  -H "Content-Type: application/json" \
  -d '{"app_name": "my-app", "image_tag": "nginx:latest"}'
```

### Solicitar rollback

```bash
curl -X POST http://localhost:8080/api/deployments/<deploy-id>/rollback \
  -H "Content-Type: application/json" \
  -d '{"target_version": "nginx:1.25"}'
```

### Reexecutar deploy falho

```bash
curl -X POST http://localhost:8080/api/deployments/<deploy-id>/retry
```

### Conectar SSE (JavaScript)

```javascript
const source = new EventSource('/api/events?deploy_id=abc-123');

source.addEventListener('deploy_log', (e) => {
  const data = JSON.parse(e.data);
  console.log(`[${data.step}] ${data.message}`);
});

source.addEventListener('deploy_completed', (e) => {
  const data = JSON.parse(e.data);
  console.log(`Deploy concluído: ${data.status}`);
  source.close();
});
```

---

## Serviços

| Serviço | Porta (Host) | URL |
|---|---|---|
| API | 8080 | http://localhost:8080 |
| Frontend (dev) | 5173 | http://localhost:5173 |
| RabbitMQ (AMQP) | 5672 | amqp://guest:guest@localhost:5672 |
| RabbitMQ (Management) | 15672 | http://localhost:15672 |
| PostgreSQL | 5433 | `postgres://gitops:gitops@localhost:5433/gitops` |

---

## Scripts Úteis

### Setup completo (Docker + infra)

```powershell
.\scripts\setup.ps1 -InitTerraform
```

### Deploy de exemplo via PowerShell

```powershell
.\scripts\deploy.ps1 -AppName my-app -ImageTag nginx:latest
```

---

## Roadmap

| Versão | Foco | Status |
|---|---|---|
| **MVP v1.0.0** | Backend: API + Worker + Terraform + RabbitMQ + PostgreSQL | ✅ Concluído |
| **v1.1.0** | Dashboard Web + Rollback + SSE + logs | ✅ Concluído |
| v1.2.0 | Observabilidade (Prometheus, Grafana, Loki, OpenTelemetry) | 📋 Planejado |
| v2.0.0 | GitOps completo (Kubernetes, Argo CD, Helm, Canary, Blue/Green) | 📋 Planejado |

---

> **Desenvolvido por Lázaro Vasconcelos** — Plataforma de estudo em Platform Engineering e DevOps
