package pubsub

import (
	"net"
	"sync"

	"github.com/gobwas/ws/wsutil"
)

type session struct {
	// Pointer to parent room allows session to signal its own disconnect
	room *Room

	connId int64
	userId int64
	conn   net.Conn

	send             chan []byte
	once             sync.Once
	closedAndCleaned bool
}

func (s *session) writePump() {
	// Safety net: Ensure cleanup happens when the loop exits (connection dies)
	defer func() {
		s.conn.Close()

		if !s.closedAndCleaned {
			s.room.disconnect <- s.connId
		}
	}()

	errCount := 0

	for msg := range s.send {

		if msg == nil {
			return
		}

		if errCount > 10 {
			s.room.disconnect <- s.connId
			return
		}

		err := wsutil.WriteServerText(s.conn, msg)
		if err != nil {

			errCount++
			if errCount > 10 {
				s.room.disconnect <- s.connId
				return
			}

			continue

		}

		if s.closedAndCleaned {
			return
		}

		errCount = 0
	}
}

func (s *session) teardown() {
	s.once.Do(func() {

		s.send <- nil

		s.conn.Close()
		s.closedAndCleaned = true
	})
}
