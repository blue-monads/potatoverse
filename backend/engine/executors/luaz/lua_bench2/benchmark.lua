-- benchmark.lua
-- Tests memory usage with either metatable bindings or direct object methods

-- Create multiple module instances to test metatable efficiency
-- Metatables are more efficient when many instances share the same metatable
local num_instances =  100
local instances = {}

for i = 1, num_instances do
    instances[i] = create_module()
end

-- Perform many method calls on all instances
local iterations = 1000
local sum = 0

for i = 1, iterations do
    for j = 1, num_instances do
        local obj = instances[j]
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
end

-- Keep the instances and result alive to measure memory
return {
    instances = instances,
    sum = sum,
    iterations = iterations,
    num_instances = num_instances
}
