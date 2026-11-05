# Plugin System

The Potatoverse Plugin System allows applications (`App` spaces) to be dynamically extended by independent plugin packages (`AppPlugin` spaces). Plugins can inject client-side UI components and execute authenticated server-side logic in their own isolated execution environments, while safely operating on the host space via cross-space capabilities.

---

## Architecture Overview

```
┌────────────────────────────────────────────────────────────────────────┐
│ Host Space (App)                                                       │
│                                                                        │
│  Browser / Client UI                                                   │
│   ├── Host Shell (index.html)                                          │
│   └── <script src="/zz/core/space/:space_key/plugin_loaders.js">       │
│         └── Injects Plugin UI (from loader.js)                         │
│                                                                        │
│  Host Backend API: /zz/api/space/:space_key/*                          │
│  Host Static Files: /zz/space/:space_key/*                             │
└────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼ Calls Plugin API
┌────────────────────────────────────────────────────────────────────────┐
│ Plugin Space (AppPlugin)                                               │
│                                                                        │
│  Plugin API Route: /zz/api/plugin/:host_space_key/:plugin_key/*        │
│                                                                        │
│  Engine Runtime:                                                       │
│   ├── Mints cryptographically signed remote_ctx_token                  │
│   └── Executes Plugin server.lua                                       │
│                                                                        │
│  Plugin server.lua:                                                    │
│   ├── req.get_claim(remote_ctx_token)     (validates host user claim)  │
│   ├── req.get_user_id(remote_ctx_token)   (gets host user ID)          │
│   └── potato.cap.execute("xRemote", ...)  (operates on host space)     │
└────────────────────────────────────────────────────────────────────────┘
```

The system operates across three interconnected layers:

1. **Host Application (`App`)**:
   Hosts declare DOM mount points (or a JavaScript registration hook) and include a single script tag pointing to `/zz/core/space/:space_key/plugin_loaders.js`.

2. **Engine Bundler & Router**:
   - Concatenates the `loader_script` from all active attached `AppPlugin` spaces into a single bundle served at `/zz/core/space/:space_key/plugin_loaders.js`.
   - Intercepts calls to `/zz/api/plugin/:space_key/:plugin_key/*subpath`, mints a signed `remote_ctx_token`, and routes execution to the target plugin's server runtime.

3. **Plugin Space (`AppPlugin`)**:
   - Ships a client-side `loader.js` to render UI inside the host frontend.
   - Runs server-side code (e.g. `server.lua`) that validates user claims against the host space identity and performs cross-space operations via `xRemote`.

---

## Plugin Package Configuration

An AppPlugin is defined in its manifest (`potato.json` or `potato.yaml`) by setting `space_type` to `"AppPlugin"`, specifying a `loader_script`, and declaring the `xRemote` capability:

```json
{
    "name": "My Plugin",
    "slug": "my-plugin",
    "info": "Extends host apps with custom widgets and features",
    "spaces": [
        {
            "namespace": "my-plugin",
            "space_type": "AppPlugin",
            "executor_type": "luaz",
            "server_file": "server.lua",
            "loader_script": "loader.js",
            "is_default": true
        }
    ],
    "capabilities": [
        "xRemote"
    ],
    "version": "0.0.1"
}
```

### Key Fields

- **`space_type`**: Must be `"AppPlugin"`.
- **`loader_script`**: Relative path to the JavaScript file bundled into `/zz/core/space/:space_key/plugin_loaders.js`.
- **`server_file`**: Entry point for server-side execution when `/zz/api/plugin/...` is called.
- **`capabilities`**: Declare `"xRemote"` to allow cross-space operations back into the host space.

---

## Client-Side Loader Script (`loader.js`)

The `loader_script` is bundled by the engine and runs directly within the host space's browser context. It can:
- Inspect host DOM elements (e.g., `#plugin-slots`).
- Interact with host registries (e.g., `window.HostApp.registerPlugin(...)`).
- Make authenticated API calls to the plugin's backend endpoints.

### Example `loader.js`

```javascript
(function() {
  console.log("[My Plugin] Initializing inside host space...");

  function mountWidget() {
    const container = document.getElementById("plugin-slots");
    if (!container) return;

    const widget = document.createElement("div");
    widget.className = "plugin-card";
    widget.innerHTML = `
      <h3>🔌 My Plugin Widget</h3>
      <button id="my-plugin-action-btn">Run Plugin Action</button>
      <pre id="my-plugin-output"></pre>
    `;
    container.appendChild(widget);

    // Derive host space key from URL or global host config
    const hostSpaceKey = window.HostApp?.spaceKey || "host-app";

    document.getElementById("my-plugin-action-btn").addEventListener("click", async () => {
      const out = document.getElementById("my-plugin-output");
      out.textContent = "Loading...";

      // Call the plugin backend endpoint via the engine's plugin API route
      const res = await fetch(`/zz/api/plugin/${hostSpaceKey}/my-plugin/action`);
      const data = await res.json();
      out.textContent = JSON.stringify(data, null, 2);
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", mountWidget);
  } else {
    mountWidget();
  }
})();
```

---

## Server-Side Plugin Execution (`server.lua`)

When an HTTP request is sent to:
```http
/zz/api/plugin/:host_space_key/:plugin_key/*subpath
```

