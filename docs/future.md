
> This is not documentation of built system, it defines eventual goal of project, most of the system has been built already in one of its previous version but they are all lost in refactor and persuit of perfection. This final refactor to rule all refactor will bring everything together.

# What is turnix

Turnix is a small app platform. It mainly host webapps with its server side code. Idea behind trunix is give a batteries includeded plaform for building hosting apps since most common parts needed is provided by platfrom apps will be much smaller. Think it as a you cloud provider in a single binary or even better analogy would be hybrid between CMS and heroku like PAAS, It accepts certain tradeoff to achive that goal which will be apparent later i doument. System is made of components like
1. Users (admin, normal, bot)
2. Auth (kinda like Oauth but for apps and its users, not for external service for now)
3. All stored in SQlite db (pg could be added in future ), with CDC system to capture and sync changes some sync/backup service.
4. App Engine is core part of platform responsible for running the apps. 
    - App are called spaces and run inside a Language VM, Lua or webassembly and call through bindings for other services as well as request for apps are handled by respective entry point. Space are created from blueprint  
    - Each blueprint has a slugified name as its namespaced key which could be used in later places for scoping resources. One of those is http route so space are served from a specific place based on its namespace key.
    - Each space also has own files scoped by its id.




## Resource Scoping
Most of the resource realstate is diviup into structure of.
1. Core Platform
2. App specific namespace.

Resources:
1. HTTP routes
    - CORE
        - /z/pages/ 
        - /z/api/core
    - APP
        - (asset serve) /z/blueprint/<blueprint_key>/ (internal) 
        - (asset serve) /z/pages/spaces/<blueprint_key>/<space_id> (external)
        - (asset serve)     /z/pages/blueprint/<blueprint_key>/

        - (api)     /z/api/space/<blueprint_key>/<space_id>     
        - (api)     /z/api/blueprint/<blueprint_key>/

        - (extra)     /z/api/space_extra/<blueprint_key>/<space_id> (extra stuff provided by runtime)



2. DB Table space
    - CORE
        - core_<table_name>
    - APP
        - z_<space_id>_<table_name>
3. Working directory
    - CORE
    - APP
        

## Extra 

- /ws/broadcast?ws_join_token=xyz
- /ws/p2p/<target_p2p_id>?ws_join_token=xyz
- /ws/room/<room_id>?room_join_token=xyz

- /kv/set
- /kv/query
- /kv/

- /file/upload
- /file/upload-presigned?presigned-key=xyz
- /file


KitchenSinkTest


## Terminology