"use client";
import React, { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { Mail, CheckCheck, Trash2, Filter, Search, ArrowLeft, MoreVertical } from 'lucide-react';
import { 
    listUserMessages, 
    setMessageAsRead, 
    setAllMessagesAsRead,
    deleteUserMessage,
    UserMessage 
} from '@/lib/api';
import { useGApp } from '@/hooks';

type FilterType = 'all' | 'unread' | 'read';

export default function Page() {
    return (<>
        <MessagesPage />
    </>)
}

const MessagesPage = () => {
    const router = useRouter();
    const gapp = useGApp();
    const [messages, setMessages] = useState<UserMessage[]>([]);
    const [filteredMessages, setFilteredMessages] = useState<UserMessage[]>([]);
    const [loading, setLoading] = useState(true);
    const [filter, setFilter] = useState<FilterType>('all');
    const [searchQuery, setSearchQuery] = useState('');
    const [lastMessageId, setLastMessageId] = useState<number | null>(null);
    const [hasMore, setHasMore] = useState(true);
    const limit = 50;
    const hasLoadedRef = useRef(false);

    useEffect(() => {
        if (!gapp.isInitialized || !gapp.isAuthenticated) {
            router.push('/auth/login');
            return;
        }
        // Only load once when initialized and authenticated
        if (!hasLoadedRef.current) {
            hasLoadedRef.current = true;
            loadMessages(true);
        }
    }, [gapp.isInitialized, gapp.isAuthenticated]);

    useEffect(() => {
        applyFilters();
    }, [messages, filter, searchQuery]);

    const loadMessages = async (reset = false) => {
        if (loading && !reset) {
            // Prevent concurrent loads unless resetting
            return;
        }
        try {
            setLoading(true);
            // Use last message ID as cursor for pagination
            const cursor = reset ? undefined : (lastMessageId ?? undefined);
            const response = await listUserMessages(cursor, limit);
            const newMessages = response.data;
            
            if (reset) {
                setMessages(newMessages);
                setLastMessageId(null);
            } else {
                setMessages(prev => [...prev, ...newMessages]);
            }
            
            // Update cursor to the last message ID (smallest ID since we're ordering by -id)
            if (newMessages.length > 0) {
                const lastId = newMessages[newMessages.length - 1].id;
                setLastMessageId(lastId);
            }
            
            setHasMore(newMessages.length === limit);
        } catch (error) {
            console.error('Failed to load messages:', error);
        } finally {
            setLoading(false);
        }
    };

    const applyFilters = () => {
        let filtered = [...messages];

        // Apply read/unread filter
        if (filter === 'unread') {
            filtered = filtered.filter(msg => !msg.is_read);
        } else if (filter === 'read') {
            filtered = filtered.filter(msg => msg.is_read);
        }

        // Apply search query
        if (searchQuery.trim()) {
            const query = searchQuery.toLowerCase();
            filtered = filtered.filter(msg => 
                msg.title.toLowerCase().includes(query) ||
                msg.contents.toLowerCase().includes(query) ||
                msg.type.toLowerCase().includes(query)
            );
        }

        setFilteredMessages(filtered);
    };

    const handleMarkAsRead = async (messageId: number) => {
        try {
            await setMessageAsRead(messageId);
            setMessages(prev => 
                prev.map(msg => msg.id === messageId ? { ...msg, is_read: true } : msg)
            );
        } catch (error) {
            console.error('Failed to mark message as read:', error);
        }
    };

    const handleMarkAllAsRead = async () => {
        try {
            await setAllMessagesAsRead();
            setMessages(prev => prev.map(msg => ({ ...msg, is_read: true })));
        } catch (error) {
            console.error('Failed to mark all as read:', error);
        }
    };

    const handleDelete = async (messageId: number) => {
        if (!confirm('Are you sure you want to delete this message?')) {
            return;
        }
        try {
            await deleteUserMessage(messageId);
            setMessages(prev => prev.filter(msg => msg.id !== messageId));
        } catch (error) {
            console.error('Failed to delete message:', error);
        }
    };

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
        return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
    };

    const getWarnLevelColor = (warnLevel: number) => {
        if (warnLevel >= 3) return 'border-l-red-500 bg-red-50/30';
        if (warnLevel >= 2) return 'border-l-orange-500 bg-orange-50/30';
        if (warnLevel >= 1) return 'border-l-yellow-500 bg-yellow-50/30';
        return 'border-l-blue-500';
    };

    const unreadCount = messages.filter(msg => !msg.is_read).length;

    return (
        <div className="min-h-screen bg-gray-50 p-8">
            <div className="max-w-6xl mx-auto">
                {/* Header */}
                <div className="mb-6">
                    <button
                        onClick={() => router.back()}
                        className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4 transition-colors"
                    >
                        <ArrowLeft className="w-4 h-4" />
                        <span>Back</span>
                    </button>
                    <div className="flex items-center justify-between">
                        <div>
                            <h1 className="text-2xl font-bold text-gray-900">Messages</h1>
                            <p className="text-sm text-gray-500 mt-1">
                                {messages.length} message{messages.length !== 1 ? 's' : ''}
                                {unreadCount > 0 && (
                                    <span className="ml-2 px-2 py-0.5 bg-red-500 text-white text-xs rounded-full">
                                        {unreadCount} unread
                                    </span>
                                )}
                            </p>
                        </div>
                        {unreadCount > 0 && (
                            <button
                                onClick={handleMarkAllAsRead}
                                className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
                            >
                                <CheckCheck className="w-4 h-4" />
                                <span>Mark all as read</span>
                            </button>
                        )}
                    </div>
                </div>

                {/* Filters and Search */}
                <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 mb-6">
                    <div className="flex flex-col md:flex-row gap-4">
                        {/* Search */}
                        <div className="flex-1 relative">
                            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
                            <input
                                type="text"
                                placeholder="Search messages..."
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                        </div>

                        {/* Filter buttons */}
                        <div className="flex items-center gap-2">
                            <Filter className="w-5 h-5 text-gray-400" />
                            <button
                                onClick={() => setFilter('all')}
                                className={`px-4 py-2 rounded-lg transition-colors ${
                                    filter === 'all'
                                        ? 'bg-blue-500 text-white'
                                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                                }`}
                            >
                                All
                            </button>
                            <button
                                onClick={() => setFilter('unread')}
                                className={`px-4 py-2 rounded-lg transition-colors ${
                                    filter === 'unread'
                                        ? 'bg-blue-500 text-white'
                                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                                }`}
                            >
                                Unread
                            </button>
                            <button
                                onClick={() => setFilter('read')}
                                className={`px-4 py-2 rounded-lg transition-colors ${
                                    filter === 'read'
                                        ? 'bg-blue-500 text-white'
                                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                                }`}
                            >
                                Read
                            </button>
                        </div>
                    </div>
                </div>

                {/* Messages List */}
                <div className="bg-white rounded-lg shadow-sm border border-gray-200">
                    {loading && messages.length === 0 ? (
                        <div className="flex items-center justify-center py-12">
                            <div className="text-gray-400">Loading messages...</div>
                        </div>
                    ) : filteredMessages.length === 0 ? (
                        <div className="flex flex-col items-center justify-center py-12">
                            <Mail className="w-12 h-12 text-gray-300 mb-4" />
                            <p className="text-gray-500 font-medium">No messages found</p>
                            <p className="text-sm text-gray-400 mt-1">
                                {searchQuery ? 'Try adjusting your search' : "You're all caught up!"}
                            </p>
                        </div>
                    ) : (
                        <>
                            <div className="divide-y divide-gray-100">
                                {filteredMessages.map((message) => (
                                    <div
                                        key={message.id}
                                        className={`
                                            p-6 hover:bg-gray-50 transition-colors
                                            border-l-4 ${getWarnLevelColor(message.warn_level)}
                                            ${!message.is_read ? 'bg-blue-50/50' : ''}
                                        `}
                                    >
                                        <div className="flex items-start gap-4">
                                            <div className="flex-1 min-w-0">
                                                <div className="flex items-start justify-between gap-4 mb-2">
                                                    <div className="flex-1">
                                                        <div className="flex items-center gap-2 mb-1">
                                                            <h3 className="text-lg font-semibold text-gray-900">
                                                                {message.title}
                                                            </h3>
                                                            {!message.is_read && (
                                                                <span className="w-2 h-2 bg-blue-500 rounded-full flex-shrink-0"></span>
                                                            )}
                                                        </div>
                                                        <div className="flex items-center gap-2 text-xs text-gray-500 mb-2">
                                                            <span className="px-2 py-0.5 bg-gray-100 rounded text-gray-600">
                                                                {message.type}
                                                            </span>
                                                            <span>{formatDate(message.created_at)}</span>
                                                            {message.warn_level > 0 && (
                                                                <span className="px-2 py-0.5 bg-red-100 text-red-600 rounded">
                                                                    Level {message.warn_level}
                                                                </span>
                                                            )}
                                                        </div>
                                                    </div>
                                                    <div className="flex items-center gap-2">
                                                        {!message.is_read && (
                                                            <button
                                                                onClick={() => handleMarkAsRead(message.id)}
                                                                className="p-2 text-gray-500 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                                                                title="Mark as read"
                                                            >
                                                                <CheckCheck className="w-4 h-4" />
                                                            </button>
                                                        )}
                                                        <button
                                                            onClick={() => handleDelete(message.id)}
                                                            className="p-2 text-gray-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                                                            title="Delete"
                                                        >
                                                            <Trash2 className="w-4 h-4" />
                                                        </button>
                                                    </div>
                                                </div>
                                                <p className="text-gray-700 leading-relaxed whitespace-pre-wrap">
                                                    {message.contents}
                                                </p>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>

                            {/* Load More */}
                            {hasMore && !loading && (
                                <div className="p-4 border-t border-gray-200 text-center">
                                    <button
                                        onClick={() => loadMessages(false)}
                                        className="px-4 py-2 text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                                    >
                                        Load more messages
                                    </button>
                                </div>
                            )}

                            {loading && messages.length > 0 && (
                                <div className="p-4 border-t border-gray-200 text-center text-gray-500">
                                    Loading...
                                </div>
                            )}
                        </>
                    )}
                </div>
            </div>
        </div>
    );
}
