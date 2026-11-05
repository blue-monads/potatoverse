package cloudy

import (
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
)

func (a *CloudyApp) Migrate() error {
	if err := a.initStore(); err != nil {
		return err
	}
	if err := a.store.migrate(); err != nil {
		return err
	}
	if err := a.tursoDB.Push(a.rootCtx); err != nil {
		return fmt.Errorf("push migrated schema: %w", err)
	}
	log.Println("cloudy migrations applied")
	return nil
}

func (a *CloudyApp) initStore() error {
	if a.store != nil {
		return nil
	}
	sqlDB, err := a.tursoDB.Connect(a.rootCtx)
	if err != nil {
		return err
	}
	store, err := newStore(sqlDB)
	if err != nil {
		return err
	}
	a.store = store
	return nil
}

func (s *Store) migrate() error {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	if err := s.execScript(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name TEXT PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return err
	}

	applied, err := s.appliedMigrations()
	if err != nil {
		return err
	}

	for _, name := range names {
		if applied[name] {
			continue
		}
		body, err := fs.ReadFile(migrationFS, "migrations/"+name)
		if err != nil {
			return err
		}
		log.Println("applying migration", name)
		if err := s.execScript(string(body)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := s.sql.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, name); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) appliedMigrations() (map[string]bool, error) {
	rows, err := s.sql.Query(`SELECT name FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out[name] = true
	}
	return out, rows.Err()
}

func (s *Store) execScript(script string) error {
	for _, stmt := range splitSQL(script) {
		if _, err := s.sql.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func splitSQL(script string) []string {
	var stmts []string
	for _, part := range strings.Split(script, ";") {
		stmt := strings.TrimSpace(part)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		stmts = append(stmts, stmt)
	}
	return stmts
}
