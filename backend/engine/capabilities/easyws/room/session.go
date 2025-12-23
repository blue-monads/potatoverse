package room

import (
	"io"
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

	qq.Println("@writePump/1", s.connId, s.userId)

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

	qq.Println("@readPump/0", s.connId, s.userId)

	errCount := 0

	for {
		if s.closedAndCleaned {
			break
		}

		if errCount > 3 {
			s.room.disconnect <- s.connId
			return
		}

		qq.Println("@readPump/1", s.connId, s.userId)

		data, msg, err := wsutil.ReadClientData(s.conn)
		if err != nil {
			// check if the error is io.EOF
			if err == io.EOF {
				s.room.disconnect <- s.connId
				return
			}

			errCount++
			return
		}

		qq.Println("@readPump/2", s.connId, s.userId, data, msg)

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
	msgFromConnId := gjson.GetBytes(data, "from_cid").String()

	if msgFromConnId == "" || msgFromConnId != string(s.connId) {
		qq.Println("@wrong_from_cid", s.connId, msgFromConnId)
		return
	}

	switch msgType {
	case ClientMessageTypeBroadcast:
		tcan := time.After(time.Second * 5)
		select {
		case s.room.broadcast <- data:
		case <-tcan:
			qq.Println("@drop_message", s.connId)
		}

	case ClientMessageTypePublish:
		msgTopic := gjson.GetBytes(data, "topic").String()
		if msgTopic == "" {
			qq.Println("@handleMessage/missing_topic", s.connId)
			return
		}

		tcan := time.After(time.Second * 5)
		select {
		case s.room.publish <- publishEvent{
			topic:   msgTopic,
			message: data,
			connId:  s.connId,
		}:
		case <-tcan:
			qq.Println("@drop_message", s.connId)
		}

	case ClientMessageTypeDirectMessage:
		msgToConnId := gjson.GetBytes(data, "to_cid").String()
		if msgToConnId == "" {
			qq.Println("@handleMessage/missing_to_cid", s.connId)
			return
		}

		tcan := time.After(time.Second * 5)

		select {
		case s.room.directMsg <- directMessageEvent{
			targetConnId: ConnId(msgToConnId),
			message:      data,
		}:
		case <-tcan:
			qq.Println("@drop_message", s.connId)
		}

	case ClientMessageTypeGetPresence:
		msgTopic := gjson.GetBytes(data, "topic").String()
		s.room.notifyPresenceUser(s.connId, msgTopic, s.userId)
	case ClientMessageTypeCommand:

		if s.room.onCommand == nil {
			qq.Println("@handleMessage/onCommand_nil", s.connId)
			return
		}

		msgSubType := gjson.GetBytes(data, "sub_type").String()

		s.room.onCommand(CommandMessage{
			SubType:    msgSubType,
			RawData:    data,
			FromConnId: s.connId,
		})

	default:
		qq.Println("@handleMessage/unknown_type", msgType, s.connId)
	}
}

func (s *session) teardown() {
	s.once.Do(func() {
		s.send <- nil
		s.conn.Close()
		s.closedAndCleaned = true

		qq.Println("@teardown/1", s.connId, s.userId)

		if s.room.onDisconnect != nil {
			qq.Println("@teardown/2", s.connId, s.userId)
			err := s.room.onDisconnect(DisconnectMessage{
				ConnId: s.connId,
				UserId: s.userId,
			})
			if err != nil {
				qq.Println("@teardown/3.1", s.connId, s.userId, err.Error())
			}

		}

	})
}
