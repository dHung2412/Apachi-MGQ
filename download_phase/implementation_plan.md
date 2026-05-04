# Lightweight Data Catalog — Implementation Plan

Xây dựng hệ thống quản lý Metadata và Data Lineage tinh gọn bằng Go, thay thế các tính năng cốt lõi của DataHub cho mục đích cá nhân.

## Confirmed Decisions

| Quyết định | Lựa chọn |
|-----------|----------|
| HTTP Framework | **Echo** (`labstack/echo/v4`) |
| Graph DB | **Memgraph** (dùng Neo4j Go driver qua Bolt protocol) |
| Protocol | **REST-only** (không gRPC) |
| Authentication | **JWT** (`echo-jwt/v5` + `golang-jwt/jwt/v5`) |
| Airflow | **2.x** (`on_success_callback` + Custom Hook) |
| Interface | **CLI + API** (không UI giai đoạn đầu) |

---

## Architecture Overview

```mermaid
graph TB
    subgraph "Airflow 2.x Cluster"
        AF[Airflow DAGs] -->|on_success_callback| HOOK[Custom Python Hook]
    end

    subgraph "Ingestion Layer"
        HOOK -->|"HTTP POST /api/v1/ingest<br/>Bearer JWT"| API[REST API - Echo]
        API --> AUTH[JWT Auth Middleware]
        AUTH --> VALID[Schema Validator]
        VALID --> QUEUE[Go Channel Queue]
    end

    subgraph "Processing Layer"
        QUEUE --> WP[Worker Pool]
        WP --> DIFF[Schema Diff Engine]
        WP --> MAP[Column Mapper]
        WP --> VER[Version Manager]
    end

    subgraph "Storage Layer"
        DIFF --> PG[(PostgreSQL)]
        MAP --> MG[(Memgraph)]
        VER --> PG
    end

    subgraph "Query Layer"
        QAPI[REST API - Query] --> PG
        QAPI --> MG
    end

    style AF fill:#e1f5fe
    style PG fill:#c8e6c9
    style MG fill:#fff3e0
    style WP fill:#f3e5f5
    style AUTH fill:#ffcdd2
```

### Luồng xử lý chính

1. **Airflow 2.x Job kết thúc** → `on_success_callback` gọi Custom Hook
2. Hook gửi **HTTP POST** với JWT token + metadata payload đến Go API
3. Echo middleware **xác thực JWT** → extract user identity
4. API **validate** payload → đẩy vào **Go Channel**
5. **Worker Pool** xử lý song song: schema diff, column mapping, versioning
6. Kết quả ghi xuống **PostgreSQL** (metadata) và **Memgraph** (graph lineage)

---

## Proposed Changes

### Phase 1: Project Foundation & Core API

#### [NEW] `go-data-catalog/cmd/server/main.go`

Entry point — load config, init DB connections, wire DI, start Echo server.

```go
func main() {
    cfg := config.Load()

    pgDB := postgres.Connect(cfg.PostgresDSN)
    mgDriver := memgraph.Connect(cfg.MemgraphURI) // Neo4j driver over Bolt

    // Repositories
    datasetRepo := postgres.NewDatasetRepository(pgDB)
    userRepo := postgres.NewUserRepository(pgDB)
    lineageRepo := memgraph.NewLineageRepository(mgDriver)

    // Services
    authSvc := service.NewAuthService(cfg.JWTSecret, userRepo)
    catalogSvc := service.NewCatalogService(datasetRepo)
    lineageSvc := service.NewLineageService(lineageRepo)
    ingestionSvc := service.NewIngestionService(catalogSvc, lineageSvc, cfg.WorkerCount)

    // Start worker pool
    ingestionSvc.Start(context.Background())

    // Router
    e := api.NewRouter(authSvc, catalogSvc, lineageSvc, ingestionSvc, cfg)
    e.Start(cfg.Port)
}
```

---

#### [NEW] `go-data-catalog/pkg/config/config.go`

```go
type Config struct {
    Port         string // :8080
    PostgresDSN  string // postgres://user:pass@localhost:5432/catalog
    MemgraphURI  string // bolt://localhost:7687
    MemgraphUser string
    MemgraphPass string
    JWTSecret    string // HMAC signing key
    WorkerCount  int    // default: 5
    LogLevel     string // debug, info, warn, error
}
```

