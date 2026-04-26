# SPUR — AGENT MASTER GUIDE
> Read this completely before writing any code.
> This is the single source of truth for the project.

---

## 1. PROJECT IDENTITY

**Name:** Spur (formerly citual)
**Go module:** github.com/spurbase/spur
**Purpose:** Composable Go backend platform — build once, reuse across projects.

---

## 2. CURRENT DIRECTORY STRUCTURE

```
spur/
├── cmd/main.go                            ← entry point
├── internal/
│   ├── app/app.go                         ← MODULE WIRING — edit here to add modules
│   ├── platform/                          ← NEVER edit per project
│   │   ├── appmodule/appmodule.go         ← Module definition + builder
│   │   ├── config/config.go               ← Config struct + Load()
│   │   ├── db/                            ← NewPool(), RunMigrations(), RunInTx()
│   │   ├── httpserver/                    ← Server, AuthGuard, VerificationGuard, APIKeyGuard
│   │   ├── logger/logger.go               ← Loggerx (zerolog)
│   │   ├── metrics/metrics.go             ← Prometheus counters — all modules use these
│   │   ├── security/                      ← Argon2id, verify
│   │   ├── temporal/client.go             ← Optional Temporal wrapper
│   │   └── utils/
│   └── modules/
│       ├── identity/                      ← AUTH MODULE — fixed, do not rebuild
│       │   ├── module.go                  ← Options, Config, Services, Module, New(), RegisterRoutes()
│       │   ├── core/domain/manifest.go   ← Manifest + ManifestPermission types
│       │   ├── core/ports/ports.go        ← All interfaces incl. TxManager, APIKeyService, RBACService
│       │   └── core/services/
│       │       └── module_sync_service.go ← RegisterManifest() — called by app.go
│       └── leadcrm/                       ← REFERENCE MODULE — copy this pattern
│           └── leadcrm-module.go
├── ui/
│   ├── shared/                            ← Theme system + reusable components
│   └── studio/                            ← Next.js 14 App Router (output: standalone)
├── deployments/
│   ├── docker-compose.yml                 ← Full stack incl. observability
│   ├── .env.example                       ← Copy to .env, fill in secrets
│   ├── loki/config.yaml
│   ├── promtail/config.yaml
│   ├── prometheus/prometheus.yml
│   ├── prometheus/alerts.yml
│   └── grafana/
│       ├── provisioning/
│       └── dashboards/spur-overview.json
├── Dockerfile.backend
├── Dockerfile.ui
├── Makefile
└── sqlc.yaml
```

---

## 3. MODULE STATUS

```
identity    FIXED   ← auth, RBAC, OAuth2, API keys, multi-tenancy
leadcrm     WORKS   ← reference module — copy this pattern exactly
metrics     BUILT   ← platform/metrics/metrics.go — use, don't rebuild
appmodule   BUILT   ← platform/appmodule/appmodule.go
temporal    BUILT   ← platform/temporal/client.go (optional)
notifications  NOT BUILT
storage        NOT BUILT
jobs           NOT BUILT
billing        NOT BUILT
audit          NOT BUILT
agent          NOT BUILT
```

---

## 4. THE MODULE PATTERN — COPY LEADCRM EXACTLY

Reference: `internal/modules/leadcrm/leadcrm-module.go`

```go
package mymodule

type Config struct { /* MYMODULE_* env vars */ }
type Options struct {
    DB  *pgxpool.Pool
    Log *logger.Loggerx
    Cfg Config
}
type Services struct { /* port interfaces only */ }
type Module struct {
    Services *Services
    Handlers *handlers.Handlers
    Manifest domain.Manifest
}
func New(ctx context.Context, opt Options) (*Module, error) { /* wire + return */ }
func (m *Module) RegisterRoutes(r chi.Router)               { /* mount routes */ }
```

Wire in `internal/app/app.go`:
```go
mod, err := mymodule.New(ctx, mymodule.Options{DB: dbPool, Log: log})
identityModule.Services.ModuleService.RegisterManifest(ctx, mod.Manifest)
mod.RegisterRoutes(r)
```

---

## 5. WHAT PLATFORM PROVIDES — DO NOT REBUILD

### Auth context (set by identity middleware on every authenticated request)
```go
httpserver.GetUserID(ctx)    → string (UUID)
httpserver.GetTenantID(ctx)  → string (UUID)
httpserver.GetRoles(ctx)     → []string
httpserver.HasPermission(ctx, "module.resource.action") → bool
httpserver.IsSuperAdmin(ctx) → bool
```

