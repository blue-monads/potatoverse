-- benchmark_script.lua

-- This script performs a series of operations to consume memory
-- and perform some dummy computations.

local total_memory = 0
local data = {}
local random_numbers = {}

-- Create a large table with random strings to consume memory
for i = 1, 100000 do
    data[i] = "string_data_" .. tostring(math.random(100000))
end

print("Created a table with 100,000 random strings.")
total_memory = total_memory + #data

-- Perform a dummy loop to simulate computation
local sum = 0
for i = 1, 1000000 do
    sum = sum + math.sqrt(i)
end

print("Completed a dummy computation loop.")

-- Create another table with random numbers
for i = 1, 50000 do
    random_numbers[i] = math.random() * 10000
end

print("Created a table with 50,000 random numbers.")
total_memory = total_memory + #random_numbers

-- Create a few deeply nested tables to test garbage collection and memory usage
local nested_table = {}
local current_table = nested_table
for i = 1, 100 do
    local new_table = { ["key" .. i] = "value" .. i }
    current_table[i] = new_table
    current_table = new_table
end

print("Created a deeply nested table.")

-- Store a few large strings
local large_string_1 = string.rep("A", 1024 * 1) -- 1KB
local large_string_2 = string.rep("B", 1024 * 5)  -- 5KB

print("Created some large strings.")

-- Return some values from the script to simulate a real-world scenario
return {
    sum_result = sum,
    total_items = total_memory
}