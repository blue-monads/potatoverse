

"use client";
import React, { useState } from 'react';
import { Heart, Users, SquareUserRound, Cog, Link2Icon } from 'lucide-react';
import { Clock, ArrowRight, Code } from 'lucide-react';
import EmptyFavorite from './sub/EmptyFavorite';
import HeroSection from './sub/HeroSection';


export default function HomePage() {
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

            <HeroSection
                searchTerm={searchTerm}
                setSearchTerm={setSearchTerm}
            />



            {/* Favorites Section */}
            {favs.length === 0 ? (<>
                <EmptyFavorite />
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