# Semki

**Communication Software for Staffbase's "Code the Future"**

An intelligent communication platform with semantic search and AI-powered user matching capabilities.

---

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Make (optional, for shortcuts)
- Add domains to /etc/hosts

### Start Application

```shell
# Start all services in detached mode
docker compose up -d

# View logs
docker compose logs -f

# Stop all services
docker compose down
```

### Access Points

| Service | URL | Description |
|---------|-----|-------------|
| **Web UI** | http://localhost:80 | Frontend application |
| **API** | http://localhost:8080 | Backend API |
| **Swagger** | http://localhost:8080/swagger/index.html | API documentation |
| **Grafana** | http://localhost:3030 | Monitoring dashboards |
| **Prometheus** | http://localhost:9090 | Metrics collection |
| **Jaeger** | http://localhost:16686 | Distributed tracing |
| **Pyroscope** | http://localhost:4040 | Continuous profiling |
| **Loki** | http://localhost:3100 | Log aggregation |
| **Mailhog** | http://localhost:8025 | Email testing |
| **MongoDB** | localhost:27017 | Database |
| **Qdrant** | http://localhost:6333 | Vector database |
| **Redis** | localhost:6379 | Cache |

---

## 🏗️ Architecture

### High-Level Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client (React)                           │
│  Vite + React 19 + TypeScript + Mantine UI + TanStack Query    │
└──────────────────────────────┬──────────────────────────────────┘
                               │ HTTP/REST
┌──────────────────────────────┴──────────────────────────────────┐
│                     Backend (Go + Gin)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │  Controller  │→ │   Service    │→ │   Adapter    │         │
│  │   (HTTP/v1)  │  │   (Business) │  │  (Repositories)│        │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└──────────────────────────────┬──────────────────────────────────┘
                               │
    ┌──────────────────────────┼──────────────────────────────┐
    │                          │                              │
┌───▼────┐  ┌────────┐  ┌─────▼─────┐  ┌────────────┐  ┌───▼─────┐
│MongoDB │  │ Qdrant │  │Embedder   │  │   Redis    │  │OpenAI   │
│(Data)  │  │(Vector)│  │(ML Model) │  │  (Tokes)   │  │  (API)  │
└────────┘  └────────┘  └───────────┘  └────────────┘  └─────────┘
```

### Backend Structure

```
internal/
├── controller/      # HTTP handlers (routes, DTOs)
│   └── http/v1/
├── service/         # Business logic
├── adapter/         # External integrations
│   └── mongo/       # Database repositories
├── model/           # Domain entities
└── utils/           # Shared utilities

pkg/                 # Shared across microservices
├── clients/         # Database clients
├── google/          # OAuth2 authentication
├── telemetry/       # Observability (logs, traces, metrics)
└── config/          # Configuration management
```

---

## 🛠️ Technologies

### Backend

| Technology | Purpose |
|------------|---------|
| **Go 1.24** | Programming language |
| **Gin** | Web framework |
| **MongoDB** | Primary database |
| **Qdrant** | Vector database for semantic search |
| **Redis** | Caching layer |
| **OpenAI API** | AI integrations |
| **JWT** | Authentication |
| **gRPC** | Service communication |
| **Gomail** | Email delivery |

### Frontend

| Technology | Purpose |
|------------|---------|
| **React 19** | UI library |
| **TypeScript** | Type safety |
| **Vite** | Build tool |
| **Mantine UI** | Component library |
| **TanStack Query** | Data fetching |
| **Zustand** | State management |
| **React Router** | Routing |
| **Axios** | HTTP client |

### Observability Stack

| Tool | Purpose |
|------|---------|
| **Prometheus** | Metrics collection |
| **Grafana** | Visualization dashboards |
| **Loki** | Log aggregation |
| **Promtail** | Log collection |
| **Jaeger** | Distributed tracing |
| **Pyroscope** | Continuous profiling |
| **OpenTelemetry** | Instrumentation |
| **Zap** | Structured logging |
| **Sentry** | Error tracking |

### Infrastructure

| Service | Purpose |
|---------|---------|
| **Docker** | Containerization |
| **Mailhog** | Email testing |
| **MongoDB Exporter** | Prometheus metrics for MongoDB |
| **Embedder** | Custom ML embedding service |

---

## 📦 Development

### Generate Swagger Documentation

Swagger docs are auto-generated from code annotations:

```shell
# Manually regenerate
sh ./resources/swagger.sh

