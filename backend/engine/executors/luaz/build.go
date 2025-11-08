package luaz

const Code = `

local db = require("db")
local math = require("math")

function im_cool(a)
	print("I'm cool")
	return a + 1
end


function on_http(ctx)
  print("Hello from lua!", ctx.type())
  local req = ctx.request()

  local rand = math.random(1, 100)

  db.add({
	group = "test",
	key = "test" .. rand,
	value = "test",
  })


  req.json(200, {
	im_cool = im_cool(18),
	message = "Hello from lua! from lua!",
	space_id = ctx.param("space_id"),
	package_id = ctx.param("package_id"),
	subpath = ctx.param("subpath"),
  })

end

`

const HandlersReference = `


function on_http(ctx)
	print("@on_http", ctx.type())
end

function on_ws_room(ctx)
	print("@on_ws_room", ctx.type())
end

function on_rmcp(ctx)
	print("@on_rmcp", ctx.type())
end




`

const ByPassPackageCode = false
