import { useEffect, useState, useCallback } from "react";
import { GAppStateHandle } from "./useGAppState";
import * as api from "@/lib/api";

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

    useEffect(() => {
        if (gapp.isInitialized && gapp.isAuthenticated) {
            loadNewMessages();
            // Optionally reload periodically
            const interval = setInterval(loadNewMessages, 30000); // every 30 seconds
            return () => clearInterval(interval);
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