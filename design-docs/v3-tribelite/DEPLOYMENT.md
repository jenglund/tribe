# TribeLite v3 - Deployment & Architecture

## Deployment Architecture

TribeLite v3 is designed as a self-hosted application that runs entirely on a local network using Docker containers. This document outlines the deployment model, configuration options, and local network setup.

## Container Architecture

### Docker Compose Stack
```yaml
# docker-compose.yml
version: '3.8'
services:
  tribelite:
    build: .
    ports:
      - "3000:3000"
    environment:
      - DATABASE_URL=postgres://tribelite:password@postgres:5432/tribelite
      - ENVIRONMENT=production
      - LOG_LEVEL=info
    depends_on:
      - postgres
    volumes:
      - ./uploads:/app/uploads
      - ./logs:/app/logs
    networks:
      - tribelite

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=tribelite
      - POSTGRES_USER=tribelite
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    networks:
      - tribelite

  # Optional: Redis for caching
  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    networks:
      - tribelite

volumes:
  postgres_data:
  redis_data:

networks:
  tribelite:
    driver: bridge
```

### Application Container (Go + React)
```dockerfile
# Dockerfile
FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.21-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
RUN go build -o tribelite cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=backend-builder /app/tribelite .
COPY --from=backend-builder /app/frontend/dist ./static
EXPOSE 3000
CMD ["./tribelite"]
```

## Local Network Configuration

### Network Discovery

#### Automatic IP Detection
```go
// pkg/network/discovery.go
package network

import (
    "net"
    "fmt"
    "log"
)

func GetLocalNetworkIPs() []string {
    var ips []string
    
    interfaces, err := net.Interfaces()
    if err != nil {
        log.Printf("Error getting network interfaces: %v", err)
        return ips
    }
    
    for _, iface := range interfaces {
        if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
            continue
        }
        
        addrs, err := iface.Addrs()
        if err != nil {
            continue
        }
        
        for _, addr := range addrs {
            if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
                if ipnet.IP.To4() != nil {
                    ips = append(ips, ipnet.IP.String())
                }
            }
        }
    }
    
    return ips
}

func LogAccessInformation() {
    ips := GetLocalNetworkIPs()
    
    fmt.Println("\n🏠 TribeLite is running!")
    fmt.Println("Access the app from any device on your network:")
    
    for _, ip := range ips {
        fmt.Printf("   📱 http://%s:3000\n", ip)
    }
    
    if len(ips) == 0 {
        fmt.Println("   ⚠️  Could not detect network IP. Try: http://localhost:3000")
    }
    
    fmt.Println("\n📋 Share these URLs with others on your network!")
    fmt.Println()
}
```

#### QR Code Generation (Optional Enhancement)
```go
// pkg/qr/generator.go
package qr

import (
    "github.com/skip2/go-qrcode"
    "fmt"
)

func GenerateAccessQR(baseURL string) ([]byte, error) {
    qrText := fmt.Sprintf("%s\n\nTribeLite\nCollaborative Decision Making", baseURL)
    
    return qrcode.Encode(qrText, qrcode.Medium, 256)
}

// Serve QR code at /qr endpoint for easy mobile access
func (h *Handler) GetQRCode(c *gin.Context) {
    ips := network.GetLocalNetworkIPs()
    if len(ips) == 0 {
        c.JSON(500, gin.H{"error": "Could not determine network IP"})
        return
    }
    
    url := fmt.Sprintf("http://%s:3000", ips[0])
    qrCode, err := GenerateAccessQR(url)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to generate QR code"})
        return
    }
    
    c.Header("Content-Type", "image/png")
    c.Data(200, "image/png", qrCode)
}
```

### Network Binding Configuration
```go
// main.go
func main() {
    router := gin.Default()
    
    // Serve static files (React app)
    router.Static("/static", "./static")
    router.StaticFile("/", "./static/index.html")
    
    // API routes
    api := router.Group("/api/v1")
    setupAPIRoutes(api)
    
    // Network discovery
    network.LogAccessInformation()
    
    // Bind to all interfaces so local network can access
    log.Println("Starting server on :3000")
    router.Run("0.0.0.0:3000")
}
```

