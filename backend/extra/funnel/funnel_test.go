package funnel

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// setupLocalServer creates a local HTTP server for testing
func setupLocalServer(t *testing.T) (*httptest.Server, int) {
	t.Log("@setupLocalServer/1")

	mux := http.NewServeMux()

	t.Log("@setupLocalServer/2")

	// Simple GET endpoint
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"test response"}`))
	})

	// POST endpoint with body
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	// WebSocket echo endpoint
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			msg, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				break
			}
			// Echo back
			wsutil.WriteServerText(conn, msg)
		}
	})

	t.Log("@setupLocalServer/3")

	server := httptest.NewServer(mux)

	t.Log("@setupLocalServer/4")

	_, portStr, err := net.SplitHostPort(server.URL[7:]) // Remove "http://"
	if err != nil {
		t.Fatalf("Failed to split host port from URL: %s", server.URL)
	}

	t.Log("@setupLocalServer/5{LOCAL_SERVER_URL}", server.URL)

	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("Failed to parse port from URL: %s", server.URL)
	}

	return server, portInt
}

// setupFunnelServer creates a funnel server with gin router
func setupFunnelServer(t *testing.T) (*Funnel, *httptest.Server, string) {
	t.Log("@setupFunnelServer/1")

	funnel := New()
	t.Log("@setupFunnelServer/2")
	gin.SetMode(gin.TestMode)
	t.Log("@setupFunnelServer/3")
	router := gin.New()

	t.Log("@setupFunnelServer/4")

	// Register server websocket endpoint
	router.GET("/funnel/register/:serverId", func(c *gin.Context) {
		serverId := c.Param("serverId")
		funnel.HandleServerWebSocket(serverId, c)
	})

	// Route endpoint
	router.NoRoute(func(c *gin.Context) {
		// http://serverid.localhost/path

		url := c.Request.URL.Host
		serverId := strings.Split(url, ".")[0]
		funnel.HandleRoute(serverId, c)
	})

	t.Log("@setupFunnelServer/5")

	server := httptest.NewServer(router)
	t.Log("@setupFunnelServer/6")

	baseURL := server.URL

	u, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("Failed to parse URL: %s", baseURL)
	}

	port := u.Port()

	t.Log("@setupFunnelServer/7{FUNNEL_SERVER_URL}", baseURL)

	t.Log("@setupFunnelServer/8{FUNNEL_SERVER_PORT}", port)

	return funnel, server, port
}

func TestFunnel_HTTP_GetRequest(t *testing.T) {
	t.Log("@TestFunnel_HTTP_GetRequest/1")
	// Setup local server
	localServer, localPort := setupLocalServer(t)
	defer localServer.Close()

	t.Log("@TestFunnel_HTTP_GetRequest/2")

	// Setup funnel server
	_, funnelServer, fport := setupFunnelServer(t)
	defer funnelServer.Close()

	t.Log("@TestFunnel_HTTP_GetRequest/3")

	// Setup funnel client
	client := NewFunnelClient(localPort, fmt.Sprintf("http://localhost:%s", fport), "test-server")
	defer client.Stop()

	t.Log("@TestFunnel_HTTP_GetRequest/4")

	// Start client in background
	clientDone := make(chan error, 1)
	go func() {
		clientDone <- client.Start()
	}()

	t.Log("@TestFunnel_HTTP_GetRequest/5")

	time.Sleep(2 * time.Second)

	t.Log("@TestFunnel_HTTP_GetRequest/6")

	// Make request through funnel with timeout
	clientHTTP := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := clientHTTP.Get(fmt.Sprintf("http://test-server.localhost:%s/test", fport))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	expected := `{"message":"test response"}`
	if string(body) != expected {
		t.Fatalf("Expected %s, got %s", expected, string(body))
	}
}

func TestFunnel_HTTP_PostRequest(t *testing.T) {
	t.Log("@TestFunnel_HTTP_PostRequest/1")
	// Setup local server
	localServer, localPort := setupLocalServer(t)
	defer localServer.Close()

	t.Log("@TestFunnel_HTTP_PostRequest/2")

	// Setup funnel server
	_, funnelServer, fport := setupFunnelServer(t)
	defer funnelServer.Close()

	t.Log("@TestFunnel_HTTP_PostRequest/3")

	// Setup funnel client
	client := NewFunnelClient(localPort, fmt.Sprintf("http://localhost:%s", fport), "test-server")
	defer client.Stop()

	t.Log("@TestFunnel_HTTP_PostRequest/4")

	// Start client in background
	clientDone := make(chan error, 1)
	go func() {
		clientDone <- client.Start()
	}()

	t.Log("@TestFunnel_HTTP_PostRequest/5")

	time.Sleep(2 * time.Second)

	t.Log("@TestFunnel_HTTP_PostRequest/6")

	// Make POST request through funnel
	clientHTTP := &http.Client{
		Timeout: 5 * time.Second,
	}

	postData := "Hello, Funnel!"
	resp, err := clientHTTP.Post(
		fmt.Sprintf("http://test-server.localhost:%s/echo", fport),
		"text/plain",
		strings.NewReader(postData),
	)
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if string(body) != postData {
		t.Fatalf("Expected %s, got %s", postData, string(body))
	}
}

func TestFunnel_WebSocket(t *testing.T) {
	t.Log("@TestFunnel_WebSocket/1")
	// Setup local server
	localServer, localPort := setupLocalServer(t)
	defer localServer.Close()

	t.Log("@TestFunnel_WebSocket/2")

	// Setup funnel server
	_, funnelServer, fport := setupFunnelServer(t)
	defer funnelServer.Close()

	t.Log("@TestFunnel_WebSocket/3")

	// Setup funnel client
	client := NewFunnelClient(localPort, fmt.Sprintf("http://localhost:%s", fport), "test-server")
	defer client.Stop()

	t.Log("@TestFunnel_WebSocket/4")

	// Start client in background
	clientDone := make(chan error, 1)
	go func() {
		clientDone <- client.Start()
	}()

	t.Log("@TestFunnel_WebSocket/5")

	time.Sleep(2 * time.Second)

	t.Log("@TestFunnel_WebSocket/6")

	// Connect to WebSocket through funnel
	u, err := url.Parse(fmt.Sprintf("ws://test-server.localhost:%s/ws", fport))
	if err != nil {
		t.Fatalf("Failed to parse WebSocket URL: %v", err)
	}

	conn, _, _, err := ws.Dial(context.Background(), u.String())
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	t.Log("@TestFunnel_WebSocket/7")

	// Send test message
	testMessage := "Hello, WebSocket!"
	err = wsutil.WriteClientText(conn, []byte(testMessage))
	if err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	t.Log("@TestFunnel_WebSocket/8")

	// Read echo response
	msg, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	if string(msg) != testMessage {
		t.Fatalf("Expected %s, got %s", testMessage, string(msg))
	}

	t.Log("@TestFunnel_WebSocket/9")
}

func TestFunnel_MultipleWebSocket(t *testing.T) {
	t.Log("@TestFunnel_MultipleWebSocket/1")
	// Setup local server
	localServer, localPort := setupLocalServer(t)
	defer localServer.Close()

	t.Log("@TestFunnel_MultipleWebSocket/2")

	// Setup funnel server
	_, funnelServer, fport := setupFunnelServer(t)
	defer funnelServer.Close()

	t.Log("@TestFunnel_MultipleWebSocket/3")

	// Setup funnel client
	client := NewFunnelClient(localPort, fmt.Sprintf("http://localhost:%s", fport), "test-server")
	defer client.Stop()

	t.Log("@TestFunnel_MultipleWebSocket/4")

	// Start client in background
	clientDone := make(chan error, 1)
	go func() {
		clientDone <- client.Start()
	}()

	t.Log("@TestFunnel_MultipleWebSocket/5")

	time.Sleep(2 * time.Second)

	t.Log("@TestFunnel_MultipleWebSocket/6")

	// Connect multiple WebSocket connections
	numConnections := 5
	var wg sync.WaitGroup
	errors := make(chan error, numConnections)

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			u, err := url.Parse(fmt.Sprintf("ws://test-server.localhost:%s/ws", fport))
			if err != nil {
				errors <- fmt.Errorf("connection %d: failed to parse URL: %v", id, err)
				return
			}

			conn, _, _, err := ws.Dial(context.Background(), u.String())
			if err != nil {
				errors <- fmt.Errorf("connection %d: failed to connect: %v", id, err)
				return
			}
			defer conn.Close()

			// Send unique message
			testMessage := fmt.Sprintf("Message from connection %d", id)
			err = wsutil.WriteClientText(conn, []byte(testMessage))
			if err != nil {
				errors <- fmt.Errorf("connection %d: failed to write: %v", id, err)
				return
			}

			// Read echo response
			msg, _, err := wsutil.ReadServerData(conn)
			if err != nil {
				errors <- fmt.Errorf("connection %d: failed to read: %v", id, err)
				return
			}

			if string(msg) != testMessage {
				errors <- fmt.Errorf("connection %d: expected %s, got %s", id, testMessage, string(msg))
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	t.Log("@TestFunnel_MultipleWebSocket/7")
}

func TestFunnel_MultipleRequests(t *testing.T) {
	t.Log("@TestFunnel_MultipleRequests/1")
	// Setup local server
	localServer, localPort := setupLocalServer(t)
	defer localServer.Close()

	t.Log("@TestFunnel_MultipleRequests/2")

	// Setup funnel server
	_, funnelServer, fport := setupFunnelServer(t)
	defer funnelServer.Close()

	t.Log("@TestFunnel_MultipleRequests/3")

	// Setup funnel client
	client := NewFunnelClient(localPort, fmt.Sprintf("http://localhost:%s", fport), "test-server")
	defer client.Stop()

	t.Log("@TestFunnel_MultipleRequests/4")

	// Start client in background
	clientDone := make(chan error, 1)
	go func() {
		clientDone <- client.Start()
	}()

	t.Log("@TestFunnel_MultipleRequests/5")

	time.Sleep(2 * time.Second)

	t.Log("@TestFunnel_MultipleRequests/6")

	// Make multiple concurrent requests
	numRequests := 10
	var wg sync.WaitGroup
	errors := make(chan error, numRequests)

	clientHTTP := &http.Client{
		Timeout: 5 * time.Second,
	}

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			resp, err := clientHTTP.Get(fmt.Sprintf("http://test-server.localhost:%s/test", fport))
			if err != nil {
				errors <- fmt.Errorf("request %d: failed to make request: %v", id, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("request %d: expected status 200, got %d", id, resp.StatusCode)
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				errors <- fmt.Errorf("request %d: failed to read response: %v", id, err)
				return
			}

			expected := `{"message":"test response"}`
			if string(body) != expected {
				errors <- fmt.Errorf("request %d: expected %s, got %s", id, expected, string(body))
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	t.Log("@TestFunnel_MultipleRequests/7")
}

func TestFunnel_ServerNotConnected(t *testing.T) {
	t.Log("@TestFunnel_ServerNotConnected/1")
	// Setup funnel server without connecting a client
	_, funnelServer, fport := setupFunnelServer(t)
	defer funnelServer.Close()

	t.Log("@TestFunnel_ServerNotConnected/2")

	// Make request to non-existent server
	clientHTTP := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := clientHTTP.Get(fmt.Sprintf("http://non-existent-server.localhost:%s/test", fport))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should get 502 Bad Gateway
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("Expected status 502, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Response should contain error message
	if !strings.Contains(string(body), "server not connected") {
		t.Fatalf("Expected error message about server not connected, got: %s", string(body))
	}

	t.Log("@TestFunnel_ServerNotConnected/3")
}
