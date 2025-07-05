# Bookstore API with Prometheus & Grafana Monitoring

A complete tutorial demonstrating Go Gin REST API with PostgreSQL, Prometheus metrics, and Grafana dashboards.

## Architecture Overview

This project consists of 4 Docker containers:
- **Go Gin API**: REST API server with CRUD operations for books
- **PostgreSQL**: Database for storing book data
- **Prometheus**: Metrics collection and storage
- **Grafana**: Metrics visualization and dashboards

## Prerequisites

- Docker & Docker Compose
- Go 1.23+ (for local development)
- Git

## Quick Start

1. **Clone and navigate to the project**:
   ```bash
   git clone <your-repo-url>
   cd gin-prometheus-grafana
   ```

2. **Start all services**:
   ```bash
   docker-compose up -d
   ```

3. **Wait for services to be ready** (about 30 seconds), then access:
   - API: http://localhost:8080
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000 (admin/admin)

## API Endpoints

### Books CRUD Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/books` | Create a new book |
| GET | `/api/v1/books` | Get all books |
| GET | `/api/v1/books/{id}` | Get book by ID |
| PUT | `/api/v1/books/{id}` | Update book |
| DELETE | `/api/v1/books/{id}` | Delete book |

### System Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrics |

## Example API Usage

### Create a Book
```bash
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "The Go Programming Language",
    "author": "Alan Donovan",
    "isbn": "9780134190440",
    "price": 49.99,
    "published_at": "2015-11-16T00:00:00Z"
  }'
```

### Get All Books
```bash
curl http://localhost:8080/api/v1/books
```

### Get Book by ID
```bash
curl http://localhost:8080/api/v1/books/1
```

### Update Book
```bash
curl -X PUT http://localhost:8080/api/v1/books/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Title",
    "price": 59.99
  }'
```

### Delete Book
```bash
curl -X DELETE http://localhost:8080/api/v1/books/1
```

## Monitoring & Metrics

### Prometheus Metrics

The API exposes the following metrics:

**HTTP Metrics**:
- `http_requests_total` - Total HTTP requests by method, path, and status code
- `http_request_duration_seconds` - HTTP request duration histogram
- `http_requests_in_flight` - Current number of HTTP requests being processed
- `http_request_size_bytes` - HTTP request size histogram
- `http_response_size_bytes` - HTTP response size histogram

**Database Metrics**:
- `db_query_total` - Total database queries by operation, table, and status
- `db_query_duration_seconds` - Database query duration histogram

### Grafana Dashboard

Access Grafana at http://localhost:3000 with credentials `admin/admin`.

The pre-configured dashboard includes:
- HTTP Request Rate
- HTTP Request Duration (95th & 50th percentiles)
- Database Query Duration
- Database Query Rate
- Current In-Flight Requests
- HTTP Status Code Distribution

## Project Structure

```
gin-prometheus-grafana/
├── cmd/server/main.go              # Main application entry point
├── internal/
│   ├── models/book.go              # Book model and DTOs
│   ├── repository/book_repository.go # Database layer with metrics
│   ├── handlers/book_handler.go     # HTTP handlers
│   └── middleware/prometheus.go     # Prometheus middleware
├── docker/
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/            # Prometheus datasource config
│   │   └── dashboards/             # Dashboard provisioning
│   └── dashboards/                 # Dashboard JSON files
├── prometheus/
│   └── prometheus.yml              # Prometheus configuration
├── docker-compose.yml              # Multi-container setup
├── Dockerfile                      # Go application container
├── go.mod                          # Go module file
└── .env                           # Environment variables
```

## Key Features

### 1. Repository Pattern
Clean separation of concerns with repository layer for database operations.

### 2. Comprehensive Logging
All requests, responses, and database queries are logged with structured logging.

### 3. Prometheus Integration
Custom metrics middleware captures:
- Request/response metrics
- Database query performance
- Business logic metrics

### 4. Grafana Dashboards
Pre-configured dashboards for:
- API performance monitoring
- Database query analysis
- Error rate tracking

### 5. Docker Orchestration
Complete containerized setup with:
- Service dependencies
- Health checks
- Volume persistence
- Network isolation

## Database Schema

