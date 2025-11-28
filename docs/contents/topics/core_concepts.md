## Core Concepts

### Platform Overview

PotatoVerse is a small app platform that hosts web applications with server-side code. It provides a batteries-included platform for building and hosting apps, combining features of a CMS and a Heroku-like PaaS in a single binary.

### Authentication

Authentication uses an OAuth-like system for apps and their users (not external services). Users authenticate to access spaces and platform resources.

### Database

Data is stored in SQLite database, where each space app is isolated from each other.

### Engine

The App Engine is the core component responsible for running applications. It manages:
- Space lifecycle
- Request routing
- Resource isolation
- Runtime execution

### Spaces

Spaces are applications that run inside language VMs like Lua or WebAssembly (future) . Each space:
- Is created from a package blueprint
- Has a namespace key (slugified name) used for resource scoping
- Is served from HTTP routes based on its namespace key
- Has its own file storage scoped by space ID
- Can access platform services through bindings

### Packages

Packages are blueprints that define spaces. They contain:
- Metadata (name, version, author, etc.)
- Artifacts (code, assets, configuration)
- Entry points for HTTP requests
- Initialization and update pages

Packages can be installed from:
- Embedded packages
- ZIP files
- Remote repositories

### Capabilities

Capabilities are extensible services that provide functionality to spaces. They:
- Can be configured per space
- Expose actions/methods that spaces can call

### Resource Scoping

Resources are divided into two scopes:

**Core Platform:**
- Core database tables (`core_*`)
- Platform HTTP routes (`/zz/api/core/*`)
- Platform working directories

**App Namespace:**
- Space-specific tables (`z_space_<space_key>_*`)
- Space HTTP routes (`/zz/space/<space_key>/*`, `/zz/api/space/<space_key>/*`)
- Space working directories (`space_wd/<space_key>`)

### Executors

Spaces run in executor environments:
- **Luaz**: Lua-based executor with bindings for platform services
- **WebAssembly**: WASM-based executor (planned)

Executors provide bindings for:
- Database operations
- File storage
- KV storage
- HTTP requests
- Capabilities
- Platform APIs