## Environment Configuration

### Environment Variables
```bash
# .env.example
# Database Configuration
DATABASE_URL=postgres://tribelite:password@postgres:5432/tribelite

# Application Settings
ENVIRONMENT=development  # development, production
LOG_LEVEL=info          # debug, info, warn, error
PORT=3000

# Feature Flags
ENABLE_QR_CODES=true
ENABLE_AVATAR_UPLOAD=true
MAX_UPLOAD_SIZE_MB=10

# Decision Making Defaults
DEFAULT_K_ELIMINATIONS=2
DEFAULT_M_FINAL_OPTIONS=1
MAX_K_ELIMINATIONS=5
MAX_M_FINAL_OPTIONS=5

# Storage Configuration
UPLOAD_DIR=/app/uploads
LOG_DIR=/app/logs
MAX_LOG_FILE_SIZE_MB=100

# Performance Settings
MAX_CONCURRENT_USERS=20
DATABASE_MAX_CONNECTIONS=10
CACHE_TTL_MINUTES=30
```

### Configuration Loading
```go
// pkg/config/config.go
package config

import (
    "os"
    "strconv"
    "log"
)

type Config struct {
    DatabaseURL string
    Environment string
    LogLevel    string
    Port        string
    
    // Feature flags
    EnableQRCodes     bool
    EnableAvatarUpload bool
    MaxUploadSizeMB   int
    
    // Decision defaults
    DefaultK int
    DefaultM int
    MaxK     int
    MaxM     int
    
    // Storage
    UploadDir string
    LogDir    string
    
    // Performance
    MaxConcurrentUsers    int
    DatabaseMaxConnections int
    CacheTTLMinutes       int
}

func Load() *Config {
    return &Config{
        DatabaseURL: getEnvOrDefault("DATABASE_URL", "postgres://tribelite:password@localhost:5432/tribelite"),
        Environment: getEnvOrDefault("ENVIRONMENT", "development"),
        LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),
        Port:        getEnvOrDefault("PORT", "3000"),
        
        EnableQRCodes:     getBoolEnvOrDefault("ENABLE_QR_CODES", true),
        EnableAvatarUpload: getBoolEnvOrDefault("ENABLE_AVATAR_UPLOAD", true),
        MaxUploadSizeMB:   getIntEnvOrDefault("MAX_UPLOAD_SIZE_MB", 10),
        
        DefaultK: getIntEnvOrDefault("DEFAULT_K_ELIMINATIONS", 2),
        DefaultM: getIntEnvOrDefault("DEFAULT_M_FINAL_OPTIONS", 1),
        MaxK:     getIntEnvOrDefault("MAX_K_ELIMINATIONS", 5),
        MaxM:     getIntEnvOrDefault("MAX_M_FINAL_OPTIONS", 5),
        
        UploadDir: getEnvOrDefault("UPLOAD_DIR", "./uploads"),
        LogDir:    getEnvOrDefault("LOG_DIR", "./logs"),
        
        MaxConcurrentUsers:    getIntEnvOrDefault("MAX_CONCURRENT_USERS", 20),
        DatabaseMaxConnections: getIntEnvOrDefault("DATABASE_MAX_CONNECTIONS", 10),
        CacheTTLMinutes:       getIntEnvOrDefault("CACHE_TTL_MINUTES", 30),
    }
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getBoolEnvOrDefault(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if result, err := strconv.ParseBool(value); err == nil {
            return result
        }
        log.Printf("Invalid boolean value for %s: %s, using default: %v", key, value, defaultValue)
    }
    return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if result, err := strconv.Atoi(value); err == nil {
            return result
        }
        log.Printf("Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
    }
    return defaultValue
}
```

## Data Storage & Persistence

