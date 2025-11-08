"use client";
import { useEffect, useRef } from "react";
import Link from "next/link";
import { X, CheckCheck, MessageCircleIcon, ExternalLink } from "lucide-react";
import { useGApp } from "@/hooks";
import useUserNotification from "@/hooks/useUserNotification";


interface MessagePanelProps {
    isOpen: boolean;
    onClose: () => void;
}

const MessagePanel = ({ isOpen, onClose }: MessagePanelProps) => {
    const gapp = useGApp();
    const notifier = useUserNotification(gapp);
    const hasLoadedRef = useRef(false);

    // Load message history when panel opens (only once per open)
    useEffect(() => {
        if (isOpen && gapp.isInitialized && gapp.isAuthenticated && !hasLoadedRef.current) {
            notifier.loadMessageHistory(50);
            hasLoadedRef.current = true;
        }
        // Reset the flag when panel closes
        if (!isOpen) {
            hasLoadedRef.current = false;
        }
    }, [isOpen, gapp.isInitialized, gapp.isAuthenticated, notifier.loadMessageHistory]);

    // Close on Escape key
    useEffect(() => {
        if (!isOpen) return;
        const handleEscape = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                onClose();
            }
        };
        window.addEventListener('keydown', handleEscape);
        return () => window.removeEventListener('keydown', handleEscape);
    }, [isOpen, onClose]);

    const formatDate = (dateString?: string) => {
        if (!dateString) return '';
        const date = new Date(dateString);
        const now = new Date();
        const diffMs = now.getTime() - date.getTime();
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMs / 3600000);
        const diffDays = Math.floor(diffMs / 86400000);

        if (diffMins < 1) return 'Just now';
        if (diffMins < 60) return `${diffMins}m ago`;
        if (diffHours < 24) return `${diffHours}h ago`;
        if (diffDays < 7) return `${diffDays}d ago`;
        return date.toLocaleDateString();
    };

    const getWarnLevelColor = (warnLevel: number) => {
        if (warnLevel >= 3) return 'border-l-red-500';
        if (warnLevel >= 2) return 'border-l-orange-500';
        if (warnLevel >= 1) return 'border-l-yellow-500';
        return 'border-l-blue-500';
    };

    return (
        <>
            {/* Backdrop - only shows when open */}
            {isOpen && (
                <div 
                    className="fixed inset-0 bg-black/20 z-40 transition-opacity"
                    onClick={onClose}
                />
            )}

            {/* Panel - always rendered but slides off-screen when closed */}
            <div className={`
                fixed left-0 top-0 h-full w-96 bg-white shadow-2xl z-50
                transform transition-transform duration-300 ease-in-out
                flex flex-col
                ${isOpen ? 'translate-x-0' : '-translate-x-full'}
            `}>
                {/* Header */}
                <div className="flex items-center justify-between p-4 border-b border-gray-200 bg-gray-50">
                    <div className="flex items-center gap-2">
                        <h2 className="text-lg font-semibold text-gray-900">Notifications</h2>
                        {notifier.unreadCount > 0 && (
                            <span className="px-2 py-0.5 text-xs font-medium text-white bg-red-500 rounded-full">
                                {notifier.unreadCount}
                            </span>
                        )}
                    </div>
                    <div className="flex items-center gap-2">
                        <Link
                            href="/portal/admin/profile/messages"
                            onClick={onClose}
                            className="p-1.5 text-gray-500 hover:text-gray-700 hover:bg-gray-200 rounded-lg transition-colors"
                            title="See all messages"
                        >
                            <ExternalLink className="w-5 h-5" />
                        </Link>
                        {notifier.unreadCount > 0 && (
                            <button
                                onClick={notifier.markAllAsRead}
                                className="p-1.5 text-gray-500 hover:text-gray-700 hover:bg-gray-200 rounded-lg transition-colors"
                                title="Mark all as read"
                            >
                                <CheckCheck className="w-5 h-5" />
                            </button>
                        )}
                        <button
                            onClick={onClose}
                            className="p-1.5 text-gray-500 hover:text-gray-700 hover:bg-gray-200 rounded-lg transition-colors"
                            title="Close"
                        >
                            <X className="w-5 h-5" />
                        </button>
                    </div>
                </div>

                {/* Messages List */}
                <div className="flex-1 overflow-y-auto">
                    {notifier.loading && notifier.notifications.length === 0 ? (
                        <div className="flex items-center justify-center h-full">
                            <div className="text-gray-400">Loading...</div>
                        </div>
                    ) : notifier.notifications.length === 0 ? (
                        <div className="flex flex-col items-center justify-center h-full p-8 text-center">
                            <MessageCircleIcon className="w-12 h-12 text-gray-300 mb-4" />
                            <p className="text-gray-500 font-medium">No notifications</p>
                            <p className="text-sm text-gray-400 mt-1">You're all caught up!</p>
                        </div>
                    ) : (
                        <div className="divide-y divide-gray-100">
                            {notifier.notifications.map((message) => (
                                <div
                                    key={message.id}
                                    className={`
                                        p-4 hover:bg-gray-50 transition-colors cursor-pointer
                                        border-l-4 ${getWarnLevelColor(message.warn_level)}
                                        ${!message.is_read ? 'bg-blue-50/50' : ''}
                                    `}
                                    onClick={() => {
                                        if (!message.is_read) {
                                            notifier.markAsRead(message.id);
                                        }
                                    }}
                                >
                                    <div className="flex items-start gap-3">
                                        <div className="flex-1 min-w-0">
                                            <div className="flex items-start justify-between gap-2 mb-1">
                                                <h3 className="text-sm font-semibold text-gray-900 truncate">
                                                    {message.title}
                                                </h3>
                                                {!message.is_read && (
                                                    <span className="w-2 h-2 bg-blue-500 rounded-full flex-shrink-0 mt-1.5"></span>
                                                )}
                                            </div>
                                            <p className="text-sm text-gray-600 line-clamp-2 mb-2">
                                                {message.contents}
                                            </p>
                                            <div className="flex items-center justify-between gap-2">
                                                <div className="flex items-center gap-2 text-xs text-gray-400">
                                                    <span className="px-1.5 py-0.5 bg-gray-100 rounded text-gray-600">
                                                        {message.type}
                                                    </span>
                                                    <span>{formatDate(message.created_at)}</span>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </>
    );
}

export default MessagePanel;
