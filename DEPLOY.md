# Deployment Guide for Tasking

This guide describes how to deploy the Tasking application to a Docker host.

## Prerequisites

- Docker host with Docker Engine installed (20.10+)
- Docker Compose installed (v2.0+)
- SSH access to the deployment host
- Git installed on the deployment host
- Ability to run `git pull` on the remote host

## Deployment Steps

### 1. Initial Setup on Deployment Host

SSH into your deployment host and clone the repository:

```bash
ssh user@your-deployment-host
cd /opt  # or your preferred deployment directory
git clone https://github.com/kristofer/TaskFolio.git
cd TaskFolio
```

### 2. Configure Environment (Optional)

The application uses sensible defaults, but you can customize settings if needed:

```bash
# Create environment file (optional)
cat > .env <<EOF
PORT=8080
DATABASE_PATH=/data/taskfolio.db
EOF
```

Available environment variables:
- `PORT`: Server port (default: 8080)
- `DATABASE_PATH`: Path to SQLite database file (default: /data/taskfolio.db)

### 3. Build and Start Services

Build the Docker image and start the container:

```bash
docker compose up -d --build
```

This will:
- Build the Go application using multi-stage Docker build
- Create a persistent volume for the SQLite database
- Start the container in detached mode
- Expose the application on port 8080

### 4. Verify Deployment

Check that the container is running:

```bash
docker compose ps
```

View logs:

```bash
docker compose logs -f tasking
```

Test the application:

```bash
curl http://localhost:8080
```

Or open in a browser: `http://your-deployment-host:8080`

## Updates and Redeployment

To deploy updates from the repository:

```bash
cd /opt/TaskFolio  # or your deployment directory
git pull
docker compose down
docker compose up -d --build
```

The database will persist across deployments in the `tasking-data` Docker volume.

## Production Considerations

### Reverse Proxy with HTTPS

For production, run behind a reverse proxy (nginx, Traefik, or Caddy) for HTTPS:

#### Example with nginx

Create `/etc/nginx/sites-available/tasking`:

```nginx
server {
    listen 80;
    server_name tasking.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name tasking.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/tasking.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/tasking.yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable and restart nginx:

```bash
sudo ln -s /etc/nginx/sites-available/tasking /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

#### Example with Caddy

Create a `Caddyfile`:

```
tasking.yourdomain.com {
    reverse_proxy localhost:8080
}
```

### Database Backups

Set up automated backups of the SQLite database:

```bash
# Create backup script
sudo tee /usr/local/bin/backup-tasking.sh > /dev/null <<'EOF'
#!/bin/bash
BACKUP_DIR="/backup/tasking"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

# Stop writes (optional, depends on write frequency)
docker compose -f /opt/TaskFolio/docker-compose.yml exec tasking sqlite3 /data/taskfolio.db ".backup /data/taskfolio-backup-${TIMESTAMP}.db"

# Copy backup out of container
docker cp tasking:/data/taskfolio-backup-${TIMESTAMP}.db ${BACKUP_DIR}/

# Keep last 30 days of backups
find $BACKUP_DIR -name "taskfolio-backup-*.db" -mtime +30 -delete

echo "Backup completed: ${BACKUP_DIR}/taskfolio-backup-${TIMESTAMP}.db"
EOF

sudo chmod +x /usr/local/bin/backup-tasking.sh
```

Schedule with cron (daily at 2 AM):

```bash
sudo crontab -e
# Add line:
0 2 * * * /usr/local/bin/backup-tasking.sh
```

### Resource Limits

Add resource limits to `docker-compose.yml` for production:

```yaml
services:
  tasking:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: tasking
    ports:
      - "8080:8080"
    volumes:
      - tasking-data:/data
    environment:
      - PORT=8080
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M

volumes:
  tasking-data:
    driver: local
```

### Security Hardening

1. **Run on non-standard port internally** and use reverse proxy
2. **Enable firewall** to only expose necessary ports:
   ```bash
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw allow 22/tcp
   sudo ufw enable
   ```

3. **Keep Docker and host updated**:
   ```bash
   sudo apt update && sudo apt upgrade -y
   ```

4. **Monitor logs** for suspicious activity:
   ```bash
   docker compose logs -f --tail=100
   ```

## Monitoring and Health Checks

### Check Container Status

```bash
docker compose ps
docker compose logs --tail=50 tasking
```

### View Resource Usage

```bash
docker stats tasking
```

### Database Size

```bash
docker compose exec tasking du -h /data/taskfolio.db
```

## Troubleshooting

### Container Won't Start

Check logs:
```bash
docker compose logs tasking
```

### Database Issues

Access the database directly:
```bash
docker compose exec tasking sqlite3 /data/taskfolio.db
```

### Port Already in Use

Change the port mapping in `docker-compose.yml`:
```yaml
ports:
  - "8081:8080"  # Map to different host port
```

### Reset Everything

To completely reset (WARNING: destroys all data):
```bash
docker compose down -v  # -v removes volumes
docker compose up -d --build
```

## Rollback Procedure

If a deployment fails:

```bash
cd /opt/TaskFolio
git log --oneline  # Find previous commit
git checkout <previous-commit-hash>
docker compose down
docker compose up -d --build
```

## Demo Accounts

After deployment, the application includes these demo accounts:

**Instructor Account:**
- Username: `instructor1`
- Password: `password123`

**Student Accounts:**
- Username: `student1`, `student2`, or `student3`
- Password: `password123`

**Note:** Change these credentials in production!

## Support

For issues or questions:
- GitHub Issues: https://github.com/kristofer/TaskFolio/issues
- Check logs: `docker compose logs -f tasking`
