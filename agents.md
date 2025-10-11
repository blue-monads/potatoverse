# Potatoverse Platform - Developer & Agent Documentation

## Project Overview

**Potatoverse** (formerly "Turnix") is a self-contained application platform for hosting web applications with server-side code execution. A single Go binary that runs web apps with Lua backends and serves frontends in isolated iframes.

### Key Features

- **Single Binary**: Everything runs from one Go executable
- **Efficient**: Built with Go and SQLite, minimal resource usage and dependencies
- **Self-Hosted**: Run your own app platform on your hardware
- **Sandboxed**: Each app runs in its own isolated environment (Lua VM + iframe)

### Example Use Cases

- **Personal Apps**: Todo lists, notes, personal tools
- **Team Tools**: Collaborative apps for small teams
- **Learning**: Build and test web apps locally
- **Offline Use**: Apps run without internet connectivity

### Technical Approach
- **Batteries Included**: Common functionality (auth, storage, file handling) provided by the platform
- **SQLite Storage**: Single database file, no separate database server needed
- **Lua Execution**: Lightweight VM for running server-side code
- **Iframe Isolation**: Frontend apps run in sandboxed iframes for security

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                 Potatoverse Platform                     │
├─────────────────────────────────────────────────────────┤
│  Frontend (Next.js)                                      │
│  - Admin Portal (/portal/admin)                         │
│  - Auth Pages                                            │
│  - User Management                                       │
├─────────────────────────────────────────────────────────┤
│  Backend (Go)                                            │
│  ┌───────────────────────────────────────────────────┐ │
│  │  Server Layer                                      │ │
│  │  - HTTP Routes (/zz/api/core, /zz/space/*)       │ │
│  │  - Authentication & Authorization                  │ │
│  └───────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────┐ │
│  │  Controller/Actions Layer                          │ │
│  │  - User Management                                 │ │
│  │  - Space Management                                │ │
│  │  - Package Management                              │ │
│  └───────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────┐ │
│  │  Engine (App Runtime)                              │ │
│  │  - Routing Index (namespace → space mapping)      │ │
│  │  - Runtime (Lua executor pool management)         │ │
│  │  - Executor Pool (LuaStatePool)                   │ │
│  └───────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────┐ │
│  │  Services                                          │ │
│  │  - DataHub (Database operations)                  │ │
│  │  - Mailer                                          │ │
│  │  - Signer (Token signing)                         │ │
│  └───────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────┐ │
│  │  Storage                                           │ │
│  │  - SQLite Database (data.db)                      │ │
│  │  - Working Directory (./tmp)                      │ │
│  └───────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## Core Components

### 1. App Structure (`backend/app/`)

**HeadLess Application** (`app.go`):
- Core application without HTTP server
- Initializes database, logger, signer, mailer
- Creates controller and engine

**Full Application**:
- HeadLess + HTTP Server
- Serves frontend and API routes
- Handles authentication

### 2. Engine (`backend/engine/`)

The Engine is the heart of the platform, responsible for running user applications (called "spaces").

#### Key Files:
- `engine.go`: Main engine with routing index and runtime management
- `runtime.go`: Manages Lua executor instances, handles HTTP requests to spaces
- `epackage.go`: Handles embedded packages (apps bundled with platform)
- `routingIndex.go`: Maps namespace keys to space IDs

#### How It Works:

1. **Routing Index**: Maps namespace keys to space/package IDs
2. **Runtime**: Maintains a pool of Lua executors per package
3. **Request Flow**:
   ```
   HTTP Request → Engine.ServeSpaceFile() or Engine.SpaceApi()
   → Runtime.ExecHttp()
   → LuaStatePool.Get()
   → Execute Lua handler (on_http)
   → Return response
   ```

### 3. Executors (`backend/engine/executors/luaz/`)

**Lua Execution Environment**:
- `luaz.go`: Main Lua executor wrapper
- `pool.go`: LuaStatePool for reusing Lua states (10-20 instances, TTL: 1 hour)
- `luah.go`: Individual Lua handler with module registration

**Bindings** (`binds/`):
- `b_kv.go`: Key-Value storage bindings
- `b_ufs.go`: User File System bindings
- `http_req.go`: HTTP request/response bindings
- `b_utils.go`: Utility functions

**Lua Handler Example**:
```lua
function on_http(ctx)
    local req = ctx.request()
    
    -- Access space data
    local space_id = ctx.param("space_id")
    
    -- Use KV store binding
    local kv = require("kv")
    kv.add({
        group = "test",
        key = "mykey",
        value = "myvalue"
    })
    
    -- Return JSON response
    req.json(200, {
        message = "Hello from Lua!",
        space_id = space_id
    })
end
```

### 4. Data Layer (`backend/services/datahub/`)

#### Database Operations (`database/`):
- `db.go`: Database connection and setup
- `dbops_user.go`: User CRUD operations
- `dbops_space.go`: Space CRUD operations
- `dbops_space_kv.go`: Key-Value store operations
- `dbops_package.go`: Package management
- `dbops_fileops.go`: File storage operations

#### Models (`dbmodels/`):
- `user.go`: User, UserGroup, UserInvite, UserDevice, UserMessage
- `space.go`: Space, SpaceUser, SpaceKV, SpacePlugin, SpaceConfig
- `package.go`: Package, PackageFile, PackageFileBlob
- `file.go`: File, FileBlob models

### 5. HTTP Server (`backend/app/server/`)

#### Routes Structure:
```
/zz
├── /api/core
│   ├── /auth          (login, invites)
│   ├── /user          (user management)
│   ├── /self          (current user info)
│   ├── /package       (install, list, delete packages)
│   ├── /space         (space management, KV, files)
│   └── /engine        (debug, space info)
├── /space/:space_key  (serve space files)
├── /plugin/:space_key/:plugin_id
├── /api/space/:space_key/*subpath  (space API endpoints)
└── /static            (static assets)
```

## Security Model

### Frontend Isolation

**Iframe Sandboxing**: Each space's frontend runs in an isolated iframe with suborigin security:
- Apps are served from unique namespace URLs (`/zz/space/<space_key>/`)
- The platform UI loads apps in `<iframe>` elements
- iframes prevent direct DOM access between apps and platform
- Suborigin headers provide additional isolation (CSP policies)
- Apps communicate with backend via scoped API endpoints only
- Each app appears at its own URL path, preventing cross-app interference

**How Apps Are Served**:
1. User navigates to space (e.g., `/zz/space/my-todo/`)
2. Platform loads app's frontend files from `PackageFiles` (from `public/` folder)
3. App loads in isolated iframe in user's browser
4. App makes API calls to `/zz/api/space/my-todo/*` for backend functionality
5. Backend Lua code handles requests in isolated VM

### Backend Isolation

**Lua VM Sandboxing**: Server-side code runs in isolated Lua virtual machines:
- No direct filesystem access (only through UFS bindings)
- No direct network access (only through provided APIs)
- Memory and CPU limits per executor
- State pooling prevents resource exhaustion
- Each package has separate executor instances

### Resource Scoping

All resources are strictly scoped by Space ID:
- **Database**: `SpaceKV` entries filtered by `space_id`
- **Files**: `Files` table filtered by `owner_space_id`
- **Working Directory**: Separate folders per space
- **API Access**: Middleware validates space ownership/authorization

### Authentication & Authorization

- Passwords hashed with bcrypt
- Session tokens signed with Branca (HMAC + XChaCha20-Poly1305)
- Token-based access control for all API endpoints
- Space-level authorization via `SpaceUsers` table
- Users can only access spaces they own or are authorized for

## Key Concepts

### Spaces

A **Space** is an instance of an application. Each space:
- Is created from a **Package** (blueprint)
- Has a unique `namespace_key` for routing
- Runs in isolated executor (Lua VM)
- Has its own key-value store
- Has its own file storage
- Can be public or private
- Can have multiple authorized users

### Packages

A **Package** is an application blueprint containing:
- Metadata (name, version, author, license)
- Frontend files (HTML, CSS, JS) in `public/` folder
- Backend code (`server.lua` or other executor files)
- Configuration (`potato.json` or `potato.toml`)

**Package Structure**:
```
simple-todo/
├── potato.json          # Package metadata
├── public/             # Frontend assets
│   └── index.html
└── server.lua          # Backend code (optional)
```

**Package Metadata** (`potato.json`):
```json
{
    "name": "Simple Todo",
    "slug": "simple-todo",
    "info": "Todo list app",
    "version": "0.0.1",
    "artifacts": [
        {
            "namespace": "simple-todo",
            "kind": "space",
            "executor_type": "luaz",
            "server_file": "server.lua",
            "route_options": {
                "serve_folder": "public",
                "force_html_extension": false,
                "force_index_html_file": false
            }
        }
    ]
}
```

### Embedded Packages

Example packages are embedded in the binary at compile time:
- `backend/engine/epackages/simple-todo/`
- `backend/engine/epackages/simple-notes/`
- `backend/engine/epackages/simple-graph-notes/`

### Resource Scoping

Resources are organized into Core Platform and App-specific namespaces:

#### HTTP Routes:
- **Core**: `/zz/api/core/*`
- **App Assets**: `/zz/space/<space_key>/*`
- **App API**: `/zz/api/space/<space_key>/*`

#### Database Tables:
- **Core**: `Users`, `Spaces`, `Packages`, `Files`
- **App KV Store**: `SpaceKV` (scoped by `space_id`)

#### Working Directory:
- **Core**: `./tmp`
- **App**: `./tmp/<package_name>/<package_id>/`

## User & Authentication System

### User Types:
- `admin`: Full platform access
- `normal`: Standard user
- `bot`: Automated users
- `api`: API-only access

### User Groups:
- Configurable groups (e.g., "admin", "normal")
- Used for authorization

### Authentication Flow:
1. User logs in with email/password
2. Server creates `UserDevice` with session token
3. Token is signed using Branca (signer service)
4. Client includes token in subsequent requests
5. Server validates token via `withAccessTokenFn` middleware

### Authorization:
- Spaces can authorize specific users via `SpaceUsers` table
- Space-specific tokens can be generated
- Users can access their authorized spaces

## API Reference

### Core APIs

#### Authentication:
- `POST /zz/api/core/auth/login` - Login with email/password
- `GET /zz/api/core/auth/invite/:token` - Get invite info
- `POST /zz/api/core/auth/invite/:token` - Accept invite

#### User Management:
- `GET /zz/api/core/user/` - List users
- `GET /zz/api/core/user/:id` - Get user by ID
- `POST /zz/api/core/user/create` - Create user directly
- `GET /zz/api/core/self/info` - Get current user info

#### Package Management:
- `GET /zz/api/core/package/list` - List available packages
- `POST /zz/api/core/package/install/embed` - Install embedded package
- `POST /zz/api/core/package/install/zip` - Install from ZIP
- `DELETE /zz/api/core/package/:id` - Delete package

#### Space Management:
- `GET /zz/api/core/space/installed` - List user's spaces
- `POST /zz/api/core/space/authorize/:space_key` - Authorize space access

#### Space KV Store:
- `GET /zz/api/core/space/:id/kv` - List KV pairs
- `POST /zz/api/core/space/:id/kv` - Create KV pair
- `PUT /zz/api/core/space/:id/kv/:kvId` - Update KV pair
- `DELETE /zz/api/core/space/:id/kv/:kvId` - Delete KV pair

#### Space Files:
- `GET /zz/api/core/space/:id/files` - List files
- `POST /zz/api/core/space/:id/files/upload` - Upload file
- `GET /zz/api/core/space/:id/files/:fileId/download` - Download file
- `POST /zz/api/core/space/:id/files/folder` - Create folder

### Space APIs (User-Defined)

Apps can define custom API endpoints handled by their Lua code:
- `GET /zz/api/space/:space_key/*subpath` - Custom app endpoints

## Lua Bindings API

Potatoverse provides Lua modules that apps can use to interact with platform services. These modules are preloaded and available via `require()`.

### Available Modules

1. **kv** - Key-Value store (space-scoped database)
2. **ufs** - User File System (file operations)
3. **Standard Lua libraries** - math, string, table, etc.

### KV Store (Key-Value Database)

```lua
local kv = require("kv")

-- Add a key-value pair
kv.add({
    group = "settings",
    key = "theme",
    value = "dark",
    tag1 = "",  -- optional tags for filtering
    tag2 = "",
    tag3 = ""
})

-- Get a specific KV pair
local data, err = kv.get("settings", "theme")
if err then
    print("Error:", err)
else
    print("Value:", data.value)
end

-- Query KV pairs by conditions
local results = kv.query({
    group = "settings",
    key = "theme"
})

-- Get all KV pairs in a group (with pagination)
local items = kv.get_by_group("settings", 0, 100)  -- offset, limit

-- Update a KV pair
kv.update("settings", "theme", {
    value = "light"
})

-- Upsert (update or insert)
kv.upsert("settings", "theme", {
    value = "auto"
})

-- Remove a KV pair
kv.remove("settings", "theme")
```

**KV Module API Reference**:

| Function | Parameters | Returns | Description |
|----------|-----------|---------|-------------|
| `add(data)` | table: {group, key, value, tag1?, tag2?, tag3?} | nil or error | Add a new KV pair |
| `get(group, key)` | string, string | table or nil, error | Get a specific KV pair |
| `query(conditions)` | table: {group?, key?, tag1?, tag2?, tag3?} | array of tables | Query KV pairs matching conditions |
| `get_by_group(group, offset, limit)` | string, number, number | array of tables | Get all KV pairs in a group (paginated) |
| `update(group, key, data)` | string, string, table | nil or error | Update an existing KV pair |
| `upsert(group, key, data)` | string, string, table | nil or error | Insert or update KV pair |
| `remove(group, key)` | string, string | nil or error | Delete a KV pair |

**Notes**:
- All KV operations are automatically scoped to the current space
- `group` is used to organize related keys (like namespaces)
- `tag1`, `tag2`, `tag3` are optional fields for additional filtering
- KV pairs are stored in the `SpaceKV` database table

### User File System

```lua
local ufs = require("ufs")

-- Write file
ufs.write("data.txt", "Hello, World!")

-- Read file
local content = ufs.read("data.txt")

-- List files
local files = ufs.list("/")

-- Delete file
ufs.delete("data.txt")
```

### HTTP Context

```lua
function on_http(ctx)
    -- Get request object
    local req = ctx.request()
    
    -- Get parameters
    local space_id = ctx.param("space_id")
    local package_id = ctx.param("package_id")
    local subpath = ctx.param("subpath")
    
    -- Return JSON response
    req.json(200, {
        message = "Success",
        data = {}
    })
    
    -- Return text response
    req.text(200, "Hello, World!")
    
    -- Return HTML response
    req.html(200, "<h1>Hello</h1>")
end
```

## Database Schema

### Core Tables:

1. **Users**: User accounts (admin, normal, bot, api)
2. **UserGroups**: User permission groups
3. **UserInvites**: Invitation system
4. **UserDevices**: Session tokens and devices
5. **Spaces**: App instances
6. **SpaceKV**: Key-value store per space
7. **SpaceUsers**: Space authorization
8. **Packages**: Installed app blueprints
9. **PackageFiles**: Package file storage (with blob support)
10. **Files**: User/space file storage

### Key Relationships:
- `Users` ←→ `UserGroups` (via `ugroup`)
- `Spaces` → `Packages` (via `package_id`)
- `Spaces` → `Users` (via `owned_by`)
- `SpaceUsers` → `Spaces` + `Users` (many-to-many)
- `SpaceKV` → `Spaces` (via `space_id`)

## Frontend Architecture

### Tech Stack:
- **Next.js** (React framework)
- **TypeScript**
- **TailwindCSS** (styling)

### Key Pages:
- `/auth/login` - Login page
- `/auth/signup/open` - Open signup
- `/auth/signup/invite-finish` - Complete invite signup
- `/portal/admin` - Admin dashboard
- `/portal/admin/users` - User management
- `/portal/admin/spaces` - Space management
- `/portal/admin/store` - Package store
- `/portal/main` - User portal

### Key Components:
- `FantasticTable`: Reusable data table component
- `GModalWrapper`: Global modal system
- `HomePage`: Landing page with hero section

### API Integration (`lib/api.ts`):
- Centralized API client
- Handles authentication tokens
- Error handling

### State Management (`hooks/`):
- `useGAppState`: Global app state
- `useGModal`: Modal state management
- `useSimpleDataLoader`: Data fetching hook

## Development Workflow

### Building the Project:

```bash
# Build backend
go build -o potatoverse main.go

# Run with seed data
./potatoverse --seed

# Frontend development
cd frontend
npm install
npm run dev
```

### Project Structure:

```
potatoverse/
├── backend/              # Go backend
│   ├── app/             # Application layer
│   ├── engine/          # App runtime engine
│   ├── services/        # Database, mailer, signer
│   ├── spaces/          # Space utilities
│   └── utils/           # Helper functions
├── frontend/            # Next.js frontend
│   ├── app/            # Next.js app directory
│   ├── contain/        # React components
│   ├── hooks/          # React hooks
│   └── lib/            # Utilities and API
├── cmd/                # CLI commands
├── tests/              # Test files
├── data.db             # SQLite database
└── tmp/                # Working directory
```

### Key Configuration Files:

- `go.mod`: Go dependencies
- `package.json`: Node dependencies
- `justfile`: Task automation (just commands)
- `ci.sh`: CI/CD script

## Testing

Test files location: `tests/`
- `tests.go`: Main test utilities
- `luaz_ufstest.go`: Lua UFS binding tests
- `ehandle_ufs.go`: Executor handle tests

## Important Implementation Details

### Lua State Pooling:
- Each package has its own Lua executor with a pool of states
- Pool size: 10-20 states
- Max concurrent requests: 50
- TTL: 1 hour (states are cleaned up periodically)
- States are reused to avoid initialization overhead

### File Storage:
- Files stored in database as BLOBs
- Support for multi-part blobs (large files)
- Three storage types: inline_blob, external_blob, multi_part_blob

### Package Storage:
- Packages can be embedded in binary
- Can be installed from ZIP files
- Files extracted to `PackageFiles` table

### Security Considerations:
- Passwords hashed with bcrypt
- Session tokens signed with Branca
- Lua sandbox isolation (no direct file system access)
- User-scoped resources

## Future Plans (from docs/future.md)

Planned features (may not be implemented yet):
- WebSocket support (`/ws/broadcast`, `/ws/room/<room_id>`)
- Plugin system
- WebAssembly executor support
- CDC (Change Data Capture) for sync/backup
- Model Context Protocol (MCP) integration
- P2P features

## Common Tasks for Agents

### Creating a New Package:

1. Create directory in `backend/engine/epackages/my-app/`
2. Add `potato.json` with metadata
3. Add frontend files in `public/`
4. Add backend code in `server.lua`
5. Install via API or embed flag

### Adding a New API Endpoint:

1. Add route in `backend/app/server/routes.go`
2. Create handler in appropriate `rt_*.go` file
3. Add controller method in `backend/app/actions/`
4. Test via frontend or API client

### Adding a New Lua Binding:

1. Create binding file in `backend/engine/executors/luaz/binds/`
2. Implement Go functions
3. Register in `luah.go` via `registerModules()`
4. Document usage

### Debugging Tips:

1. Check logs for executor errors
2. Use `/zz/api/core/engine/debug` for engine state
3. Use `pp.Println()` for detailed debugging (k0kubun/pp)
4. Check Lua state pool status in debug data

## Module Dependencies

### Go Modules:
- `gin-gonic/gin`: HTTP framework
- `mattn/go-sqlite3`: SQLite driver
- `upper/db/v4`: Database ORM
- `yuin/gopher-lua`: Lua VM
- `hako/branca`: Token signing
- `nfnt/resize`: Image processing

### Frontend Dependencies:
- `next`: Next.js framework
- `react`: React library
- `tailwindcss`: CSS framework
- `typescript`: Type checking

## Terminology

- **Potatoverse**: The platform name (formerly "Turnix")
- **Space**: An instance of an application
- **Package**: A blueprint/template for creating spaces
- **Namespace Key**: Unique identifier for routing to a space
- **Executor**: The runtime environment (Lua VM) for a space
- **Artifact**: A deployable component defined in a package
- **KV Store**: Key-Value storage system per space
- **UFS**: User File System (file storage bindings)
- **HeadLess**: Backend without HTTP server
- **Portal**: Admin/user dashboard interface
- **Potato**: Project name (potatoverse = universe of app platforms)

**Note**: The codebase still contains many references to "Turnix" (especially in module paths like `github.com/blue-monads/turnix`). This is legacy from the rename and both names refer to the same project.

## Contact & Contribution

- Project: Potatoverse (formerly Turnix, blue-monads organization)
- Module: `github.com/blue-monads/turnix` (legacy module name, directory is `potatoverse`)
- Language: Go 1.24+
- Frontend: Next.js

---

**This documentation is intended for AI agents and developers to understand the Potatoverse platform architecture, components, and development workflow.**

