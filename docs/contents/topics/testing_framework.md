# Testing Framework & Developer Shell

Potatoverse includes a built-in API-level testing framework and developer shell tooling designed for rapid, automated testing of applications (spaces). It utilizes the embedded Lua VM (`gopher-lua`), preloaded HTTP and JSON modules, and local UNIX socket communication to automatically provision access tokens and test against running server instances.

---

## Commands Overview

The testing and developer shell tools are exposed under the `potatoverse dev` CLI:

| Command | Description |
|---|---|
| `potatoverse dev tests <lua_file>` | Runs API tests against an **already running** development server. |
| `potatoverse dev shell [<command> ...]` | Executes a shell command (or opens an interactive shell) with exposed environment variables against an **already running** server. |
| `potatoverse dev run-and-tests <lua_file>` | Starts the dev server, pushes the current app, waits, runs Lua tests, and terminates the server when done. |
| `potatoverse dev run-and-shell [<command> ...]` | Starts the dev server, pushes the current app, waits, executes a shell command with exposed environment variables, and terminates the server when done. |

---

## 1. Running Tests Against Existing Server

### `potatoverse dev tests <lua_file>`

Connects to a running Potatoverse development server via its local UNIX socket (`.pdata/potatoverse.sock`), automatically acquires an admin token and a space token for the target space, sets up the test execution context, and runs `on_test_run(ctx)` inside the specified Lua file.

```bash
# Run tests on default/root space
potatoverse dev tests tests/api_test.lua

# Target a specific space
potatoverse dev tests tests/api_test.lua --use-space=my-custom-space

# Specify custom working directory or potato.yaml
potatoverse dev tests tests/api_test.lua --working-dir=./.pdata --potato-yaml-file=./potato.yaml
```

---

## 2. All-in-One: Run Server and Tests

### `potatoverse dev run-and-tests <lua_file>`

Starts the local development server in the background, pushes the current package (auto-building if needed), waits for the server to be ready (default `10s`), runs the test file, and terminates the dev server upon completion.

```bash
# Start server, wait 10s, run tests, and exit
potatoverse dev run-and-tests tests/api_test.lua

# Custom delay before tests run
potatoverse dev run-and-tests tests/api_test.lua --delay=5s

# Keep server running after tests finish
potatoverse dev run-and-tests tests/api_test.lua --keep-server
```

---

## 3. Developer Shell & Environment Variables

### `potatoverse dev shell [<shell_command> ...]`

Connects to a running server and injects authenticated environment variables into the subshell or command execution. If no command is provided, an interactive shell session is opened.

```bash
# Inspect environment tokens
potatoverse dev shell "echo Server: $POTATO_SERVER_URL Token: $POTATO_DEV_SPACE_TOKEN"

# Make authenticated curl requests
potatoverse dev shell "curl -H 'Authorization: Bearer $POTATO_DEV_SPACE_TOKEN' $POTATO_SERVER_URL/zz/space/cimple-books/api/items"

# Run tests written in other languages/frameworks (Node, Bun, Python, etc.)
potatoverse dev shell "bun test"

# Launch an interactive shell with environment tokens loaded
potatoverse dev shell
```

### `potatoverse dev run-and-shell [<shell_command> ...]`

Starts the development server, pushes the app, waits (default `10s`), injects the environment variables, executes the command, and cleans up the server upon exit.

```bash
potatoverse dev run-and-shell "curl -s -H 'Authorization: Bearer $POTATO_DEV_SPACE_TOKEN' $POTATO_SERVER_URL/zz/space/cimple-books/api/health"
```

### Exposed Environment Variables

The shell commands expose the following environment variables:

| Variable | Description |
|---|---|
| `POTATO_SERVER_URL` | Base URL of the development server (e.g. `http://localhost:7777`). |
| `POTATO_DEV_SPACE_TOKEN` | Bearer token for the root space (or space specified by `--use-space`). |
| `POTATO_DEV_ADMIN_USER_TOKEN` | Admin access token for core management API (`/zz/api/core/*`). |
| `POTATO_DEV_NAMESPACE_KEY` | Namespace key of the target space. |
| `POTATO_DEV_SPACE_ID` | Numeric space ID of the target space. |

---

## 4. Writing Lua API Tests

Tests are written in Lua files. The runner automatically executes the script and invokes the `on_test_run(ctx)` global function with a test context object (`ctx`).

### Test Script Lifecycle

```lua
local gluahttp = require("gluahttp")
local ptest = require("potato-test")

function on_test_run(ctx)
    -- Your tests run here!
    -- Use standard Lua assertions (assert)
    local res, err = ctx.rootSpace().get("/zz/space/" .. ctx.namespace_key .. "/items")
    assert(err == nil, "Request failed: " .. tostring(err))
    assert(res.status_code == 200, "Expected status 200, got: " .. tostring(res.status_code))
end
```

If `assert` fails or any error is thrown during `on_test_run`, the command exits with non-zero exit code (suitable for CI/CD).

---

## 5. Lua Test Context (`ctx`) API

The `ctx` object passed to `on_test_run(ctx)` provides the following properties and methods:

### Properties

- `ctx.token`: Space token string for the root space.
- `ctx.space_token`: Alias for `ctx.token`.
- `ctx.admin_token`: Admin user token string.
- `ctx.server_url`: Base URL of the server (e.g. `http://localhost:7777`).
- `ctx.namespace_key`: Namespace key of the root space.
- `ctx.space_id`: Numeric space ID.

### Methods

#### `ctx.rootSpace()` / `ctx.root_space()`
Returns an HTTP client configured with the root space's token and namespace key.

```lua
local root = ctx.rootSpace()
```

#### `ctx.createSpaceHttp(options)` / `ctx.create_space_http(options)`
Dynamically resolves and authorizes another space, returning a space HTTP client configured with that space's token:

```lua
local althttp, err = ctx.createSpaceHttp({
    namespace_key = "another-app-namespace"
})
assert(err == nil, tostring(err))
```

#### `ctx.getSpaceToken(options)` / `ctx.get_space_token(options)`
Returns the raw space token string for another space:

```lua
local token, err = ctx.getSpaceToken({
    namespace_key = "another-app-namespace"
})
```

---

## 6. Space HTTP Client API

The client returned by `ctx.rootSpace()` or `ctx.createSpaceHttp(...)` provides helper methods for making HTTP requests:

### HTTP Methods
- `client.get(path, [options])`
- `client.post(path, [options])`
- `client.put(path, [options])`
- `client.delete(path, [options])`
- `client.patch(path, [options])`
- `client.head(path, [options])`
- `client.options(path, [options])`
- `client.request(method, path, [options])`

*(Both dot syntax `client.post(...)` and colon syntax `client:post(...)` are supported).*

### Path Resolution
- **Absolute Paths**: E.g. `"/zz/space/my-app/items"` are automatically prepended with `ctx.server_url`.
- **Space-Relative Paths**: E.g. `"items"` are automatically prefixed with `/zz/space/<namespace_key>/`.
- **Full URLs**: E.g. `"http://localhost:7777/..."` are used as-is.

### Request Options Table
The options table can contain:
- `payload`: Lua table or string. Tables are automatically serialized to JSON with `Content-Type: application/json`.
- `json`: Table or string serialized as JSON.
- `body`: Raw body string.
- `form`: Key-value table URL-encoded as `application/x-www-form-urlencoded`.
- `headers`: Key-value table of custom HTTP headers. By default, `Authorization: Bearer <space_token>` is added automatically.
- `query`: Key-value table or query string appended to URL.
- `cookies`: Key-value table of HTTP cookies.
- `timeout`: Duration in seconds (e.g. `10`) or string (`"5s"`).

### Response Object
Requests return `(res, err)` where `res` contains:
- `res.status_code`: Numeric HTTP status code (e.g. `200`).
- `res.status`: Status text (e.g. `"200 OK"`).
- `res.ok`: Boolean (`true` if status code is in 200–299 range).
- `res.body`: Raw response body string.
- `res.headers`: Table of response headers.
- `res.cookies`: Table of response cookies.
- `res.data`: Automatically parsed Lua representation if the body is JSON.
- `res.json()` / `res:json()`: Method to decode response body as JSON, returning `(data, err)`.

---

## 7. Preloaded Lua Modules

Every test runner Lua state comes preloaded with standard libraries and:
- `gluahttp` (also aliased as `phttp` and `http`): Standard `gluahttp` HTTP client.
- `json`: Fast JSON encoder and decoder (`layeh.com/gopher-json`).
- `potato-test`: Preloaded module exposing `ctx` and testing helpers.

---

## 8. Complete Test Example

```lua
-- tests/app_api_test.lua

function on_test_run(ctx)
    local root = ctx.rootSpace()

    -- 1. Test POST request with JSON payload
    local create_res, err = root.post("/zz/space/" .. ctx.namespace_key .. "/api/books", {
        payload = {
            title = "The Great Potato",
            author = "Bornjre",
            year = 2026
        }
    })
    assert(err == nil, "POST failed: " .. tostring(err))
    assert(create_res.status_code == 201 or create_res.status_code == 200, "Unexpected status: " .. create_res.status_code)

    local book = create_res.json()
    assert(book ~= nil, "Expected JSON response")
    assert(book.title == "The Great Potato", "Book title mismatch")

    -- 2. Test GET request
    local get_res, gerr = root.get("/zz/space/" .. ctx.namespace_key .. "/api/books/" .. tostring(book.id))
    assert(gerr == nil, "GET failed: " .. tostring(gerr))
    assert(get_res.status_code == 200, "Expected 200 OK")
    assert(get_res.data.id == book.id, "Book ID mismatch")

    -- 3. Test interacting with a secondary space
    local auth_client, aerr = ctx.createSpaceHttp({
        namespace_key = "auth-space"
    })
    if aerr == nil and auth_client ~= nil then
        local user_res, uerr = auth_client.get("/zz/space/auth-space/api/me")
        assert(uerr == nil, "Cross-space request failed")
        assert(user_res.ok, "Expected ok response from auth space")
    end

    print("All tests passed successfully!")
end
```