# Or use IDE run configuration (recommended)
```

### Environment Variables

Copy `.env.example` to `.env.production` and configure:

```env
# MongoDB
MONGO_USER=admin
MONGO_PASSWORD=password
MONGO_DATABASE=semki

# Redis
REDIS_PASSWORD=your-redis-password

# OpenAI
OPENAI_API_KEY=your-api-key

# JWT
JWT_SECRET=your-secret-key

# ... (see .env.example for full list)
```

### Project Commands

```shell
# Build application
make build

# Generate swagger
make swagger

# Restore grafana
make restore-grafana-file FILE=./backups/grafana-20251107-025757.db

# Get list of the backups
restore-grafana

# Create a backup
make backup-grafana
```

---

## 🎯 Key Features

### 1. Semantic User Search
- Vector-based similarity search using Qdrant
- AI-powered user matching based on roles, teams, and descriptions
- Custom embedding service for user profiles

### 2. Authentication & Authorization
- JWT-based authentication
- Google OAuth2 integration
- Role-based access control (OWNER, ADMIN, USER)

### 3. Organization Management
- Multi-tenant architecture
- Teams, Levels, and Locations hierarchy
- Mock data generation for testing

### 4. Observability
- Full request tracing with Jaeger
- Centralized logging with Loki
- Real-time metrics in Grafana
- Continuous profiling with Pyroscope

### 5. Email Communication
- Email delivery via SMTP
- MailHog for local testing

---

## 🚢 Deployment

### Docker Compose (Production)

```shell
docker compose -f docker-compose.yml up -d
```

### Kubernetes (Advanced)

See deployment steps in comments for K8S setup with Minikube:

1. Install Docker, K8S, and Minikube
2. Start Docker Engine and Minikube
3. Enable ingress: `minikube addons enable ingress`
4. Configure `/etc/hosts` entries
5. Apply K8S manifests from `/k8s` directory

---

## 📝 API Documentation

Interactive API documentation is available via Swagger UI:

**Local:** http://localhost:8080/swagger/index.html

### Example Endpoints

```
POST   /api/v1/auth/login              # User login
POST   /api/v1/auth/register           # User registration
GET    /api/v1/organization            # Get organization details
POST   /api/v1/organization/teams      # Create team
GET    /api/v1/users/search            # Semantic user search
POST   /api/v1/organization/insert-mock # Generate mock data
```

---

## 🧪 Testing

### Mock Data Generation

Insert realistic test data:

```shell
curl -X POST http://localhost:8080/api/v1/organization/insert-mock \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Generates:
- 5 Teams (Engineering, Product, Design, Marketing, Sales)
- 5 Levels (Junior, Middle, Senior, Lead, Principal)
- 6 Locations (SF, NYC, London, Berlin, Tokyo, Remote)
- 18 Users with diverse characteristics

---

## 📊 Monitoring

### View Logs

```shell
# Loki/Promtail
{container="semki-server"}
{container="semki-server", level="ERROR"}

# Docker logs
docker compose logs -f semki-server
```

### Metrics

- **Grafana:** Pre-configured dashboards at http://localhost:3030
- **Prometheus:** Raw metrics at http://localhost:9090

### Tracing

- **Jaeger:** Distributed traces at http://localhost:16686
- Search by service: `semki-server`

### Profiling

- **Pyroscope:** CPU/Memory profiles at http://localhost:4040

---

## 🤝 Contributing

1. Follow the existing code structure
2. Add Swagger comments for new endpoints
3. Write tests for business logic
4. Update this README if adding new services

---

## 📄 License

See LICENSE file for details.

---

## 🔗 Links

- [Swagger UI](http://localhost:8080/swagger/index.html)
- [Grafana Dashboards](http://localhost:3030)
