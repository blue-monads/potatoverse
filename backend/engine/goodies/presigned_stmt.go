package goodies

type PresignedStmtHub struct {
	stmts map[string]*PresignedStmt
}

/*

insert into Events (name, description, tenant_id, created_at) values (?, ?, ?, {{tenant_id}})

*/

type PresignedClaim struct {
	PresignedIds []string
	PinnedParams map[string]any
}

type PresignedStmt struct {
	Stmt    string
	Params  map[string]any
	Selects []string

	// insert, batch_insert, update, delete, query
	Mode            string
	TTL             int64
	LastRefreshedAt int64
}
