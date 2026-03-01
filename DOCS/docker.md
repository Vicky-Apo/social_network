# Docker Deployment Guide

This document explains the Docker containerization setup for the social-network project.

## Overview

The application is fully containerized using Docker and Docker Compose, making it easy to deploy and run in any environment. The setup includes:

- **Backend** - Go API server with automatic migrations
- **Frontend** - Next.js web application in standalone mode
- **Database** - PostgreSQL 16 with persistent storage

## Architecture

### Services

All services are defined in `docker-compose.yml` and communicate through a dedicated Docker network:

```
┌─────────────────┐
│    Frontend     │  Port 3000
│   (Next.js)     │
└────────┬────────┘
         │
         │ HTTP
         │
┌────────▼────────┐
│    Backend      │  Port 8080
│   (Go API)      │
└────────┬────────┘
         │
         │ PostgreSQL
         │
┌────────▼────────┐
│   PostgreSQL    │  Port 5433
│   (Database)    │
└─────────────────┘
```

### Network Configuration

- All services run on a custom bridge network: `social-network`
- Services communicate using container names as hostnames
- Only necessary ports are exposed to the host machine

## File Structure

```
social-network/
├── docker-compose.yml           # Orchestration configuration
├── backend/
│   ├── Dockerfile              # Backend container definition
│   └── .dockerignore           # Files to exclude from build
├── frontend/
│   ├── Dockerfile              # Frontend container definition
│   └── .dockerignore           # Files to exclude from build
└── DOCS/
    └── docker.md               # This file
```

## Backend Container

### Dockerfile Structure

The backend uses a **multi-stage build** to optimize image size:

1. **Builder Stage** (`golang:1.23.5-alpine3.21`)
   - Downloads Go dependencies
   - Compiles the application binary
   - Full Go toolchain available

2. **Production Stage** (`alpine:3.21`)
   - Minimal base image (security + size)
   - Copies only the compiled binary
   - Copies migration files
   - Installs security updates
   - No source code or build tools

### Environment Variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `SERVER_ADDR` | Server bind address | `0.0.0.0:8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@postgres:5432/db?sslmode=disable` |
| `MIGRATIONS_PATH` | Path to migration files | `/root/pkg/db/migrations/postgres` |
| `UPLOAD_DIR` | Directory for uploaded files | `/root/uploads` |
| `CORS_ALLOWED_ORIGINS` | Frontend URL for CORS | `http://localhost:3000` |

### Key Features

- **Automatic migrations** - Database schema is applied on startup
- **Persistent uploads** - Files stored in Docker volume `backend-uploads`
- **Health checks** - Container readiness via service dependencies
- **Security** - Non-root user, minimal attack surface

### Build Optimizations

```dockerfile
# .dockerignore excludes:
- .env files
- Test files
- Git history
- Development tools
```

This reduces build context from ~50MB to ~5MB, speeding up builds significantly.

## Frontend Container

### Dockerfile Structure

The frontend also uses a **multi-stage build**:

1. **Builder Stage** (`node:20-alpine`)
   - Installs npm dependencies
   - Builds Next.js application
   - Generates standalone output

2. **Production Stage** (`node:20-alpine`)
   - Copies standalone server
   - Copies static assets
   - Installs only production dependencies
   - No development packages

### Configuration

The frontend requires specific Next.js configuration for Docker deployment:

**`next.config.ts`:**
```typescript
const nextConfig: NextConfig = {
  output: 'standalone',  // Required for Docker
  // ... other config
};
```

**Why `standalone` mode?**
- Produces self-contained output
- Includes all necessary Node.js code
- Smaller image size (excludes dev dependencies)
- Better for containerized deployments

### Dynamic Page Rendering

Some pages (like `/messages`) use runtime features that can't be statically generated:

**`frontend/src/app/messages/page.tsx`:**
```typescript
export const dynamic = 'force-dynamic';
```

This prevents build errors when pages use:
- `useSearchParams()` without Suspense boundaries
- Server-side data fetching
- Real-time features (WebSockets)

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `NEXT_PUBLIC_API_URL` | Backend API URL | `http://localhost:8080` |
| `NODE_ENV` | Node environment | `production` |
| `NEXT_TELEMETRY_DISABLED` | Disable Next.js telemetry | `1` |

## Database Container

### Configuration

Uses the official PostgreSQL 16 image with:

