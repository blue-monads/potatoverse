package higher

import (
	"encoding/json"
	"errors"
	"maps"
	"net"
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
		}
	}
}

func (r *Room) AddConn(userId int64, conn net.Conn, connId int64) (int64, error) {
	sess := &session{
		room:   r,
		connId: connId,
		userId: userId,
		conn:   conn,
		send:   make(chan []byte, 16),
	}

	r.sLock.Lock()
	existingSess := r.sessions[sess.connId]
	r.sessions[sess.connId] = sess
	r.sLock.Unlock()

	if existingSess != nil {
		existingSess.teardown()
	}

	go sess.writePump()
	go sess.readPump()

	return sess.connId, nil
}

func (r *Room) RemoveConn(userId int64, connId int64) error {
	tcan := time.After(time.Second * 10)

	select {
	case r.disconnect <- connId:
		return nil
	case <-tcan:
		return errors.New("room is very busy or dead")
	}
}

// Broadcast sends a message to all users in the room
func (r *Room) Broadcast(message []byte) error {
	msg := Message{
		Id:   time.Now().UnixNano(),
		Type: MessageTypeBroadcast,
		Data: message,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	r.handleBroadcast(data, time.Second*2)

	return nil
}

// Publish sends a message to all subscribers of a topic in the room
func (r *Room) Publish(topicName string, message []byte) error {
	msg := Message{
		Id:   time.Now().UnixNano(),
		Type: MessageTypePublish,
		Data: message,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	r.handlePublish(topicName, data, time.Second*5)

	return nil

}

// DirectMessage sends a message to a specific user
func (r *Room) DirectMessage(targetConnId int64, message []byte) error {
	msg := Message{
		Id:   time.Now().UnixNano(),
		Type: MessageTypeDirectMessage,
		Data: message,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	r.handleDirectMessage(targetConnId, data, time.Second*5)

	return nil

}

// Subscribe adds a connection to a topic subscription
func (r *Room) Subscribe(topicName string, connId int64) error {
	r.tLock.Lock()
	if r.topics[topicName] == nil {
		r.topics[topicName] = make(map[int64]bool)
	}
	r.topics[topicName][connId] = true
	r.tLock.Unlock()

	return nil
}

// Unsubscribe removes a connection from a topic subscription
func (r *Room) Unsubscribe(topicName string, connId int64) error {

	r.tLock.Lock()
	if subMap, ok := r.topics[topicName]; ok {
		delete(subMap, connId)
		if len(subMap) == 0 {
			delete(r.topics, topicName)
		}
	}
	r.tLock.Unlock()

	return nil
}

// private

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
	UserId   int64   `json:"user_id"`
	Identity string  `json:"identity"`
	ConnIds  []int64 `json:"conn_ids"`
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
