package pubsub

import (
	"errors"
	"net"
	"sync"
	"testing"
	"time"
)

// mockConn is a mock implementation of net.Conn for testing
type mockConn struct {
	readChan  chan []byte
	writeChan chan []byte
	closed    bool
	mu        sync.Mutex
	closeChan chan struct{}
}

func newMockConn() *mockConn {
	return &mockConn{
		readChan:  make(chan []byte, 10),
		writeChan: make(chan []byte, 10),
		closeChan: make(chan struct{}),
	}
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return 0, errors.New("connection closed")
	}
	m.mu.Unlock()

	select {
	case data := <-m.readChan:
		n = copy(b, data)
		return n, nil
	case <-m.closeChan:
		return 0, errors.New("connection closed")
	}
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return 0, errors.New("connection closed")
	}
	m.mu.Unlock()

	select {
	case m.writeChan <- b:
		return len(b), nil
	case <-m.closeChan:
		return 0, errors.New("connection closed")
	}
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil
	}
	m.closed = true
	close(m.closeChan)
	return nil
}

func (m *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (m *mockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestNewSockd(t *testing.T) {
	sockd := NewSockd()
	if sockd.rooms == nil {
		t.Error("Expected rooms map to be initialized")
	}
	if len(sockd.rooms) != 0 {
		t.Error("Expected empty rooms map")
	}
}

func TestAddConn(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	connId, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}
	if connId != 100 {
		t.Errorf("Expected connId 100, got %d", connId)
	}

	// Verify room was created
	sockd.mu.RLock()
	room, exists := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	if !exists {
		t.Error("Expected room to be created")
	}
	if room == nil {
		t.Error("Expected room to not be nil")
	}

	// Verify session was added
	room.sLock.RLock()
	sess, exists := room.sessions[100]
	room.sLock.RUnlock()

	if !exists {
		t.Error("Expected session to be added")
	}
	if sess == nil {
		t.Error("Expected session to not be nil")
	}
	if sess.connId != 100 {
		t.Errorf("Expected connId 100, got %d", sess.connId)
	}
	if sess.userId != 1 {
		t.Errorf("Expected userId 1, got %d", sess.userId)
	}
}

func TestAddConn_DuplicateConnId(t *testing.T) {
	sockd := NewSockd()
	conn1 := newMockConn()
	conn2 := newMockConn()
	defer conn1.Close()
	defer conn2.Close()

	_, err := sockd.AddConn(1, conn1, 100, "test-room")
	if err != nil {
		t.Fatalf("First AddConn failed: %v", err)
	}

	_, err = sockd.AddConn(2, conn2, 100, "test-room")
	if err == nil {
		t.Error("Expected error for duplicate connId")
	}
	if err.Error() != "connId collision" {
		t.Errorf("Expected 'connId collision' error, got: %v", err)
	}
}

func TestAddConn_MultipleRooms(t *testing.T) {
	sockd := NewSockd()
	conn1 := newMockConn()
	conn2 := newMockConn()
	defer conn1.Close()
	defer conn2.Close()

	_, err := sockd.AddConn(1, conn1, 100, "room1")
	if err != nil {
		t.Fatalf("AddConn to room1 failed: %v", err)
	}

	_, err = sockd.AddConn(2, conn2, 200, "room2")
	if err != nil {
		t.Fatalf("AddConn to room2 failed: %v", err)
	}

	sockd.mu.RLock()
	if len(sockd.rooms) != 2 {
		t.Errorf("Expected 2 rooms, got %d", len(sockd.rooms))
	}
	room1, exists1 := sockd.rooms["room1"]
	room2, exists2 := sockd.rooms["room2"]
	sockd.mu.RUnlock()

	if !exists1 || !exists2 {
		t.Error("Expected both rooms to exist")
	}

	room1.sLock.RLock()
	if len(room1.sessions) != 1 {
		t.Errorf("Expected 1 session in room1, got %d", len(room1.sessions))
	}
	room1.sLock.RUnlock()

	room2.sLock.RLock()
	if len(room2.sessions) != 1 {
		t.Errorf("Expected 1 session in room2, got %d", len(room2.sessions))
	}
	room2.sLock.RUnlock()
}

func TestAddSub(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	connId, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	// Subscribe to a topic
	err = sockd.AddSub("test-room", "topic1", 1, connId, conn)
	if err != nil {
		t.Fatalf("AddSub failed: %v", err)
	}

	// Verify subscription
	sockd.mu.RLock()
	room := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	room.tLock.RLock()
	subMap, exists := room.topics["topic1"]
	if !exists {
		t.Error("Expected topic1 to exist")
	}
	if subMap == nil {
		t.Error("Expected subscription map to not be nil")
	}
	if !subMap[connId] {
		t.Error("Expected connId to be subscribed to topic1")
	}
	room.tLock.RUnlock()
}

func TestAddSub_NonExistentRoom(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	err := sockd.AddSub("non-existent", "topic1", 1, 100, conn)
	if err == nil {
		t.Error("Expected error for non-existent room")
	}
	if err.Error() != "room not found" {
		t.Errorf("Expected 'room not found' error, got: %v", err)
	}
}

func TestAddSub_MultipleTopics(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	connId, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	// Subscribe to multiple topics
	err = sockd.AddSub("test-room", "topic1", 1, connId, conn)
	if err != nil {
		t.Fatalf("AddSub topic1 failed: %v", err)
	}

	err = sockd.AddSub("test-room", "topic2", 1, connId, conn)
	if err != nil {
		t.Fatalf("AddSub topic2 failed: %v", err)
	}

	// Verify both subscriptions
	sockd.mu.RLock()
	room := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	room.tLock.RLock()
	if len(room.topics) != 2 {
		t.Errorf("Expected 2 topics, got %d", len(room.topics))
	}
	if !room.topics["topic1"][connId] {
		t.Error("Expected subscription to topic1")
	}
	if !room.topics["topic2"][connId] {
		t.Error("Expected subscription to topic2")
	}
	room.tLock.RUnlock()
}

func TestAddSub_MultipleSubscribers(t *testing.T) {
	sockd := NewSockd()
	conn1 := newMockConn()
	conn2 := newMockConn()
	defer conn1.Close()
	defer conn2.Close()

	connId1, err := sockd.AddConn(1, conn1, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn 1 failed: %v", err)
	}

	connId2, err := sockd.AddConn(2, conn2, 200, "test-room")
	if err != nil {
		t.Fatalf("AddConn 2 failed: %v", err)
	}

	// Both subscribe to the same topic
	err = sockd.AddSub("test-room", "topic1", 1, connId1, conn1)
	if err != nil {
		t.Fatalf("AddSub conn1 failed: %v", err)
	}

	err = sockd.AddSub("test-room", "topic1", 2, connId2, conn2)
	if err != nil {
		t.Fatalf("AddSub conn2 failed: %v", err)
	}

	// Verify both are subscribed
	sockd.mu.RLock()
	room := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	room.tLock.RLock()
	subMap := room.topics["topic1"]
	if len(subMap) != 2 {
		t.Errorf("Expected 2 subscribers, got %d", len(subMap))
	}
	if !subMap[connId1] || !subMap[connId2] {
		t.Error("Expected both connections to be subscribed")
	}
	room.tLock.RUnlock()
}

func TestPublish(t *testing.T) {
	sockd := NewSockd()
	conn1 := newMockConn()
	conn2 := newMockConn()
	defer conn1.Close()
	defer conn2.Close()

	// Add two connections
	connId1, err := sockd.AddConn(1, conn1, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn 1 failed: %v", err)
	}

	_, err = sockd.AddConn(2, conn2, 200, "test-room")
	if err != nil {
		t.Fatalf("AddConn 2 failed: %v", err)
	}

	// Only conn1 subscribes to topic1
	err = sockd.AddSub("test-room", "topic1", 1, connId1, conn1)
	if err != nil {
		t.Fatalf("AddSub failed: %v", err)
	}

	// Give goroutines time to start
	time.Sleep(100 * time.Millisecond)

	// Publish to topic1
	message := []byte("test message")
	err = sockd.Publish("test-room", "topic1", message)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Wait for messages to be written
	time.Sleep(200 * time.Millisecond)

	// Only conn1 should receive the message
	select {
	case msg := <-conn1.writeChan:
		if len(msg) == 0 {
			t.Error("conn1: Expected non-empty message")
		}
	case <-time.After(1 * time.Second):
		t.Error("conn1: Timeout waiting for message")
	}

	// conn2 should not receive the message
	select {
	case <-conn2.writeChan:
		t.Error("conn2: Should not have received message")
	case <-time.After(200 * time.Millisecond):
		// Expected - conn2 is not subscribed
	}
}

func TestPublish_MultipleSubscribers(t *testing.T) {
	sockd := NewSockd()
	numSubscribers := 5
	conns := make([]*mockConn, numSubscribers)

	// Create multiple connections
	for i := 0; i < numSubscribers; i++ {
		conns[i] = newMockConn()
		defer conns[i].Close()

		connId, err := sockd.AddConn(int64(i+1), conns[i], int64(100+i), "test-room")
		if err != nil {
			t.Fatalf("AddConn %d failed: %v", i, err)
		}

		// All subscribe to the same topic
		err = sockd.AddSub("test-room", "topic1", int64(i+1), connId, conns[i])
		if err != nil {
			t.Fatalf("AddSub %d failed: %v", i, err)
		}
	}

	// Give goroutines time to start
	time.Sleep(200 * time.Millisecond)

	// Publish a message
	message := []byte("broadcast message")
	err := sockd.Publish("test-room", "topic1", message)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Wait for messages to be written
	time.Sleep(500 * time.Millisecond)

	// Verify all subscribers received the message
	// Note: wsutil.WriteServerText writes WebSocket frames, so we just verify
	// that data was written (non-empty) rather than checking exact content
	for i, conn := range conns {
		select {
		case msg := <-conn.writeChan:
			if len(msg) == 0 {
				t.Errorf("conn %d: Expected non-empty message", i)
			}
		case <-time.After(2 * time.Second):
			t.Errorf("conn %d: Timeout waiting for message", i)
		}
	}
}

func TestPublish_NonExistentRoom(t *testing.T) {
	sockd := NewSockd()
	message := []byte("test message")

	// Publishing to non-existent room should not error
	err := sockd.Publish("non-existent", "topic1", message)
	if err != nil {
		t.Errorf("Expected no error for non-existent room, got: %v", err)
	}
}

func TestPublish_NonExistentTopic(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	_, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	// Publish to non-existent topic
	message := []byte("test message")
	err = sockd.Publish("test-room", "non-existent", message)
	if err != nil {
		t.Errorf("Expected no error for non-existent topic, got: %v", err)
	}

	// Connection should not receive message
	select {
	case <-conn.writeChan:
		t.Error("Should not have received message for non-existent topic")
	case <-time.After(200 * time.Millisecond):
		// Expected
	}
}

func TestPublish_NoSubscribers(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	_, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	// Create topic but don't subscribe
	sockd.mu.RLock()
	room := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	room.tLock.Lock()
	room.topics["topic1"] = make(map[int64]bool)
	room.tLock.Unlock()

	// Publish to topic with no subscribers
	message := []byte("test message")
	err = sockd.Publish("test-room", "topic1", message)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Connection should not receive message
	select {
	case <-conn.writeChan:
		t.Error("Should not have received message (no subscribers)")
	case <-time.After(200 * time.Millisecond):
		// Expected
	}
}

func TestPublish_MultipleTopics(t *testing.T) {
	sockd := NewSockd()
	conn1 := newMockConn()
	conn2 := newMockConn()
	defer conn1.Close()
	defer conn2.Close()

	connId1, err := sockd.AddConn(1, conn1, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn 1 failed: %v", err)
	}

	connId2, err := sockd.AddConn(2, conn2, 200, "test-room")
	if err != nil {
		t.Fatalf("AddConn 2 failed: %v", err)
	}

	// conn1 subscribes to topic1, conn2 subscribes to topic2
	err = sockd.AddSub("test-room", "topic1", 1, connId1, conn1)
	if err != nil {
		t.Fatalf("AddSub conn1 failed: %v", err)
	}

	err = sockd.AddSub("test-room", "topic2", 2, connId2, conn2)
	if err != nil {
		t.Fatalf("AddSub conn2 failed: %v", err)
	}

	// Give goroutines time to start
	time.Sleep(100 * time.Millisecond)

	// Publish to topic1 - only conn1 should receive
	message1 := []byte("topic1 message")
	err = sockd.Publish("test-room", "topic1", message1)
	if err != nil {
		t.Fatalf("Publish topic1 failed: %v", err)
	}

	// Publish to topic2 - only conn2 should receive
	message2 := []byte("topic2 message")
	err = sockd.Publish("test-room", "topic2", message2)
	if err != nil {
		t.Fatalf("Publish topic2 failed: %v", err)
	}

	// Wait for messages
	time.Sleep(300 * time.Millisecond)

	// Verify conn1 received topic1 message (and only topic1)
	receivedConn1 := 0
	receivedConn2 := 0

	// Collect messages from both connections
	for i := 0; i < 10; i++ {
		select {
		case <-conn1.writeChan:
			receivedConn1++
		case <-conn2.writeChan:
			receivedConn2++
		case <-time.After(100 * time.Millisecond):
			goto doneCollecting
		}
	}
doneCollecting:

	if receivedConn1 == 0 {
		t.Error("conn1: Expected to receive topic1 message")
	}
	if receivedConn2 == 0 {
		t.Error("conn2: Expected to receive topic2 message")
	}

	// Each connection should have received exactly 1 message (their respective topic)
	// Since we can't easily distinguish message content in WebSocket frames,
	// we verify that each received at least one message
	// In a real scenario, conn1 should only receive topic1 and conn2 only topic2
}

func TestRemoveConn(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	connId, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	// Subscribe to a topic
	err = sockd.AddSub("test-room", "topic1", 1, connId, conn)
	if err != nil {
		t.Fatalf("AddSub failed: %v", err)
	}

	// Verify session and subscription exist
	sockd.mu.RLock()
	room := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	room.sLock.RLock()
	if _, exists := room.sessions[connId]; !exists {
		t.Error("Expected session to exist before removal")
	}
	room.sLock.RUnlock()

	room.tLock.RLock()
	if !room.topics["topic1"][connId] {
		t.Error("Expected subscription to exist before removal")
	}
	room.tLock.RUnlock()

	// Remove connection
	err = sockd.RemoveConn(1, connId, "test-room")
	if err != nil {
		t.Fatalf("RemoveConn failed: %v", err)
	}

	// Wait for cleanup
	time.Sleep(300 * time.Millisecond)

	// Verify session was removed
	room.sLock.RLock()
	if _, exists := room.sessions[connId]; exists {
		t.Error("Expected session to be removed")
	}
	room.sLock.RUnlock()

	// Verify subscription was removed
	room.tLock.RLock()
	if subMap, exists := room.topics["topic1"]; exists {
		if subMap[connId] {
			t.Error("Expected subscription to be removed")
		}
		// Topic should be deleted if empty
		if len(subMap) == 0 {
			if _, stillExists := room.topics["topic1"]; stillExists {
				t.Error("Expected empty topic to be deleted")
			}
		}
	}
	room.tLock.RUnlock()
}

func TestRemoveConn_NonExistentRoom(t *testing.T) {
	sockd := NewSockd()

	// Removing from non-existent room should not error
	err := sockd.RemoveConn(1, 100, "non-existent")
	if err != nil {
		t.Errorf("Expected no error for non-existent room, got: %v", err)
	}
}

func TestRemoveConn_NonExistentConnId(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	_, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	// Remove non-existent connection
	err = sockd.RemoveConn(1, 999, "test-room")
	if err != nil {
		t.Errorf("Expected no error for non-existent connId, got: %v", err)
	}
}

func TestSession_WritePump(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	connId, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	// Get the session
	sockd.mu.RLock()
	room := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	room.sLock.RLock()
	sess := room.sessions[connId]
	room.sLock.RUnlock()

	// Send a message to the session's send channel
	message := []byte("test message")
	sess.send <- message

	// Wait for write
	time.Sleep(200 * time.Millisecond)

	// Check that message was written to connection
	// Note: wsutil.WriteServerText writes WebSocket frames, so we just verify
	// that data was written (non-empty) rather than checking exact content
	select {
	case msg := <-conn.writeChan:
		if len(msg) == 0 {
			t.Error("Expected non-empty message")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message to be written")
	}
}

func TestSession_Teardown(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	connId, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	// Get the session
	sockd.mu.RLock()
	room := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	room.sLock.RLock()
	sess := room.sessions[connId]
	room.sLock.RUnlock()

	// Teardown should be idempotent
	sess.teardown()
	sess.teardown()
	sess.teardown()

	// Verify send channel is closed
	select {
	case _, ok := <-sess.send:
		if ok {
			t.Error("Expected send channel to be closed")
		}
	default:
		t.Error("Send channel should be closed")
	}
}

func TestRoom_Cleanup(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	connId, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	// Subscribe to a topic
	err = sockd.AddSub("test-room", "topic1", 1, connId, conn)
	if err != nil {
		t.Fatalf("AddSub failed: %v", err)
	}

	sockd.mu.RLock()
	room := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	// Verify session and subscription exist
	room.sLock.RLock()
	if _, exists := room.sessions[connId]; !exists {
		t.Error("Expected session to exist")
	}
	room.sLock.RUnlock()

	room.tLock.RLock()
	if !room.topics["topic1"][connId] {
		t.Error("Expected subscription to exist")
	}
	room.tLock.RUnlock()

	// Trigger cleanup
	room.disconnect <- connId

	// Wait for cleanup
	time.Sleep(300 * time.Millisecond)

	// Verify session was removed
	room.sLock.RLock()
	if _, exists := room.sessions[connId]; exists {
		t.Error("Expected session to be removed")
	}
	room.sLock.RUnlock()

	// Verify subscription was removed
	room.tLock.RLock()
	if subMap, exists := room.topics["topic1"]; exists {
		if subMap[connId] {
			t.Error("Expected subscription to be removed")
		}
	}
	room.tLock.RUnlock()
}

func TestRoom_Cleanup_NonExistentSession(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	_, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	sockd.mu.RLock()
	room := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	// Cleanup non-existent session should not panic
	room.disconnect <- 999

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Original session should still exist
	room.sLock.RLock()
	if _, exists := room.sessions[100]; !exists {
		t.Error("Expected original session to still exist")
	}
	room.sLock.RUnlock()
}

func TestRoom_Cleanup_RemovesEmptyTopics(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	connId, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	// Subscribe to a topic
	err = sockd.AddSub("test-room", "topic1", 1, connId, conn)
	if err != nil {
		t.Fatalf("AddSub failed: %v", err)
	}

	sockd.mu.RLock()
	room := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	// Verify topic exists
	room.tLock.RLock()
	if _, exists := room.topics["topic1"]; !exists {
		t.Error("Expected topic1 to exist")
	}
	room.tLock.RUnlock()

	// Remove connection (which should remove subscription and delete empty topic)
	err = sockd.RemoveConn(1, connId, "test-room")
	if err != nil {
		t.Fatalf("RemoveConn failed: %v", err)
	}

	// Wait for cleanup
	time.Sleep(300 * time.Millisecond)

	// Verify topic was deleted (since it's now empty)
	room.tLock.RLock()
	if _, exists := room.topics["topic1"]; exists {
		t.Error("Expected empty topic to be deleted")
	}
	room.tLock.RUnlock()
}

func TestPublish_Concurrent(t *testing.T) {
	sockd := NewSockd()
	conn1 := newMockConn()
	conn2 := newMockConn()
	defer conn1.Close()
	defer conn2.Close()

	connId1, err := sockd.AddConn(1, conn1, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn 1 failed: %v", err)
	}

	connId2, err := sockd.AddConn(2, conn2, 200, "test-room")
	if err != nil {
		t.Fatalf("AddConn 2 failed: %v", err)
	}

	// Both subscribe to the same topic
	err = sockd.AddSub("test-room", "topic1", 1, connId1, conn1)
	if err != nil {
		t.Fatalf("AddSub conn1 failed: %v", err)
	}

	err = sockd.AddSub("test-room", "topic1", 2, connId2, conn2)
	if err != nil {
		t.Fatalf("AddSub conn2 failed: %v", err)
	}

	// Give goroutines time to start
	time.Sleep(100 * time.Millisecond)

	// Send multiple publishes concurrently
	numPublishes := 10
	var wg sync.WaitGroup
	wg.Add(numPublishes)

	for i := 0; i < numPublishes; i++ {
		go func(id int) {
			defer wg.Done()
			message := []byte{byte(id)}
			err := sockd.Publish("test-room", "topic1", message)
			if err != nil {
				t.Errorf("Publish %d failed: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	// Wait for all messages to be written
	time.Sleep(500 * time.Millisecond)

	// Verify both connections received messages
	// (exact count may vary due to timing, but should receive some)
	received1 := 0
	received2 := 0

	for i := 0; i < numPublishes*2; i++ {
		select {
		case <-conn1.writeChan:
			received1++
		case <-conn2.writeChan:
			received2++
		case <-time.After(100 * time.Millisecond):
			goto done
		}
	}
done:

	if received1 == 0 {
		t.Error("conn1: Expected to receive at least one message")
	}
	if received2 == 0 {
		t.Error("conn2: Expected to receive at least one message")
	}
}

func TestSession_WritePump_ErrorHandling(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()

	_, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	// Get the session
	sockd.mu.RLock()
	room := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	room.sLock.RLock()
	sess := room.sessions[100]
	room.sLock.RUnlock()

	// Close the connection to trigger errors
	conn.Close()

	// Send a message - should trigger error handling
	message := []byte("test message")
	sess.send <- message

	// Wait for error handling
	time.Sleep(300 * time.Millisecond)

	// The session should eventually be cleaned up due to errors
	// (after 10 errors)
}

func TestPublish_Timeout(t *testing.T) {
	sockd := NewSockd()

	// Create a connection with a blocked write channel
	conn := newMockConn()
	defer conn.Close()

	connId, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	// Subscribe to topic
	err = sockd.AddSub("test-room", "topic1", 1, connId, conn)
	if err != nil {
		t.Fatalf("AddSub failed: %v", err)
	}

	sockd.mu.RLock()
	room := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	room.sLock.RLock()
	sess := room.sessions[connId]
	room.sLock.RUnlock()

	// Fill up the send channel to block it
	for i := 0; i < 20; i++ {
		select {
		case sess.send <- []byte("block"):
		default:
			// Channel is full
		}
	}

	// Publish a message - should timeout
	message := []byte("timeout test")
	err = sockd.Publish("test-room", "topic1", message)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Wait for timeout handling
	time.Sleep(600 * time.Millisecond)

	// The message should have timed out (we can't easily verify this
	// without checking logs, but at least it shouldn't crash)
}

func TestAddConn_SameConnId_DifferentRooms(t *testing.T) {
	sockd := NewSockd()
	conn1 := newMockConn()
	conn2 := newMockConn()
	defer conn1.Close()
	defer conn2.Close()

	// Same connId in different rooms should be allowed
	_, err := sockd.AddConn(1, conn1, 100, "room1")
	if err != nil {
		t.Fatalf("AddConn to room1 failed: %v", err)
	}

	_, err = sockd.AddConn(2, conn2, 100, "room2")
	if err != nil {
		t.Fatalf("AddConn to room2 failed: %v", err)
	}

	// Verify both sessions exist
	sockd.mu.RLock()
	room1 := sockd.rooms["room1"]
	room2 := sockd.rooms["room2"]
	sockd.mu.RUnlock()

	room1.sLock.RLock()
	if _, exists := room1.sessions[100]; !exists {
		t.Error("Expected session 100 in room1")
	}
	room1.sLock.RUnlock()

	room2.sLock.RLock()
	if _, exists := room2.sessions[100]; !exists {
		t.Error("Expected session 100 in room2")
	}
	room2.sLock.RUnlock()
}

func TestRemoveConn_BusyRoom(t *testing.T) {
	sockd := NewSockd()
	conn := newMockConn()
	defer conn.Close()

	connId, err := sockd.AddConn(1, conn, 100, "test-room")
	if err != nil {
		t.Fatalf("AddConn failed: %v", err)
	}

	sockd.mu.RLock()
	room := sockd.rooms["test-room"]
	sockd.mu.RUnlock()

	// Fill up the disconnect channel to simulate busy room
	for i := 0; i < 33; i++ {
		select {
		case room.disconnect <- int64(i + 1000):
		default:
			// Channel is full
		}
	}

	// RemoveConn should handle busy channel gracefully
	err = sockd.RemoveConn(1, connId, "test-room")
	if err != nil {
		// It's okay if it returns an error for busy room
		// The test verifies it doesn't panic
	}

	// Wait a bit
	time.Sleep(200 * time.Millisecond)
}

