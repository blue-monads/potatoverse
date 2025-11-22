package higher

import (
	"encoding/json"
	"maps"
	"sync"
	"time"

	"github.com/blue-monads/turnix/backend/utils/qq"
)

type Room struct {
	name string

	disconnect chan int64
	broadcast  chan []byte
	publish    chan publishEvent
	directMsg  chan directMessageEvent

	// topics: TopicName -> ConnId -> bool
	topics map[string]map[int64]bool
	tLock  sync.RWMutex

	// sessions: ConnId -> Session Object
	sessions map[int64]*session
	sLock    sync.RWMutex
}

func (r *Room) run() {
	for {
		select {
		case msg := <-r.broadcast:
			r.handleBroadcast(msg, time.Second*1)
		case pub := <-r.publish:
			r.handlePublish(pub.topic, pub.message, time.Second*2)
		case dm := <-r.directMsg:
			r.handleDirectMessage(dm.targetConnId, dm.message, time.Second*2)

		case connId := <-r.disconnect:

			r.cleanup(connId)

			//			r.notifyPresenceAll(r.name)
		}
	}
}

func (r *Room) handleBroadcast(message []byte, maxWait time.Duration) {
	copySess := make([]*session, 0, len(r.sessions))

	r.sLock.RLock()
	for _, sess := range r.sessions {
		copySess = append(copySess, sess)
	}
	r.sLock.RUnlock()

	for _, sess := range copySess {
		if sess.closedAndCleaned {
			continue
		}

		tcan := time.After(maxWait)
		select {
		case sess.send <- message:
			continue
		case <-tcan:
			qq.Println("@drop_message", sess.connId)
			continue
		}
	}
}

func (r *Room) handlePublish(topic string, message []byte, maxWait time.Duration) {

	r.tLock.Lock()
	topicSubscribers := r.topics[topic]
	if len(topicSubscribers) == 0 {
		r.tLock.Unlock()
		return
	}

	topicCopy := maps.Clone(topicSubscribers)

	r.tLock.Unlock()

	copySess := make([]*session, 0, len(topicCopy))

	r.sLock.RLock()
	for connId := range topicCopy {
		sess, exists := r.sessions[connId]
		if !exists || sess == nil || sess.closedAndCleaned {
			continue
		}
		copySess = append(copySess, sess)
	}
	r.sLock.RUnlock()

	for _, sess := range copySess {
		if sess.closedAndCleaned {
			continue
		}

		tcan := time.After(maxWait)
		select {
		case sess.send <- message:
			continue
		case <-tcan:
			qq.Println("@drop_message", sess.connId)
			continue
		}
	}
}

func (r *Room) handleDirectMessage(targetConnId int64, message []byte, maxWait time.Duration) {
	sess, exists := r.sessions[targetConnId]
	if !exists || sess == nil || sess.closedAndCleaned {
		return
	}

	tcan := time.After(maxWait)
	select {
	case sess.send <- message:
		return
	case <-tcan:
		qq.Println("@drop_message", targetConnId)
		return
	}

}

// cleanup performs the heavy lifting of removing the user from all maps
func (r *Room) cleanup(connId int64) {
	r.sLock.Lock()
	delete(r.sessions, connId)
	r.sLock.Unlock()

	sess, exists := r.sessions[connId]
	if !exists {
		return
	}

	userTopics := make([]string, 0, len(r.topics))

	r.tLock.Lock()
	for topic := range r.topics {
		topicSubscribers := r.topics[topic]
		if len(topicSubscribers) == 0 {
			continue
		}

		if _, ok := topicSubscribers[connId]; ok {
			userTopics = append(userTopics, topic)
			delete(topicSubscribers, connId)
			if len(topicSubscribers) == 0 {
				delete(r.topics, topic)
			}
		}
	}

	r.tLock.Unlock()

	sess.teardown()

	for _, topic := range userTopics {
		r.notifyPresenceAll(topic)
	}
}

// presence

type PresenceInfo struct {
	Topic string              `json:"topic"`
	Users map[int64]*UserInfo `json:"users"`
}

type UserInfo struct {
	UserId   int64
	Identity string
	ConnIds  []int64
}

func (r *Room) notifyPresenceAll(topic string) error {
	presenceInfo := r.buildPresenceInfo(topic)

	data, err := json.Marshal(presenceInfo)
	if err != nil {
		return err
	}

	r.handlePublish(topic, data, time.Second*2)

	return nil
}

func (r *Room) notifyPresenceUser(topic string, userId int64) error {
	presenceInfo := r.buildPresenceInfo(topic)

	user := presenceInfo.Users[userId]
	if user == nil {
		return nil
	}

	data, err := json.Marshal(presenceInfo)
	if err != nil {
		return err
	}

	r.handleDirectMessage(userId, data, time.Second*2)

	return nil
}

func (r *Room) buildPresenceInfo(topic string) *PresenceInfo {

	users := make(map[int64]*UserInfo)

	r.tLock.RLock()
	topicSubscribers := r.topics[topic]
	r.tLock.RUnlock()

	r.sLock.RLock()
	for connId := range topicSubscribers {

		sess, exists := r.sessions[connId]
		if !exists || sess == nil {
			continue
		}

		uInfo := users[sess.userId]
		if uInfo == nil {
			uInfo = &UserInfo{
				UserId:   sess.userId,
				Identity: "Todo",
				ConnIds:  []int64{connId},
			}
			users[sess.userId] = uInfo
		} else {
			uInfo.ConnIds = append(uInfo.ConnIds, connId)
		}

	}

	r.sLock.RUnlock()

	return &PresenceInfo{
		Topic: topic,
		Users: users,
	}

}
