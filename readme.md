# PotatoVerse 🥔

> ⚠️ **Alpha Software**: This is early alpha software being developed in spare weekend time. Expect bugs, breaking changes, and incomplete features.


[DEMO ](https://tubersalltheway.top/zz/pages) 


PotatoVerse is a small app platform that hosts web applications with server-side code. Think of it as a hybrid between a CMS and Heroku-like PaaS, all in one binary.

https://github.com/user-attachments/assets/120b826e-c5f9-4e1f-829e-816e7d2982ea

## Features

- 🚀 **Spaces**: Isolated app environments running in Lua VMs (WASM planned)
- 📦 **Packages**: Blueprint-based app deployment from embedded/ZIP/remote sources
- 🔐 **Auth**: OAuth-like authentication for users and apps
- 💾 **Database**: SQLite with per-space isolation
- 🛠️ **Capabilities**: Extensible services (files, websockets, users, etc.)
- 🎨 **Frontend**: Next.js/React admin portal
- ⚡ **CLI**: Package management, server operations, backups

## Future features
- [ ] Buddy backup (backup your maindb.sqlite file through litestream and webdav; there's a POC in code but not integrated)
- [ ] WASM executor (current lua runtime is much easier for testing APIs and ideas)

## Terminologies

- **Spaces**: Apps created from packages, isolated by namespace. Frontend can also run on its own suborigin for better isolation.
- **Engine**: Manages space lifecycle, routing, and execution
- **Executor**: Executor is responsible for running server code (Lua VM or WASM in the future). It's an interface which is registered similar to how SQL drivers are registered, so you can bring your own or write apps in native Go code too.
- **Capabilities**: Platform services exposed to spaces
- **Packages**: Blueprints containing spaces (apps), code, and assets. Imagine if we had an SPK (server package file) similar to APK for Android apps. A simple example would contain the following in a zip:
    - `potato.json` (manifest file)
    - `public/{index.html, style.css, client.js}` (folder served to users running the space)
    - `server.lua`

# Developing
[How to develop](./docs/how_to_develop.md)