---

#### [NEW] `go-data-catalog/internal/api/router.go`

Echo router with JWT middleware:

```go
func NewRouter(authSvc, catalogSvc, lineageSvc, ingestionSvc, cfg) *echo.Echo {
    e := echo.New()

    // Global middleware
    e.Use(middleware.Logger(), middleware.Recover(), middleware.CORS())

    // Public routes
    e.POST("/api/v1/auth/login", authHandler.Login)
    e.POST("/api/v1/auth/register", authHandler.Register)
    e.GET("/api/v1/health", healthHandler.Check)

    // Protected routes (JWT required)
    v1 := e.Group("/api/v1", echojwt.WithConfig(jwtConfig))
    v1.POST("/ingest", ingestHandler.Ingest)
    v1.GET("/datasets", datasetHandler.List)
    v1.GET("/datasets/:id", datasetHandler.Get)
    v1.GET("/datasets/:id/schema", datasetHandler.GetSchema)
    v1.GET("/datasets/:id/lineage", lineageHandler.GetLineage)
    v1.GET("/lineage/trace", lineageHandler.Trace)
    v1.PUT("/datasets/:id/owner", datasetHandler.UpdateOwner)
    v1.PUT("/datasets/:id/tags", datasetHandler.UpdateTags)

    return e
}
```

---

#### [NEW] `go-data-catalog/internal/api/middleware/jwt_auth.go`

JWT authentication using `echo-jwt/v5`:
- Extract user identity (user_id, username, role) from claims
- Inject into Echo context for downstream handlers
- Support `Authorization: Bearer <token>` header

---

#### [NEW] `go-data-catalog/internal/api/handlers/auth_handler.go`

| Endpoint | Description |
|----------|-------------|
| `POST /api/v1/auth/register` | Tạo user mới, hash password (bcrypt) |
| `POST /api/v1/auth/login` | Xác thực credentials, trả JWT token |

---

#### [NEW] `go-data-catalog/internal/api/handlers/ingest_handler.go`

Payload JSON mẫu từ Airflow:

```json
{
  "event_type": "job_completed",
  "job": {
    "dag_id": "etl_sales_pipeline",
    "task_id": "transform_sales",
    "run_id": "manual__2026-04-29T00:00:00",
    "execution_date": "2026-04-29T00:00:00Z",
    "duration_seconds": 120
  },
  "source_datasets": [
    {
      "name": "raw_sales", "database": "bronze_db", "schema": "public",
      "columns": [
        {"name": "id", "type": "BIGINT", "nullable": false},
        {"name": "amount", "type": "DECIMAL(10,2)", "nullable": false}
      ]
    }
  ],
  "target_datasets": [
    {
      "name": "fact_sales", "database": "gold_db", "schema": "public",
      "columns": [
        {"name": "sale_id", "type": "BIGINT", "nullable": false},
        {"name": "total_amount", "type": "DECIMAL(12,2)", "nullable": false}
      ]
    }
  ],
  "column_mappings": [
    {"source": "raw_sales.id", "target": "fact_sales.sale_id", "transformation": "direct_copy"},
    {"source": "raw_sales.amount", "target": "fact_sales.total_amount", "transformation": "SUM(amount)"}
  ],
  "tags": ["sales", "etl", "daily"]
}
```

---

### Phase 2: Data Models & Storage

#### PostgreSQL Entity Models

```mermaid
erDiagram
    DATASET {
        uuid id PK
        string name
        string database_name
        string schema_name
        string description
        uuid owner_id FK
        jsonb tags
        timestamp created_at
        timestamp updated_at
    }
    COLUMN_DEF {
        uuid id PK
        uuid dataset_id FK
        string name
        string data_type
        boolean nullable
        int ordinal_position
        int schema_version
    }
    SCHEMA_VERSION {
        uuid id PK
        uuid dataset_id FK
        int version_number
        jsonb diff_summary
        uuid changed_by FK
        timestamp created_at
    }
    USER_ENTITY {
        uuid id PK
        string username
        string email
        string password_hash
        string role
        timestamp created_at
    }
    JOB {
        uuid id PK
        string dag_id
        string task_id
        string run_id
        uuid triggered_by FK
        float duration_seconds
        string status
        timestamp execution_date
    }
    TAG {
        uuid id PK
        string key
        string value
    }
    DATASET ||--o{ COLUMN_DEF : "has columns"
    DATASET ||--o{ SCHEMA_VERSION : "has versions"
    DATASET }o--|| USER_ENTITY : "owned by"
    JOB }o--|| USER_ENTITY : "triggered by"
```

