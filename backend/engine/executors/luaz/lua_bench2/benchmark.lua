-- benchmark.lua
-- Tests memory usage with either metatable bindings or direct object methods

-- Create the module object
local obj = create_module()

-- Perform many method calls
local iterations = 100000
local sum = 0

for i = 1, iterations do
    -- Use colon syntax for metatable, dot syntax for direct
    -- Both work the same way from Lua's perspective
    sum = sum + obj:method1()
    sum = sum + obj:method2()
    sum = sum + obj:method3()
    sum = sum + obj:method4()
    sum = sum + obj:method5()
    sum = sum + obj:method6()
    sum = sum + obj:method7()
    sum = sum + obj:method8()
    sum = sum + obj:method9()
    sum = sum + obj:method10()
    sum = sum + obj:method11()
    sum = sum + obj:method12()
end

-- Keep the object and result alive to measure memory
return {
    obj = obj,
    sum = sum,
    iterations = iterations
}
