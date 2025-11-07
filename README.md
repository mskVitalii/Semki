# Semki

**Intelligent Communication Platform with AI-Powered User Matching**

Communication Software for Staffbase's "Code the Future" competition.

---

## 📖 Overview

Semki is a modern communication platform that uses semantic search and AI to intelligently match users based on their roles, expertise, and communication preferences. Built with a microservices architecture, it features comprehensive observability, vector-based search, and multi-tenant organization management.

## 🚀 Quick Start

### Prerequisites

- **Docker & Docker Compose** (required)
- **Make** (optional, for shortcuts)

### Clone & Start

```bash
# Clone the repository
git clone <repository-url>
cd Semki

# Start all services
cd server
docker compose up -d

# View logs
docker compose logs -f

# Access the application
open http://localhost:80
```

### Development Mode

```bash
# Hot reload for both frontend and backend
cd server
docker compose watch
```

---

## 🏗️ Project Structure

```
Semki/
├── server/                 # Backend (Go)
│   ├── cmd/               # Application entrypoints
│   ├── internal/          # Internal packages
│   │   ├── controller/    # HTTP handlers
│   │   ├── service/       # Business logic
│   │   ├── adapter/       # Database repositories
│   │   ├── model/         # Domain models
│   │   └── utils/         # Utilities
│   ├── pkg/               # Shared packages
│   ├── resources/         # Config files (Loki, Promtail, etc.)
│   ├── docker-compose.yml # Infrastructure setup
│   └── README.md          # Server documentation
│
├── client/                # Frontend (React)
│   ├── src/
│   │   ├── pages/        # Page components
│   │   ├── api/          # API client
│   │   └── ...
│   ├── Dockerfile
│   └── README.md          # Client documentation
│
├── Makefile               # Project shortcuts
└── README.md              # This file
```

---

## 🛠️ Tech Stack

### Backend

| Component | Technology |
|-----------|-----------|
| Language | Go 1.24 |
| Framework | Gin |
| Database | MongoDB |
| Vector DB | Qdrant |
| Cache | Redis |
| AI | OpenAI API |
| Auth | JWT + OAuth2 |
| API Docs | Swagger/OpenAPI |

### Frontend

| Component | Technology |
|-----------|-----------|
| Framework | React 19 |
| Language | TypeScript |
| Build Tool | Vite |
| UI Library | Mantine UI |
| State | Zustand |
| Data Fetching | TanStack Query |
| Routing | React Router v7 |

### Infrastructure & Observability

| Service | Purpose |
|---------|---------|
| Prometheus | Metrics collection |
| Grafana | Dashboards & visualization |
| Loki + Promtail | Log aggregation |
| Jaeger | Distributed tracing |
| Pyroscope | Continuous profiling |
| Mailhog | Email testing |
| MongoDB Exporter | Database metrics |

---

## 🌐 Access Points

After starting with `docker compose up -d`, access services at:

| Service | URL | Credentials |
|---------|-----|-------------|
| **Web Application** | http://localhost:80 | - |
| **API** | http://localhost:8000 | - |
| **API Docs (Swagger)** | http://localhost:8000/swagger/index.html | - |
| **Grafana** | http://localhost:3030 | admin/admin |
| **Prometheus** | http://localhost:9090 | - |
| **Jaeger** | http://localhost:16686 | - |
| **Pyroscope** | http://localhost:4040 | - |
| **Mailhog** | http://localhost:8025 | - |
| **Qdrant Dashboard** | http://localhost:6333/dashboard | - |

---

## ✨ Key Features

### 🔍 Semantic User Search
- **Vector-based matching** using Qdrant vector database
- **AI-powered recommendations** based on user characteristics
- **Custom embedding service** for profile vectorization
- Search by skills, communication style, availability

### 👥 Organization Management
- **Multi-tenant architecture** with organization isolation
- **Hierarchical structure**: Teams → Levels → Locations
- **Role-based access control** (OWNER, ADMIN, USER)
- **Mock data generation** for testing and demos

### 🔐 Authentication & Security
- **JWT-based authentication** with secure token management
- **Google OAuth2 integration** for social login
- **Encrypted user data** at rest
- **Email verification** workflow

### 📊 Full Observability
- **Distributed tracing** with Jaeger (OpenTelemetry)
- **Centralized logging** with Loki + Promtail
- **Real-time metrics** in Grafana dashboards
- **Continuous profiling** with Pyroscope
- **Error tracking** with Sentry

### 📧 Communication
- **Email delivery** via SMTP
- **Template system** for notifications
- **Mailhog integration** for development

---

## 📝 Documentation

- **[Server README](./server/README.md)** - Detailed backend documentation
- **[Client README](./client/README.md)** - Frontend setup and structure
- **[API Documentation](http://localhost:8080/swagger/index.html)** - Interactive API docs (after starting)

---

## 🧪 Testing

### Generate Mock Data

After logging in, generate realistic test data:

```bash
# Via API
curl -X POST http://localhost:8080/api/v1/organization/insert-mock \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Or use the UI button in Organization Settings
```

This creates:
- 5 Teams (Engineering, Product, Design, Marketing, Sales)
- 5 Levels (Junior, Middle, Senior, Lead, Principal)
- 6 Locations (San Francisco, New York, London, Berlin, Tokyo, Remote)
- 18 Users with diverse characteristics and communication styles

---

## 🚢 Deployment

### Docker Compose (Recommended)

```bash
cd server
docker compose up -d
```

All services (frontend, backend, databases, monitoring) are orchestrated together.

## 🔧 Development

### Environment Setup

```bash
# 1. Copy environment template
cd server
cp .env.example .env.production

# 2. Configure required variables
# - MongoDB credentials
# - Redis password
# - OpenAI API key
# - JWT secret
# - etc.

# 3. Start development environment
docker compose watch
```

### Code Structure Guidelines

- **Backend**: Follow clean architecture (controller → service → adapter)
- **Frontend**: Component-based structure with hooks
- **API**: Document all endpoints with Swagger annotations
- **Tests**: Write unit tests for business logic
- **Logging**: Use structured logging (Zap)

### Generate Swagger Docs

```bash
cd server
sh ./resources/swagger.sh
```

Or configure your IDE to run this automatically before starting the server.

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style

- **Go**: Follow standard Go conventions (gofmt, golint)
- **React**: Use ESLint + Prettier configuration
- **Commits**: Use conventional commits format

---

## 📊 Monitoring & Observability

### View Application Logs

```bash
# Via Loki (in Grafana)
{container="semki-server"}
{container="semki-server", level="ERROR"}

# Via Docker
docker compose logs -f semki-server
```

### Check Metrics

- Open Grafana: http://localhost:3030
- Navigate to pre-configured dashboards
- View system, application, and database metrics

### Trace Requests

- Open Jaeger: http://localhost:16686
- Search by service: `semki-server`
- View distributed trace spans

### Profile Performance

- Open Pyroscope: http://localhost:4040
- View CPU and memory flame graphs
- Identify performance bottlenecks

---

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

## 🔗 Links

- [Habr Article (Russian)](https://habr.com/ru/articles/826508/)
- [Medium Article (English)](https://medium.com/@msk.vitalii/from-firebase-to-self-hosted-4ddb01c539e1)
- [Staffbase Code the Future](https://www.staffbase.com/en/code-the-future/)
- [Server Documentation](./server/README.md)
- [Client Documentation](./client/README.md)

---

## 📧 Contact

For questions or support, please open an issue in this repository.

---

**Built with ❤️ for Staffbase's "Code the Future" competition**
