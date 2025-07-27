"use client";
import React, { useState } from 'react';
import { User, Mail, Calendar, MapPin, Link, Edit, Settings, Heart, Eye, Code, Image, Star, Users, Clock, Shield, Award, Github, Twitter, Globe } from 'lucide-react';


export default function Page() {
    return (<>
        <UserProfile />
    </>)
}


const UserProfile = () => {
    const [isEditing, setIsEditing] = useState(false);
    const [activeTab, setActiveTab] = useState('apps');

    // Mock user data
    const user = {
        id: 1,
        name: 'Alex Johnson',
        username: '@alexjohnson',
        email: 'alex.johnson@email.com',
        bio: 'Full-stack developer passionate about AI and machine learning. Building the future one app at a time. Love exploring new technologies and sharing knowledge with the community.',
        location: 'San Francisco, CA',
        website: 'https://alexjohnson.dev',
        joinDate: 'January 2024',
        avatar: 'AJ',
        gradient: 'from-blue-500 to-purple-600',
        stats: {
            appsCreated: 12,
            totalLikes: 2847,
            followers: 156,
            following: 89
        },
        socialLinks: {
            github: 'alexjohnson',
            twitter: 'alex_codes',
            website: 'alexjohnson.dev'
        }
    };

    const userApps = [
        {
            id: 1,
            title: 'AI Image Generator Pro',
            description: 'Advanced image generation with custom prompts',
            likes: 234,
            views: 5600,
            gradient: 'from-pink-500 to-purple-600',
            category: 'Image Generation',
            createdAt: '2 days ago'
        },
        {
            id: 2,
            title: 'Code Review Assistant',
            description: 'AI-powered code analysis and suggestions',
            likes: 189,
            views: 3200,
            gradient: 'from-green-500 to-blue-500',
            category: 'Code Generation',
            createdAt: '1 week ago'
        },
        {
            id: 3,
            title: 'Voice Translator',
            description: 'Real-time speech translation in 50+ languages',
            likes: 156,
            views: 4100,
            gradient: 'from-orange-500 to-red-500',
            category: 'Language Translation',
            createdAt: '2 weeks ago'
        }
    ];

    const AppCard = ({ app }: any) => (
        <div className={`relative overflow-hidden rounded-xl bg-gradient-to-br ${app.gradient} p-5 text-white hover:scale-105 transition-all duration-200 cursor-pointer group`}>
            <div className="flex flex-col h-full justify-between min-h-[160px]">
                <div>
                    <h3 className="text-lg font-bold mb-2">{app.title}</h3>
                    <p className="text-sm text-white/90 mb-3 line-clamp-2">{app.description}</p>
                    <span className="text-xs bg-white/20 px-2 py-1 rounded-full">{app.category}</span>
                </div>

                <div className="flex items-center justify-between mt-4 text-sm">
                    <div className="flex items-center gap-3">
                        <div className="flex items-center gap-1">
                            <Heart className="w-3 h-3" />
                            <span>{app.likes}</span>
                        </div>
                        <div className="flex items-center gap-1">
                            <Eye className="w-3 h-3" />
                            <span>{app.views}</span>
                        </div>
                    </div>
                    <span className="text-white/80">{app.createdAt}</span>
                </div>
            </div>
        </div>
    );

    return (
        <div className="min-h-screen bg-gray-50">
            {/* Header */}
            <header className="bg-white border-b border-gray-200 px-6 py-4">
                <div className="max-w-7xl mx-auto flex items-center justify-between">
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
                        <button className="border border-gray-300 px-4 py-2 rounded-lg font-medium hover:bg-gray-50 transition-colors">
                            Share Profile
                        </button>
                        <button
                            onClick={() => setIsEditing(!isEditing)}
                            className="bg-blue-600 text-white px-4 py-2 rounded-lg font-medium hover:bg-blue-700 transition-colors flex items-center gap-2"
                        >
                            <Edit className="w-4 h-4" />
                            {isEditing ? 'Save' : 'Edit Profile'}
                        </button>
                    </div>
                </div>
            </header>

            <div className="max-w-7xl mx-auto px-6 py-8">
                <div className="flex md:flex-col justify-between">
                    <div className="bg-white rounded-xl border border-gray-200 p-6 mb-6">
                        <div className="text-center mb-6">
                            <div className={`w-24 h-24 bg-gradient-to-br ${user.gradient} rounded-full flex items-center justify-center text-white text-2xl font-bold mx-auto mb-4`}>
                                {user.avatar}
                            </div>
                            <h2 className="text-2xl font-bold text-gray-900 mb-1">{user.name}</h2>
                            <p className="text-blue-600 font-medium mb-2">{user.username}</p>

                            {/* Stats */}
                            <div className="grid grid-cols-2 gap-4 mt-6">
                                <div className="text-center">
                                    <div className="text-2xl font-bold text-gray-900">{user.stats.appsCreated}</div>
                                    <div className="text-sm text-gray-500">Apps</div>
                                </div>
                                <div className="text-center">
                                    <div className="text-2xl font-bold text-gray-900">{user.stats.totalLikes}</div>
                                    <div className="text-sm text-gray-500">Likes</div>
                                </div>
                                <div className="text-center">
                                    <div className="text-2xl font-bold text-gray-900">{user.stats.followers}</div>
                                    <div className="text-sm text-gray-500">Followers</div>
                                </div>
                                <div className="text-center">
                                    <div className="text-2xl font-bold text-gray-900">{user.stats.following}</div>
                                    <div className="text-sm text-gray-500">Following</div>
                                </div>
                            </div>
                        </div>

                        {/* Bio */}
                        <div className="mb-6">
                            <h3 className="font-semibold text-gray-900 mb-2">About</h3>
                            {isEditing ? (
                                <textarea
                                    className="w-full p-3 border border-gray-300 rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    rows={4}
                                    defaultValue={user.bio}
                                />
                            ) : (
                                <p className="text-gray-600 leading-relaxed">{user.bio}</p>
                            )}
                        </div>

                        {/* Contact Info */}
                        <div className="space-y-3">
                            <div className="flex items-center gap-3 text-gray-600">
                                <Mail className="w-4 h-4" />
                                <span className="text-sm">{user.email}</span>
                            </div>
                            <div className="flex items-center gap-3 text-gray-600">
                                <MapPin className="w-4 h-4" />
                                <span className="text-sm">{user.location}</span>
                            </div>
                            <div className="flex items-center gap-3 text-gray-600">
                                <Link className="w-4 h-4" />
                                <a href={user.website} className="text-sm text-blue-600 hover:underline">{user.website}</a>
                            </div>
                            <div className="flex items-center gap-3 text-gray-600">
                                <Calendar className="w-4 h-4" />
                                <span className="text-sm">Joined {user.joinDate}</span>
                            </div>
                        </div>

                        {/* Social Links */}
                        <div className="mt-6 pt-6 border-t border-gray-200">
                            <h4 className="font-medium text-gray-900 mb-3">Connect</h4>
                            <div className="flex items-center gap-3">
                                <a href="#" className="p-2 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors">
                                    <Github className="w-4 h-4 text-gray-600" />
                                </a>
                                <a href="#" className="p-2 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors">
                                    <Twitter className="w-4 h-4 text-gray-600" />
                                </a>
                                <a href="#" className="p-2 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors">
                                    <Globe className="w-4 h-4 text-gray-600" />
                                </a>
                            </div>
                        </div>
                    </div>




                    <div className="bg-white rounded-xl border border-gray-200 mb-6">
                        <div className="border-b border-gray-200">
                            <div className="flex">
                                <button
                                    onClick={() => setActiveTab('apps')}
                                    className={`px-6 py-4 font-medium text-sm border-b-2 transition-colors ${activeTab === 'apps'
                                            ? 'border-blue-500 text-blue-600'
                                            : 'border-transparent text-gray-500 hover:text-gray-700'
                                        }`}
                                >
                                    My Apps ({userApps.length})
                                </button>
                                <button
                                    onClick={() => setActiveTab('activity')}
                                    className={`px-6 py-4 font-medium text-sm border-b-2 transition-colors ${activeTab === 'activity'
                                            ? 'border-blue-500 text-blue-600'
                                            : 'border-transparent text-gray-500 hover:text-gray-700'
                                        }`}
                                >
                                    Activity
                                </button>
                                <button
                                    onClick={() => setActiveTab('favorites')}
                                    className={`px-6 py-4 font-medium text-sm border-b-2 transition-colors ${activeTab === 'favorites'
                                            ? 'border-blue-500 text-blue-600'
                                            : 'border-transparent text-gray-500 hover:text-gray-700'
                                        }`}
                                >
                                    Favorites
                                </button>
                            </div>
                        </div>

                        <div className="p-6">
                            {activeTab === 'apps' && (
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    {userApps.map((app) => (
                                        <AppCard key={app.id} app={app} />
                                    ))}
                                </div>
                            )}

                            {activeTab === 'activity' && (
                                <div className="space-y-4">
                                    <div className="flex items-center gap-3 p-4 bg-gray-50 rounded-lg">
                                        <div className="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center">
                                            <Star className="w-5 h-5 text-green-600" />
                                        </div>
                                        <div className="flex-1">
                                            <p className="text-sm text-gray-900">Created <strong>AI Image Generator Pro</strong></p>
                                            <p className="text-xs text-gray-500">2 days ago</p>
                                        </div>
                                    </div>
                                    <div className="flex items-center gap-3 p-4 bg-gray-50 rounded-lg">
                                        <div className="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
                                            <Heart className="w-5 h-5 text-blue-600" />
                                        </div>
                                        <div className="flex-1">
                                            <p className="text-sm text-gray-900">Liked <strong>Voice Synthesis App</strong></p>
                                            <p className="text-xs text-gray-500">3 days ago</p>
                                        </div>
                                    </div>
                                </div>
                            )}

                            {activeTab === 'favorites' && (
                                <div className="text-center py-12">
                                    <Heart className="w-12 h-12 text-gray-300 mx-auto mb-4" />
                                    <p className="text-gray-500">No favorite apps yet</p>
                                </div>
                            )}
                        </div>
                    </div>

                </div>
            </div>
        </div>
    );
};

