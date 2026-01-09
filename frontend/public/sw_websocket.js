// Shared Worker for Admin Portal
console.log('hello from shared worker');

// WebSocket Manager for Shared Worker
class SharedWorkerWebSocketManager {
  constructor() {
    this.ws = null;
    this.reconnectTimeout = null;
    this.reconnectAttempts = 0;
    this.isConnecting = false;
    this.shouldBeConnected = false;
    this.connectionId = 0;
    this.MAX_RECONNECT_ATTEMPTS = 5;
    this.RECONNECT_DELAY = 3000; // 3 seconds
    this.connectedPorts = new Set(); // Track all connected ports
    this.storedToken = null;
    this.storedHost = null;
  }

  addPort(port) {
    this.connectedPorts.add(port);
    port.onmessage = (event) => {
      this.handleMessage(event.data, port);
    };
    port.onclose = () => {
      this.connectedPorts.delete(port);
    };
    // Start port
    port.start();
  }

  handleMessage(data, sourcePort) {
    console.log('[SW] Shared Worker received message:', data);
    
    const { type, token, host } = data;
    
    if (type === 'ping') {
      // Respond with pong but don't connect
      sourcePort.postMessage({ type: 'pong' });
    } else if (type === 'ws-connect') {
      // Connect to websocket with provided token
      const hostname = host || new URL(self.location.origin).host;
      this.storedToken = token;
      this.storedHost = hostname;
      this.connect(token, hostname);
    } else if (type === 'ws-disconnect') {
      // Disconnect websocket
      this.disconnect();
    } else if (type === 'ws-status') {
      // Return connection status
      sourcePort.postMessage({ 
        type: 'ws-status-response', 
        isConnected: this.isConnected() 
      });
    }
  }

  connect(token, host) {
    // If already connected and open, do nothing
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('[SW] WebSocket already connected, skipping connection attempt');
      this.shouldBeConnected = true;
      return;
    }

    // If currently connecting, do nothing
    if (this.isConnecting) {
      console.log('[SW] WebSocket connection already in progress, skipping');
      return;
    }

    // If connection is in CONNECTING state, wait for it
    if (this.ws?.readyState === WebSocket.CONNECTING) {
      console.log('[SW] WebSocket is connecting, skipping duplicate connection');
      this.shouldBeConnected = true;
      return;
    }

    this.shouldBeConnected = true;

    if (!token) {
      console.error('[SW] No access token available for WebSocket connection');
      return;
    }

    // Only close existing connection if it's in a bad state (CLOSED or CLOSING)
    if (this.ws) {
      const state = this.ws.readyState;
      if (state === WebSocket.CLOSED || state === WebSocket.CLOSING) {
        this.ws = null;
      } else {
        // Connection is in a valid state, don't close it
        console.log('[SW] WebSocket exists in valid state, not closing');
        return;
      }
    }

