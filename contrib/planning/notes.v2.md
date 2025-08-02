# Notes


## DIVI UP Resources
### ROUTES:
 /z/startup/
 /z/pages/startup/
 /z/pages/portal/spaces/<space_type>/
 /z/spaces/<space_type>/
 /z/api/spaces/<space_type>/
 /z/api/space_static/<space_type> ??
 /z/auth/portal_redirrect -> if spaces should be authed then it should redirrect with /z/auth/portal_redirrect?redirect_url=""&space=mnop

### TABLE:
 zSpaceXXX
 zspace_pid_xxx
 zBuddyXXX
 zbuddy_bid_xxx
 zCaptureXXX
 zcapture_chash_xxx
 zLogXXX
### WORKING_DIR:
 /working_space/<space_type>
### CLI: (lip gloss)
 turnix app [start|stop|restart]
 turnix space [create|delete|list|info|start|stop|restart|logs|exec|run]
 turnix p2pProxy [start|stop|restart]
 turnix buddy
 



app.toml
space.toml

## TOOODOOO

1. generate llm.txt docs
    - introduction
    - core and arch docs
    - api docs
    - lua bindings docs
    - cli docs
