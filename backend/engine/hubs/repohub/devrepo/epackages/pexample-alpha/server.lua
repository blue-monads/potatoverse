function on_http(ctx)
    local req = ctx.request()
    local subpath = ctx.param("subpath") or ""
    if subpath:sub(1, 1) == "/" then
        subpath = subpath:sub(2)
    end

    if subpath == "status" or subpath == "api/status" then
        req.json(200, {
            app = "pexample-alpha",
            type = "host",
            status = "running",
            timestamp = os.time()
        })
        return
    end

    if subpath == "info" or subpath == "api/info" then
        req.json(200, {
            app = "pexample-alpha",
            name = "Example Alpha Host",
            description = "Host application supporting AppPlugins"
        })
        return
    end

    req.json(200, {
        app = "pexample-alpha",
        subpath = subpath
    })
end
