- dev/cli
- repo system
- sugify things (check validate entitity names if key has "<a>|<b>" then it has to disallow |) 
- go through all services and fix logging
- devices and sessions, hash password
- low db txn support
- low db testing
- event hub testing


## capabilities
- user_system_ws / sockethub
- controller
- ws_server_broadcast(room_id) (only server broadcast) 
- ws_p2p(room_id) (server broadcast and one user could send to another)
- user(user_group/all)
- system
- remote_space
- lazydb

https://github.com/pluveto/flydav

// root_<pub_key_hash>.freehttptunnel.com
// <s-x>_<pub_key_hash>.freehttptunnel.com

// zz-10-funnel