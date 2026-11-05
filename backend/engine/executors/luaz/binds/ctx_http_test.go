package binds

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/remotehub"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

type mockTestApp struct {
	xtypes.App
	sig    *signer.Signer
	engine any
}

func (m *mockTestApp) Signer() *signer.Signer {
	return m.sig
}

func (m *mockTestApp) Engine() any {
	return m.engine
}

type mockTestEngine struct {
	xtypes.Engine
	rhub *remotehub.RemoteHub
}

func (m *mockTestEngine) GetRemoteHub() *remotehub.RemoteHub {
	return m.rhub
}

func setupTestCtxHttp(t *testing.T, spaceId int64, authSpaceId int64, userId int64) (*lua.LState, *luaHttpRequestContext, *remotehub.RemoteHub, string) {
	gin.SetMode(gin.TestMode)
	L := lua.NewState()
	L.OpenLibs()

	secret := "test-secret-key-must-be-long-enough-for-signing"
	sig := signer.New([]byte(secret))
	rhub := remotehub.NewRemoteHub()

	app := &mockTestApp{
		sig:    sig,
		engine: &mockTestEngine{rhub: rhub},
	}

	userToken, err := sig.SignSpace(&signer.SpaceClaim{
		UserId:  userId,
		SpaceId: authSpaceId,
	})
	if err != nil {
		t.Fatalf("failed to sign space claim: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	c.Request = req

	reqUserData := NewHttpRequestContext(L, app, spaceId, c)
	L.SetGlobal("req", reqUserData)

	reqCtx := reqUserData.Value.(*luaHttpRequestContext)

	return L, reqCtx, rhub, userToken
}

func TestGetClaim_StandardSpace(t *testing.T) {
	L, _, _, _ := setupTestCtxHttp(t, 100, 100, 42)
	defer L.Close()

	err := L.DoString(`
		local claim, err = req.get_claim()
		assert(err == nil, "expected no error, got: " .. tostring(err))
		assert(claim ~= nil, "expected claim")
		assert(claim.u == 42, "expected u 42")
		assert(claim.s == 100, "expected s 100")
	`)
	if err != nil {
		t.Fatalf("Lua execution failed: %v", err)
	}
}

func TestGetClaim_MismatchWithoutToken(t *testing.T) {
	// Plugin is space 200, but user token is for space 100
	L, _, _, _ := setupTestCtxHttp(t, 200, 100, 42)
	defer L.Close()

	err := L.DoString(`
		local claim, err = req.get_claim()
		assert(claim == nil, "expected nil claim")
		assert(err ~= nil, "expected error for space mismatch")
	`)
	if err != nil {
		t.Fatalf("Lua execution failed: %v", err)
	}
}

func TestGetClaim_WithValidRemoteCtxToken(t *testing.T) {
	// Plugin is space 200, host space is 100
	L, _, rhub, _ := setupTestCtxHttp(t, 200, 100, 42)
	defer L.Close()

	token, err := rhub.SignRemoteCtxToken(&remotehub.RemoteCtxClaim{
		TargetSpaceId:   100,
		TargetPackageId: 10,
		PluginSpaceId:   200,
		PluginId:        1,
	})
	if err != nil {
		t.Fatalf("failed to sign remote ctx token: %v", err)
	}

	L.SetGlobal("remote_token", lua.LString(token))

	// Test dot syntax with remote_ctx_token
	err = L.DoString(`
		local claim, err = req.get_claim(remote_token)
		assert(err == nil, "expected no error, got: " .. tostring(err))
		assert(claim ~= nil, "expected claim")
		assert(claim.u == 42, "expected u 42")
		assert(claim.s == 100, "expected s 100")

		local uid, uerr = req.get_user_id(remote_token)
		assert(uerr == nil, "expected no error, got: " .. tostring(uerr))
		assert(uid == 42, "expected uid 42")
	`)
	if err != nil {
		t.Fatalf("Lua execution failed: %v", err)
	}

	// Test colon syntax
	err = L.DoString(`
		local claim, err = req:get_claim(remote_token)
		assert(err == nil, "expected no error, got: " .. tostring(err))
		assert(claim ~= nil, "expected claim")
		assert(claim.u == 42, "expected u 42")

		local uid, uerr = req:get_user_id(remote_token)
		assert(uerr == nil, "expected no error")
		assert(uid == 42, "expected uid 42")
	`)
	if err != nil {
		t.Fatalf("Lua colon syntax execution failed: %v", err)
	}
}

func TestGetClaim_WithTamperedRemoteCtxToken(t *testing.T) {
	// Plugin is space 200, but token was issued for plugin space 999
	L, _, rhub, _ := setupTestCtxHttp(t, 200, 100, 42)
	defer L.Close()

	token, err := rhub.SignRemoteCtxToken(&remotehub.RemoteCtxClaim{
		TargetSpaceId:   100,
		TargetPackageId: 10,
		PluginSpaceId:   999, // wrong plugin space
		PluginId:        1,
	})
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	L.SetGlobal("remote_token", lua.LString(token))

	err = L.DoString(`
		local claim, err = req.get_claim(remote_token)
		assert(claim == nil, "expected nil claim for wrong plugin space")
		assert(err ~= nil, "expected error")
	`)
	if err != nil {
		t.Fatalf("Lua execution failed: %v", err)
	}
}

func TestGetClaim_NoImplicitAutoDetect(t *testing.T) {
	// Plugin is space 200, host space is 100
	L, reqCtx, rhub, _ := setupTestCtxHttp(t, 200, 100, 42)
	defer L.Close()

	token, err := rhub.SignRemoteCtxToken(&remotehub.RemoteCtxClaim{
		TargetSpaceId:   100,
		TargetPackageId: 10,
		PluginSpaceId:   200,
		PluginId:        1,
	})
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	reqCtx.ctx.Set("remote_ctx_token", token)
	L.SetGlobal("remote_token", lua.LString(token))

	// Without argument, it defaults to current space (200) and fails because user claim is for 100
	err = L.DoString(`
		local claim, err = req.get_claim()
		assert(claim == nil, "expected nil claim when token not passed")
		assert(err == "invalid space id", "expected invalid space id, got: " .. tostring(err))

		-- Explicitly passing the token succeeds
		local validClaim, verr = req.get_claim(remote_token)
		assert(verr == nil, "expected no error with explicit token, got: " .. tostring(verr))
		assert(validClaim ~= nil, "expected claim")
		assert(validClaim.u == 42, "expected u 42")
	`)
	if err != nil {
		t.Fatalf("Lua execution failed: %v", err)
	}
}

func TestGetClaim_RejectRawSpaceId(t *testing.T) {
	L, _, _, _ := setupTestCtxHttp(t, 200, 100, 42)
	defer L.Close()

	// Passing a non-token string space id should return invalid remote_ctx_token error
	err := L.DoString(`
		local claim, err = req.get_claim("100")
		assert(claim == nil, "expected nil claim when passing string space id")
		assert(err == "invalid remote_ctx_token", "expected invalid remote_ctx_token error, got: " .. tostring(err))

		local uid, uerr = req.get_user_id("100")
		assert(uid == nil, "expected nil uid when passing string space id")
		assert(uerr == "invalid remote_ctx_token", "expected invalid remote_ctx_token error, got: " .. tostring(uerr))

		-- Passing a number should fail because remote_ctx_token must be a string
		local ok, callErr = pcall(function() req.get_claim(100) end)
		assert(not ok, "expected number argument to be rejected")
	`)
	if err != nil {
		t.Fatalf("Lua execution failed: %v", err)
	}
}
