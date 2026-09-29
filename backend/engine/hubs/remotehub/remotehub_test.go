package remotehub

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/gin-gonic/gin"
)

func TestRContextInterfaces(t *testing.T) {
	// Verify that both HttpBindContext and SimpleRContext implement RContext
	var _ RContext = (*HttpBindContext)(nil)
	var _ RContext = (*SimpleRContext)(nil)
}

func TestSimpleRContext(t *testing.T) {
	ctx := &SimpleRContext{
		SpaceId:        10,
		PackageId:      20,
		PackageVersion: 30,
		Data:           []byte(`{"key":"hello"}`),
		Meta: map[string]string{
			"param1": "value1",
		},
	}

	if ctx.GetSpaceId() != 10 {
		t.Fatalf("expected SpaceId 10, got %d", ctx.GetSpaceId())
	}
	if ctx.GetPackageId() != 20 {
		t.Fatalf("expected PackageId 20, got %d", ctx.GetPackageId())
	}
	if ctx.GetPackageVersion() != 30 {
		t.Fatalf("expected PackageVersion 30, got %d", ctx.GetPackageVersion())
	}

	data, err := ctx.GetData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `{"key":"hello"}` {
		t.Fatalf("expected data %s, got %s", `{"key":"hello"}`, string(data))
	}

	val, err := ctx.GetMeta("param1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "value1" {
		t.Fatalf("expected value1, got %s", val)
	}

	_, err = ctx.GetMeta("missing")
	if err == nil {
		t.Fatal("expected error for missing meta key")
	}

	err = ctx.SetData([]byte("response"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(ctx.ResponseData) != "response" {
		t.Fatalf("expected response, got %s", string(ctx.ResponseData))
	}
}

func TestNewCapRContext(t *testing.T) {
	// Import lazydata
	params := lazydata.LazyDataBytes([]byte(`{"foo":"bar","num":42}`))
	capCtx, err := NewCapRContext(1, 2, 3, params)
	if err != nil {
		t.Fatalf("unexpected error creating CapRContext: %v", err)
	}

	if capCtx.GetSpaceId() != 1 {
		t.Fatalf("expected SpaceId 1, got %d", capCtx.GetSpaceId())
	}
	if capCtx.GetPackageId() != 2 {
		t.Fatalf("expected PackageId 2, got %d", capCtx.GetPackageId())
	}
	if capCtx.GetPackageVersion() != 3 {
		t.Fatalf("expected PackageVersion 3, got %d", capCtx.GetPackageVersion())
	}

	val, err := capCtx.GetMeta("foo")
	if err != nil || val != "bar" {
		t.Fatalf("expected meta foo=bar, got %s (err: %v)", val, err)
	}
}


func TestHttpBindContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody := `{"action":"test"}`
	req, err := http.NewRequest(http.MethodPost, "/test?queryKey=queryVal", bytes.NewBufferString(reqBody))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req.Header.Set("X-Custom-Header", "headerVal")
	c.Request = req

	// Add a path param
	c.Params = gin.Params{gin.Param{Key: "pathParam", Value: "paramVal"}}

	hCtx := &HttpBindContext{
		Http:           c,
		PackageId:      100,
		PackageVersion: 200,
		SpaceId:        300,
		RequestId:      "req-1",
	}

	if hCtx.GetSpaceId() != 300 {
		t.Fatalf("expected SpaceId 300, got %d", hCtx.GetSpaceId())
	}
	if hCtx.GetPackageId() != 100 {
		t.Fatalf("expected PackageId 100, got %d", hCtx.GetPackageId())
	}
	if hCtx.GetPackageVersion() != 200 {
		t.Fatalf("expected PackageVersion 200, got %d", hCtx.GetPackageVersion())
	}

	// Test GetData
	data, err := hCtx.GetData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != reqBody {
		t.Fatalf("expected data %s, got %s", reqBody, string(data))
	}

	// Call GetData second time (should return cached data)
	data2, err := hCtx.GetData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data2) != reqBody {
		t.Fatalf("expected cached data %s, got %s", reqBody, string(data2))
	}

	// Test GetMeta with param
	val, err := hCtx.GetMeta("pathParam")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "paramVal" {
		t.Fatalf("expected paramVal, got %s", val)
	}

	// Test GetMeta with query
	val, err = hCtx.GetMeta("queryKey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "queryVal" {
		t.Fatalf("expected queryVal, got %s", val)
	}

	// Test GetMeta with header
	val, err = hCtx.GetMeta("X-Custom-Header")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "headerVal" {
		t.Fatalf("expected headerVal, got %s", val)
	}

	// Test GetMeta with missing key
	_, err = hCtx.GetMeta("missingKey")
	if err == nil {
		t.Fatal("expected error for missing key")
	}

	// Test SetData
	err = hCtx.SetData([]byte("out"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(hCtx.GetResponseData()) != "out" {
		t.Fatalf("expected out, got %s", string(hCtx.GetResponseData()))
	}
}

func TestRemoteHubTokenAndAuthed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rhub := NewRemoteHub()
	if rhub == nil {
		t.Fatal("expected rhub not nil")
	}

	token, err := rhub.GetExecToken(12, 34, 56, "req-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected token not empty")
	}

	r := gin.New()
	var capturedSpaceId int64
	var capturedPkgId int64
	var capturedVerId int64
	var capturedMeta string
	var capturedData []byte

	handler := func(ctx RContext) (any, error) {
		capturedSpaceId = ctx.GetSpaceId()
		capturedPkgId = ctx.GetPackageId()
		capturedVerId = ctx.GetPackageVersion()
		capturedMeta, _ = ctx.GetMeta("myparam")
		capturedData, _ = ctx.GetData()
		return map[string]string{"status": "ok"}, nil
	}

	r.POST("/test/:myparam", rhub.Authed(handler))

	body := `{"msg":"hi"}`
	req, _ := http.NewRequest(http.MethodPost, "/test/apple", bytes.NewBufferString(body))
	req.Header.Set(XExecHeader, token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d (%s)", w.Code, w.Body.String())
	}
	if capturedSpaceId != 56 {
		t.Fatalf("expected capturedSpaceId 56, got %d", capturedSpaceId)
	}
	if capturedPkgId != 12 {
		t.Fatalf("expected capturedPkgId 12, got %d", capturedPkgId)
	}
	if capturedVerId != 34 {
		t.Fatalf("expected capturedVerId 34, got %d", capturedVerId)
	}
	if capturedMeta != "apple" {
		t.Fatalf("expected capturedMeta 'apple', got '%s'", capturedMeta)
	}
	if string(capturedData) != body {
		t.Fatalf("expected capturedData %s, got %s", body, string(capturedData))
	}

	var resp map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", resp["status"])
	}
}

