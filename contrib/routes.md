[GIN-debug] GET    /zz/ping                  --> github.com/blue-monads/turnix/backend/app/server.(*Server).bindRoutes.func1 (3 handlers)
"@using_dev_proxy" "http://localhost:7779"
[GIN-debug] GET    /zz/pages                 --> github.com/blue-monads/turnix/backend/app/server/assets.PagesRoutesServer.func1 (3 handlers)
[GIN-debug] GET    /zz/pages/*files          --> github.com/blue-monads/turnix/backend/app/server/assets.PagesRoutesServer.func1 (3 handlers)
[GIN-debug] GET    /zz/lib/*file             --> github.com/blue-monads/turnix/backend/app/server.(*Server).pages.func1 (3 handlers)
[GIN-debug] GET    /zz/profileImage/:id/:name --> github.com/blue-monads/turnix/backend/app/server.(*Server).userSvgProfileIcon-fm (3 handlers)
[GIN-debug] GET    /zz/profileImage/:id      --> github.com/blue-monads/turnix/backend/app/server.(*Server).userSvgProfileIconById-fm (3 handlers)
[GIN-debug] GET    /zz/api/gradients         --> github.com/blue-monads/turnix/backend/app/server.(*Server).ListGradients-fm (3 handlers)
[GIN-debug] GET    /zz/api/core/user/        --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func1 (3 handlers)
[GIN-debug] GET    /zz/api/core/user/:id     --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func2 (3 handlers)
[GIN-debug] GET    /zz/api/core/user/invites --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func3 (3 handlers)
[GIN-debug] GET    /zz/api/core/user/invites/:id --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func4 (3 handlers)
[GIN-debug] POST   /zz/api/core/user/invites --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func5 (3 handlers)
[GIN-debug] PUT    /zz/api/core/user/invites/:id --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func6 (3 handlers)
[GIN-debug] DELETE /zz/api/core/user/invites/:id --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func7 (3 handlers)
[GIN-debug] POST   /zz/api/core/user/invites/:id/resend --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func8 (3 handlers)
[GIN-debug] POST   /zz/api/core/user/create  --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func9 (3 handlers)
[GIN-debug] GET    /zz/api/core/user/groups  --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func10 (3 handlers)
[GIN-debug] GET    /zz/api/core/user/groups/:name --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func11 (3 handlers)
[GIN-debug] POST   /zz/api/core/user/groups  --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func12 (3 handlers)
[GIN-debug] PUT    /zz/api/core/user/groups/:name --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func13 (3 handlers)
[GIN-debug] DELETE /zz/api/core/user/groups/:name --> github.com/blue-monads/turnix/backend/app/server.(*Server).userRoutes.(*Server).withAccessTokenFn.func14 (3 handlers)
[GIN-debug] POST   /zz/api/core/auth/login   --> github.com/blue-monads/turnix/backend/app/server.(*Server).login-fm (3 handlers)
[GIN-debug] GET    /zz/api/core/auth/invite/:token --> github.com/blue-monads/turnix/backend/app/server.(*Server).getInviteInfo-fm (3 handlers)
[GIN-debug] POST   /zz/api/core/auth/invite/:token --> github.com/blue-monads/turnix/backend/app/server.(*Server).acceptInvite-fm (3 handlers)
[GIN-debug] GET    /zz/api/core/self/portalData/:portal_type --> github.com/blue-monads/turnix/backend/app/server.(*Server).selfUserRoutes.(*Server).withAccessTokenFn.func1 (3 handlers)
[GIN-debug] GET    /zz/api/core/self/info    --> github.com/blue-monads/turnix/backend/app/server.(*Server).selfUserRoutes.(*Server).withAccessTokenFn.func2 (3 handlers)
[GIN-debug] PUT    /zz/api/core/self/bio     --> github.com/blue-monads/turnix/backend/app/server.(*Server).selfUserRoutes.(*Server).withAccessTokenFn.func3 (3 handlers)


[GIN-debug] POST   /zz/api/core/package/install --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func2 (3 handlers)
[GIN-debug] POST   /zz/api/core/package/install/zip --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func3 (3 handlers)
[GIN-debug] POST   /zz/api/core/package/install/embed --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func4 (3 handlers)

[GIN-debug] DELETE /zz/api/core/package/:id  --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func5 (3 handlers)
[GIN-debug] POST   /zz/api/core/package/:id/dev-token --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func6 (3 handlers)
[GIN-debug] POST   /zz/api/core/package/push --> github.com/blue-monads/turnix/backend/app/server.(*Server).PushPackage-fm (3 handlers)
[GIN-debug] GET    /zz/api/core/package/list --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func7 (3 handlers)
[GIN-debug] GET    /zz/api/core/space/installed --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func8 (3 handlers)
[GIN-debug] POST   /zz/api/core/space/authorize/:space_key --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func9 (3 handlers)

[GIN-debug] GET    /zz/api/core/package/:id/files --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func10 (3 handlers)
[GIN-debug] GET    /zz/api/core/package/:id/files/:fileId --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func11 (3 handlers)
[GIN-debug] GET    /zz/api/core/package/:id/files/:fileId/download --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func12 (3 handlers)
[GIN-debug] DELETE /zz/api/core/package/:id/files/:fileId --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func13 (3 handlers)
[GIN-debug] POST   /zz/api/core/package/:id/files/upload --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func14 (3 handlers)

[GIN-debug] GET    /zz/api/core/space/:id/kv --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func15 (3 handlers)
[GIN-debug] GET    /zz/api/core/space/:id/kv/:kvId --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func16 (3 handlers)
[GIN-debug] POST   /zz/api/core/space/:id/kv --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func17 (3 handlers)
[GIN-debug] PUT    /zz/api/core/space/:id/kv/:kvId --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func18 (3 handlers)
[GIN-debug] DELETE /zz/api/core/space/:id/kv/:kvId --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func19 (3 handlers)
[GIN-debug] GET    /zz/api/core/space/:id/files --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func20 (3 handlers)
[GIN-debug] GET    /zz/api/core/space/:id/files/:fileId --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func21 (3 handlers)
[GIN-debug] GET    /zz/api/core/space/:id/files/:fileId/download --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func22 (3 handlers)
[GIN-debug] DELETE /zz/api/core/space/:id/files/:fileId --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func23 (3 handlers)
[GIN-debug] POST   /zz/api/core/space/:id/files/upload --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func24 (3 handlers)
[GIN-debug] POST   /zz/api/core/space/:id/files/folder --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func25 (3 handlers)
[GIN-debug] POST   /zz/api/core/space/:id/files/presigned --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).withAccessTokenFn.func26 (3 handlers)
[GIN-debug] POST   /zz/file/upload-presigned --> github.com/blue-monads/turnix/backend/app/server.(*Server).UploadFileWithPresigned-fm (3 handlers)
[GIN-debug] GET    /zz/api/core/engine/debug --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleEngineDebugData-fm (3 handlers)
[GIN-debug] GET    /zz/api/core/engine/space_info/:space_key --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceInfo-fm (3 handlers)





// addon-root

[GIN-debug] GET    /zz/addon-root/:addon_name/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleAddonsRoot-fm (3 handlers)
[GIN-debug] POST   /zz/addon-root/:addon_name/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleAddonsRoot-fm (3 handlers)
[GIN-debug] PUT    /zz/addon-root/:addon_name/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleAddonsRoot-fm (3 handlers)
[GIN-debug] PATCH  /zz/addon-root/:addon_name/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleAddonsRoot-fm (3 handlers)
[GIN-debug] HEAD   /zz/addon-root/:addon_name/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleAddonsRoot-fm (3 handlers)
[GIN-debug] OPTIONS /zz/addon-root/:addon_name/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleAddonsRoot-fm (3 handlers)
[GIN-debug] DELETE /zz/addon-root/:addon_name/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleAddonsRoot-fm (3 handlers)
[GIN-debug] CONNECT /zz/addon-root/:addon_name/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleAddonsRoot-fm (3 handlers)
[GIN-debug] TRACE  /zz/addon-root/:addon_name/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleAddonsRoot-fm (3 handlers)

// space

[GIN-debug] GET    /zz/api/space/:space_key  --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] POST   /zz/api/space/:space_key  --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] PUT    /zz/api/space/:space_key  --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] PATCH  /zz/api/space/:space_key  --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] HEAD   /zz/api/space/:space_key  --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] OPTIONS /zz/api/space/:space_key  --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] DELETE /zz/api/space/:space_key  --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] CONNECT /zz/api/space/:space_key  --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] TRACE  /zz/api/space/:space_key  --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] GET    /zz/api/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] POST   /zz/api/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] PUT    /zz/api/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] PATCH  /zz/api/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] HEAD   /zz/api/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] OPTIONS /zz/api/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] DELETE /zz/api/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] CONNECT /zz/api/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)
[GIN-debug] TRACE  /zz/api/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceApi-fm (3 handlers)

[GIN-debug] GET    /zz/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceFile.func1 (3 handlers)
[GIN-debug] POST   /zz/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceFile.func1 (3 handlers)
[GIN-debug] PUT    /zz/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceFile.func1 (3 handlers)
[GIN-debug] PATCH  /zz/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceFile.func1 (3 handlers)
[GIN-debug] HEAD   /zz/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceFile.func1 (3 handlers)
[GIN-debug] OPTIONS /zz/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceFile.func1 (3 handlers)
[GIN-debug] DELETE /zz/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceFile.func1 (3 handlers)
[GIN-debug] CONNECT /zz/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceFile.func1 (3 handlers)
[GIN-debug] TRACE  /zz/space/:space_key/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handleSpaceFile.func1 (3 handlers)



// 

[GIN-debug] GET    /zz/api/plugin/:space_key/:plugin_id --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] POST   /zz/api/plugin/:space_key/:plugin_id --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] PUT    /zz/api/plugin/:space_key/:plugin_id --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] PATCH  /zz/api/plugin/:space_key/:plugin_id --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] HEAD   /zz/api/plugin/:space_key/:plugin_id --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] OPTIONS /zz/api/plugin/:space_key/:plugin_id --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] DELETE /zz/api/plugin/:space_key/:plugin_id --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] CONNECT /zz/api/plugin/:space_key/:plugin_id --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] TRACE  /zz/api/plugin/:space_key/:plugin_id --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] GET    /zz/api/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] POST   /zz/api/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] PUT    /zz/api/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] PATCH  /zz/api/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] HEAD   /zz/api/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] OPTIONS /zz/api/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] DELETE /zz/api/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] CONNECT /zz/api/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)
[GIN-debug] TRACE  /zz/api/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).handlePluginApi-fm (3 handlers)

[GIN-debug] GET    /zz/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).handlePluginFile.func1 (3 handlers)
[GIN-debug] POST   /zz/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).handlePluginFile.func1 (3 handlers)
[GIN-debug] PUT    /zz/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).handlePluginFile.func1 (3 handlers)
[GIN-debug] PATCH  /zz/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).handlePluginFile.func1 (3 handlers)
[GIN-debug] HEAD   /zz/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).handlePluginFile.func1 (3 handlers)
[GIN-debug] OPTIONS /zz/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).handlePluginFile.func1 (3 handlers)
[GIN-debug] DELETE /zz/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).handlePluginFile.func1 (3 handlers)
[GIN-debug] CONNECT /zz/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).handlePluginFile.func1 (3 handlers)
[GIN-debug] TRACE  /zz/plugin/:space_key/:plugin_id/*subpath --> github.com/blue-monads/turnix/backend/app/server.(*Server).engineRoutes.(*Server).handlePluginFile.func1 (3 handlers)


// addons





