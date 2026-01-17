"use client";
import React, { useState, useEffect } from 'react';
import { User, Mail, ArrowLeft, Calendar, MessageSquare, X } from 'lucide-react';
import { getUser, User as UserType, sendUserMessage } from '../../../../../lib/api';
import { useGApp } from '../../../../../hooks/contexts/GAppStateContext';
import { useSearchParams, useRouter } from 'next/navigation';

export default function Page() {
    return (<>
        <UserProfileViewer />
    </>)
}

const UserProfileViewer = () => {
    const { loaded, isInitialized, isAuthenticated } = useGApp();
    const searchParams = useSearchParams();
    const router = useRouter();
    const [user, setUser] = useState<UserType | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [showMessageModal, setShowMessageModal] = useState(false);
    const [sending, setSending] = useState(false);
    const [messageError, setMessageError] = useState<string | null>(null);
    const [messageSuccess, setMessageSuccess] = useState(false);
    const [messageForm, setMessageForm] = useState({
        title: '',
        contents: '',
    });

    const userId = searchParams.get('user');

    useEffect(() => {
        // Only load user info when the app state is fully loaded and initialized
        if (loaded && isInitialized && isAuthenticated && userId) {
            loadUserInfo();
        }
    }, [loaded, isInitialized, isAuthenticated, userId]);

    const loadUserInfo = async () => {
        if (!userId) {
            setError('No user ID provided');
            setLoading(false);
            return;
        }

        try {
            setLoading(true);
            setError(null);
            const response = await getUser(parseInt(userId));
            setUser(response.data);
        } catch (error: any) {
            console.error('Failed to load user info:', error);
            setError(error.response?.data?.message || 'Failed to load user profile');
        } finally {
            setLoading(false);
        }
    };

    const formatDate = (dateString: string) => {
        if (!dateString) return 'Unknown';
        try {
            const date = new Date(dateString);
            return date.toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'long',
                day: 'numeric'
            });
        } catch {
            return 'Unknown';
        }
    };

    const handleBack = () => {
        router.back();
    };

    const handleSendMessage = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!user || !messageForm.title.trim() || !messageForm.contents.trim()) {
            setMessageError('Title and message are required');
            return;
        }

        try {
            setSending(true);
            setMessageError(null);
            setMessageSuccess(false);
            
            await sendUserMessage({
                title: messageForm.title.trim(),
                type: 'info',
                contents: messageForm.contents.trim(),
                to_user: user.id,
            });

            setMessageSuccess(true);
            setMessageForm({ title: '', contents: '' });
            
            // Close modal after 1.5 seconds
            setTimeout(() => {
                setShowMessageModal(false);
                setMessageSuccess(false);
            }, 1500);
        } catch (error: any) {
            console.error('Failed to send message:', error);
            setMessageError(error.response?.data?.message || 'Failed to send message');
        } finally {
            setSending(false);
        }
    };

    const handleCloseModal = () => {
        setShowMessageModal(false);
        setMessageError(null);
        setMessageSuccess(false);
        setMessageForm({ title: '', contents: '' });
    };

    // Show loading if app state is not ready or if we're loading user data
    if (!loaded || !isInitialized || loading) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
                    <p className="text-gray-600">
                        {!loaded || !isInitialized ? 'Initializing...' : 'Loading user profile...'}
                    </p>
                </div>
            </div>
        );
    }

    // Show error if not authenticated
    if (!isAuthenticated) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-center">
                    <p className="text-gray-600">Please log in to view user profiles</p>
                </div>
            </div>
        );
    }

    // Show error if no user ID provided
    if (!userId) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-center">
                    <p className="text-gray-600">No user ID provided</p>
                    <button
                        onClick={handleBack}
                        className="mt-4 text-blue-600 hover:text-blue-800"
                    >
                        Go Back
                    </button>
                </div>
            </div>
        );
    }

    // Show error if failed to load user
    if (error) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-center">
                    <p className="text-red-600 mb-4">{error}</p>
                    <button
                        onClick={handleBack}
                        className="text-blue-600 hover:text-blue-800"
                    >
                        Go Back
                    </button>
                </div>
            </div>
        );
    }

    if (!user) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-center">
                    <p className="text-gray-600">User not found</p>
                    <button
                        onClick={handleBack}
                        className="mt-4 text-blue-600 hover:text-blue-800"
                    >
                        Go Back
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gray-50">
            {/* Header */}
            <header className="bg-white border-b border-gray-200 px-6 py-4">
                <div className="max-w-4xl mx-auto flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <button
                            onClick={handleBack}
                            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                        >
                            <ArrowLeft className="w-5 h-5 text-gray-600" />
                        </button>
                        <div className="flex items-center gap-2">
                            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                                <User className="w-5 h-5 text-white" />
                            </div>
                            <div>
                                <h1 className="text-xl font-bold">User Profile</h1>
                                <p className="text-sm text-gray-600">View User Information</p>
                            </div>
                        </div>
                    </div>
                    {user && (
                        <button
                            onClick={() => setShowMessageModal(true)}
                            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                        >
                            <MessageSquare className="w-4 h-4" />
                            <span>Send Message</span>
                        </button>
                    )}
                </div>
            </header>

            <div className="max-w-4xl mx-auto px-6 py-8">
                <div className="bg-white rounded-xl border border-gray-200 p-8">
                    <div className="text-center mb-8">
                        <div className=" rounded-full flex items-center justify-center text-white text-2xl font-bold mx-auto mb-4">
                            <img src={`/zz/profileImage/${user.id}/${user.name}`} alt="User Profile" className="w-20 h-20 rounded-full" />
                        </div>
                        <h2 className="text-2xl font-bold text-gray-900 mb-1">{user.name}</h2>
                        {user.username && (
                            <p className="text-blue-600 font-medium mb-2">@{user.username}</p>
                        )}
                        <p className="text-sm text-gray-500 mb-4">
                            {user.utype} • {user.ugroup}
                        </p>
                    </div>

                    {/* Bio Section */}
                    <div className="mb-8">
                        <h3 className="font-semibold text-gray-900 mb-3">About</h3>
                        <div className="p-4 bg-gray-50 rounded-lg">
                            {user.bio ? (
                                <p className="text-gray-700 leading-relaxed whitespace-pre-wrap">{user.bio}</p>
                            ) : (
                                <p className="text-gray-500 italic">No bio available</p>
                            )}
                        </div>
                    </div>

                    {/* Contact Information */}
                    <div className="space-y-4">
                        <h3 className="font-semibold text-gray-900 mb-3">Contact Information</h3>

                        <div className="flex items-center gap-3 text-gray-600">
                            <Mail className="w-5 h-5" />
                            <span className="text-sm">{user.email}</span>
                        </div>

                        {user.phone && (
                            <div className="flex items-center gap-3 text-gray-600">
                                <User className="w-5 h-5" />
                                <span className="text-sm">{user.phone}</span>
                            </div>
                        )}

                        <div className="flex items-center gap-3 text-gray-600">
                            <div className="w-5 h-5 flex items-center justify-center">
                                <div className={`w-3 h-3 rounded-full ${user.is_verified ? 'bg-green-500' : 'bg-gray-400'}`}></div>
                            </div>
                            <span className="text-sm">
                                {user.is_verified ? 'Verified Account' : 'Unverified Account'}
                            </span>
                        </div>

                        {user.createdAt && (
                            <div className="flex items-center gap-3 text-gray-600">
                                <Calendar className="w-5 h-5" />
                                <span className="text-sm">Joined {formatDate(user.createdAt)}</span>
                            </div>
                        )}
                    </div>

                    {/* Account Status */}
                    <div className="mt-8 pt-6 border-t border-gray-200">
                        <h3 className="font-semibold text-gray-900 mb-3">Account Status</h3>
                        <div className="grid grid-cols-2 gap-4">
                            <div className="p-3 bg-gray-50 rounded-lg">
                                <p className="text-sm text-gray-600">User ID</p>
                                <p className="font-medium">{user.id}</p>
                            </div>
                            <div className="p-3 bg-gray-50 rounded-lg">
                                <p className="text-sm text-gray-600">Status</p>
                                <p className="font-medium">
                                    {user.disabled ? 'Disabled' : 'Active'}
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Message Modal */}
            {showMessageModal && (
                <div className="fixed inset-0 bg-black/55 backdrop-blur-sm transition-opacity flex items-center justify-center z-50 p-4">
                    <div className="bg-white rounded-xl shadow-xl max-w-md w-full">
                        <div className="flex items-center justify-between p-6 border-b border-gray-200">
                            <h2 className="text-xl font-bold text-gray-900">Send Message</h2>
                            <button
                                onClick={handleCloseModal}
                                className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                            >
                                <X className="w-5 h-5 text-gray-600" />
                            </button>
                        </div>
                        
                        <form onSubmit={handleSendMessage} className="p-6">
                            <div className="space-y-4">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-2">
                                        To
                                    </label>
                                    <div className="p-3 bg-gray-50 rounded-lg text-sm text-gray-700">
                                        {user?.name} (@{user?.username})
                                    </div>
                                </div>

                                <div>
                                    <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-2">
                                        Title <span className="text-red-500">*</span>
                                    </label>
                                    <input
                                        type="text"
                                        id="title"
                                        value={messageForm.title}
                                        onChange={(e) => setMessageForm({ ...messageForm, title: e.target.value })}
                                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                        placeholder="Message title"
                                        required
                                    />
                                </div>

                                <div>
                                    <label htmlFor="contents" className="block text-sm font-medium text-gray-700 mb-2">
                                        Message <span className="text-red-500">*</span>
                                    </label>
                                    <textarea
                                        id="contents"
                                        value={messageForm.contents}
                                        onChange={(e) => setMessageForm({ ...messageForm, contents: e.target.value })}
                                        rows={4}
                                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                                        placeholder="Enter your message..."
                                        required
                                    />
                                </div>

                                {messageError && (
                                    <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
                                        <p className="text-sm text-red-600">{messageError}</p>
                                    </div>
                                )}

                                {messageSuccess && (
                                    <div className="p-3 bg-green-50 border border-green-200 rounded-lg">
                                        <p className="text-sm text-green-600">Message sent successfully!</p>
                                    </div>
                                )}
                            </div>

                            <div className="flex gap-3 mt-6">
                                <button
                                    type="button"
                                    onClick={handleCloseModal}
                                    className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
                                    disabled={sending}
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                    disabled={sending}
                                >
                                    {sending ? 'Sending...' : 'Send Message'}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
};