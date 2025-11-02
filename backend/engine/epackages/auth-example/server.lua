
-- handle_http

function on_http(ctx)
    local req = ctx.request()

    req.json(200, {
        message = "Hello, world!"
    })
end