### Database Initialization
```sql
-- migrations/001_initial_schema.sql
-- This file is automatically run when the postgres container starts

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(500),
    preferences JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Lists table
CREATE TABLE lists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    is_community BOOLEAN DEFAULT FALSE,
    category VARCHAR(100),
    k_default INTEGER DEFAULT 2,
    m_default INTEGER DEFAULT 1,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- [Additional tables as defined in DATA-MODEL.md]
-- ...

-- Create indexes
CREATE INDEX idx_users_name ON users(name);
CREATE INDEX idx_lists_created_by ON lists(created_by);
-- [Additional indexes as defined in DATA-MODEL.md]
-- ...
```

### File Storage Strategy
```go
// pkg/storage/files.go
package storage

import (
    "os"
    "path/filepath"
    "fmt"
    "io"
    "mime/multipart"
)

type FileStorage struct {
    uploadDir string
    maxSizeMB int
}

func NewFileStorage(uploadDir string, maxSizeMB int) *FileStorage {
    // Ensure upload directory exists
    os.MkdirAll(uploadDir, 0755)
    
    return &FileStorage{
        uploadDir: uploadDir,
        maxSizeMB: maxSizeMB,
    }
}

func (fs *FileStorage) SaveAvatar(userID string, file *multipart.FileHeader) (string, error) {
    // Validate file size
    if file.Size > int64(fs.maxSizeMB)*1024*1024 {
        return "", fmt.Errorf("file too large: %d bytes (max: %d MB)", file.Size, fs.maxSizeMB)
    }
    
    // Generate file path
    ext := filepath.Ext(file.Filename)
    filename := fmt.Sprintf("avatar_%s%s", userID, ext)
    filePath := filepath.Join(fs.uploadDir, "avatars", filename)
    
    // Ensure directory exists
    os.MkdirAll(filepath.Dir(filePath), 0755)
    
    // Save file
    src, err := file.Open()
    if err != nil {
        return "", err
    }
    defer src.Close()
    
    dst, err := os.Create(filePath)
    if err != nil {
        return "", err
    }
    defer dst.Close()
    
    if _, err := io.Copy(dst, src); err != nil {
        return "", err
    }
    
    // Return relative URL for serving
    return fmt.Sprintf("/uploads/avatars/%s", filename), nil
}
```

## Mobile PWA Configuration

### Progressive Web App Manifest
```json
{
  "name": "TribeLite",
  "short_name": "TribeLite",
  "description": "Collaborative Decision Making for Small Groups",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#ffffff",
  "theme_color": "#2563eb",
  "orientation": "portrait-primary",
  "icons": [
    {
      "src": "/static/icons/icon-72x72.png",
      "sizes": "72x72",
      "type": "image/png"
    },
    {
      "src": "/static/icons/icon-96x96.png",
      "sizes": "96x96",
      "type": "image/png"
    },
    {
      "src": "/static/icons/icon-128x128.png",
      "sizes": "128x128",
      "type": "image/png"
    },
    {
      "src": "/static/icons/icon-144x144.png",
      "sizes": "144x144",
      "type": "image/png"
    },
    {
      "src": "/static/icons/icon-152x152.png",
      "sizes": "152x152",
      "type": "image/png"
    },
    {
      "src": "/static/icons/icon-192x192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "/static/icons/icon-384x384.png",
      "sizes": "384x384",
      "type": "image/png"
    },
    {
      "src": "/static/icons/icon-512x512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ]
}
```

### Service Worker for Offline Capability
```typescript
// frontend/public/sw.js
const CACHE_NAME = 'tribelite-v1';
const urlsToCache = [
  '/',
  '/static/js/bundle.js',
  '/static/css/main.css',
  '/manifest.json'
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => cache.addAll(urlsToCache))
  );
});

self.addEventListener('fetch', (event) => {
  event.respondWith(
    caches.match(event.request)
      .then((response) => {
        // Return cached version or fetch from network
        return response || fetch(event.request);
      })
  );
});
```

## Performance Optimization