func TestRemoteCtxToken(t *testing.T) {
	rhub := NewRemoteHub()
	if rhub == nil {
		t.Fatal("expected rhub not nil")
	}

	claim := &RemoteCtxClaim{
		TargetSpaceId:          100,
		TargetPackageId:        200,
		TargetPackageVersionId: 300,
		PluginSpaceId:          400,
		PluginId:               500,
		RequestID:              "req-abc",
	}

	token, err := rhub.SignRemoteCtxToken(claim)
	if err != nil {
		t.Fatalf("unexpected error signing token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	parsed, err := rhub.ParseRemoteCtxToken(token)
	if err != nil {
		t.Fatalf("unexpected error parsing token: %v", err)
	}

	if parsed.TargetSpaceId != 100 {
		t.Fatalf("expected TargetSpaceId 100, got %d", parsed.TargetSpaceId)
	}
	if parsed.TargetPackageId != 200 {
		t.Fatalf("expected TargetPackageId 200, got %d", parsed.TargetPackageId)
	}
	if parsed.TargetPackageVersionId != 300 {
		t.Fatalf("expected TargetPackageVersionId 300, got %d", parsed.TargetPackageVersionId)
	}
	if parsed.PluginSpaceId != 400 {
		t.Fatalf("expected PluginSpaceId 400, got %d", parsed.PluginSpaceId)
	}
	if parsed.PluginId != 500 {
		t.Fatalf("expected PluginId 500, got %d", parsed.PluginId)
	}
	if parsed.RequestID != "req-abc" {
		t.Fatalf("expected RequestID req-abc, got %s", parsed.RequestID)
	}

	// Test invalid token
	_, err = rhub.ParseRemoteCtxToken("invalid-garbage-token")
	if err == nil {
		t.Fatal("expected error parsing invalid token")
	}
}
