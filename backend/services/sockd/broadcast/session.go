package broadcast

import (
	"net"
	"sync"
	"time"

	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type session struct {
	// Pointer to parent room allows session to signal its own disconnect
	room *Room

	connId int64
	userId int64
	conn   net.Conn

	send   chan []byte
	once   sync.Once
	closed bool
}

func (s *session) writePump() {
	// Safety net: Ensure cleanup happens when the loop exits (connection dies)
	defer func() {
		s.conn.Close()
		// Trigger the Room Event Loop
		s.room.disconnect <- s.connId
	}()

	errCount := 0

	for msg := range s.send {
		err := wsutil.WriteServerText(s.conn, msg)
		if err != nil {

			qq.Println("@writePump/1{ERROR}", err.Error())

			errCount++
			if errCount > 10 {
				s.room.disconnect <- s.connId
				return
			}

			continue

		}
		errCount = 0
	}
}

func (s *session) readPump() {

	errCount := 0

	for {

		if s.closed {
			break
		}

		if errCount > 10 {
			s.room.disconnect <- s.connId
			return
		}

		data, msg, err := wsutil.ReadClientData(s.conn)
		if err != nil {
			errCount++
			return
		}

		errCount = 0

		if msg == ws.OpClose {
			s.room.disconnect <- s.connId
			return
		}

		if msg == ws.OpPing {
			wsutil.WriteServerMessage(s.conn, ws.OpPong, nil)
			continue
		}

		if msg == ws.OpPong {
			continue
		}

		if msg == ws.OpText {
			s.room.broadcast <- data
			continue
		}

		if msg == ws.OpBinary {

			tcan := time.After(time.Second * 5)

			select {
			case s.room.broadcast <- data:
				continue
			case <-tcan:
				qq.Println("@drop_message", s.connId)
				continue
			}

		}

	}
}

func (s *session) teardown() {
	s.once.Do(func() {
		close(s.send)
		s.conn.Close()
		s.closed = true
	})
}
