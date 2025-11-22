package sockd

import (
	"net"
	"sync"

	"github.com/gobwas/ws/wsutil"
)

type session struct {
	// Pointer to parent room allows session to signal its own disconnect
	room *Room

	connId int32
	userId int64
	conn   net.Conn

	send chan []byte
	once sync.Once
}

func (s *session) writePump() {
	// Safety net: Ensure cleanup happens when the loop exits (connection dies)
	defer func() {
		s.conn.Close()
		// Trigger the Room Event Loop
		s.room.disconnect <- s.connId
	}()

	for msg := range s.send {
		err := wsutil.WriteServerText(s.conn, msg)
		if err != nil {
			// Error writing (broken pipe, etc.) -> Loop breaks -> Defer runs
			return
		}
	}
}

func (s *session) teardown() {
	s.once.Do(func() {
		close(s.send)
		s.conn.Close()
	})
}
