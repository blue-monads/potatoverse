package room

import (
	"net"
	"sync"
	"time"

	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/tidwall/gjson"
)

type session struct {
	room *Room

	connId ConnId
	userId int64
	conn   net.Conn

	send             chan []byte
	once             sync.Once
	closedAndCleaned bool
}

func (s *session) writePump() {

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
			qq.Println("@writePump/1{ERROR}", err.Error())
			errCount++
			if errCount > 10 {
				s.room.disconnect <- s.connId
				return
			}
			continue
		}
		errCount = 0

		if s.closedAndCleaned {
			return
		}

	}
}

func (s *session) readPump() {
	errCount := 0

	for {
		if s.closedAndCleaned {
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
			s.handleMessage(data)
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

func (s *session) handleMessage(data []byte) {

	msgType := gjson.GetBytes(data, "type").String()
	msgTopic := gjson.GetBytes(data, "topic").String()
	msgTarget := ConnId(gjson.GetBytes(data, "target").String())

	switch msgType {
	case ClientMessageTypeBroadcast:
		tcan := time.After(time.Second * 5)
		select {
		case s.room.broadcast <- data:
		case <-tcan:
			qq.Println("@drop_message", s.connId)
		}

	case ClientMessageTypePublish:
		if msgTopic == "" {
			qq.Println("@handleMessage/missing_topic", s.connId)
			return
		}

		tcan := time.After(time.Second * 5)
		select {
		case s.room.publish <- publishEvent{
			topic:   msgTopic,
			message: data,
		}:
		case <-tcan:
			qq.Println("@drop_message", s.connId)
		}

	case ClientMessageTypeDirectMessage:
		if msgTarget == "" {
			qq.Println("@handleMessage/missing_target", s.connId)
			return
		}

		tcan := time.After(time.Second * 5)

		select {
		case s.room.directMsg <- directMessageEvent{
			targetConnId: msgTarget,
			message:      data,
		}:
		case <-tcan:
			qq.Println("@drop_message", s.connId)
		}

	case ClientMessageTypeGetPresence:
		s.room.notifyPresenceUser(s.connId, msgTopic, s.userId)
	case ClientMessageTypeCommand:
		s.room.cmdChan <- Message{
			Type:  msgType,
			Data:  data,
			Topic: msgTopic,
		}

	default:
		qq.Println("@handleMessage/unknown_type", msgType, s.connId)
	}
}

func (s *session) teardown() {
	s.once.Do(func() {
		s.send <- nil
		s.conn.Close()
		s.closedAndCleaned = true

		s.room.onDisconnect <- UserConnInfo{
			ConnId: s.connId,
			UserId: s.userId,
		}

	})
}
