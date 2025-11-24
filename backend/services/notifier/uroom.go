package notifier

import "sync"

type UserRoom struct {
	userId      int64
	group       string
	maxMsgId    int64
	connections map[int64]*Connection
	mu          sync.RWMutex
}
