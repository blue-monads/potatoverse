function on_http(ctx)
    local req = ctx.request()
    local subpath = ctx.param("subpath") or ""
    if subpath:sub(1, 1) == "/" then
        subpath = subpath:sub(2)
    end

    if subpath == "api/status" or subpath == "status" then
        req.json(200, {
            status = "ok",
            space = "pexample-multi",
            role = "main_space",
            is_default = false,
            message = "Main space running (non-default entrypoint)",
            timestamp = os.time()
        })
        return
    end

    if subpath == "api/info" or subpath == "info" then
        req.json(200, {
            app = "pexample-multi",
            name = "Example Multi Main Space",
            role = "main_space",
            is_default = false,
            description = "Primary main space for pexample-multi package. Namespace matches package slug.",
            spaces = {
                {
                    namespace = "pexample-multi",
                    role = "main_space",
                    is_default = false
                },
                {
                    namespace = "pexample-multi:sub",
                    role = "sub_space",
                    is_default = true
                }
            }
        })
        return
    end

    req.json(200, {
        space = "pexample-multi",
        role = "main_space",
        subpath = subpath,
        message = "Handled by Main Space server"
    })
end

