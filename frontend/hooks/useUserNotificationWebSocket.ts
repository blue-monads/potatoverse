import { useRef, useCallback, useEffect } from "react";
import { getLoginData } from "@/lib/utils";
import * as api from "@/lib/api";

interface UseUserNotificationWebSocketOptions {
    isInitialized: boolean;
    isAuthenticated: boolean;
    onMessage: (message: api.UserMessage) => void;
    onFallback?: () => void;
}

interface UseUserNotificationWebSocketReturn {
    connect: () => void;
    disconnect: () => void;
    isConnected: () => boolean;
}

const MAX_RECONNECT_ATTEMPTS = 5;
const RECONNECT_DELAY = 3000; // 3 seconds
const FALLBACK_POLL_INTERVAL = 30000; // 30 seconds

// Singleton WebSocket manager to share connection across all hook instances
class WebSocketManager {
    private ws: WebSocket | null = null;
    private subscribers: Set<(message: api.UserMessage) => void> = new Set();
    private reconnectTimeout: NodeJS.Timeout | null = null;
    private fallbackInterval: NodeJS.Timeout | null = null;
    private reconnectAttempts = 0;
    private isConnecting = false;
    private shouldBeConnected = false;
    private connectionId: number = 0; // Track connection attempts to prevent duplicates

    subscribe(callback: (message: api.UserMessage) => void): () => void {
        this.subscribers.add(callback);
        return () => {
            this.subscribers.delete(callback);
        };
    }