### Database Optimization
```sql
-- Regular maintenance queries for PostgreSQL
-- These should be run periodically (weekly/monthly)

-- Update table statistics
ANALYZE;

-- Reclaim storage and update statistics
VACUUM ANALYZE;

-- Check for unused indexes
SELECT schemaname, tablename, attname, n_distinct, correlation 
FROM pg_stats 
WHERE tablename IN ('users', 'lists', 'list_items', 'activities');

-- Monitor query performance
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;
```

### Caching Strategy (Optional Redis)
```go
// pkg/cache/redis.go
package cache

import (
    "encoding/json"
    "time"
    "github.com/go-redis/redis/v8"
    "context"
)

type Cache struct {
    client *redis.Client
    ttl    time.Duration
}

func NewCache(redisURL string, ttl time.Duration) *Cache {
    opts, _ := redis.ParseURL(redisURL)
    client := redis.NewClient(opts)
    
    return &Cache{
        client: client,
        ttl:    ttl,
    }
}

func (c *Cache) GetUserLists(userID string) ([]List, error) {
    ctx := context.Background()
    key := fmt.Sprintf("user_lists:%s", userID)
    
    val, err := c.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, nil // Cache miss
    } else if err != nil {
        return nil, err
    }
    
    var lists []List
    err = json.Unmarshal([]byte(val), &lists)
    return lists, err
}

func (c *Cache) SetUserLists(userID string, lists []List) error {
    ctx := context.Background()
    key := fmt.Sprintf("user_lists:%s", userID)
    
    data, err := json.Marshal(lists)
    if err != nil {
        return err
    }
    
    return c.client.Set(ctx, key, data, c.ttl).Err()
}
```

## Backup & Recovery

### Automated Backup Script
```bash
#!/bin/bash
# scripts/backup.sh

BACKUP_DIR="/app/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="tribelite_backup_${DATE}.sql"

# Create backup directory
mkdir -p $BACKUP_DIR

# Create database backup
docker exec tribelite_postgres_1 pg_dump -U tribelite tribelite > "$BACKUP_DIR/$BACKUP_FILE"

# Compress backup
gzip "$BACKUP_DIR/$BACKUP_FILE"

# Keep only last 7 days of backups
find $BACKUP_DIR -name "tribelite_backup_*.sql.gz" -mtime +7 -delete

echo "Backup created: $BACKUP_FILE.gz"
```

### Docker Volume Backup
```bash
#!/bin/bash
# scripts/backup-volumes.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/app/backups/volumes"

mkdir -p $BACKUP_DIR

# Backup database volume
docker run --rm \
  -v tribelite_postgres_data:/data \
  -v $BACKUP_DIR:/backup \
  alpine tar czf /backup/postgres_data_${DATE}.tar.gz -C /data .

# Backup upload volume  
docker run --rm \
  -v $(pwd)/uploads:/data \
  -v $BACKUP_DIR:/backup \
  alpine tar czf /backup/uploads_${DATE}.tar.gz -C /data .

echo "Volume backups created in $BACKUP_DIR"
```

## Quick Start Guide

### Installation Steps
1. **Clone Repository**
   ```bash
   git clone https://github.com/your-repo/tribelite.git
   cd tribelite
   ```

2. **Configure Environment**
   ```bash
   cp .env.example .env
   # Edit .env with your preferences
   ```

3. **Start Services**
   ```bash
   docker-compose up -d
   ```

4. **Access Application**
   - Check console output for local network IP addresses
   - Open browser to `http://[YOUR_LOCAL_IP]:3000`
   - Create first user profile and start adding lists!

### Troubleshooting
- **Can't access from other devices**: Check firewall settings and ensure Docker is binding to 0.0.0.0
- **Database connection issues**: Verify PostgreSQL container is running with `docker-compose ps`
- **Slow performance**: Consider enabling Redis cache and adjusting resource limits
- **Storage issues**: Monitor disk usage and set up regular backup rotation

This deployment model prioritizes simplicity and local network optimization while providing the foundation for a reliable, self-hosted collaborative decision-making platform. 