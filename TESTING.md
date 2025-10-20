# Testing the Docker Setup

This document provides instructions for testing the Docker deployment and database persistence.

## Prerequisites

Ensure you have:
- Docker 20.10+ installed
- Docker Compose 2.0+ installed
- Internet access to pull base images

## Test 1: Basic Docker Build and Run

```bash
# Build and start the container
docker compose up -d

# Check if container is running
docker ps | grep tasking

# View logs
docker compose logs -f tasking

# You should see: "Tasking server starting on port 8080..."

# Test the application
curl http://localhost:8080/login

# You should see HTML with "Tasking" as the title
```

## Test 2: Database Persistence

This test verifies that data persists between container restarts:

```bash
# 1. Start the application
docker compose up -d

# 2. Login and create some tasks
# Open http://localhost:8080
# Login as instructor1 / password123
# Create a few tasks

# 3. Stop the container
docker compose down

# 4. Verify the volume still exists
docker volume ls | grep tasking-data
# You should see: taskfolio_tasking-data

# 5. Restart the container
docker compose up -d

# 6. Login again
# Open http://localhost:8080
# Login as instructor1 / password123
# Verify your tasks are still there!
```

## Test 3: Volume Inspection

```bash
# Inspect the volume
docker volume inspect taskfolio_tasking-data

# You should see output like:
# [
#     {
#         "Driver": "local",
#         "Mountpoint": "/var/lib/docker/volumes/taskfolio_tasking-data/_data",
#         ...
#     }
# ]
```

## Test 4: Manual Database Backup

```bash
# List files in the data volume
docker compose exec tasking ls -la /data/

# You should see taskfolio.db

# Create a backup
docker compose exec tasking cp /data/taskfolio.db /data/backup-$(date +%Y%m%d).db

# Copy backup to host
docker cp tasking:/data/backup-*.db ./
```

## Test 5: Environment Variable Configuration

```bash
# Stop current container
docker compose down

# Edit docker-compose.yml to change the port
# Change: - "3000:8080"  # Use port 3000 on host

docker compose up -d

# Test on new port
curl http://localhost:3000/login
```

## Test 6: Complete Cleanup

```bash
# Stop and remove containers
docker compose down

# Remove the volume (WARNING: deletes all data)
docker volume rm taskfolio_tasking-data

# Verify volume is gone
docker volume ls | grep tasking-data
# Should return no results

# Clean rebuild
docker compose up -d --build
```

## Expected Results

✅ Container builds successfully  
✅ Application is accessible at http://localhost:8080  
✅ Login page shows "Tasking" branding  
✅ Tasks persist after container restart  
✅ Volume exists and contains database file  
✅ Environment variables are respected  

## Troubleshooting

### Container won't start
```bash
docker compose logs tasking
```

### Port already in use
```bash
# Find what's using port 8080
lsof -i :8080
# or
netstat -tulpn | grep 8080

# Either stop that service or change the port in docker-compose.yml
```

### Database corruption
```bash
# Reset everything
docker compose down
docker volume rm taskfolio_tasking-data
docker compose up -d
```

### Build fails
```bash
# Clean build cache
docker builder prune

# Rebuild without cache
docker compose build --no-cache
```

## Notes

- The database file is named `taskfolio.db` inside the container (at `/data/taskfolio.db`)
- The Docker volume is named `taskfolio_tasking-data` (prefixed with project name)
- Container name is `tasking`
- Service name in docker-compose is `tasking`
