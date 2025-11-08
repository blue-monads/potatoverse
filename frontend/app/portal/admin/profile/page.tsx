"use client";
import React, { useState, useEffect } from 'react';
import { User, Mail, Edit, Save, X } from 'lucide-react';
import { getSelfInfo, updateSelfBio, User as UserType } from '../../../../lib/api';
import { useGApp } from '../../../../hooks/contexts/GAppStateContext';

export default function Page() {
    return (<>
        <UserProfile />
    </>)
}

const UserProfile = () => {
    const { loaded, isInitialized, isAuthenticated } = useGApp();
    const [user, setUser] = useState<UserType | null>(null);
    const [isEditing, setIsEditing] = useState(false);
    const [bio, setBio] = useState('');
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        // Only load user info when the app state is fully loaded and initialized
        if (loaded && isInitialized && isAuthenticated) {
            loadUserInfo();
        }
    }, [loaded, isInitialized, isAuthenticated]);

    const loadUserInfo = async () => {
        try {
            setLoading(true);
            const response = await getSelfInfo();
            setUser(response.data);
            setBio(response.data.bio || '');
        } catch (error) {
            console.error('Failed to load user info:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleSaveBio = async () => {
        try {
            setSaving(true);
            await updateSelfBio(bio);
            if (user) {
                setUser({ ...user, bio });
            }
            setIsEditing(false);
        } catch (error) {
            console.error('Failed to update bio:', error);
        } finally {
            setSaving(false);
        }
    };

    const handleCancelEdit = () => {
        setBio(user?.bio || '');
        setIsEditing(false);
    };

    // Show loading if app state is not ready or if we're loading user data
    if (!loaded || !isInitialized || loading) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
                    <p className="text-gray-600">
                        {!loaded || !isInitialized ? 'Initializing...' : 'Loading profile...'}
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
                    <p className="text-gray-600">Please log in to view your profile</p>
                </div>
            </div>
        );
    }

    if (!user) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-center">
                    <p className="text-gray-600">Failed to load profile</p>
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
                        <div className="flex items-center gap-2">
                            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                                <User className="w-5 h-5 text-white" />
                            </div>
                            <div>
                                <h1 className="text-xl font-bold">Profile</h1>
                                <p className="text-sm text-gray-600">User Information</p>
                            </div>
                        </div>
                    </div>

                    <div className="flex items-center gap-3">
                        {isEditing ? (
                            <>
                                <button
                                    onClick={handleCancelEdit}
                                    className="border border-gray-300 px-4 py-2 rounded-lg font-medium hover:bg-gray-50 transition-colors flex items-center gap-2"
                                >
                                    <X className="w-4 h-4" />
                                    Cancel
                                </button>
                                <button
                                    onClick={handleSaveBio}
                                    disabled={saving}
                                    className="bg-green-600 text-white px-4 py-2 rounded-lg font-medium hover:bg-green-700 transition-colors flex items-center gap-2 disabled:opacity-50"
                                >
                                    <Save className="w-4 h-4" />
                                    {saving ? 'Saving...' : 'Save'}
                                </button>
                            </>
                        ) : (
                            <button
                                onClick={() => setIsEditing(true)}
                                className="bg-blue-600 text-white px-4 py-2 rounded-lg font-medium hover:bg-blue-700 transition-colors flex items-center gap-2"
                            >
                                <Edit className="w-4 h-4" />
                                Edit Profile
                            </button>
                        )}
                    </div>
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
                        {isEditing ? (
                            <textarea
                                value={bio}
                                onChange={(e) => setBio(e.target.value)}
                                className="w-full p-4 border border-gray-300 rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
                                rows={4}
                                placeholder="Tell us about yourself..."
                            />
                        ) : (
                            <div className="p-4 bg-gray-50 rounded-lg">
                                {user.bio ? (
                                    <p className="text-gray-700 leading-relaxed whitespace-pre-wrap">{user.bio}</p>
                                ) : (
                                    <p className="text-gray-500 italic">No bio added yet. Click "Edit Profile" to add one.</p>
                                )}
                            </div>
                        )}
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
                    </div>
                </div>
            </div>
        </div>
    );
};