    // Use stored host if available, otherwise use provided host
    const hostname = host || this.storedHost || new URL(self.location.origin).host;
    const protocol = self.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${hostname}/zz/api/core/user/messages/ws?token=${encodeURIComponent(token)}`;

    try {
      this.isConnecting = true;
      const currentConnectionId = ++this.connectionId;
      console.log(`[SW Connection #${currentConnectionId}] Creating new WebSocket connection`);
      const ws = new WebSocket(wsUrl);
      
      ws.onopen = () => {
        // Only proceed if this is still the current connection attempt
        if (currentConnectionId === this.connectionId) {
          console.log(`[SW Connection #${currentConnectionId}] WebSocket connected for user notifications`);
          this.reconnectAttempts = 0;
          this.isConnecting = false;
          this.shouldBeConnected = true;
          this.broadcastToClients({ type: 'ws-connected' });
        } else {
          console.log(`[SW Connection #${currentConnectionId}] WebSocket opened but superseded by newer connection, closing`);
          ws.close();
        }
      };

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          // Forward message to all clients
          console.log('[SW] Received WebSocket message, forwarding to clients:', message);
          this.broadcastToClients({ 
            type: 'ws-message', 
            message: message 
          });
        } catch (error) {
          console.error('[SW] Failed to parse WebSocket message:', error);
        }
      };

      ws.onerror = (error) => {
        if (currentConnectionId === this.connectionId) {
          console.error(`[SW Connection #${currentConnectionId}] WebSocket error:`, error);
          this.isConnecting = false;
          this.broadcastToClients({ type: 'ws-error', error: 'WebSocket error occurred' });
        }
      };

      ws.onclose = (event) => {
        // Only handle close if this is the current connection
        if (currentConnectionId === this.connectionId) {
          console.log(`[SW Connection #${currentConnectionId}] WebSocket closed`, event.code, event.reason, "shouldBeConnected:", this.shouldBeConnected);
          this.ws = null;
          this.isConnecting = false;
          this.broadcastToClients({ type: 'ws-closed', code: event.code, reason: event.reason });

          // Only attempt to reconnect if:
          // 1. Not a normal closure (code 1000)
          // 2. We haven't exceeded max attempts
          // 3. We should still be connected (auth state hasn't changed)
          if (event.code !== 1000 && 
              this.reconnectAttempts < this.MAX_RECONNECT_ATTEMPTS &&
              this.shouldBeConnected) {
            this.reconnectAttempts += 1;
            console.log(`[SW Connection #${currentConnectionId}] Attempting to reconnect (${this.reconnectAttempts}/${this.MAX_RECONNECT_ATTEMPTS})...`);
            
            this.reconnectTimeout = setTimeout(() => {
              // Double-check we should still be connected before reconnecting
              if (this.shouldBeConnected && currentConnectionId === this.connectionId) {
                // Use stored token and host for reconnection
                if (this.storedToken && this.storedHost) {
                  this.connect(this.storedToken, this.storedHost);
                } else {
                  // Request reconnection from clients
                  this.broadcastToClients({ type: 'ws-reconnect-needed' });
                }
              }
            }, this.RECONNECT_DELAY);
          } else if (this.reconnectAttempts >= this.MAX_RECONNECT_ATTEMPTS) {
            console.error(`[SW Connection #${currentConnectionId}] Max reconnection attempts reached.`);
            this.broadcastToClients({ type: 'ws-max-reconnect-reached' });
          } else if (event.code === 1000) {
            console.log(`[SW Connection #${currentConnectionId}] WebSocket closed normally (code 1000)`);
          }
        } else {
          console.log(`[SW Connection #${currentConnectionId}] WebSocket closed but was superseded, ignoring`);
        }
      };

      // Only set as current connection if this is still the latest attempt
      if (currentConnectionId === this.connectionId) {
        this.ws = ws;
      } else {
        console.log(`[SW Connection #${currentConnectionId}] Connection attempt superseded, closing`);
        ws.close();
      }
    } catch (error) {
      console.error('[SW] Failed to create WebSocket connection:', error);
      this.isConnecting = false;
      this.broadcastToClients({ type: 'ws-error', error: error.message });
    }
  }

  disconnect() {
    this.shouldBeConnected = false;
    
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    this.reconnectAttempts = 0;
    this.isConnecting = false;
    this.broadcastToClients({ type: 'ws-disconnected' });
  }

  isConnected() {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  broadcastToClients(message) {
    // Send message to all connected ports
    this.connectedPorts.forEach(port => {
      try {
        port.postMessage(message);
      } catch (error) {
        console.error('[SW] Error sending message to port:', error);
        // Remove dead port
        this.connectedPorts.delete(port);
      }
    });
  }
}

// Global singleton instance
const swWsManager = new SharedWorkerWebSocketManager();

// Connect event - fired when a new port connects to the shared worker
self.addEventListener('connect', (event) => {
  console.log('[SW] New port connecting to shared worker');
  const port = event.ports[0];
  swWsManager.addPort(port);
});
