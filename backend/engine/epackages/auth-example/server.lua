
function get_tasks(ctx)
    local kv = require("kv")
    local req = ctx.request()

    local data, err = kv.query({
        group = "TASKS",
    })

    if err then
        req.json(500, {
            error = tostring(err)
        })
        return
    end

    -- data should be a table/array, but ensure it's not nil
    if data == nil then
        data = {}
    end
    
    req.json(200, data)
end

function add_task(ctx)
    local kv = require("kv")
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
    
    -- Return all tasks
    local data, err = kv.query({
        group = "TASKS",
    })
    
    if err then
        req.json(500, {
            error = tostring(err)
        })
        return
    end
    
    -- Ensure data is a table
    if data == nil then
        data = {}
    end
    
    req.json(200, data)
end

function delete_task(ctx)
    local kv = require("kv")
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
    
    -- Delete the task directly by key
    local _, err = kv.remove("TASKS", key)
    if err then
        req.json(500, {
            error = tostring(err)
        })
        return
    end
    
    -- Return remaining tasks
    local remainingData, err = kv.query({
        group = "TASKS",
    })
    
    if err then
        req.json(500, {
            error = tostring(err)
        })
        return
    end
    
    -- Ensure data is a table
    if remainingData == nil then
        remainingData = {}
    end
    
    req.json(200, remainingData)
end

function on_http(ctx)
    local req = ctx.request()

    local path = ctx.param("subpath")
    local method = ctx.param("method")

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