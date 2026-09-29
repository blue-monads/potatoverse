package xremote

import (
	"errors"
	"fmt"
	"strings"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/remotehub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

type RemoteCapability struct {
	app                     xtypes.App
	remoteHub               *remotehub.RemoteHub
	capHandle               xcapability.XCapabilityHandle
	defaultSpaceId          int64
	defaultPackageId        int64
	defaultPackageVersionId int64
}

var actions = []string{
	// DB actions
	"run_query",
	"run_query_one",
	"insert",
	"update_by_id",
	"delete_by_id",
	"find_by_id",
	"update_by_cond",
	"delete_by_cond",
	"find_all_by_cond",
	"find_one_by_cond",
	"find_all_by_query",
	"find_by_join",
	"list_tables",
	"list_table_columns",

	// KV actions
	"kv_add",
	"kv_get",
	"kv_query",
	"kv_remove",
	"kv_update",
	"kv_upsert",

	// Core actions
	"publish_event",
	"file_token",
	"sign_advisery_token",
	"parse_advisery_token",
	"read_package_file",
	"list_files",
	"decode_file_id",
	"encode_file_id",
	"get_env",

	// Capability actions
	"cap_list",
	"cap_execute",
	"cap_methods",
	"cap_token_sign",
}

func (r *RemoteCapability) ListActions() ([]string, error) {
	return actions, nil
}

func (r *RemoteCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	return r, nil
}

func (r *RemoteCapability) Close() error {
	return nil
}

func (r *RemoteCapability) Handle(ctx *gin.Context) {
	action := ctx.Param("action")
	if action == "" {
		action = ctx.Param("subpath")
	}
	if action == "" {
		httpx.WriteErrString(ctx, "action is required")
		return
	}
	if strings.HasPrefix(action, "/") {
		action = action[1:]
	}

	result, err := r.Execute(action, lazydata.NewLazyHTTP(ctx))
	httpx.WriteJSON(ctx, result, err)
}

func (r *RemoteCapability) resolveTarget(params lazydata.LazyData) (int64, int64, int64, error) {
	token := params.GetFieldAsString("remote_ctx_token")
	if token == "" {
		token = params.GetFieldAsString("token")
	}

	if token != "" && r.remoteHub != nil {
		claim, err := r.remoteHub.ParseRemoteCtxToken(token)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid remote_ctx_token: %w", err)
		}
		return claim.TargetSpaceId, claim.TargetPackageId, claim.TargetPackageVersionId, nil
	}

	// Fallback to static configuration if available
	if r.defaultSpaceId != 0 || r.defaultPackageId != 0 {
		return r.defaultSpaceId, r.defaultPackageId, r.defaultPackageVersionId, nil
	}

	return 0, 0, 0, errors.New("missing remote_ctx_token or target space configuration")
}

func (r *RemoteCapability) buildRContext(targetSpaceId, targetPkgId, targetVerId int64, params lazydata.LazyData) (remotehub.RContext, error) {
	return remotehub.NewCapRContext(targetSpaceId, targetPkgId, targetVerId, params)
}

func normalizeAction(name string) string {
	name = strings.TrimPrefix(name, "/")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "/", "_")
	return name
}

func (r *RemoteCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	if r.remoteHub == nil {
		return nil, errors.New("remotehub is not available")
	}

	targetSpaceId, targetPkgId, targetVerId, err := r.resolveTarget(params)
	if err != nil {
		return nil, err
	}

	rctx, err := r.buildRContext(targetSpaceId, targetPkgId, targetVerId, params)
	if err != nil {
		return nil, err
	}

	action := normalizeAction(name)

	switch action {
	// DB actions
	case "run_query", "db_run_query":
		return r.remoteHub.DBRunQuery(rctx)
	case "run_query_one", "db_run_query_one":
		return r.remoteHub.DBRunQueryOne(rctx)
	case "insert", "db_insert":
		return r.remoteHub.DBInsert(rctx)
	case "update_by_id", "db_update_by_id":
		return r.remoteHub.DBUpdateById(rctx)
	case "delete_by_id", "db_delete_by_id":
		return r.remoteHub.DBDeleteById(rctx)
	case "find_by_id", "db_find_by_id":
		return r.remoteHub.DBFindById(rctx)
	case "update_by_cond", "db_update_by_cond":
		return r.remoteHub.DBUpdateByCond(rctx)
	case "delete_by_cond", "db_delete_by_cond":
		return r.remoteHub.DBDeleteByCond(rctx)
	case "find_all_by_cond", "db_find_all_by_cond":
		return r.remoteHub.DBFindAllByCond(rctx)
	case "find_one_by_cond", "db_find_one_by_cond":
		return r.remoteHub.DBFindOneByCond(rctx)
	case "find_all_by_query", "db_find_all_by_query":
		return r.remoteHub.DBFindAllByQuery(rctx)
	case "find_by_join", "db_find_by_join":
		return r.remoteHub.DBFindByJoin(rctx)
	case "list_tables", "db_list_tables":
		return r.remoteHub.DBListTables(rctx)
	case "list_table_columns", "db_list_columns", "list_columns":
		return r.remoteHub.DBListColumns(rctx)

	// KV actions
	case "kv_add", "add":
		return r.remoteHub.KVAdd(rctx)
	case "kv_get", "get":
		return r.remoteHub.KVGet(rctx)
	case "kv_query", "query":
		return r.remoteHub.KVQuery(rctx)
	case "kv_remove", "remove":
		return r.remoteHub.KVRemove(rctx)
	case "kv_update", "update":
		return r.remoteHub.KVUpdate(rctx)
	case "kv_upsert", "upsert":
		return r.remoteHub.KVUpsert(rctx)

	// Core actions
	case "publish_event", "core_publish_event":
		return r.remoteHub.CorePublishEvent(rctx)
	case "file_token", "core_file_token":
		return r.remoteHub.CoreFileToken(rctx)
	case "sign_advisery_token", "core_sign_advisery_token":
		return r.remoteHub.CoreSignAdviseryToken(rctx)
	case "parse_advisery_token", "core_parse_advisery_token":
		return r.remoteHub.CoreParseAdviseryToken(rctx)
	case "read_package_file", "core_read_package_file":
		return r.remoteHub.CoreReadPackageFile(rctx)
	case "list_files", "core_list_files":
		return r.remoteHub.CoreListFiles(rctx)
	case "decode_file_id", "core_decode_file_id":
		return r.remoteHub.CoreDecodeFileId(rctx)
	case "encode_file_id", "core_encode_file_id":
		return r.remoteHub.CoreEncodeFileId(rctx)
	case "get_env", "core_get_env":
		return r.remoteHub.CoreGetEnv(rctx)

	// Cap actions
	case "cap_list":
		return r.remoteHub.CapList(rctx)
	case "cap_execute":
		return r.remoteHub.CapExecute(rctx)
	case "cap_methods":
		return r.remoteHub.CapMethods(rctx)
	case "cap_token_sign":
		return r.remoteHub.CapTokenSign(rctx)

	default:
		return nil, fmt.Errorf("unknown action: %s", name)
	}
}
