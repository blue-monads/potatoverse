"use client";
import React, { useEffect, useRef, useState } from 'react';
import { Search, Filter, ArrowUpDown, Heart, Users, Zap, Image, Box, Octagon, SquareUserRound, BadgeDollarSign, BookOpenText, BookHeart, BriefcaseBusiness, Drama, Bolt, CloudLightning, ScrollText, Files, Grid2x2Plus, Cog, Link2Icon } from 'lucide-react';
import { createPortal } from 'react-dom';

import { Clock, TrendingUp, Star, ArrowRight, Sparkles, Code, Mic, Globe, Eye, Type } from 'lucide-react';


export default function Page() {
    const [searchTerm, setSearchTerm] = useState('');

    const favs = [
        {
            id: 1,
            title: 'Addit ⚡',
            description: 'Add objects to images using text prompts',
            author: 'nvidia',
            lastUsed: '2 hours ago',
            gradient: 'from-pink-500 to-orange-500',
            category: 'Image Generation'
        },
        {
            id: 2,
            title: 'ChatGPT Clone 💬',
            description: 'A powerful conversational AI interface',
            author: 'openai-community',
            lastUsed: '1 day ago',
            gradient: 'from-green-500 to-teal-500',
            category: 'Text Generation'
        },

    ];

   // favs.length = 0; // For testing empty state


    return (<>
        <div className="min-h-screen bg-gray-50 w-full">
            {/* Hero Section */}
            <div className="bg-gradient-to-br from-blue-600 via-purple-600 to-pink-600 text-white w-full flex items-center justify-center">
                <div className="max-w-7xl mx-auto px-6 py-16">
                    <div className="text-center max-w-4xl mx-auto">
                        <div className="flex items-center justify-center gap-2 mb-4">
                            <Sparkles className="w-6 h-6" />
                            <span className="text-lg font-medium">Welcome to admin portal!</span>
                        </div>
                        <h1 className="text-5xl font-bold mb-6">
                            Discover Apps and Tools
                        </h1>
                        <p className="text-xl text-white/90 mb-8 leading-relaxed">
                            Find apps to fit your needs or create your own quickly.
                        </p>

                        {/* Search Bar */}
                        <div className="relative max-w-2xl mx-auto">
                            <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                            <input
                                type="text"
                                placeholder="Search recent apps"
                                className="w-full pl-12 pr-16 py-4 text-gray-900 bg-white rounded-xl focus:outline-none focus:ring-4 focus:ring-white/30 shadow-lg text-lg"
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                            />
                            <button className="absolute right-2 top-1/2 transform -translate-y-1/2 bg-blue-600 text-white p-2 rounded-lg hover:bg-blue-700 transition-colors">
                                <Zap className="w-5 h-5" />
                            </button>
                        </div>

                        <div className="flex items-center justify-center gap-4 mt-6 text-sm text-white/80">
                            <span>Popular searches:</span>
                            <button className="bg-white/20 px-3 py-1 rounded-full hover:bg-white/30 transition-colors">
                                Game
                            </button>
                            <button className="bg-white/20 px-3 py-1 rounded-full hover:bg-white/30 transition-colors">
                                Calendar
                            </button>

                        </div>
                    </div>
                </div>
            </div>


            {/* Favorites Section */}
            {favs.length === 0 ? (<>
                <FavoritesEmpty />
            </>) : (<>
                <div className="max-w-7xl mx-auto px-6 py-12">
                    <div className="flex items-center justify-between mb-6">
                        <div className="flex items-center gap-3">
                            <Heart className="w-6 h-6 text-pink-600" />
                            <h2 className="text-2xl font-bold text-gray-900">Favorites</h2>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 w-full">
                        <div className="lg:col-span-2">

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 w-full">

                                {favs.map((app) => (
                                    <FavCard key={app.id} app={app} />
                                ))}

                            </div>


                        </div>


                    </div>
                </div>


            </>)}


            {/* Quick Links (users, profile, setting, dev console) */}
            <div className="max-w-7xl mx-auto px-6 py-12">

                <div className="flex items-center justify-between mb-6">
                    <div className="flex items-center gap-3">
                        <Link2Icon className="w-6 h-6 text-pink-600" />
                        <h2 className="text-2xl font-bold text-gray-900">Quick Links</h2>
                    </div>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">


                    <div className="bg-white rounded-lg shadow-lg p-6 flex items-center gap-4 hover:shadow-xl transition-shadow border border-gray-200">
                        <Users className="w-8 h-8 text-blue-600" />
                        <div>
                            <h3 className="text-lg font-semibold">Users</h3>
                            <p className="text-sm text-gray-600">Manage users and permissions</p>
                        </div>
                    </div>

                    <div className="bg-white rounded-lg shadow-lg p-6 flex items-center gap-4 hover:shadow-xl transition-shadow border border-gray-200">
                        <SquareUserRound className="w-8 h-8 text-green-600" />
                        <div>
                            <h3 className="text-lg font-semibold">Profile</h3>
                            <p className="text-sm text-gray-600">View and edit your profile</p>
                        </div>
                    </div>

                    <div className="bg-white rounded-lg shadow-lg p-6 flex items-center gap-4 hover:shadow-xl transition-shadow border border-gray-200">
                        <Cog className="w-8 h-8 text-yellow-600" />
                        <div>
                            <h3 className="text-lg font-semibold">Settings</h3>
                            <p className="text-sm text-gray-600">Configure application settings</p>
                        </div>
                    </div>

                    <div className="bg-white rounded-lg shadow-lg p-6 flex items-center gap-4 hover:shadow-xl transition-shadow border border-gray-200">
                        <Code className="w-8 h-8 text-purple-600" />
                        <div>
                            <h3 className="text-lg font-semibold">Dev Console</h3>
                            <p className="text-sm text-gray-600">Access developer tools</p>
                        </div>
                    </div>

                </div>
            </div>







        </div>
    </>)
}


