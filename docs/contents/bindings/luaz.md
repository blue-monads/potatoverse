# Lua Bindings

## potato

Main module providing access to system functionality.

### potato.kv

Key-value storage operations.

- `query(opts)` - Query KV entries. `opts`: `{group, cond, offset, limit, include_value}`
- `add(data)` - Add new KV entry
- `get(group, key)` - Get KV entry
- `get_by_group(group, offset, limit)` - Get entries by group
- `remove(group, key)` - Remove KV entry
- `update(group, key, data)` - Update KV entry
- `upsert(group, key, data)` - Upsert KV entry

### potato.cap

Capability operations.

- `list()` - List available capabilities
- `execute(name, method, params)` - Execute capability method
- `methods(name)` - List methods for capability
- `sign_token(name, opts)` - Sign capability token. `opts`: `{resource_id, extrameta, user_id}`

### potato.db

Database operations.

- `run_ddl(sql)` - Execute DDL statement
- `run_query(sql, ...args)` - Execute query, returns array of rows
- `run_query_one(sql, ...args)` - Execute query, returns single row
- `insert(table, data)` - Insert record, returns ID
- `update_by_id(table, id, data)` - Update by ID
- `delete_by_id(table, id)` - Delete by ID
- `find_by_id(table, id)` - Find by ID
- `update_by_cond(table, cond, data)` - Update by condition
- `delete_by_cond(table, cond)` - Delete by condition
- `find_all_by_cond(table, cond)` - Find all by condition
- `find_one_by_cond(table, cond)` - Find one by condition
- `find_all_by_query(query)` - Find all by query object
- `find_by_join(query)` - Find by join query
- `list_tables()` - List all tables
- `list_columns(table)` - List table columns
- `start_txn()` - Start transaction, returns `potato.txn` object

### potato.txn

Transaction operations (same methods as `potato.db` plus):

- `commit()` - Commit transaction
- `rollback()` - Rollback transaction

### potato.core

Core system operations.

- `publish_event(opts)` - Publish event. `opts`: `{name, payload, resource_id}`
- `file_token(opts)` - Generate file presigned token. `opts`: `{path, file_name, user_id}`
- `advisery_token(opts)` - Generate advisery token. `opts`: `{token_sub_type, user_id, data}`
- `read_package_file(fpath)` - read package file contents

## http.request

HTTP request context object.

### Request methods

- `param(key)` - Get URL parameter
- `get_query(key)` - Get query param, returns `(value, exists)`
- `get_post_form(key)` - Get POST form param, returns `(value, exists)`
- `default_query(key, default)` - Get query with default
- `default_post_form(key, default)` - Get POST form with default
- `get_query_map(key)` - Get query map, returns `(table, exists)`
- `get_query_array(key)` - Get query array, returns `(table, exists)`
- `get_post_form_map(key)` - Get POST form map, returns `(table, exists)`
- `get_post_form_array(key)` - Get POST form array, returns `(table, exists)`
- `get_header(key)` - Get request header
- `cookie(name)` - Get cookie, returns `(value, error)`
- `content_type()` - Get content type
- `client_ip()` - Get client IP
- `remote_ip()` - Get remote IP
- `full_path()` - Get full path
- `get_raw_data()` - Get raw request body
- `form_file(name)` - Get uploaded file, returns `{filename, size}`
- `bind_json()` - Bind JSON body to table
- `bind_header()` - Bind headers to table
- `bind_query()` - Bind query params to table

### Response methods

- `status(code)` - Set status code
- `header(key, value)` - Set response header
- `json(code, data)` - Send JSON response
- `json_array(code, data)` - Send JSON array response
- `html(code, template, data)` - Render HTML template
- `string(code, format, ...)` - Send formatted string
- `data(code, content_type, data)` - Send raw data
- `redirect(code, location)` - Redirect
- `set_cookie(name, value, max_age, path, domain, secure, http_only)` - Set cookie
- `sse_event(name, message)` - Send SSE event

### Abort methods

- `abort()` - Abort request
- `abort_with_status(code)` - Abort with status
- `abort_with_status_json(code, data)` - Abort with JSON

### Auth methods

- `get_claim()` - Get space claim, returns `(claim, error)`
- `get_user_id()` - Get user ID, returns `(user_id, error)`

### State methods

- `state_keys()` - Get all state keys
- `state_get(key)` - Get state value
- `state_set(key, value)` - Set state value
- `state_set_all(data)` - Set all state from table

## generic.context

Generic request context.

- `list_actions()` - List available actions
- `execute_action(name, params)` - Execute action with params

## mcp

Model Context Protocol client.

- `create_http_client(endpoint, name, transport_type?)` - Create MCP client. Returns client object with:
  - `list_tools(params)` - List available tools
  - `list_resources(params)` - List available resources
  - `call_tool(params)` - Call tool with params

