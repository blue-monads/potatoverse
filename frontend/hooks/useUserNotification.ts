import { useEffect, useState, useCallback } from "react";
import { GAppStateHandle } from "./useGAppState";
import * as api from "@/lib/api";
import { useUserNotificationWebSocket } from "./useUserNotificationWebSocket";

const useUserNotification = (gapp: GAppStateHandle) => {
    const [notifications, setNotifications] = useState<api.UserMessage[]>([]);
    const [unreadCount, setUnreadCount] = useState(0);
    const [loading, setLoading] = useState(false);

    const loadNewMessages = useCallback(async () => {
        if (!gapp.isInitialized || !gapp.isAuthenticated) {
            return;
        }

        try {
            setLoading(true);
            const response = await api.queryNewMessages();
            const newMessages = response.data;
            setNotifications(newMessages);
            setUnreadCount(newMessages.filter(msg => !msg.is_read).length);
        } catch (error) {
            console.error("Failed to load new messages:", error);
        } finally {
            setLoading(false);
        }
    }, [gapp.isInitialized, gapp.isAuthenticated]);

    const loadMessageHistory = useCallback(async (limit: number = 100) => {
        if (!gapp.isInitialized || !gapp.isAuthenticated) {
            return;
        }

        try {
            setLoading(true);
            const response = await api.queryMessageHistory(limit);
            const messages = response.data;
            setNotifications(messages);
            setUnreadCount(messages.filter(msg => !msg.is_read).length);
        } catch (error) {
            console.error("Failed to load message history:", error);
        } finally {
            setLoading(false);
        }
    }, [gapp.isInitialized, gapp.isAuthenticated]);

    const markAsRead = useCallback(async (messageId: number) => {
        try {
            await api.setMessageAsRead(messageId);
            setNotifications(prev => 
                prev.map(msg => msg.id === messageId ? { ...msg, is_read: true } : msg)
            );
            setUnreadCount(prev => Math.max(0, prev - 1));
        } catch (error) {
            console.error("Failed to mark message as read:", error);
        }
    }, []);

    const markAllAsRead = useCallback(async () => {
        try {
            await api.setAllMessagesAsRead();
            setNotifications(prev => prev.map(msg => ({ ...msg, is_read: true })));
            setUnreadCount(0);
        } catch (error) {
            console.error("Failed to mark all messages as read:", error);
        }
    }, []);

    // Handle incoming WebSocket messages
    const handleWebSocketMessage = useCallback((message: api.UserMessage) => {
        // Add or update the message in notifications
        setNotifications(prev => {
            // Check if message already exists
            const existingIndex = prev.findIndex(msg => msg.id === message.id);
            if (existingIndex >= 0) {
                // Update existing message
                const existing = prev[existingIndex];
                const updated = [...prev];
                updated[existingIndex] = message;
                
                // Update unread count based on read status change
                setUnreadCount(current => {
                    let newCount = current;
                    if (!existing.is_read && message.is_read) {
                        // Message was marked as read
                        newCount = Math.max(0, newCount - 1);
                    } else if (existing.is_read && !message.is_read) {
                        // Message was marked as unread (unlikely but handle it)
                        newCount = newCount + 1;
                    }
                    return newCount;
                });
                
                return updated;
            } else {
                // Add new message at the beginning
                // Update unread count if message is unread
                if (!message.is_read) {
                    setUnreadCount(prev => prev + 1);
                }
                return [message, ...prev];
            }
        });
    }, []);

    // Use WebSocket hook
    useUserNotificationWebSocket({
        isInitialized: gapp.isInitialized,
        isAuthenticated: gapp.isAuthenticated,
        onMessage: handleWebSocketMessage,
        onFallback: loadNewMessages,
    });

    // Load initial messages when authenticated
    useEffect(() => {
        if (gapp.isInitialized && gapp.isAuthenticated) {
            loadNewMessages();
        }
    }, [gapp.isInitialized, gapp.isAuthenticated, loadNewMessages]);

    return {
        notifications,
        unreadCount,
        loading,
        loadNewMessages,
        loadMessageHistory,
        markAsRead,
        markAllAsRead,
    }
}

export default useUserNotification;