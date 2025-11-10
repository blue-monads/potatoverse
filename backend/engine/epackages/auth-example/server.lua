
function get_tasks(ctx)
    local kv = require("pkv")
    local req = ctx.request()

    local data, err = kv.query({
        group = "TASKS",
        include_value = true,
    })

    if err then
        print("get_tasks error:", tostring(err))

        req.json(500, {
            error = tostring(err)
        })
        return
    end

    -- Ensure data is a table (bindings now return proper arrays)
    if data == nil then
        data = {}
    end
    
    req.jsonArray(200, data)
end

function add_task(ctx)
    local kv = require("pkv")
    local req = ctx.request()
    
    -- Parse request body
    local body, err = req.bindJSON()
    if err then
        req.json(400, {
            error = "Invalid JSON: " .. tostring(err)
        })
        return
    end
    
    if not body or not body.task then
        req.json(400, {
            error = "task field is required"
        })
        return
    end
    
    -- Generate a unique key (using timestamp or counter)
    local taskText = body.task
    local key = tostring(os.time()) .. "_" .. math.random(1000, 9999)
    
    -- Add task to KV store using upsert
    local _, err = kv.upsert("TASKS", key, {
        value = taskText
    })
    if err then
        req.json(500, {
            error = tostring(err)
        })
        return
    end
    
    
    req.json(201, {})
end

function delete_task(ctx)
    local kv = require("pkv")
    local req = ctx.request()
    local path = ctx.param("subpath")
    
    -- Extract key from path like "/api/tasks/{key}"
    local key = string.match(path, "/api/tasks/(.+)")
    if not key or key == "" then
        req.json(400, {
            error = "Task key is required"
        })
        return
    end
    
    local err = kv.remove("TASKS", key)
    if err then
        req.json(500, {
            error = tostring(err)
        })
        return
    end
    

    
    req.json(200, {})
end

function on_http(ctx)
    local req = ctx.request()

    local path = ctx.param("subpath")
    local method = ctx.param("method")

    print("SOMETHING NEW")

    print("on_http - path:", path, "method:", method)

    if path == "/api/tasks" and method == "GET" then
        return get_tasks(ctx)
    end
    
    if path == "/api/tasks" and method == "POST" then
        return add_task(ctx)
    end
    
    if string.match(path, "^/api/tasks/.+$") and method == "DELETE" then
        return delete_task(ctx)
    end

    req.json(200, {
        message = "Hello, world!"
    })
end