    connect() {
        // If already connected and open, do nothing
        if (this.ws?.readyState === WebSocket.OPEN) {
            console.log("WebSocket already connected, skipping connection attempt");
            this.shouldBeConnected = true;
            return;
        }

        // If currently connecting, do nothing
        if (this.isConnecting) {
            console.log("WebSocket connection already in progress, skipping");
            return;
        }

        // If connection is in CONNECTING state, wait for it
        if (this.ws?.readyState === WebSocket.CONNECTING) {
            console.log("WebSocket is connecting, skipping duplicate connection");
            this.shouldBeConnected = true;
            return;
        }

        this.shouldBeConnected = true;

        const loginData = getLoginData();
        if (!loginData?.accessToken) {
            console.error("No access token available for WebSocket connection");
            return;
        }

        // Only close existing connection if it's in a bad state (CLOSED or CLOSING)
        if (this.ws) {
            const state = this.ws.readyState;
            if (state === WebSocket.CLOSED || state === WebSocket.CLOSING) {
                this.ws = null;
            } else {
                // Connection is in a valid state, don't close it
                console.log("WebSocket exists in valid state, not closing");
                return;
            }
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const token = loginData.accessToken;
        const wsUrl = `${protocol}//${window.location.host}/zz/api/core/user/messages/ws?token=${encodeURIComponent(token)}`;

        try {
            this.isConnecting = true;
            const currentConnectionId = ++this.connectionId;
            console.log(`[Connection #${currentConnectionId}] Creating new WebSocket connection to:`, wsUrl.replace(/token=[^&]+/, 'token=***'));
            const ws = new WebSocket(wsUrl);
            
            ws.onopen = () => {
                // Only proceed if this is still the current connection attempt
                if (currentConnectionId === this.connectionId) {
                    console.log(`[Connection #${currentConnectionId}] WebSocket connected for user notifications`);
                    this.reconnectAttempts = 0;
                    this.isConnecting = false;
                    this.shouldBeConnected = true;
                } else {
                    console.log(`[Connection #${currentConnectionId}] WebSocket opened but superseded by newer connection, closing`);
                    ws.close();
                }
            };

            ws.onmessage = (event) => {
                try {
                    const message: api.UserMessage = JSON.parse(event.data);
                    // Notify all subscribers
                    this.subscribers.forEach(callback => {
                        try {
                            callback(message);
                        } catch (error) {
                            console.error("Error in WebSocket subscriber:", error);
                        }
                    });
                } catch (error) {
                    console.error("Failed to parse WebSocket message:", error);
                }
            };

            ws.onerror = (error) => {
                if (currentConnectionId === this.connectionId) {
                    console.error(`[Connection #${currentConnectionId}] WebSocket error:`, error);
                    this.isConnecting = false;
                }
            };

            ws.onclose = (event) => {
                // Only handle close if this is the current connection
                if (currentConnectionId === this.connectionId) {
                    console.log(`[Connection #${currentConnectionId}] WebSocket closed`, event.code, event.reason, "shouldBeConnected:", this.shouldBeConnected);
                    this.ws = null;
                    this.isConnecting = false;

                    // Only attempt to reconnect if:
                    // 1. Not a normal closure (code 1000)
                    // 2. We haven't exceeded max attempts
                    // 3. We should still be connected (auth state hasn't changed)
                    if (event.code !== 1000 && 
                        this.reconnectAttempts < MAX_RECONNECT_ATTEMPTS &&
                        this.shouldBeConnected) {
                        this.reconnectAttempts += 1;
                        console.log(`[Connection #${currentConnectionId}] Attempting to reconnect (${this.reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})...`);
                        
                        this.reconnectTimeout = setTimeout(() => {
                            // Double-check we should still be connected before reconnecting
                            if (this.shouldBeConnected && currentConnectionId === this.connectionId) {
                                this.connect();
                            }
                        }, RECONNECT_DELAY);
                    } else if (this.reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
                        console.error(`[Connection #${currentConnectionId}] Max reconnection attempts reached. Falling back to polling.`);
                    } else if (event.code === 1000) {
                        console.log(`[Connection #${currentConnectionId}] WebSocket closed normally (code 1000)`);
                    }
                } else {
                    console.log(`[Connection #${currentConnectionId}] WebSocket closed but was superseded, ignoring`);
                }
            };

            // Only set as current connection if this is still the latest attempt
            if (currentConnectionId === this.connectionId) {
                this.ws = ws;
            } else {
                console.log(`[Connection #${currentConnectionId}] Connection attempt superseded, closing`);
                ws.close();
            }
        } catch (error) {
            console.error("Failed to create WebSocket connection:", error);
            this.isConnecting = false;
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
        if (this.fallbackInterval) {
            clearInterval(this.fallbackInterval);
            this.fallbackInterval = null;
        }
        this.reconnectAttempts = 0;
        this.isConnecting = false;
    }

    isConnected(): boolean {
        return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
    }
}

// Global singleton instance
const wsManager = new WebSocketManager();

export const useUserNotificationWebSocket = (
    options: UseUserNotificationWebSocketOptions
): UseUserNotificationWebSocketReturn => {
    const { isInitialized, isAuthenticated, onMessage, onFallback } = options;
    
    const prevAuthStateRef = useRef<{ initialized: boolean; authenticated: boolean } | null>(null);
    const unsubscribeRef = useRef<(() => void) | null>(null);
    const fallbackIntervalRef = useRef<NodeJS.Timeout | null>(null);

    // Subscribe to WebSocket messages
    useEffect(() => {
        const unsubscribe = wsManager.subscribe(onMessage);
        unsubscribeRef.current = unsubscribe;
        return () => {
            unsubscribe();
        };
    }, [onMessage]);

    // Handle fallback polling
    useEffect(() => {
        if (onFallback && !wsManager.isConnected()) {
            // Set up fallback polling if WebSocket is not connected
            fallbackIntervalRef.current = setInterval(() => {
                if (!wsManager.isConnected() && onFallback) {
                    onFallback();
                }
            }, FALLBACK_POLL_INTERVAL);
        }
        return () => {
            if (fallbackIntervalRef.current) {
                clearInterval(fallbackIntervalRef.current);
                fallbackIntervalRef.current = null;
            }
        };
    }, [onFallback]);

    const connect = useCallback(() => {
        wsManager.connect();
    }, []);

    const disconnect = useCallback(() => {
        wsManager.disconnect();
    }, []);

    const isConnected = useCallback(() => {
        return wsManager.isConnected();
    }, []);

    // Auto-connect when authenticated, disconnect when not
    // Only track auth state changes
    useEffect(() => {
        const prevState = prevAuthStateRef.current;
        const currentState = { initialized: isInitialized, authenticated: isAuthenticated };

        setTimeout(() => {


            if (!prevState || 
                prevState.initialized !== isInitialized || 
                prevState.authenticated !== isAuthenticated) {
                
                console.log("Auth state changed:", {
                    prev: prevState,
                    current: currentState,
                    isConnected: wsManager.isConnected()
                });
                
                if (isInitialized && isAuthenticated) {
                    // Only connect if not already connected
                    if (!wsManager.isConnected()) {
                        connect();
                    } else {
                        console.log("WebSocket already connected, skipping connect()");
                    }
                } else {
                    disconnect();
                }
                
                prevAuthStateRef.current = currentState;
            }

        }, 2000);
        
        
        
        // Don't disconnect in cleanup - let the connection persist across route changes
        // It will only disconnect when auth state explicitly changes to false
    }, [isInitialized, isAuthenticated, connect, disconnect]);

    return {
        connect,
        disconnect,
        isConnected,
    };
};
