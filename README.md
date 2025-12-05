# TTT Monorepo

A production-ready monorepo using **Turborepo** and **pnpm workspaces** with React + Go.

## 🏗️ Architecture

```
ttt/
├── apps/
│   ├── api/          # Go + Gin backend
│   ├── web/          # React + Vite frontend
│   └── db/           # Database scripts
├── packages/         # Shared packages (future)
├── docker-compose.yml
├── turbo.json
└── pnpm-workspace.yaml
```

## 🚀 Quick Start

### Prerequisites

- **Node.js** >= 20.x
- **pnpm** >= 9.x
- **Go** >= 1.24
- **Docker** & **Docker Compose**

### Installation

```bash
# Clone and install dependencies
pnpm install

# Setup Husky hooks
pnpm prepare
```

### Development

```bash
# Run all apps in development mode
pnpm dev

# Run specific app
pnpm dev:web   # Frontend only (port 3000)
pnpm dev:api   # Backend only (port 8080)
```

### Build

```bash
# Build all apps (React + Go)
pnpm build

# Build will:
# - Compile React app to dist/
# - Compile Go binary to tmp/main
```

### Lint & Format

```bash
# Lint all apps
pnpm lint

# Format code
pnpm format

# Check formatting
pnpm format:check
```

### Docker

```bash
# Production build
docker compose up -d --build

# Development with hot reload
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# View logs
docker compose logs -f

# Stop all services
docker compose down
```

## 📦 Polyglot Build System

This monorepo integrates Go into Turborepo's JavaScript-centric pipeline:

```json
// apps/api/package.json
{
  "scripts": {
    "build": "go build -o tmp/main ./cmd/api/main.go",
    "dev": "air -c .air.toml"
  }
}
```

Running `pnpm turbo run build` will simultaneously:
1. Build React app (`tsc && vite build`)
2. Compile Go binary (`go build`)

## 🔒 Git Hooks (Husky)

Pre-configured hooks ensure code quality:

- **pre-commit**: Runs full build (React + Go must compile)
- **pre-push**: Runs lint, build, and tests

If either React or Go has errors, the commit/push is rejected.

## 🐳 Docker Services

| Service  | Port | Description           |
|----------|------|-----------------------|
| web      | 3000 | React frontend (Nginx)|
| api      | 8080 | Go Gin backend        |
| postgres | 5432 | PostgreSQL database   |

## 📁 Key Files

| File                    | Purpose                              |
|-------------------------|--------------------------------------|
| `turbo.json`            | Turborepo pipeline configuration     |
| `pnpm-workspace.yaml`   | pnpm workspace definition            |
| `docker-compose.yml`    | Production Docker orchestration      |
| `docker-compose.dev.yml`| Development overrides (hot reload)   |
| `.husky/pre-commit`     | Build check before commit            |

## 🛠️ Scripts

| Command           | Description                          |
|-------------------|--------------------------------------|
| `pnpm dev`        | Start all apps in dev mode           |
| `pnpm build`      | Build all apps                       |
| `pnpm lint`       | Lint all apps                        |
| `pnpm test`       | Run all tests                        |
| `pnpm clean`      | Clean build artifacts                |
| `pnpm format`     | Format code with Prettier            |
| `pnpm typecheck`  | TypeScript type checking             |

## 📄 License

ISC
