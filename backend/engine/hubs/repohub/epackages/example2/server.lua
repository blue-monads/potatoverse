-- Example 2: Dynamic Router Server Script
-- This demonstrates handling template and API routes with path parameters

-- Sample data for categories
local categories_data = {
    electronics = {
        id = "electronics",
        name = "Electronics",
        description = "Electronic devices, gadgets, and accessories",
        item_count = 156,
        created_at = "2024-01-15",
        items = {
            {name = "Laptop Pro", price = "1299.99", description = "High-performance laptop"},
            {name = "Wireless Mouse", price = "29.99", description = "Ergonomic wireless mouse"},
            {name = "USB-C Cable", price = "12.99", description = "Fast charging cable"}
        }
    },
    books = {
        id = "books",
        name = "Books",
        description = "Literature, textbooks, and e-books",
        item_count = 432,
        created_at = "2024-01-10",
        items = {
            {name = "The Go Programming Language", price = "39.99", description = "Comprehensive Go guide"},
            {name = "Clean Code", price = "34.99", description = "A Handbook of Agile Software Craftsmanship"},
            {name = "Design Patterns", price = "44.99", description = "Elements of Reusable OO Software"}
        }
    }
}

-- Sample user data
local users_data = {
    john123 = {
        user_id = "john123",
        username = "John Doe",
        email = "john@example.com",
        bio = "Software developer and tech enthusiast. Love coding in Go and Lua!",
        location = "San Francisco, CA",
        website = "https://johndoe.dev",
        member_since = "January 2023",
        avatar_emoji = "👨‍💻",
        post_count = 42,
        follower_count = 1337,
        following_count = 256
    },
    alice456 = {
        user_id = "alice456",
        username = "Alice Johnson",
        email = "alice@example.com",
        bio = "Full-stack developer | Open source contributor | Coffee addict ☕",
        location = "New York, NY",
        website = "https://alice.codes",
        member_since = "March 2023",
        avatar_emoji = "👩‍💻",
        post_count = 87,
        follower_count = 2048,
        following_count = 512
    }
}

-- Template handler: Get category page
function get_category_page(ctx)
    -- Get the category ID from path parameter
    local category_id = ctx.param("id")
    local req = ctx.request()
    
    
    print("get_category_page called with ID:", category_id)
    
    -- Get category data
    local category = categories_data[category_id]
    
    if not category then
        -- Default category if not found
        category = {
            category_id = category_id,
            category_name = string.upper(string.sub(category_id, 1, 1)) .. string.sub(category_id, 2),
            description = "This is a demo category",
            item_count = 0,
            created_at = "2025-10-17",
            items = {
                {name = "Sample Item 1", price = "19.99", description = "Demo item"},
                {name = "Sample Item 2", price = "29.99", description = "Another demo item"}
            }
        }
    else
        category.category_id = category.id
        category.category_name = category.name
    end

    req.state_set_all({
        category_id = category.category_id,
        category_name = category.category_name,
        description = category.description,
        item_count = category.item_count,
        created_at = category.created_at,
        items = category.items
    })
end

-- Template handler: Get user profile
function get_user_profile(ctx)
    -- Get the user ID from path parameter
    local user_id = ctx.param("userId")
    local req = ctx.request()
    
    print("get_user_profile called with userId:", user_id)
    
    -- Get user data
    local user = users_data[user_id]
    
    if not user then
        -- Default user if not found
        user = {
            user_id = user_id,
            username = string.upper(string.sub(user_id, 1, 1)) .. string.sub(user_id, 2),
            email = user_id .. "@example.com",
            bio = "Demo user profile",
            location = "Unknown",
            website = "https://example.com",
            member_since = "October 2025",
            avatar_emoji = "👤",
            post_count = 0,
            follower_count = 0,
            following_count = 0
        }
    end
    
    -- Set template variables using setAll

    req.state_set_all({
        user_id = user.user_id,
        username = user.username,
        email = user.email,
        bio = user.bio,
        location = user.location,
        website = user.website,
        member_since = user.member_since,
        avatar_emoji = user.avatar_emoji,
        post_count = user.post_count,
        follower_count = user.follower_count,
        following_count = user.following_count
    })
end

-- API handler: Get category
function get_category(ctx)
    local category_id = ctx.param("id")
    
    print("get_category API called with ID:", category_id)
    
    local category = categories_data[category_id]
    
    if not category then
        ctx.request():json(404, {
            success = false,
            error = "Category not found",
            category_id = category_id
        })
        return
    end
    
    ctx.request():json(200, {
        success = true,
        data = category
    })
end

-- API handler: Create category
function create_category(ctx)
    print("create_category API called")
    
    local new_id = "new-" .. tostring(math.random(1000, 9999))
    
    ctx.request():json(201, {
        success = true,
        message = "Category created successfully",
        data = {
            id = new_id,
            name = "New Category",
            description = "This is a newly created category",
            created_at = "2025-10-17",
            item_count = 0
        }
    })
end

-- API handler: Delete category
function delete_category(ctx)
    local category_id = ctx.param("id")
    
    print("delete_category API called with ID:", category_id)

    
    ctx.request():json(200, {
        success = true,
        message = "Category deleted successfully",
        deleted_id = category_id
    })
end

print("Example 2 server script loaded successfully!")

