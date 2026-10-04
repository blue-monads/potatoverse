package engine

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/gin-gonic/gin"
)

func TestRootRouter_MatchDomain(t *testing.T) {
	eng := &Engine{}
	rr := NewRootRouter(eng)

	routes := map[string]*RootRouteTarget{
		"*":              {SpaceId: 12},
		"example.com":    {SpaceId: 11},
		"ee.example":     {SpaceId: 14},
		"localhost":      {SpaceId: 15},
		"*.doomsday.com": {SpaceId: 13},
	}
	rr.setRoutes(routes)

	tests := []struct {
		host        string
		wantSpaceId int64
		wantFound   bool
	}{
		{"example.com", 11, true},
		{"EXAMPLE.COM", 11, true},
		{"example.com:8080", 11, true},
		{"ee.example", 14, true},
		{"ee.example:3000", 14, true},
		{"localhost", 15, true},
		{"localhost:7777", 15, true},
		{"foo.doomsday.com", 13, true},
		{"bar.baz.doomsday.com", 13, true},
		{"foo.doomsday.com:9000", 13, true},
		{"doomsday.com", 13, true},
		{"other-domain.net", 12, true},
		{"", 0, false},
	}

	for _, tt := range tests {
		spaceId, found := rr.MatchDomain(tt.host)
		if found != tt.wantFound {
			t.Errorf("MatchDomain(%q) found = %v, want %v", tt.host, found, tt.wantFound)
		}
		if spaceId != tt.wantSpaceId {
			t.Errorf("MatchDomain(%q) spaceId = %d, want %d", tt.host, spaceId, tt.wantSpaceId)
		}
	}
}

func TestRootRouter_WildcardSpecificity(t *testing.T) {
	eng := &Engine{}
	rr := NewRootRouter(eng)

	routes := map[string]*RootRouteTarget{
		"*.doomsday.com":     {SpaceId: 13},
		"*.sub.doomsday.com": {SpaceId: 20},
		"exact.doomsday.com": {SpaceId: 30},
	}
	rr.setRoutes(routes)

	// Exact match should win over wildcard
	id, ok := rr.MatchDomain("exact.doomsday.com")
	if !ok || id != 30 {
		t.Errorf("expected 30 for exact.doomsday.com, got %d, %v", id, ok)
	}

	// More specific wildcard should win
	id, ok = rr.MatchDomain("app.sub.doomsday.com")
	if !ok || id != 20 {
		t.Errorf("expected 20 for app.sub.doomsday.com, got %d, %v", id, ok)
	}

	// Less specific wildcard
	id, ok = rr.MatchDomain("other.doomsday.com")
	if !ok || id != 13 {
		t.Errorf("expected 13 for other.doomsday.com, got %d, %v", id, ok)
	}
}

func TestRootRouter_ReloadIndex_FromDatabase(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	eng := &Engine{
		db:     db,
		logger: slog.Default(),
	}
	rr := NewRootRouter(eng)

	// Initially empty
	err := rr.ReloadIndex()
	if err != nil {
		t.Fatalf("ReloadIndex failed: %v", err)
	}
	if _, ok := rr.MatchDomain("example.com"); ok {
		t.Fatal("expected no match initially")
	}

	// Insert ROOT_ROUTING_INDEX config into GlobalConfig
	jsonConfig := `{
		"*": {"space_id": 12},
		"example.com": {"space_id": 11},
		"ee.example": 14,
		"localhost": {"space_id": 15},
		"*.doomsday.com": 13
	}`
	_, err = db.GetGlobalOps().AddGlobalConfig(&dbmodels.GlobalConfig{
		Key:       RootRoutingIndexConfigKey,
		GroupName: RootRoutingIndexConfigGroup,
		Value:     jsonConfig,
	})
	if err != nil {
		t.Fatalf("AddGlobalConfig failed: %v", err)
	}

	// Reload index
	err = rr.ReloadIndex()
	if err != nil {
		t.Fatalf("ReloadIndex failed: %v", err)
	}

	// Verify matches
	if id, ok := rr.MatchDomain("example.com"); !ok || id != 11 {
		t.Errorf("expected 11 for example.com, got %d, %v", id, ok)
	}
	if id, ok := rr.MatchDomain("ee.example"); !ok || id != 14 {
		t.Errorf("expected 14 for ee.example, got %d, %v", id, ok)
	}
	if id, ok := rr.MatchDomain("localhost:8080"); !ok || id != 15 {
		t.Errorf("expected 15 for localhost:8080, got %d, %v", id, ok)
	}
	if id, ok := rr.MatchDomain("test.doomsday.com"); !ok || id != 13 {
		t.Errorf("expected 13 for test.doomsday.com, got %d, %v", id, ok)
	}
	if id, ok := rr.MatchDomain("any-other.org"); !ok || id != 12 {
		t.Errorf("expected 12 for any-other.org, got %d, %v", id, ok)
	}
}

func TestRootRouter_Serve_NotFound404(t *testing.T) {
	eng := &Engine{}
	rr := NewRootRouter(eng)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/some/path", nil)
	req.Host = "unknown.example.com"
	c.Request = req

	rr.Serve(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