> [!NOTE]
> `USER_ENTITY` now includes `password_hash` for JWT authentication. Passwords hashed with bcrypt.

---

#### Memgraph Graph Model

Memgraph is Bolt-protocol compatible — we use the **Neo4j Go driver** (`neo4j-go-driver/v5`) to connect. All Cypher queries work identically.

**Node Types:**
| Label | Properties |
|-------|-----------|
| `:Dataset` | `id`, `name`, `database`, `schema` |
| `:Column` | `id`, `name`, `data_type`, `nullable`, `dataset_id` |
| `:Job` | `id`, `dag_id`, `task_id`, `run_id` |

**Relationships:**
| Relationship | From → To | Properties |
|-------------|-----------|-----------|
| `HAS_COLUMN` | Dataset → Column | `ordinal_position` |
| `DERIVED_FROM` | Column → Column | `transformation`, `job_id`, `timestamp` |
| `READS` | Job → Dataset | `timestamp` |
| `WRITES` | Job → Dataset | `timestamp` |
| `TRANSFORMS_TO` | Dataset → Dataset | `job_id`, `timestamp` |

**Core Cypher:**
```cypher
-- Upsert Dataset + Columns
MERGE (d:Dataset {id: $id}) SET d.name = $name, d.database = $db

-- Column lineage
MATCH (src:Column {id: $srcId}), (tgt:Column {id: $tgtId})
MERGE (tgt)-[r:DERIVED_FROM]->(src)
SET r.transformation = $transform, r.job_id = $jobId

-- Trace upstream (recursive)
MATCH path = (t:Column {id: $colId})-[:DERIVED_FROM*1..10]->(src)
RETURN path

-- Trace downstream impact
MATCH path = (s:Column {id: $colId})<-[:DERIVED_FROM*1..10]-(tgt)
RETURN path
```

---

#### [NEW] `go-data-catalog/internal/repository/memgraph/`

Files:
- `connection.go` — Neo4j driver over Bolt to Memgraph
- `lineage_repo.go` — Graph CRUD + lineage queries
- `queries.go` — Cypher constants

```go
// Connection example
driver, _ := neo4j.NewDriverWithContext(
    "bolt://localhost:7687",
    neo4j.BasicAuth(cfg.MemgraphUser, cfg.MemgraphPass, ""),
)
```

---

### Phase 3: Processing & Business Logic

#### [NEW] `go-data-catalog/internal/service/ingestion.go`

Worker Pool (Go Channels):

```go
type IngestionService struct {
    eventChan   chan IngestEvent   // buffered, cap=1000
    catalogSvc  CatalogService
    lineageSvc  LineageService
    workerCount int               // default=5
}

func (s *IngestionService) Start(ctx context.Context) {
    for i := 0; i < s.workerCount; i++ {
        go s.worker(ctx, i)
    }
}

func (s *IngestionService) processEvent(event IngestEvent) {
    // 1. Upsert datasets & columns → PostgreSQL
    // 2. Detect schema changes → create SchemaVersion with diff
    // 3. Create/update lineage graph → Memgraph
    // 4. Update ownership & tags
}
```

#### Other Services
- `catalog.go` — Dataset CRUD, schema diff, ownership, tags, search
- `lineage.go` — Record lineage, trace upstream/downstream, full path
- `schema_diff.go` — Compare schema version N vs N-1
- `auth.go` — JWT token generation, password hashing/verification

---

### Phase 4: Airflow 2.x Integration

#### [NEW] `airflow-provider/hooks/catalog_hook.py`

```python
from airflow.hooks.base import BaseHook
import requests

class DataCatalogHook(BaseHook):
    conn_name_attr = 'catalog_conn_id'
    default_conn_name = 'data_catalog_default'

    def __init__(self, catalog_conn_id='data_catalog_default'):
        super().__init__()
        self.conn = self.get_connection(catalog_conn_id)
        self.base_url = f"http://{self.conn.host}:{self.conn.port}"
        # JWT token stored in Connection's password field
        self.token = self.conn.password

    def push_lineage(self, payload: dict):
        response = requests.post(
            f"{self.base_url}/api/v1/ingest",
            json=payload,
            headers={"Authorization": f"Bearer {self.token}"},
            timeout=30
        )
        response.raise_for_status()
        return response.json()
```

