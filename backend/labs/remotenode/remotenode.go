package eventhub

type RemoteNode interface {
	Start() error
}

/*

/zz/remote/sysevent
/zz/remote/dbhub/
/zz/remote/engine/

*/

type SystemEventHub struct {
	nodes map[string]RemoteNode
}
