package xremote

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/remotehub"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/gin-gonic/gin"
)

func TestXRemoteListActions(t *testing.T) {
	rc := &RemoteCapability{}
	acts, err := rc.ListActions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	required := []string{
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
		"kv_add",
		"kv_get",
		"kv_query",
		"kv_remove",
		"kv_update",
		"kv_upsert",
		"publish_event",
		"file_token",
		"sign_advisery_token",
		"parse_advisery_token",
		"read_package_file",
		"list_files",
		"decode_file_id",
		"encode_file_id",
		"get_env",
		"cap_list",
		"cap_execute",
		"cap_methods",
		"cap_token_sign",
	}

	actMap := make(map[string]bool)
	for _, a := range acts {
		actMap[a] = true
	}

	for _, req := range required {
		if !actMap[req] {
			t.Fatalf("missing required action: %s", req)
		}
	}
}

func TestXRemoteExecuteMissingTarget(t *testing.T) {
	rhub := remotehub.NewRemoteHub()
	rc := &RemoteCapability{
		remoteHub: rhub,
	}

	params := lazydata.LazyDataBytes([]byte(`{"query":"SELECT 1"}`))
	_, err := rc.Execute("run_query", params)
	if err == nil {
		t.Fatal("expected error for missing remote_ctx_token or target configuration")
	}
}

func TestXRemoteExecuteInvalidToken(t *testing.T) {
	rhub := remotehub.NewRemoteHub()
	rc := &RemoteCapability{
		remoteHub: rhub,
	}

	params := lazydata.LazyDataBytes([]byte(`{"remote_ctx_token":"bad-token","query":"SELECT 1"}`))
	_, err := rc.Execute("run_query", params)
	if err == nil {
		t.Fatal("expected error for invalid remote_ctx_token")
	}
}

func TestXRemoteExecuteWithToken(t *testing.T) {
	rhub := remotehub.NewRemoteHub()
	rc := &RemoteCapability{
		remoteHub: rhub,
	}

	claim := &remotehub.RemoteCtxClaim{
		TargetSpaceId:          101,
		TargetPackageId:        202,
		TargetPackageVersionId: 303,
		PluginSpaceId:          404,
		PluginId:               505,
		RequestID:              "req-123",
	}

	token, err := rhub.SignRemoteCtxToken(claim)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	targetSpace, targetPkg, targetVer, err := rc.resolveTarget(lazydata.LazyDataBytes([]byte(`{"remote_ctx_token":"` + token + `"}`)))
	if err != nil {
		t.Fatalf("resolveTarget failed: %v", err)
	}

	if targetSpace != 101 {
		t.Fatalf("expected targetSpace 101, got %d", targetSpace)
	}
	if targetPkg != 202 {
		t.Fatalf("expected targetPkg 202, got %d", targetPkg)
	}
	if targetVer != 303 {
		t.Fatalf("expected targetVer 303, got %d", targetVer)
	}
}

func TestXRemoteExecuteWithStaticConfig(t *testing.T) {
	rhub := remotehub.NewRemoteHub()
	rc := &RemoteCapability{
		remoteHub:               rhub,
		defaultSpaceId:          111,
		defaultPackageId:        222,
		defaultPackageVersionId: 333,
	}

	targetSpace, targetPkg, targetVer, err := rc.resolveTarget(lazydata.LazyDataBytes([]byte(`{}`)))
	if err != nil {
		t.Fatalf("resolveTarget failed: %v", err)
	}

	if targetSpace != 111 {
		t.Fatalf("expected targetSpace 111, got %d", targetSpace)
	}
	if targetPkg != 222 {
		t.Fatalf("expected targetPkg 222, got %d", targetPkg)
	}
	if targetVer != 333 {
		t.Fatalf("expected targetVer 333, got %d", targetVer)
	}
}

func TestNormalizeAction(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"run_query", "run_query"},
		{"/run_query", "run_query"},
		{"db/run_query", "db_run_query"},
		{"db.run_query", "db_run_query"},
		{"kv/get", "kv_get"},
		{"core/publish_event", "core_publish_event"},
	}

	for _, tt := range tests {
		got := normalizeAction(tt.input)
		if got != tt.expected {
			t.Fatalf("normalizeAction(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestXRemoteHandleHttp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rhub := remotehub.NewRemoteHub()
	rc := &RemoteCapability{
		remoteHub: rhub,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Call handle without action
	rc.Handle(c)
	if w.Code != http.StatusBadRequest && !strings.Contains(w.Body.String(), "error") {
		t.Fatalf("expected error response when action is empty, got %s", w.Body.String())
	}

	// Call handle with action and missing token
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Params = gin.Params{gin.Param{Key: "action", Value: "run_query"}}
	c2.Request = httptest.NewRequest(http.MethodPost, "/run_query", strings.NewReader(`{}`))
	rc.Handle(c2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d (%s)", w2.Code, w2.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(w2.Body.Bytes(), &resp)
	if resp["message"] == nil {
		t.Fatal("expected message in error response")
	}
}
