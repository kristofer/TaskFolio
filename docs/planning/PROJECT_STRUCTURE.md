# TaskFolio Project Structure

## Overview
This document outlines the recommended directory structure for the TaskFolio project following Go best practices.

## Backend Structure (Go)

```
TaskFolio/
├── cmd/
│   ├── api/                      # API server entry point
│   │   └── main.go
│   ├── migrate/                  # Database migration tool
│   │   └── main.go
│   └── worker/                   # Background worker (optional)
│       └── main.go
│
├── internal/                     # Private application code
│   ├── api/                      # API layer
│   │   ├── handlers/            # HTTP handlers
│   │   │   ├── auth.go
│   │   │   ├── task.go
│   │   │   ├── user.go
│   │   │   ├── comment.go
│   │   │   ├── tag.go
│   │   │   ├── notification.go
│   │   │   └── team.go
│   │   ├── middleware/          # HTTP middleware
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── logger.go
│   │   │   ├── ratelimit.go
│   │   │   └── recovery.go
│   │   ├── routes/              # Route definitions
│   │   │   └── routes.go
│   │   └── validators/          # Request validators
│   │       └── validators.go
│   │
│   ├── service/                 # Business logic layer
│   │   ├── auth.go
│   │   ├── task.go
│   │   ├── user.go
│   │   ├── comment.go
│   │   ├── tag.go
│   │   ├── notification.go
│   │   └── team.go
│   │
│   ├── repository/              # Data access layer
│   │   ├── postgres/           # PostgreSQL implementations
│   │   │   ├── user.go
│   │   │   ├── task.go
│   │   │   ├── comment.go
│   │   │   ├── tag.go
│   │   │   ├── notification.go
│   │   │   └── team.go
│   │   └── repository.go       # Repository interfaces
│   │
│   ├── models/                  # Domain models
│   │   ├── user.go
│   │   ├── task.go
│   │   ├── comment.go
│   │   ├── tag.go
│   │   ├── notification.go
│   │   ├── team.go
│   │   └── common.go
│   │
│   ├── auth/                    # Authentication utilities
│   │   ├── jwt.go
│   │   ├── password.go
│   │   └── permissions.go
│   │
│   ├── config/                  # Configuration management
│   │   └── config.go
│   │
│   ├── database/                # Database utilities
│   │   ├── connection.go
│   │   └── migrations/         # SQL migration files
│   │       ├── 000001_init.up.sql
│   │       ├── 000001_init.down.sql
│   │       ├── 000002_add_teams.up.sql
│   │       └── 000002_add_teams.down.sql
│   │
│   ├── notification/            # Notification services
│   │   ├── email.go
│   │   ├── push.go
│   │   └── websocket.go
│   │
│   ├── storage/                 # File storage utilities
│   │   ├── local.go
│   │   └── s3.go
│   │
│   └── utils/                   # Utility functions
│       ├── logger.go
│       ├── errors.go
│       ├── response.go
│       └── pagination.go
│
├── pkg/                         # Public, reusable packages
│   ├── validator/              # Custom validators
│   │   └── validator.go
│   ├── timeutil/               # Time utilities
│   │   └── timeutil.go
│   └── crypto/                 # Crypto utilities
│       └── crypto.go
│
├── api/                         # API documentation
│   ├── swagger.yaml            # OpenAPI/Swagger spec
│   └── postman/                # Postman collections
│       └── taskfolio.json
│
├── web/                         # Frontend application
│   ├── public/                 # Static assets
│   │   ├── index.html
│   │   ├── favicon.ico
│   │   └── assets/
│   │
│   ├── src/
│   │   ├── components/         # React/Vue components
│   │   │   ├── common/        # Reusable components
│   │   │   │   ├── Button/
│   │   │   │   ├── Input/
│   │   │   │   ├── Modal/
│   │   │   │   └── Dropdown/
│   │   │   ├── layout/        # Layout components
│   │   │   │   ├── Header/
│   │   │   │   ├── Sidebar/
│   │   │   │   └── Footer/
│   │   │   ├── auth/          # Auth components
│   │   │   │   ├── LoginForm/
│   │   │   │   └── RegisterForm/
│   │   │   ├── tasks/         # Task components
│   │   │   │   ├── TaskList/
│   │   │   │   ├── TaskCard/
│   │   │   │   ├── TaskForm/
│   │   │   │   ├── KanbanBoard/
│   │   │   │   └── CalendarView/
│   │   │   ├── user/          # User components
│   │   │   │   ├── UserProfile/
│   │   │   │   └── UserSettings/
│   │   │   └── dashboard/     # Dashboard components
│   │   │       └── DashboardStats/
│   │   │
│   │   ├── pages/              # Page components
│   │   │   ├── Login.jsx
│   │   │   ├── Register.jsx
│   │   │   ├── Dashboard.jsx
│   │   │   ├── Tasks.jsx
│   │   │   ├── TaskDetail.jsx
│   │   │   ├── Profile.jsx
│   │   │   └── Settings.jsx
│   │   │
│   │   ├── services/           # API services
│   │   │   ├── api.js         # API client setup
│   │   │   ├── auth.js
│   │   │   ├── tasks.js
│   │   │   ├── users.js
│   │   │   └── notifications.js
│   │   │
│   │   ├── store/              # State management
│   │   │   ├── index.js
│   │   │   ├── auth/
│   │   │   ├── tasks/
│   │   │   └── notifications/
│   │   │
│   │   ├── hooks/              # Custom React hooks
│   │   │   ├── useAuth.js
│   │   │   ├── useTasks.js
│   │   │   └── useWebSocket.js
│   │   │
│   │   ├── utils/              # Utility functions
│   │   │   ├── date.js
│   │   │   ├── validation.js
│   │   │   └── formatting.js
│   │   │
│   │   ├── styles/             # Global styles
│   │   │   ├── globals.css
│   │   │   ├── variables.css
│   │   │   └── theme.js
│   │   │
│   │   ├── App.jsx             # Root component
│   │   └── main.jsx            # Entry point
│   │
│   ├── package.json
│   ├── vite.config.js          # Build config (if using Vite)
│   └── .eslintrc.js            # ESLint config
│
├── tests/                       # Test files
│   ├── integration/            # Integration tests
│   │   ├── api/
│   │   │   ├── auth_test.go
│   │   │   ├── task_test.go
│   │   │   └── user_test.go
│   │   └── database/
│   │       └── repository_test.go
│   │
│   ├── e2e/                    # End-to-end tests
│   │   ├── auth.spec.js
│   │   ├── tasks.spec.js
│   │   └── dashboard.spec.js
│   │
│   └── fixtures/               # Test fixtures
│       ├── users.json
│       └── tasks.json
│
├── scripts/                     # Utility scripts
│   ├── setup.sh                # Setup script
│   ├── seed.sh                 # Database seeding
│   ├── backup.sh               # Backup script
│   └── deploy.sh               # Deployment script
│
├── docs/                        # Documentation
│   ├── architecture.md         # Architecture docs
│   ├── development.md          # Development guide
│   ├── deployment.md           # Deployment guide
│   └── api/                    # API documentation
│       └── README.md
│
├── configs/                     # Configuration files
│   ├── dev.env                 # Development config
│   ├── staging.env             # Staging config
│   └── prod.env.example        # Production config template
│
├── deployments/                 # Deployment configurations
│   ├── docker/
│   │   ├── Dockerfile.api
│   │   ├── Dockerfile.web
│   │   └── docker-compose.yml
│   ├── kubernetes/             # K8s manifests
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── ingress.yaml
│   └── terraform/              # Infrastructure as code
│       ├── main.tf
│       └── variables.tf
│
├── .github/                     # GitHub specific files
│   ├── workflows/              # GitHub Actions
│   │   ├── ci.yml
│   │   ├── cd.yml
│   │   └── test.yml
│   └── ISSUE_TEMPLATE/
│       ├── bug_report.md
│       └── feature_request.md
│
├── .gitignore                   # Git ignore file
├── .env.example                 # Environment variables template
├── go.mod                       # Go module file
├── go.sum                       # Go dependencies
├── Makefile                     # Build automation
├── README.md                    # Project readme
├── LICENSE                      # License file
├── IMPLEMENTATION_PLAN.md       # Implementation plan
├── DATABASE_SCHEMA.md           # Database schema
└── API_SPECIFICATION.md         # API specification
```

