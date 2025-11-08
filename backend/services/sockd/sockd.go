package sockd

import (
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Sockd struct {
	rooms map[string]*SockdRoom
	rLock sync.RWMutex
}

type Connection struct {
	id        int64
	conn      net.Conn
	selfAttrs map[string]string
}

type SockdRoom struct {
	connections       map[int64]*Connection
	connectionsLock   sync.RWMutex
	roomType          string // server_process, client_broadcast, client_p2p, client_cond
	broadcastPresence bool
}

type Condition struct {
	Key   string
	Value string
	Op    string
	Sub   []Condition
	Or    bool
}

func (s *SockdRoom) SendWithCondition(cond Condition, data []byte) {

	s.connectionsLock.RLock()
	defer s.connectionsLock.RUnlock()

	matches := make(map[int64]bool)

	for _, conn := range s.connections {
		if checkCondition(conn.selfAttrs, cond) {
			matches[conn.id] = true
		}
	}

	for _, conn := range s.connections {
		if matches[conn.id] {
			conn.conn.Write(data)
		}
	}

}

func checkCondition(attrs map[string]string, cond Condition) bool {

	if cond.Sub != nil {
		if cond.Or {
			match := false
			for _, sub := range cond.Sub {
				if checkCondition(attrs, sub) {
					match = true
				}
			}

			if !match {
				return false
			}
		} else {
			match := true
			for _, sub := range cond.Sub {
				if !checkCondition(attrs, sub) {
					match = false
					break
				}
			}
			if !match {
				return false
			}
		}

	}

	targetValue := attrs[cond.Key]

	if cond.Op == ">" {
		fval, _ := strconv.ParseInt(targetValue, 10, 64)
		cval, _ := strconv.ParseInt(cond.Value, 10, 64)
		return fval > cval
	}
	if cond.Op == "<" {
		fval, _ := strconv.ParseInt(targetValue, 10, 64)
		cval, _ := strconv.ParseInt(cond.Value, 10, 64)

		return fval < cval
	}

	if cond.Op == ">=" {
		fval, _ := strconv.ParseInt(targetValue, 10, 64)
		cval, _ := strconv.ParseInt(cond.Value, 10, 64)
		return fval >= cval
	}

	if cond.Op == "<=" {
		fval, _ := strconv.ParseInt(targetValue, 10, 64)
		cval, _ := strconv.ParseInt(cond.Value, 10, 64)
		return fval <= cval
	}

	if cond.Op == "==" {
		return targetValue == cond.Value
	}

	if cond.Op == "!=" {
		return targetValue != cond.Value
	}

	if cond.Op == "contains" {
		return strings.Contains(targetValue, cond.Value)
	}

	if cond.Op == "notcontains" {
		return !strings.Contains(targetValue, cond.Value)
	}

	if cond.Op == "startswith" {
		return strings.HasPrefix(targetValue, cond.Value)
	}

	if cond.Op == "endswith" {
		return strings.HasSuffix(targetValue, cond.Value)
	}

	if cond.Op == "notstartswith" {
		return !strings.HasPrefix(targetValue, cond.Value)
	}

	if cond.Op == "notendswith" {
		return !strings.HasSuffix(targetValue, cond.Value)
	}

	if cond.Op == "regex" {
		re, err := regexp.Compile(cond.Value)
		if err != nil {
			return false
		}

		return re.MatchString(targetValue)
	}

	if cond.Op == "notregex" {
		re, err := regexp.Compile(cond.Value)
		if err != nil {
			return false
		}
		return !re.MatchString(targetValue)
	}

	return false
}

/*

{
	key: "age"
	op: ">"
	value: "10"
	or: false
	sub: [
		{
			key: "city"
			op: "=="
			value: "new york"
		},
		{
			key: "country"
			op: "=="
			value: "usa"
		}
	]
}

*/

/*

--metadata-start-5643126--
{meta_data: 1}
--metadata-end-5643126--
{actual_payload: 1}



selfAttrs:
	age -> 11
	city -> new york
	country -> usa

*/
