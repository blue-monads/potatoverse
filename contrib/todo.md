- sugify things (check validate entitity names if key has "<a>|<b>" then it has to disallow |) 
- go through all services and fix logging
- devices and sessions, hash password
- low db txn support, add opentxn to closer in lua bindings
- low db testing
- event hub testing
- add subprocess bashed exec binary (cli server start-actual)
- integrate litestream
- cli package validate
- claude skill (potato skill repo)
- remove unsafe func from glua (maybe softfork it ?)
- see wal2/bw2 sqlite is worth effort


## capabilities
- user_system_ws / sockethub
- controller
- ws_server_broadcast(room_id) (only server broadcast) 
- ws_p2p(room_id) (server broadcast and one user could send to another)
- user(user_group/all)
- system
- remote_space
- lazydb



- Zero to 100 building app
- tell invalid zip if manifest is empty
- show error if potato.yaml is missing for packaging command (instead of unkown panic ) and other 
- add lua linter
- disabled app in single domain mode or ns collision icon
- expose flag (--expose-hq-tunnel)
- refactor notification fronetend and fix bug
- normal user portal


[Potatoverse](https://github.com/blue-monads/potatoverse)