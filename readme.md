# PotatoVerse 🥔


<div align="center">
    <img src="contrib/potatoverse_banner.png" >
    
[FIDDLE WITH DEMO SITE](https://tubersalltheway.top/zz/pages/auth/login) 

</div>

PotatoVerse is a small app platform that hosts web applications with server-side code. Think of it as a hybrid between a CMS and Heroku-like PaaS, all in one static go binary (ui assets bundled) and SQLite db.

## Features

- Apps called spaces run in isolated environment i.e. suborigin `zz-<app_id>.myapps.com/zz/space/my_app_namespace` and backend in language VM (lua for now, WASM planned) or you also extend/write in native go code or bring your own executor for maybe another language 🤷 . **Apps can also run without suborigin isolation / wildcard origin but you can only run one instance of app or apps sharing common namespace.**
- Custom behaviour and resources can be registered as capabilities and used by apps. 
```lua
potato.cap.execute(
    "xEasyWS", 
    "broadcast", 
    {
        type = "event_created", 
        data = event
    })
```
- Whole platform can be used as a libary to build own CMS or whatever, Code uses go pattern of registering builder in `func init() {...}` like sql drivers for things like `Executor`, `Capability`, `RepoProvider` etc.  means you can write your own executor or capability and register it with the platform just using `import _ "mything"` pattern. Also platform does not eat your root http namespace just `/zz/*` so you can mount your own app and just route `/zz/*` traffic to platform.
- Users and User groups (permission and role future plan).
- Installing and upgrading apps from repo (store), zip upload or URL directly. Very easy to host own repository. So we donot have a single repo fiasco like wordpress.com vs wordpress.org. Ref stores: [Official Store](https://github.com/blue-monads/store)   [Third party store](https://github.com/blue-monads/store-thirdparty)
- Apps has simple KV store or sqlite db for complex apps, Each apps only have access to own isolated set of tables (enforced by parsing sql statement and enforcing access ) unless you go through explict capability.
- Simple package development loop, you can build package and push using cli `potatoverse package push`. Some sample apps ref [Potato Apps](https://github.com/blue-monads/potato-apps) **They are not very useful apps but just to give basic example of working with platform and bindings**
- Apps can emit async events that other apps react to kinda like signal slot like system so you can build side apps that extend functionality of other apps.



> ⚠️ **🚨🚨🚨Alpha Software🚨🚨🚨**: This is early alpha software being developed in spare weekend time. Expect bugs, breaking changes, and incomplete features.





https://github.com/user-attachments/assets/120b826e-c5f9-4e1f-829e-816e7d2982ea


## Quick Install


```bash
curl https://raw.githubusercontent.com/blue-monads/potatoverse/refs/heads/main/contrib/installer.sh | bash
```

```bash
potatoverse server init-and-start 
# Access locally: http://localhost:7777/zz/pages
# Access via tunnel: http://buddy-<nodeid>.tubersalltheway.top/zz/pages
```

> The tunneling system currently has limitations with WebSockets and isolated origin modes.


## Future
- [ ] Polish stuff and write documentation.
- [ ] Buddy backup (WIP)
- [ ] Http Tunnel (WIP, http://buddy-<nodeid>.tubersalltheway.top/zz/pages )
- [ ] WASM executor (current lua runtime is much easier for testing APIs and ideas)
- [ ] Postgres support. (technically possible cz undelying orm supports it but sqlite is just easier for now)


## Terminologies

- **Spaces**: Apps created from packages run in isolated environment.
- **Engine**: Manages space lifecycle, routing, and execution of spaces.
- **Executor**: Executor is responsible for running server code (Lua VM or WASM in the future). It's an interface which is registered similar to how SQL drivers are registered, so you can bring your own or write apps in native Go code too.
- **Capabilities**: Platform services exposed to spaces
- **Packages**: Blueprints containing spaces (apps), code, and assets. Imagine if we had an SPK (server package file) similar to APK for Android apps. A simple example would contain the following in a zip:
    - `potato.json` (manifest file)
    - `public/{index.html, style.css, client.js}` (folder served to users running the space)
    - `server.lua`

# Developing
[How to develop](./docs/contents/topics/how_to_develop.md)