## Key Design Patterns

### Backend (Go)

1. **Repository Pattern**: Abstracts data access logic
2. **Service Layer**: Contains business logic, independent of HTTP handlers
3. **Dependency Injection**: Services receive dependencies through constructors
4. **Clean Architecture**: Separation of concerns between layers
5. **Interface-Based Design**: Enables easy testing and mocking

### Frontend

1. **Component-Based Architecture**: Reusable, composable components
2. **Container/Presentation Pattern**: Separate logic from presentation
3. **Custom Hooks**: Encapsulate reusable logic
4. **State Management**: Centralized state with Redux/Vuex/Context
5. **Service Layer**: API calls abstracted into service modules

## File Naming Conventions

### Backend (Go)
- Files: `snake_case.go`
- Test files: `*_test.go`
- Interfaces: Descriptive names (e.g., `UserRepository`)
- Implementations: Implementation-specific names (e.g., `PostgresUserRepository`)

### Frontend
- Components: `PascalCase/` (folders) with `index.jsx`
- Services: `camelCase.js`
- Utilities: `camelCase.js`
- Styles: `ComponentName.module.css` or `ComponentName.styles.js`

## Environment Variables

### Backend (.env)
```bash
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=taskfolio
DB_USER=postgres
DB_PASSWORD=password
DB_SSL_MODE=disable

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# Email
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=noreply@taskfolio.com
SMTP_PASSWORD=password

# Storage
STORAGE_TYPE=local # or s3
S3_BUCKET=taskfolio-uploads
S3_REGION=us-east-1

# Rate Limiting
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=15m
```

### Frontend (.env)
```bash
VITE_API_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080/ws
VITE_APP_NAME=TaskFolio
```

## Build Commands

### Makefile
```makefile
.PHONY: help build run test lint migrate

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o bin/api cmd/api/main.go

run: ## Run the application
	go run cmd/api/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: ## Show test coverage
	go tool cover -html=coverage.out

lint: ## Run linter
	golangci-lint run

migrate-up: ## Run database migrations
	go run cmd/migrate/main.go up

migrate-down: ## Rollback database migrations
	go run cmd/migrate/main.go down

docker-build: ## Build Docker images
	docker-compose build

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

seed: ## Seed database with sample data
	go run scripts/seed/main.go
```

## Testing Structure

### Unit Tests
- Located alongside source files (`*_test.go`)
- Test individual functions and methods
- Use table-driven tests where applicable

### Integration Tests
- Located in `tests/integration/`
- Test interactions between components
- Use test database

### E2E Tests
- Located in `tests/e2e/`
- Test complete user workflows
- Use Cypress or Playwright

## Documentation

Each major component should have:
1. **README.md**: Overview and usage
2. **Code comments**: For complex logic
3. **API documentation**: OpenAPI/Swagger specs
4. **Architecture diagrams**: System design
5. **Decision records**: Major architectural decisions