The `books` table schema:
```sql
CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    isbn VARCHAR(13) UNIQUE NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    published_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Configuration

### Environment Variables
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password (default: postgres)
- `DB_NAME`: Database name (default: bookstore)
- `DB_SSL_MODE`: SSL mode (default: disable)
- `SERVER_PORT`: API server port (default: 8080)

### Prometheus Configuration
Scrapes metrics from the API every 5 seconds at `/metrics` endpoint.

### Grafana Configuration
- Auto-provisioned Prometheus datasource
- Pre-loaded dashboard for API monitoring
- Admin credentials: admin/admin

## Development

### Running Locally
```bash
# Start dependencies
docker-compose up -d postgres prometheus grafana

# Install dependencies
go mod download

# Run the application
go run cmd/server/main.go
```

### Building
```bash
# Build binary
go build -o bookstore cmd/server/main.go

# Build Docker image
docker build -t bookstore-api .
```

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Ensure PostgreSQL container is running
   - Check database credentials in .env file

2. **Metrics Not Showing in Grafana**
   - Verify Prometheus is scraping metrics: http://localhost:9090/targets
   - Check API is exposing metrics: http://localhost:8080/metrics

3. **Grafana Dashboard Not Loading**
   - Ensure Prometheus datasource is configured
   - Check dashboard provisioning in logs

### Viewing Logs
```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f api
docker-compose logs -f prometheus
docker-compose logs -f grafana
```

## Load Testing & Monitoring Demo

### Built-in Load Testing Scripts

The project includes several load testing scripts to generate realistic traffic for monitoring demonstrations:

#### 1. Simple Load Test (Bash + curl)
```bash
# Quick tests
./scripts/simple_load_test.sh light    # 30s, 10 req/s (300 requests)
./scripts/simple_load_test.sh medium   # 60s, 15 req/s (900 requests) 
./scripts/simple_load_test.sh heavy    # 120s, 20 req/s (2400 requests)

# Perfect for dashboard demo (5 minutes of steady traffic)
./scripts/simple_load_test.sh demo     # 300s, 5 req/s (1500 requests)

# Check service status
./scripts/simple_load_test.sh check
```

#### 2. Advanced Load Test (Go)
```bash
# Using the Go-based load tester
./scripts/load_test.sh light    # Quick test
./scripts/load_test.sh medium   # Standard test
./scripts/load_test.sh heavy    # Stress test

# Custom parameters: requests, concurrency, duration
./scripts/load_test.sh custom 500 15 90
```

#### 3. Manual Testing with curl
```bash
# Create books
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Book","author":"Test Author","isbn":"1234567890","price":29.99,"published_at":"2024-01-01T00:00:00Z"}'

# Read all books  
curl http://localhost:8080/api/v1/books

# Update book
curl -X PUT http://localhost:8080/api/v1/books/1 \
  -H "Content-Type: application/json" \
  -d '{"price":39.99}'

# Delete book
curl -X DELETE http://localhost:8080/api/v1/books/1
```

### Load Test Features

The scripts generate **realistic e-commerce traffic patterns**:

- **35% Create operations** - New book additions
- **30% Read operations** - Browse books (mix of list all + individual reads)
- **25% Update operations** - Price changes, inventory updates
- **10% Delete operations** - Remove books

**Smart behavior:**
- Maintains minimum book inventory (prevents empty database)
- Uses realistic book data from popular programming titles
- Handles concurrent operations safely
- Tracks book IDs dynamically
- Random delays between requests (simulates real user behavior)

### Watching the Dashboards

1. **Start load test in background:**
   ```bash
   ./scripts/simple_load_test.sh demo &
   ```

2. **Open monitoring dashboards:**
   - Grafana: http://localhost:3000 (admin/admin)
   - Prometheus: http://localhost:9090
   - API Metrics: http://localhost:8080/metrics

3. **Watch real-time metrics:**
   - HTTP request rates and latencies
   - Database query performance  
   - Error rates and status codes
   - System resource usage

### Performance Testing with External Tools

You can also use external tools:
```bash
# Install hey (load testing tool)
go install github.com/rakyll/hey@latest

# Generate load
hey -n 1000 -c 10 http://localhost:8080/api/v1/books
```

## Next Steps

1. Add authentication middleware
2. Implement rate limiting
3. Add more business metrics
4. Set up alerting rules
5. Add distributed tracing
6. Implement caching layer

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.