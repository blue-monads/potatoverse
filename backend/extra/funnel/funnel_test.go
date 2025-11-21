package funnel

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
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

	resp, err := clientHTTP.Get(fmt.Sprintf("http://test-server.localhost:%d/test", localPort))
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

/*

func TestFunnel_HTTP_PostRequest(t *testing.T) {
	// Setup local server
	localServer, localPort := setupLocalServer(t)
	defer localServer.Close()

	// Setup funnel server
	funnel, funnelServer, funnelURL := setupFunnelServer(t)
	defer funnelServer.Close()

	// Setup funnel client
	client := NewFunnelClient(localPort, funnelURL, "test-server")
	defer client.Stop()

	// Start client in background
	clientDone := make(chan error, 1)
	go func() {
		clientDone <- client.Start()
	}()

	// Wait for client to connect and verify connection
	connected := false
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		funnel.scLock.RLock()
		_, exists := funnel.serverConnections["test-server"]
		funnel.scLock.RUnlock()
		if exists {
			connected = true
			break
		}
	}

	if !connected {
		t.Fatalf("Client failed to connect to funnel")
	}

	// Make POST request through funnel with timeout
	clientHTTP := &http.Client{
		Timeout: 5 * time.Second,
	}
	body := strings.NewReader("test body content")
	resp, err := clientHTTP.Post(fmt.Sprintf("%s/route/test-server/echo", funnelURL), "text/plain", body)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	expected := "test body content"
	if string(respBody) != expected {
		t.Fatalf("Expected %s, got %s", expected, string(respBody))
	}
}

func TestFunnel_WebSocket(t *testing.T) {
	// Setup local server
	localServer, localPort := setupLocalServer(t)
	defer localServer.Close()

	// Setup funnel server
	funnel, funnelServer, funnelURL := setupFunnelServer(t)
	defer funnelServer.Close()

	// Setup funnel client
	client := NewFunnelClient(localPort, funnelURL, "test-server")
	defer client.Stop()

	// Start client in background
	clientDone := make(chan error, 1)
	go func() {
		clientDone <- client.Start()
	}()

	// Wait for client to connect and verify connection
	connected := false
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		funnel.scLock.RLock()
		_, exists := funnel.serverConnections["test-server"]
		funnel.scLock.RUnlock()
		if exists {
			connected = true
			break
		}
	}

	if !connected {
		t.Fatalf("Client failed to connect to funnel")
	}

	// Connect to funnel via websocket
	wsURL := strings.Replace(funnelURL, "http://", "ws://", 1)
	wsURL = fmt.Sprintf("%s/route/test-server/ws", wsURL)

	conn, _, _, err := ws.Dial(context.TODO(), wsURL)
	if err != nil {
		t.Fatalf("Failed to connect to websocket: %v", err)
	}
	defer conn.Close()

	// Send message
	testMsg := []byte("hello websocket")
	err = wsutil.WriteClientText(conn, testMsg)
	if err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	// Read echo response
	msg, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	if !bytes.Equal(msg, testMsg) {
		t.Fatalf("Expected %s, got %s", string(testMsg), string(msg))
	}
}

func TestFunnel_MultipleRequests(t *testing.T) {
	// Setup local server
	localServer, localPort := setupLocalServer(t)
	defer localServer.Close()

	// Setup funnel server
	funnel, funnelServer, funnelURL := setupFunnelServer(t)
	defer funnelServer.Close()

	// Setup funnel client
	client := NewFunnelClient(localPort, funnelURL, "test-server")
	defer client.Stop()

	// Start client in background
	clientDone := make(chan error, 1)
	go func() {
		clientDone <- client.Start()
	}()

	// Wait for client to connect and verify connection
	connected := false
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		funnel.scLock.RLock()
		_, exists := funnel.serverConnections["test-server"]
		funnel.scLock.RUnlock()
		if exists {
			connected = true
			break
		}
	}

	if !connected {
		t.Fatalf("Client failed to connect to funnel")
	}

	// Make multiple requests with timeout
	clientHTTP := &http.Client{
		Timeout: 5 * time.Second,
	}
	for i := 0; i < 5; i++ {
		resp, err := clientHTTP.Get(fmt.Sprintf("%s/route/test-server/test", funnelURL))
		if err != nil {
			t.Fatalf("Failed to make request %d: %v", i, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Request %d: Expected status 200, got %d", i, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Request %d: Failed to read response: %v", i, err)
		}

		expected := `{"message":"test response"}`
		if string(body) != expected {
			t.Fatalf("Request %d: Expected %s, got %s", i, expected, string(body))
		}
	}
}

func TestFunnel_ServerNotConnected(t *testing.T) {
	// Setup funnel server without client
	_, funnelServer, funnelURL := setupFunnelServer(t)
	defer funnelServer.Close()

	// Try to make request to non-existent server
	resp, err := http.Get(fmt.Sprintf("%s/route/non-existent-server/test", funnelURL))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("Expected status 502, got %d", resp.StatusCode)
	}
}

*/