- **Persistent storage** - Data stored in Docker volume `pgdata`
- **Health checks** - Ensures database is ready before starting backend
- **Custom port** - Exposed on 5433 (to avoid conflicts with local PostgreSQL)

### Environment Variables

| Variable | Purpose | Value |
|----------|---------|-------|
| `POSTGRES_USER` | Database user | `social-network-role` |
| `POSTGRES_PASSWORD` | Database password | `123456` |
| `POSTGRES_DB` | Database name | `social-network-db` |

### Health Check

```yaml
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U social-network-role -d social-network-db"]
  interval: 5s
  timeout: 5s
  retries: 10
```

This ensures the backend waits for PostgreSQL to be fully ready before starting.

## Docker Compose

### Service Dependencies

```yaml
frontend:
  depends_on:
    - backend

backend:
  depends_on:
    postgres:
      condition: service_healthy  # Wait for health check
```

This ensures services start in the correct order:
1. PostgreSQL starts and becomes healthy
2. Backend starts and runs migrations
3. Frontend starts and connects to backend

### Volumes

Two persistent volumes are created:

1. **`pgdata`** - PostgreSQL data directory
   - Survives container restarts
   - Ensures data persistence

2. **`backend-uploads`** - User uploaded files
   - Avatars, post images, etc.
   - Shared across backend restarts

### Port Mapping

| Service | Container Port | Host Port |
|---------|----------------|-----------|
| Frontend | 3000 | 3000 |
| Backend | 8080 | 8080 |
| PostgreSQL | 5432 | 5433 |

## Usage Commands

### Start Everything

```bash
docker-compose up -d
```

This command:
- Builds images if they don't exist
- Creates the network
- Creates volumes
- Starts all containers in background

### View Status

```bash
docker-compose ps
```

Shows:
- Running containers
- Port mappings
- Health status

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend

# Last 50 lines
docker-compose logs --tail=50 backend
```

### Rebuild After Changes

```bash
# Rebuild specific service
docker-compose build backend

# Rebuild without cache
docker-compose build --no-cache backend

# Rebuild and restart
docker-compose up -d --build backend
```

### Stop Everything

```bash
# Stop containers (keeps volumes)
docker-compose down

# Stop and remove volumes (DELETES DATA)
docker-compose down -v
```

### Access Container Shell

```bash
# Backend
docker-compose exec backend sh

# Database
docker-compose exec postgres psql -U social-network-role -d social-network-db

# Frontend
docker-compose exec frontend sh
```

## Troubleshooting

### Backend Won't Start

**Check logs:**
```bash
docker-compose logs backend
```

**Common issues:**
- Missing environment variables
- Database connection failed
- Migration errors
- Port already in use

**Solutions:**
```bash
# Check if port 8080 is available
lsof -i :8080

# Verify database is healthy
docker-compose ps postgres

# Restart with fresh database
docker-compose down -v
docker-compose up -d
```

### Frontend Build Fails

**Error:** `TypeError: fetch failed`

**Cause:** Next.js trying to fetch from backend during build

**Solution:** Use `dynamic = 'force-dynamic'` for pages that need runtime data

**Error:** `useSearchParams() should be wrapped in a suspense boundary`

**Solution:** Add to page.tsx:
```typescript
export const dynamic = 'force-dynamic';
```

### Database Connection Issues

**Check connection string format:**
```
postgres://user:password@host:port/database?sslmode=disable
```

**Inside containers, use container name:**
```
postgres://social-network-role:123456@postgres:5432/social-network-db?sslmode=disable
```

**From host machine, use localhost:**
```
postgres://social-network-role:123456@localhost:5433/social-network-db?sslmode=disable
```

### Slow Build Times

**Solutions:**
1. Use `.dockerignore` to exclude unnecessary files
2. Leverage Docker layer caching (don't change deps frequently)
3. Use `--parallel` flag: `docker-compose build --parallel`
4. Increase Docker resources (CPU/Memory) in Docker Desktop

### Volume Permissions

If you see permission errors with uploads:

```bash
# Check volume
docker volume inspect social-network_backend-uploads

# Fix permissions in container
docker-compose exec backend sh -c "chmod -R 755 /root/uploads"
```

## Security Considerations

### Production Recommendations

1. **Change default passwords**
   ```yaml
   POSTGRES_PASSWORD: "use-strong-random-password"
   ```

2. **Use secrets management**
   - Docker secrets
   - External secret stores (Vault, AWS Secrets Manager)

3. **Enable SSL/TLS**
   ```yaml
   DATABASE_URL: postgres://...?sslmode=require
   ```

4. **Scan images for vulnerabilities**
   ```bash
   docker scout cves social-network_backend
   docker scout cves social-network_frontend
   ```

5. **Run as non-root user**
   ```dockerfile
   USER nobody
   ```

6. **Limit container resources**
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '1.0'
         memory: 512M
   ```

