# Production Docker Setup

This directory contains production-ready Docker configuration for the web application.

## Files Overview

- `Dockerfile.prod` - Multi-stage production Dockerfile
- `entrypoint.prod.sh` - Production entrypoint script
- `vite.config.prod.ts` - Production Vite configuration with CSP and caching
- `docker-compose.prod.yml` - Production docker-compose override
- `README.prod.md` - This documentation

## Production Features

### Security
- **Content Security Policy (CSP)** - Strict CSP headers for XSS protection
- **Security Headers** - X-Frame-Options, X-Content-Type-Options, etc.
- **Non-root User** - Runs as user 1001:1001 for security
- **Resource Limits** - Memory and CPU limits to prevent resource exhaustion

### Performance
- **Multi-stage Build** - Optimized build process with separate runtime stage
- **Chunk Splitting** - Vendor chunks for better caching
- **Minification** - Terser minification with console.log removal
- **Caching Headers** - Proper cache headers for static assets
- **Health Checks** - Built-in health monitoring

### Caching Strategy
- **Static Assets**: `Cache-Control: public, max-age=31536000, immutable`
- **HTML/API**: `Cache-Control: public, max-age=3600`
- **Vendor Chunks**: Separate chunks for React, Query, UI libraries

## Usage

### Build Production Image
```bash
# Build the production image
docker build -f Dockerfile.prod -t recsys-web:prod .

# Or use the npm script
pnpm docker:build:prod
```

### Run Production Container
```bash
# Run with docker-compose
docker-compose -f docker-compose.yml -f web/docker-compose.prod.yml up

# Or run standalone
docker run -p 3000:3000 \
  -e VITE_API_HOST=http://localhost:8081 \
  -e VITE_ALLOWED_HOSTS=localhost,0.0.0.0,127.0.0.1 \
  recsys-web:prod
```

### Local Production Testing
```bash
# Build and preview locally
pnpm build:prod
pnpm preview:prod
```

## Environment Variables

| Variable               | Default                       | Description                   |
|------------------------|-------------------------------|-------------------------------|
| `VITE_ALLOWED_HOSTS`   | `localhost,0.0.0.0,127.0.0.1` | Allowed hosts for Vite        |
| `VITE_API_HOST`        | `http://localhost:8081`       | API server URL                |
| `API_HEALTH_CHECK_URL` | -                             | Optional API health check URL |
| `NODE_ENV`             | `production`                  | Node environment              |

## CSP Configuration

The production build includes a strict Content Security Policy:

```
default-src 'self';
script-src 'self' 'unsafe-inline' 'unsafe-eval';
style-src 'self' 'unsafe-inline';
img-src 'self' data: blob: https:;
font-src 'self' data:;
connect-src 'self' ws: wss: https:;
object-src 'none';
base-uri 'self';
form-action 'self';
frame-ancestors 'none';
```

## Caching Strategy

### Static Assets (JS, CSS, Images)
- **Cache-Control**: `public, max-age=31536000, immutable`
- **Purpose**: Long-term caching for versioned assets

### HTML and API Responses
- **Cache-Control**: `public, max-age=3600`
- **Purpose**: Reasonable caching for dynamic content

### Vendor Chunks
- **React/React-DOM**: Separate chunk for framework code
- **TanStack Query**: Separate chunk for data fetching
- **UI Libraries**: Separate chunk for markdown, ML libraries
- **Purpose**: Better cache utilization across deployments

## Health Monitoring

The production container includes health checks:
- **Endpoint**: `http://localhost:3000`
- **Interval**: 30 seconds
- **Timeout**: 10 seconds
- **Retries**: 3 attempts
- **Start Period**: 40 seconds

## Resource Limits

Production containers have resource limits:
- **Memory Limit**: 512MB
- **Memory Reservation**: 256MB
- **CPU Limit**: 0.5 cores
- **CPU Reservation**: 0.25 cores

## Development vs Production

| Aspect            | Development  | Production           |
|-------------------|--------------|----------------------|
| **Build**         | Single stage | Multi-stage          |
| **Dependencies**  | All deps     | Production only      |
| **Source Maps**   | Full         | Optimized            |
| **Minification**  | None         | Terser               |
| **CSP**           | None         | Strict               |
| **Caching**       | None         | Aggressive           |
| **Security**      | Basic        | Hardened             |
| **User**          | Root         | Non-root (1001:1001) |
| **Health Checks** | None         | Built-in             |

## Troubleshooting

### Build Issues
```bash
# Check build logs
docker build -f Dockerfile.prod -t recsys-web:prod . --no-cache

# Verify build output
docker run --rm recsys-web:prod ls -la /app/dist
```

### Runtime Issues
```bash
# Check container logs
docker logs <container-id>

# Check health status
docker inspect <container-id> | grep -A 10 Health
```

### CSP Issues
If you encounter CSP violations, check the browser console for:
- Blocked script sources
- Blocked style sources
- Blocked image sources

Adjust the CSP in `vite.config.prod.ts` as needed.

## Security Considerations

1. **CSP Violations**: Monitor for CSP violations in production
2. **Dependencies**: Regularly update dependencies for security patches
3. **Secrets**: Never commit secrets to the repository
4. **Network**: Use HTTPS in production environments
5. **Updates**: Keep base images updated (node:20-alpine)
