---
name: potatoverse_app
description: Standard workflow for generating a Potatoverse app with migrations and seeding.
---

# Potatoverse App Generation Skill

Use this skill to scaffold a new Potatoverse app. This workflow ensures the app follows the latest architectural patterns, including SQL migrations and JSON seeding.

## Prerequisites

- Read the documentation in `./docs/` for API and Lua bindings reference.
- Use the templates in `./templates/` as starting points.

## Scaffolding Workflow

### 1. Project Initialization
Create the following directory structure:
```text
your-app/
├── server/
│   ├── migration/    # .sql files
│   └── seed/         # .json files
├── ui/               # Frontend source (Vanilla or React)
├── public/           # Static assets (if vanilla)
├── potato.yaml
├── server.lua        # (actually saved as server/server.lua but mapped to server.lua)
└── justfile
```

### 2. Configure `potato.yaml`
Use `templates/potato.yaml.tmpl`. Ensure the following:
- `slug` is unique.
- `capabilities` include `xMigrator` (folder: `migration`) and `xStaticSeeder` (folder: `seed`).
- `developer.include_files` correctly maps local files to the package structure.

### 3. Implement `server.lua`
Use `templates/server.lua.tmpl`.
- Implement `on_http(ctx)` to handle routing.
- Ensure `/setup` endpoint calls `xMigrator` and `xStaticSeeder`.
- Use `req.get_user_id()` for authentication.
- Use `potato.db` for database operations.

### 4. Database Schema
- Place initial schema in `server/migration/0001_initial.sql`.
- Subsequent changes go in `0002_xxx.sql`, etc.

### 5. Static Seeding
- Place initial data in `server/seed/0001_xxx.json`.
- The format should be matching the table names (e.g., `{"Authors": [...]}`).

### 6. UI Implementation
- **Vanilla**: Place files in `ui/`. Map `ui/**/*` to `public` in `potato.yaml`.
- **React**: Initialize a Vite/React app in `ui/`. Ensure `build_command` is set in `potato.yaml` and maps `ui/dist/**/*` to `public`.

## Reference Documentation
- [Lua Bindings](./docs/luaz.md)
- [Space API](./docs/space.md)
- [Auth API](./docs/auth.md)
- [User API](./docs/user.md)
- [Package API](./docs/package.md)

## Example Scaffolding Command
```sh
# (To be used by AI during execution)
# 1. mkdir app && cd app
# 2. cp {SKILL_PATH}/templates/* .
# 3. Rename and fill placeholders
```
