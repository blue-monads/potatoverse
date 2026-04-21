package server

import "github.com/gin-gonic/gin"

func (s *Server) BindRemoteBinds(rg *gin.RouterGroup) {

	rhub := s.engine.GetRemoteHub()

	rg.GET("/capability/list", rhub.Authed(rhub.CapList))
	rg.POST("/capability/:cap/sign", rhub.Authed(rhub.CapTokenSign))
	rg.GET("/capability/:cap/methods", rhub.Authed(rhub.CapMethods))
	rg.POST("/capability/:cap/execute/:method", rhub.Authed(rhub.CapExecute))

	rg.GET("/core/read_package_file/*path", rhub.Authed(rhub.CoreReadPackageFile))
	rg.GET("/core/list_files/*path", rhub.Authed(rhub.CoreListFiles))
	rg.GET("/core/decode_file_id/:id", rhub.Authed(rhub.CoreDecodeFileId))
	rg.GET("/core/encode_file_id/:id", rhub.Authed(rhub.CoreEncodeFileId))
	rg.GET("/core/env/:key", rhub.Authed(rhub.CoreGetEnv))

	rg.POST("/core/publish_event", rhub.Authed(rhub.CorePublishEvent))
	rg.POST("/core/file_token", rhub.Authed(rhub.CoreFileToken))
	rg.POST("/core/sign_advisery_token", rhub.Authed(rhub.CoreSignAdviseryToken))
	rg.POST("/core/parse_advisery_token", rhub.Authed(rhub.CoreParseAdviseryToken))

	rg.POST("/db/run_query", rhub.Authed(rhub.DBRunQuery))
	rg.POST("/db/run_query_one", rhub.Authed(rhub.DBRunQueryOne))
	rg.POST("/db/insert", rhub.Authed(rhub.DBInsert))
	rg.POST("/db/update_by_id", rhub.Authed(rhub.DBUpdateById))
	rg.POST("/db/delete_by_id", rhub.Authed(rhub.DBDeleteById))
	rg.POST("/db/find_by_id", rhub.Authed(rhub.DBFindById))
	rg.POST("/db/update_by_cond", rhub.Authed(rhub.DBUpdateByCond))
	rg.POST("/db/delete_by_cond", rhub.Authed(rhub.DBDeleteByCond))
	rg.POST("/db/find_all_by_cond", rhub.Authed(rhub.DBFindAllByCond))
	rg.POST("/db/find_one_by_cond", rhub.Authed(rhub.DBFindOneByCond))
	rg.POST("/db/find_all_by_query", rhub.Authed(rhub.DBFindAllByQuery))
	rg.POST("/db/find_by_join", rhub.Authed(rhub.DBFindByJoin))
	rg.GET("/db/list_tables", rhub.Authed(rhub.DBListTables))
	rg.GET("/db/table/:table/columns", rhub.Authed(rhub.DBListColumns))

	rg.POST("/kv/add", rhub.Authed(rhub.KVAdd))
	rg.GET("/kv/:group/:key", rhub.Authed(rhub.KVGet))
	rg.POST("/kv/query", rhub.Authed(rhub.KVQuery))
	rg.POST("/kv/remove", rhub.Authed(rhub.KVRemove))
	rg.POST("/kv/update", rhub.Authed(rhub.KVUpdate))
	rg.POST("/kv/upsert", rhub.Authed(rhub.KVUpsert))

}