### Logging — ALWAYS use this, NEVER fmt.Printf/Println
```go
opt.Log.Info(ctx).Str("module", "mymodule").Str("entity_id", id.String()).Msg("what happened")
opt.Log.Error(ctx).Err(err).Str("module", "mymodule").Msg("what failed")
```

### Metrics
```go
import "github.com/spurbase/spur/internal/platform/metrics"
defer metrics.TrackDBQuery("ListJobsByTenant", "mymodule", time.Now())
metrics.ModuleErrorsTotal.WithLabelValues("mymodule", "db_error").Inc()
metrics.RecordAuthAttempt("password", "success")  // identity only
```

### DB — use sqlc generated functions, not raw SQL
```go
opt.DB  // *pgxpool.Pool — pass to postgres.NewStore(opt.DB)
```

---

## 6. SQL RULES

All application tables go in **public schema** (not identity schema).

Migration template:
```sql
CREATE TABLE IF NOT EXISTS entities (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID        NOT NULL,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_entities_tenant ON entities(tenant_id);
```

Add sqlc block to sqlc.yaml:
```yaml
  - name: "mymodule"
    engine: "postgresql"
    schema: "internal/modules/mymodule/sql/migrations"
    queries: "internal/modules/mymodule/sql/queries"
    gen:
      go:
        package: "sqlc"
        out: "internal/modules/mymodule/adapters/postgres/sqlc"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_interface: true
        emit_pointers_for_null_types: true
```

Run: `sqlc generate` or `make sqlc`

---

## 7. HANDLER PATTERN

```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. Permission — always first
    if !httpserver.HasPermission(r.Context(), "mymodule.entities.create") {
        http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
        return
    }
    // 2. Parse
    var cmd ports.CreateCmd
    json.NewDecoder(r.Body).Decode(&cmd)
    // 3. Always inject from context — never trust request body for these
    cmd.TenantID  = uuid.MustParse(httpserver.GetTenantID(r.Context()))
    cmd.CreatedBy = uuid.MustParse(httpserver.GetUserID(r.Context()))
    // 4. Call service
    result, err := h.svc.Create(r.Context(), cmd)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // 5. Respond
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(result)
}
```

---

## 8. CROSS-MODULE RULES — NEVER VIOLATE

- NEVER import another module's package
- NEVER FK to identity schema — use UUID soft refs
- NEVER call identity services from application modules
- NEVER rebuild auth, sessions, or JWT parsing
- NEVER use fmt.Printf/Println — use opt.Log
- NEVER call os.Exit or log.Fatal inside modules
- ALWAYS return errors from constructors

---

## 9. ENVIRONMENT VARIABLES

```bash
APP_ENV=development
HTTP_ADDR=:8080
DATABASE_URL=postgres://spur:changeme@localhost:5432/spur?sslmode=disable
REDIS_URL=redis://localhost:6379
OAUTH_ISSUER=http://localhost:8080
FOSITE_GLOBAL_SECRET=<openssl rand -hex 32>
AUTH_CLIENT_ID=<uuidgen>
AUTH_CLIENT_SECRET=<openssl rand -hex 24>
JWT_PRIVATE_KEY_PATH=./keys/private.pem
```

---

## 10. BUILD + VERIFY

```bash
make keys          # generate RSA key (first time)
make sqlc          # regenerate sqlc after query changes
go build ./...     # must pass with zero errors
make docker-up     # full stack including observability
curl http://localhost:8080/healthz   # expect 200
```

---

## 11. OBSERVABILITY

Logs: Grafana at http://localhost:3001 (admin/admin first login)

```logql
{app="spur"} | json | level="error"
{app="spur"} | json | tenant_id="<uuid>"
{app="spur"} | json | module="mymodule" | level="error"
{app="spur"} | json | duration_ms > 500
```

Alerts pre-configured: HighErrorRate, DBPoolExhausted, SlowResponses, AuthFailureSpike

---

## 12. PROJECT BRIEF FORMAT

When starting a new project, provide this structure:

```
PROJECT: <Name>
TAGLINE: <One sentence>
TENANCY: multi-tenant B2B | multi-tenant B2C | single-tenant

SPUR MODULES IN USE:
- identity   ← always

NEW MODULES TO BUILD:
- <name>: <description>

DATA MODEL:
<table>: <columns>

PERMISSIONS:
<module>.<resource>.<action>: <description>

UI PAGES:
<role> sees /<path> → <description>

SUCCESS CRITERIA:
<measurable outcome>
```