#### [NEW] `airflow-provider/callbacks/catalog_callback.py`

```python
def on_job_complete(context):
    ti = context['task_instance']
    hook = DataCatalogHook()
    payload = {
        "event_type": "job_completed",
        "job": {"dag_id": ti.dag_id, "task_id": ti.task_id, ...},
        **ti.xcom_pull(task_ids=ti.task_id, key='lineage_metadata')
    }
    hook.push_lineage(payload)
```

---

### Phase 5: Optimization

- **Worker Pool tuning:** Channel buffer=1000, workers=CPU cores
- **Graceful shutdown:** `context.WithCancel` + `sync.WaitGroup`
- **Caching (optional):** Redis for hot schema/lineage queries (only if bottleneck found)

---

## Project Structure

```
go-data-catalog/
├── cmd/server/main.go
├── internal/
│   ├── api/
│   │   ├── router.go                  # Echo router + JWT config
│   │   ├── handlers/
│   │   │   ├── auth_handler.go        # Login, Register
│   │   │   ├── ingest_handler.go      # POST /ingest
│   │   │   ├── dataset_handler.go     # Dataset CRUD
│   │   │   ├── lineage_handler.go     # Lineage queries
│   │   │   └── health_handler.go
│   │   └── middleware/
│   │       └── jwt_auth.go            # JWT middleware config
│   ├── model/
│   │   ├── dataset.go, lineage.go, job.go, user.go, tag.go, event.go
│   ├── repository/
│   │   ├── interfaces.go
│   │   ├── postgres/                  # GORM-based repos
│   │   └── memgraph/                  # Neo4j driver → Memgraph
│   └── service/
│       ├── auth.go, ingestion.go, catalog.go, lineage.go, schema_diff.go
├── pkg/config/config.go
├── scripts/migrations/, scripts/memgraph/
├── docker-compose.yml                 # PostgreSQL + Memgraph
├── Makefile
├── .env.example
├── go.mod
└── README.md
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/labstack/echo/v4` | HTTP framework |
| `github.com/labstack/echo-jwt/v5` | JWT middleware for Echo |
| `github.com/golang-jwt/jwt/v5` | JWT token handling |
| `gorm.io/gorm` + `gorm.io/driver/postgres` | ORM |
| `github.com/neo4j/neo4j-go-driver/v5` | Memgraph via Bolt protocol |
| `github.com/google/uuid` | UUID generation |
| `github.com/rs/zerolog` | Structured logging |
| `golang.org/x/crypto` | bcrypt password hashing |

---

## Verification Plan

### Automated Tests
```bash
go test ./internal/... -v -cover
docker-compose up -d postgres memgraph
go test ./internal/repository/... -tags=integration -v
```

### Manual Verification
```bash
# Register + Login
curl -X POST localhost:8080/api/v1/auth/register -d '{"username":"hoshi","password":"..."}'
curl -X POST localhost:8080/api/v1/auth/login -d '{"username":"hoshi","password":"..."}'

# Ingest with JWT
curl -X POST localhost:8080/api/v1/ingest -H "Authorization: Bearer <token>" -d @test.json

# Query lineage
curl localhost:8080/api/v1/datasets/{id}/lineage -H "Authorization: Bearer <token>"
```

### Memgraph Lab
Access `http://localhost:3000` to visualize lineage graph.

---

## Execution Timeline

| Phase | Scope | Effort |
|-------|-------|--------|
| **Phase 1** | Project setup, Echo router, JWT auth, Models | 2-3 ngày |
| **Phase 2** | PostgreSQL repos (GORM), Memgraph repos, Migrations | 2-3 ngày |
| **Phase 3** | Worker pool, Catalog/Lineage/Auth services, Schema diff | 3-4 ngày |
| **Phase 4** | Airflow 2.x Hook/Callback (Python) | 1-2 ngày |
| **Phase 5** | Optimization, tuning | 1-2 ngày |

**Tổng: ~10-14 ngày**
