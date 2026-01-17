package enforcer

import (
	"strings"
	"testing"
)

func TestTransformQuery(t *testing.T) {
	ownerType := "user"
	ownerID := "123"
	expectedPrefix := "zz_user__123__"

	tests := []struct {
		name             string
		input            string
		shouldContain    []string // Table names that should be present in output
		shouldNotContain []string // Table names that should NOT be present in output
		wantErr          bool
	}{
		{
			name:             "SELECT with single table",
			input:            "SELECT * FROM users",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{"FROM users", "FROM \"users\""},
			wantErr:          false,
		},
		{
			name:             "SELECT with WHERE clause",
			input:            "SELECT id, name FROM users WHERE id = 1",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{"FROM users", "FROM \"users\""},
			wantErr:          false,
		},
		{
			name:             "SELECT with JOIN",
			input:            "SELECT u.id, p.title FROM users u JOIN posts p ON u.id = p.user_id",
			shouldContain:    []string{expectedPrefix + "users", expectedPrefix + "posts"},
			shouldNotContain: []string{"FROM users", "JOIN posts"},
			wantErr:          false,
		},
		{
			name:             "SELECT with LEFT JOIN",
			input:            "SELECT * FROM users LEFT JOIN posts ON users.id = posts.user_id",
			shouldContain:    []string{expectedPrefix + "users", expectedPrefix + "posts"},
			shouldNotContain: []string{"FROM users", "JOIN posts"},
			wantErr:          false,
		},
		{
			name:             "SELECT with qualified column reference",
			input:            "SELECT users.name, posts.title FROM users, posts",
			shouldContain:    []string{expectedPrefix + "users", expectedPrefix + "posts"},
			shouldNotContain: []string{"users.name", "posts.title", "FROM users", "FROM posts"},
			wantErr:          false,
		},
		{
			name:             "INSERT statement",
			input:            "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{"INTO users", "INTO \"users\""},
			wantErr:          false,
		},
		{
			name:             "UPDATE statement",
			input:            "UPDATE users SET name = 'Jane' WHERE id = 1",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{"UPDATE users", "UPDATE \"users\""},
			wantErr:          false,
		},
		{
			name:             "DELETE statement",
			input:            "DELETE FROM users WHERE id = 1",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{"FROM users", "FROM \"users\""},
			wantErr:          false,
		},
		{
			name:             "CREATE TABLE statement",
			input:            "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{"TABLE users", "TABLE \"users\""},
			wantErr:          false,
		},
		{
			name:             "CREATE TABLE IF NOT EXISTS",
			input:            "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY)",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{"TABLE users", "TABLE \"users\""},
			wantErr:          false,
		},
		{
			name:             "CREATE VIRTUAL TABLE statement",
			input:            "CREATE VIRTUAL TABLE EventLocations USING geopoly (event_id)",
			shouldContain:    []string{expectedPrefix + "EventLocations"},
			shouldNotContain: []string{"TABLE EventLocations", "TABLE \"EventLocations\""},
			wantErr:          false,
		},
		{
			name:             "CREATE VIRTUAL TABLE IF NOT EXISTS",
			input:            "CREATE VIRTUAL TABLE IF NOT EXISTS FeatureLocations USING geopoly (feature_id)",
			shouldContain:    []string{expectedPrefix + "FeatureLocations"},
			shouldNotContain: []string{"TABLE FeatureLocations", "TABLE \"FeatureLocations\""},
			wantErr:          false,
		},
		{
			name:             "CREATE INDEX statement",
			input:            "CREATE INDEX idx_name ON users (name)",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{"ON users", "ON \"users\""},
			wantErr:          false,
		},
		{
			name:             "CREATE UNIQUE INDEX",
			input:            "CREATE UNIQUE INDEX idx_email ON users (email)",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{"ON users", "ON \"users\""},
			wantErr:          false,
		},
		{
			name:             "DROP TABLE statement",
			input:            "DROP TABLE users",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{"TABLE users", "TABLE \"users\""},
			wantErr:          false,
		},
		{
			name:             "DROP TABLE IF EXISTS",
			input:            "DROP TABLE IF EXISTS users",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{"TABLE users", "TABLE \"users\""},
			wantErr:          false,
		},
		{
			name:             "ALTER TABLE RENAME",
			input:            "ALTER TABLE users RENAME TO new_users",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{"TABLE users", "TABLE \"users\""},
			wantErr:          false,
		},
		{
			name:             "SELECT with already scoped table (should not double-scope)",
			input:            "SELECT * FROM " + expectedPrefix + "users",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{expectedPrefix + expectedPrefix + "users"},
			wantErr:          false,
		},
		{
			name:             "SELECT with multiple tables",
			input:            "SELECT * FROM users, posts, comments",
			shouldContain:    []string{expectedPrefix + "users", expectedPrefix + "posts", expectedPrefix + "comments"},
			shouldNotContain: []string{"FROM users", "FROM posts", "FROM comments"},
			wantErr:          false,
		},
		{
			name:             "SELECT with subquery",
			input:            "SELECT * FROM users WHERE id IN (SELECT user_id FROM posts)",
			shouldContain:    []string{expectedPrefix + "users"},
			shouldNotContain: []string{"FROM users"},
			wantErr:          false,
			// Note: Subquery table transformation may depend on parser implementation
		},
		{
			name:             "INSERT with SELECT",
			input:            "INSERT INTO users SELECT * FROM temp_users",
			shouldContain:    []string{expectedPrefix + "users", expectedPrefix + "temp_users"},
			shouldNotContain: []string{"INTO users", "FROM temp_users"},
			wantErr:          false,
		},
		{
			name:             "CREATE TABLE with FOREIGN KEY",
			input:            "CREATE TABLE orders (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, FOREIGN KEY (user_id) REFERENCES users(id))",
			shouldContain:    []string{expectedPrefix + "orders", expectedPrefix + "users"},
			shouldNotContain: []string{"TABLE orders", "REFERENCES users"},
			wantErr:          false,
		},
		{
			name:             "CREATE TABLE with FOREIGN KEY and multiple columns",
			input:            "CREATE TABLE order_items (id INTEGER PRIMARY KEY, order_id INTEGER NOT NULL, product_id INTEGER NOT NULL, FOREIGN KEY (order_id) REFERENCES orders(id), FOREIGN KEY (product_id) REFERENCES products(id))",
			shouldContain:    []string{expectedPrefix + "order_items", expectedPrefix + "orders", expectedPrefix + "products"},
			shouldNotContain: []string{"TABLE order_items", "REFERENCES orders", "REFERENCES products"},
			wantErr:          false,
		},
		{
			name:    "Invalid SQL",
			input:   "SELECT * FROM",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transformQuery(ownerType, ownerID, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("transformQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotLower := strings.ToLower(got)
				// Check that all expected table names are present
				for _, expected := range tt.shouldContain {
					if !strings.Contains(gotLower, strings.ToLower(expected)) {
						t.Errorf("transformQuery() output should contain %q, got: %v", expected, got)
					}
				}
				// Check that unscoped table names are not present
				for _, notExpected := range tt.shouldNotContain {
					if strings.Contains(gotLower, strings.ToLower(notExpected)) {
						t.Errorf("transformQuery() output should not contain %q, got: %v", notExpected, got)
					}
				}
			}
		})
	}
}

func TestTransformQueryDifferentOwners(t *testing.T) {
	tests := []struct {
		name          string
		ownerType     string
		ownerID       string
		input         string
		shouldContain []string
	}{
		{
			name:          "space owner",
			ownerType:     "space",
			ownerID:       "456",
			input:         "SELECT * FROM users",
			shouldContain: []string{"zz_space__456__users"},
		},
		{
			name:          "user owner with special characters",
			ownerType:     "user",
			ownerID:       "abc-123",
			input:         "SELECT * FROM posts",
			shouldContain: []string{"zz_user__abc-123__posts"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transformQuery(tt.ownerType, tt.ownerID, tt.input)
			if err != nil {
				t.Errorf("transformQuery() error = %v", err)
				return
			}
			gotLower := strings.ToLower(got)
			for _, expected := range tt.shouldContain {
				if !strings.Contains(gotLower, strings.ToLower(expected)) {
					t.Errorf("transformQuery() output should contain %q, got: %v", expected, got)
				}
			}
		})
	}
}