7. **Use specific image tags** (not `latest`)
   ```dockerfile
   FROM golang:1.23.5-alpine3.21  # ✓ Specific
   FROM golang:latest             # ✗ Unpredictable
   ```

## CI/CD Integration

### Building Images

```bash
# Build with version tag
docker build -t social-network-backend:1.0.0 ./backend
docker build -t social-network-frontend:1.0.0 ./frontend

# Push to registry
docker push your-registry/social-network-backend:1.0.0
docker push your-registry/social-network-frontend:1.0.0
```

### Automated Testing

```bash
# Start services
docker-compose up -d

# Wait for health checks
sleep 10

# Run integration tests
docker-compose exec backend go test ./...

# Cleanup
docker-compose down -v
```

## Performance Optimization

### Image Size Comparison

| Service | Before Optimization | After Optimization |
|---------|--------------------|--------------------|
| Backend | ~800MB (full Go image) | ~30MB (Alpine + binary) |
| Frontend | ~1.2GB (all deps) | ~200MB (standalone) |

### Build Time Optimization

- **Layer caching** - Dependencies changed less frequently than code
- **Multi-stage builds** - Build tools not in production image
- **`.dockerignore`** - Reduces build context transfer time

### Runtime Optimization

- **Connection pooling** - Backend configures DB connection limits
- **Static assets** - Frontend serves from CDN-ready structure
- **Health checks** - Fast startup without waiting unnecessarily

## Deployment Environments

### Development

```bash
docker-compose up -d
# Hot reload not supported, rebuild after changes
docker-compose up -d --build
```

### Staging/Production

Use environment-specific compose files:

```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

**`docker-compose.prod.yml`:**
```yaml
services:
  backend:
    environment:
      - DEBUG=false
      - RATE_LIMIT_REQUESTS_PER_MINUTE=100
  
  postgres:
    volumes:
      - /data/postgres:/var/lib/postgresql/data
```

## Monitoring

### View Resource Usage

```bash
docker stats social-network-backend social-network-frontend social-network-postgres
```

### Check Health

```bash
# Backend API
curl http://localhost:8080/auth/me

# Frontend
curl http://localhost:3000

# Database
docker-compose exec postgres pg_isready
```

## Backup and Restore

### Backup Database

```bash
docker-compose exec postgres pg_dump -U social-network-role social-network-db > backup.sql
```

### Restore Database

```bash
cat backup.sql | docker-compose exec -T postgres psql -U social-network-role -d social-network-db
```

### Backup Uploads

```bash
docker run --rm -v social-network_backend-uploads:/data -v $(pwd):/backup alpine tar czf /backup/uploads.tar.gz -C /data .
```

### Restore Uploads

```bash
docker run --rm -v social-network_backend-uploads:/data -v $(pwd):/backup alpine tar xzf /backup/uploads.tar.gz -C /data
```

## Migration from Development Setup

If you were running locally without Docker:

1. **Export development data**
   ```bash
   pg_dump -h localhost -p 5433 -U social-network-role social-network-db > dev-data.sql
   ```

2. **Stop local services**
   ```bash
   # Stop local PostgreSQL if running
   # Stop backend server
   # Stop frontend dev server
   ```

3. **Start Docker environment**
   ```bash
   docker-compose up -d
   ```

4. **Import data**
   ```bash
   cat dev-data.sql | docker-compose exec -T postgres psql -U social-network-role -d social-network-db
   ```

5. **Verify migration**
   ```bash
   docker-compose logs -f
   curl http://localhost:8080/auth/me
   ```

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Next.js Docker Documentation](https://nextjs.org/docs/deployment#docker-image)
- [PostgreSQL Docker Hub](https://hub.docker.com/_/postgres)

## Summary

The Docker setup provides:

✅ **Easy deployment** - Single command to start everything  
✅ **Environment consistency** - Same setup everywhere  
✅ **Isolation** - Services don't conflict with local tools  
✅ **Scalability** - Ready for orchestration (Kubernetes, etc.)  
✅ **Security** - Minimal images, automatic updates  
✅ **Maintainability** - Clear configuration, easy updates