const FavoritesEmpty = () => {
    return (
        <div className="flex flex-col items-center justify-center max-w-7xl mx-auto px-6 py-16">
            <div className="mb-8">
                <svg
                    width="200"
                    height="160"
                    viewBox="0 0 200 160"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                    className="drop-shadow-sm"
                >

                    <circle cx="50" cy="40" r="3" fill="#e5e7eb" opacity="0.5" />
                    <circle cx="150" cy="30" r="2" fill="#e5e7eb" opacity="0.3" />
                    <circle cx="170" cy="60" r="2.5" fill="#e5e7eb" opacity="0.4" />

                    <path
                        d="M100 130c-2-1.5-45-35-45-65 0-20 15-35 35-35 10 0 18 5 23 12 5-7 13-12 23-12 20 0 35 15 35 35 0 30-43 63.5-45 65z"
                        fill="url(#heartGradient)"
                        className="animate-pulse"
                    />

                    <path
                        d="M100 130c-2-1.5-45-35-45-65 0-20 15-35 35-35 10 0 18 5 23 12 5-7 13-12 23-12 20 0 35 15 35 35 0 30-43 63.5-45 65z"
                        stroke="#f3f4f6"
                        strokeWidth="3"
                        fill="none"
                    />

                    <g className="animate-bounce" style={{ animationDelay: '0.5s' }}>
                        <path d="M65 25l2 6 6 2-6 2-2 6-2-6-6-2 6-2 2-6z" fill="#fbbf24" />
                    </g>
                    <g className="animate-bounce" style={{ animationDelay: '1s' }}>
                        <path d="M140 20l1.5 4.5 4.5 1.5-4.5 1.5-1.5 4.5-1.5-4.5-4.5-1.5 4.5-1.5 1.5-4.5z" fill="#f59e0b" />
                    </g>
                    <g className="animate-bounce" style={{ animationDelay: '1.5s' }}>
                        <path d="M30 80l1 3 3 1-3 1-1 3-1-3-3-1 3-1 1-3z" fill="#fbbf24" />
                    </g>

                    <defs>
                        <linearGradient id="heartGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                            <stop offset="0%" stopColor="#fce7f3" />
                            <stop offset="50%" stopColor="#f9a8d4" />
                            <stop offset="100%" stopColor="#ec4899" />
                        </linearGradient>
                    </defs>
                </svg>
            </div>

            {/* Content */}
            <div className="text-center max-w-md">


                <h3 className="text-xl font-semibold text-gray-700 mb-3">
                    No favorites yet!
                </h3>


                <div className="flex flex-col sm:flex-row gap-3 justify-center">
                    <button className="bg-gradient-to-r from-pink-500 to-purple-600 text-white px-6 py-3 rounded-lg font-medium hover:from-pink-600 hover:to-purple-700 transition-all transform hover:scale-105 shadow-lg">
                        Explore Apps
                    </button>
                    <button className="border border-gray-300 text-gray-700 px-6 py-3 rounded-lg font-medium hover:bg-gray-50 transition-colors">
                        Store
                    </button>
                </div>
            </div>

            {/* Bottom decoration */}
            <div className="mt-12 flex items-center gap-2 text-sm text-gray-400">
                <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M13 3c-4.97 0-9 4.03-9 9H1l3.89 3.89.07.14L9 12H6c0-3.87 3.13-7 7-7s7 3.13 7 7-3.13 7-7 7c-1.93 0-3.68-.79-4.94-2.06l-1.42 1.42C8.27 19.99 10.51 21 13 21c4.97 0 9-4.03 9-9s-4.03-9-9-9zm-1 5v5l4.28 2.54.72-1.21-3.5-2.08V8H12z" />
                </svg>
                <span>Start adding favorites to see them here</span>
            </div>
        </div>
    );
};

const FavCard = ({ app }: any) => (
    <div className={`relative overflow-hidden rounded-xl bg-gradient-to-br ${app.gradient} p-5 text-white hover:scale-105 transition-all duration-200 cursor-pointer group`}>
        <div className="flex flex-col h-full justify-between">
            <div>
                <h3 className="text-lg font-bold mb-2">{app.title}</h3>
                <p className="text-sm text-white/90 mb-3 line-clamp-2">{app.description}</p>
                <span className="text-xs bg-white/20 px-2 py-1 rounded-full">{app.category}</span>
            </div>

            <div className="flex items-center justify-between mt-4 text-sm">
                <div className="flex items-center gap-2">
                    <div className="w-5 h-5 bg-white/20 rounded-full flex items-center justify-center">
                        <Users className="w-3 h-3" />
                    </div>
                    <span className="font-medium">{app.author}</span>
                </div>
                <div className="flex items-center gap-1 text-white/80">
                    <Clock className="w-3 h-3" />
                    <span>{app.lastUsed}</span>
                </div>
            </div>
        </div>
        <div className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity">
            <ArrowRight className="w-5 h-5" />
        </div>
    </div>
);