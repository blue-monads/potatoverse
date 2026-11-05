function on_http(ctx)
    local req = ctx.request()
    local subpath = ctx.param("subpath")
    local remote_token = ctx.param("remote_ctx_token")

    -- Clean up subpath
    if subpath and subpath:sub(1, 1) == "/" then
        subpath = subpath:sub(2)
    end

    -- Verify claim using remote_ctx_token if provided
    local claim, claim_err = nil, nil
    if remote_token and remote_token ~= "" then
        claim, claim_err = req.get_claim(remote_token)
    else
        claim, claim_err = req.get_claim()
    end

    if subpath == "ping" then
        req.json(200, {
            status = "ok",
            plugin = "pexample-beta",
            action = "ping",
            message = "Pong from Example Beta Plugin!",
            has_remote_ctx_token = (remote_token ~= nil and remote_token ~= ""),
            authenticated = (claim ~= nil),
            user_id = claim and claim.u or nil,
            target_space_id = claim and claim.s or nil
        })
        return
    end

    if subpath == "info" then
        req.json(200, {
            name = "Example Beta Plugin",
            version = "0.0.1",
            space_type = "AppPlugin",
            has_remote_ctx_token = (remote_token ~= nil and remote_token ~= ""),
            authenticated = (claim ~= nil)
        })
        return
    end

    if subpath == "remote_exec" then
        -- Showcase calling xRemote capability with remote_ctx_token
        if not remote_token or remote_token == "" then
            req.json(400, {
                error = "remote_ctx_token is required for cross-space execution"
            })
            return
        end

        local res, err = potato.cap.execute("xRemote", "list_tables", {
            remote_ctx_token = remote_token
        })

        if err then
            req.json(500, {
                error = tostring(err)
            })
            return
        end

        req.json(200, {
            status = "success",
            tables = res
        })
        return
    end

    req.json(200, {
        plugin = "pexample-beta",
        subpath = subpath or "",
        has_remote_ctx_token = (remote_token ~= nil and remote_token ~= "")
    })
end
