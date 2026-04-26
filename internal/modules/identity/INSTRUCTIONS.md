# Identity Module Build Plan

## Phase 1: Domain & Database (The Foundation)
- [ ] **1.1 Define Domain Models** (`core/domain/`)
    - Define structs: `User`, `Tenant`, `Role`, `Permission`.
    - Define strict types for `TenantKind`.
- [ ] **1.2 Create Migrations** (`sql/migrations/`)
    - `tenants` (Constraint: only one ops tenant).
    - `users` (Constraint: unique super admin).
    - `roles`, `permissions`, `role_permissions`, `user_roles`.
    - `fosite_clients`, `fosite_tokens` (access/refresh/code/pkce).
    - `audit_events`.
- [ ] **1.3 Write SQL Queries** (`sql/queries/`)
    - `queries.sql`: standard CRUD.
    - `fosite.sql`: specific queries to store/retrieve OAuth sessions.
    - Run `sqlc generate`.

## Phase 2: Core Business Logic (Service Layer)
- [ ] **2.1 Implement Tenant Service**
    - Logic to ensure only one Ops tenant exists.
    - Tenant creation flow (with default admin user).
- [ ] **2.2 Implement User/Auth Service**
    - Password hashing (Argon2id).
    - Login validation.
- [ ] **2.3 Implement RBAC Service**
    - `HasPermission(ctx, userID, permissionKey)`
    - Caching layer (permissions by role).
    - Logic to bump `authz_version` on role updates.

## Phase 3: Fosite (OAuth2) Integration
- [ ] **3.1 Fosite Storage Adapter** (`adapters/fosite_store`)
    - Implement `fosite.Storage` interface using `sqlc` generated code.
- [ ] **3.2 OAuth2 Handlers** (`adapters/http/oauth.go`)
    - `GET /oauth/authorize` (PKCE handling).
    - `POST /oauth/token` (Exchange code for tokens).
    - `POST /oauth/introspect` (Resource server validation).

## Phase 4: HTTP API & Middleware
- [ ] **4.1 Middleware** (`adapters/http/middleware`)
    - `AuthMiddleware`: Validates JWT, extracts claims to Context.
    - `TenantIsolationMiddleware`: Ensures `JWT.tid == URL.tenant_id`.
    - `AuditMiddleware`: Logs sensitive actions to DB.
- [ ] **4.2 REST Handlers**
    - Ops: `POST /ops/tenants` (Create Tenant).
    - Users: `POST /tenants/{id}/users` (Invite User).
    - Roles: `POST /tenants/{id}/roles`.

## Phase 5: Seeding & Testing
- [ ] **5.1 Seeding**
    - Create the "Bootstrap" function to generate the default Ops Tenant and Super Admin if DB is empty.
- [ ] **5.2 Integration Tests**
    - Test full OAuth2 flow (Login -> Code -> Token).
    - Test Tenant Isolation (User A cannot access Tenant B).