package notifier

import (
	"net"
	"sync"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/gobwas/ws/wsutil"
	"github.com/tidwall/gjson"
)

type Connection struct {
	connId           int64
	conn             net.Conn
	send             chan []byte
	once             sync.Once
	closedAndCleaned bool
	userRoom         *UserRoom
}

func (c *Connection) writePump() {
	defer func() {
		c.userRoom.notifier.cleanConnChan <- c.connId
	}()

	errCount := 0
	for msg := range c.send {
		if msg == nil {
			return
		}

		id := gjson.GetBytes(msg, "id").Int()
		if id > c.userRoom.maxMsgId {
			c.userRoom.maxMsgId = id
		}

		err := wsutil.WriteServerText(c.conn, msg)
		if err != nil {
			qq.Println("@writePump/1{ERROR}", err.Error())
			errCount++
			if errCount > 10 {
				return
			}
			continue
		}
		errCount = 0
	}
}

func (c *Connection) teardown() {
	c.once.Do(func() {
		c.closedAndCleaned = true
		if c.send != nil {
			c.send <- nil
		}
		if c.conn != nil {
			c.conn.Close()
		}
	})
}
