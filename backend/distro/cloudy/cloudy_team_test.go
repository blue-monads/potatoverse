package cloudy

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestTeamStructure(t *testing.T) {
	dir, err := os.MkdirTemp("", "cloudy_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	dbPath := filepath.Join(dir, "test.db")
	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer sqlDB.Close()

	store, err := newStore(sqlDB)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	if err := store.migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Verify all 4 tables exist
	tables := []string{"CloudyUsers", "CloudyTeams", "CloudyTeamMembers", "CloudyTeamPotatoInstances"}
	for _, tbl := range tables {
		var count int
		err := sqlDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tbl).Scan(&count)
		if err != nil {
			t.Fatalf("check table %s: %v", tbl, err)
		}
		if count == 0 {
			t.Fatalf("expected table %s to exist", tbl)
		}
	}

	// Insert user with team (user is added to CloudyTeamMembers, no instance created upon signup)
	user, team, err := store.insertUserWithTeam("Alice Smith", "alice@example.com", "secret_hash", "acme-team", UTypeNormal, true)
	if err != nil {
		t.Fatalf("insert user with team: %v", err)
	}

	if user.ID == 0 || user.Email != "alice@example.com" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if team.ID == 0 || team.OwnerID != user.ID || team.Name != "acme-team" {
		t.Fatalf("unexpected team: %+v", team)
	}

	// Verify user is added to the team (CloudyTeamMembers)
	inTeam, err := store.isUserInTeam(user.ID, team.ID)
	if err != nil || !inTeam {
		t.Fatalf("expected user %d to be in team %d", user.ID, team.ID)
	}

	// Verify user's teams list contains team
	uTeams, err := store.getTeamsForUser(user.ID)
	if err != nil || len(uTeams) != 1 || uTeams[0].ID != team.ID {
		t.Fatalf("expected user teams to contain team %d, got %+v", team.ID, uTeams)
	}

	// Verify no instance exists yet for user
	pInst, _, err := store.getPrimaryInstanceForUser(user.ID)
	if err != nil {
		t.Fatalf("get primary instance error: %v", err)
	}
	if pInst != nil {
		t.Fatalf("expected no instance upon signup, got %+v", pInst)
	}

	// User creates subapp instance for team (instance is added to team in CloudyTeamPotatoInstances)
	inst, err := store.createPotatoInstance(team.ID, "acme-app", "Acme Main App")
	if err != nil {
		t.Fatalf("create potato instance: %v", err)
	}
	if inst.ID == 0 || inst.TeamID != team.ID || inst.Slug != "acme-app" {
		t.Fatalf("unexpected instance: %+v", inst)
	}

	// Verify instance is linked to team
	tInsts, err := store.getPotatoInstancesByTeamID(team.ID)
	if err != nil || len(tInsts) != 1 || tInsts[0].ID != inst.ID {
		t.Fatalf("expected team instances to contain instance %d, got %+v", inst.ID, tInsts)
	}

	// Lookup primary instance
	pInst, pTeam, err := store.getPrimaryInstanceForUser(user.ID)
	if err != nil {
		t.Fatalf("get primary instance: %v", err)
	}
	if pInst == nil || pInst.Slug != "acme-app" {
		t.Fatalf("expected primary instance slug acme-app, got %+v", pInst)
	}
	if pTeam == nil || pTeam.ID != team.ID {
		t.Fatalf("expected primary team id %d, got %+v", team.ID, pTeam)
	}
}
