function on_http(ctx)
    local req = ctx.request()
    local subpath = ctx.param("subpath") or ""
    if subpath:sub(1, 1) == "/" then
        subpath = subpath:sub(2)
    end

    if subpath == "api/status" or subpath == "status" then
        req.json(200, {
            status = "ok",
            space = "pexample-multi:sub",
            role = "sub_space",
            is_default = true,
            parent_main_space = "pexample-multi",
            message = "Default sub-space is healthy and serving primary requests",
            timestamp = os.time()
        })
        return
    end

    if subpath == "api/info" or subpath == "info" then
        req.json(200, {
            app = "pexample-multi:sub",
            name = "Example Multi Default Sub-Space",
            role = "sub_space",
            is_default = true,
            parent_main_space = "pexample-multi",
            description = "Default sub-space (is_default: true) running alongside non-default main space (pexample-multi)"
        })
        return
    end

    req.json(200, {
        space = "pexample-multi:sub",
        role = "sub_space",
        is_default = true,
        subpath = subpath,
        message = "Handled by Default Sub-Space server"
    })
end

