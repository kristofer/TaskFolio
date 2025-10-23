# Docker Deployment Guide

This guide explains how to deploy Tasking using Docker and Docker Compose.

## Prerequisites

- Docker 20.10 or higher
- Docker Compose 2.0 or higher

## Quick Start

### 1. Build and Run with Docker Compose

```bash
# Build and start the container
docker compose up -d

# View logs
docker compose logs -f

# Stop the container
docker compose down
```

The application will be available at `http://localhost:8080`

### 2. Database Persistence

The database is stored in a Docker named volume called `tasking-data`. This ensures that your data persists even when containers are stopped or removed.

To verify the volume exists:
```bash
docker volume ls | grep tasking-data
```

### 3. Testing Database Persistence

To test that data persists between container restarts:

```bash
# Start the application
docker compose up -d

# Access the application and create some tasks
# Open http://localhost:8080 and log in

# Stop the container
docker compose down

# Start again - your data should still be there
docker compose up -d
```

## Manual Docker Build

If you prefer to build manually without Docker Compose:

```bash
# Build the image
docker build -t tasking:latest .

# Run the container with a persistent volume
docker run -d \
  --name tasking \
  -p 8080:8080 \
  -v tasking-data:/data \
  tasking:latest
```

## Environment Variables

The following environment variables can be configured:

- `PORT` - Server port (default: 8080)
- `DATABASE_PATH` - Path to SQLite database file (default: /data/taskfolio.db)

To set custom environment variables with Docker Compose, edit `docker-compose.yml`:

```yaml
services:
  tasking:
    environment:
      - PORT=3000
      - DATABASE_PATH=/data/custom.db
```

## Production Deployment

For production deployment on `https://tasking.zipcode.rocks`, you'll want to:

1. **Use a reverse proxy** (nginx, Caddy, or Traefik) to handle HTTPS
2. **Set up automated backups** of the Docker volume
3. **Configure resource limits** in docker-compose.yml:

```yaml
services:
  tasking:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
```

### Example with Caddy (recommended for automatic HTTPS)

Create a `Caddyfile`:

```
tasking.zipcode.rocks {
    reverse_proxy tasking:8080
}
```

Update `docker-compose.yml` to add Caddy:

```yaml
services:
  tasking:
    # existing config...
    ports: []  # Remove public port exposure
    
  caddy:
    image: caddy:2-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy-data:/data
      - caddy-config:/config
    restart: unless-stopped

volumes:
  tasking-data:
  caddy-data:
  caddy-config:
```

## Backup and Restore

### Backup Database

```bash
# Create a backup of the database
docker compose exec tasking cp /data/taskfolio.db /data/backup-$(date +%Y%m%d).db

# Copy backup to host
docker cp tasking:/data/backup-$(date +%Y%m%d).db ./
```

### Restore Database

```bash
# Copy backup to container
docker cp ./backup.db tasking:/data/taskfolio.db

# Restart container
docker compose restart
```

## Troubleshooting

### Container won't start

Check logs:
```bash
docker compose logs tasking
```

### Database issues

Reset the database (WARNING: deletes all data):
```bash
docker compose down
docker volume rm taskfolio_tasking-data
docker compose up -d
```

### Port already in use

Change the port in `docker-compose.yml`:
```yaml
ports:
  - "3000:8080"  # Use port 3000 on host
```

## Architecture

```
┌─────────────────┐
│   Host Machine  │
│                 │
│  ┌───────────┐  │
│  │  Docker   │  │
│  │ Container │  │
│  │           │  │
│  │  Tasking  │  │
│  │   :8080   │  │
│  └─────┬─────┘  │
│        │        │
│  ┌─────▼─────┐  │
│  │  Volume   │  │
│  │ tasking-  │  │
│  │   data    │  │
│  │           │  │
│  │ taskfolio │  │
│  │   .db     │  │
│  └───────────┘  │
└─────────────────┘
```

The SQLite database file is stored in the Docker volume, ensuring data persists across container recreations.