The Engine automatically:
1. Resolves `:host_space_key` to the target host space.
2. Verifies that `:plugin_key` is attached to that host space.
3. Mints a secure, cryptographically signed `remote_ctx_token`.
4. Executes the plugin's `server.lua` with the request context and token parameters.

### Security Model & `remote_ctx_token`

The `remote_ctx_token` encodes:
- **`TargetSpaceId`**: The host space ID.
- **`TargetPackageId`**: The host package ID.
- **`TargetPackageVersionId`**: The active host package version ID.
- **`PluginSpaceId`**: The plugin space ID this token was issued to.

Because user sessions authenticate against the **host** space (not the plugin space), the plugin backend must validate incoming claims against the host space identity using `remote_ctx_token`.

### Claim Validation Functions

- **`req.get_claim(remote_ctx_token)`**:
  - Validates that `remote_ctx_token` was signed for this plugin.
  - Verifies the user's incoming JWT token against the target host space.
  - Returns the user's claim table (`user_id`, `space_id`, `install_id`, etc.) or `nil, err`.
  - Passing no arguments strictly checks against the current space (no implicit cross-space elevation). Passing raw numeric IDs is strictly rejected.

- **`req.get_user_id(remote_ctx_token)`**:
  - Convenience helper returning the verified numeric user ID for the claim in the target space context.

### Example `server.lua`

```lua
function on_http(ctx)
    local req = ctx.request()
    local subpath = ctx.param("subpath") or ""
    local remote_token = ctx.param("remote_ctx_token")

    -- Normalize subpath
    if subpath:sub(1, 1) == "/" then
        subpath = subpath:sub(2)
    end

    -- Authenticate user claim in the context of the host space
    local claim, err = req.get_claim(remote_token)
    if err or not claim then
        req.json(401, { error = "Unauthorized: invalid user claim for host space" })
        return
    end

    if subpath == "action" then
        req.json(200, {
            status = "ok",
            message = "Action executed successfully",
            user_id = claim.user_id,
            host_space_id = claim.space_id
        })
        return
    end

    if subpath == "host_data" then
        -- Execute cross-space action via xRemote capability
        local res, err = potato.cap.execute("xRemote", "query_db", {
            remote_ctx_token = remote_token,
            query = "SELECT * FROM items LIMIT 10"
        })

        if err then
            req.json(500, { error = tostring(err) })
            return
        end

        req.json(200, { data = res })
        return
    end

    req.json(404, { error = "Endpoint not found" })
end
```

---

## Cross-Space Operations via `xRemote`

To maintain strict isolation and auditability:
- Plugins cannot access host databases, key-value stores, or files directly.
- Direct passing of raw space IDs to database or binding calls is rejected.
- All cross-space read/write operations must go through the **`xRemote`** capability using the verified `remote_ctx_token`:

```lua
local res, err = potato.cap.execute("xRemote", "<action_name>", {
    remote_ctx_token = remote_token,
    ...
})
```

---

## Host Space Integration

To enable plugins in a host application (`App`):

### 1. Add Mount Points in UI

In your HTML template or component tree (e.g. `public/index.html`):

```html
<div class="main-layout">
  <div class="content">...</div>

  <!-- Dedicated slot for plugged extensions -->
  <div id="plugin-slots"></div>
</div>
```

### 2. Include the Plugin Loader Script

Add the dynamic script tag at the bottom of the page:

```html
<script src="/zz/core/space/:space_key/plugin_loaders.js"></script>
```

Replace `:space_key` with the space's namespace key (e.g., `pexample-alpha`). The engine dynamically responds with the concatenated scripts of all attached plugins. If no plugins are plugged, it returns `// no plugins loaded`.

---

## Space Plugin REST API

Manage plugin connections via the Core API endpoints:

### List Plugins Attached to a Space
```http
GET /zz/api/core/space/:install_id/plugins?space_id=:space_id
```
Returns an array of attached plugin records (`SpacePlugin`).

### List Available AppPlugins
```http
GET /zz/api/core/space/:install_id/plugins/available?space_id=:space_id
```
Returns all installed spaces of type `AppPlugin` eligible to be attached to this host space.

### Get Attached Plugin
```http
GET /zz/api/core/space/:install_id/plugins/:pluginId
```
Returns the specific `SpacePlugin` record.

### Attach Plugin to Space
```http
POST /zz/api/core/space/:install_id/plugins?space_id=:space_id
Content-Type: application/json

{
  "target_space_id": 42,
  "priority": 10,
  "active": true,
  "init_order": 1,
  "extra_meta": {}
}
```

### Update Attached Plugin
```http
PUT /zz/api/core/space/:install_id/plugins/:pluginId
Content-Type: application/json

{
  "priority": 20,
  "active": true
}
```

### Detach Plugin from Space
```http
DELETE /zz/api/core/space/:install_id/plugins/:pluginId
```
Unplugs the plugin from the host space.

---

## Reference Examples

Working examples are provided in the embedded packages:
- **`pexample-alpha`** (`backend/engine/hubs/repohub/devrepo/epackages/pexample-alpha`):
  Sample host application with `#plugin-slots` and the loader script tag.
- **`pexample-beta`** (`backend/engine/hubs/repohub/devrepo/epackages/pexample-beta`):
  Sample `AppPlugin` with client-side DOM injection (`loader.js`) and authenticated backend handlers (`server.lua`).
