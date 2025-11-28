package eventhub

type RemoteNode interface {
	Start() error
}

/*

/zz/remote/sysevent
/zz/remote/dbhub/
/zz/remote/engine/

exec stretegy ->
- always_remote
- pefer_remote
- always_local
- perfer_local

*/

type SystemEventHub struct {
	nodes map[string]RemoteNode
}
