"use client";
import React, { useState } from 'react';
import { Search, Filter, ArrowUpDown, Heart, Users, Zap, Image, Box, Octagon, SquareUserRound, BadgeDollarSign, BookOpenText, BookHeart, BriefcaseBusiness, Drama } from 'lucide-react';



export default function Page() {
    return (<>
        <SpacesDirectory />
    </>)
}





const SpacesDirectory = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedFilter, setSelectedFilter] = useState('Relevance');

    const categories = [
        { name: 'Personal', icon: BookHeart },
        { name: 'AI', icon: Octagon },
        { name: 'Productivity', icon: BriefcaseBusiness },
        { name: 'Entertainment', icon: Drama },
        { name: 'Finance', icon: BadgeDollarSign },
        { name: 'Education', icon: BookOpenText },
        { name: 'Social', icon: SquareUserRound },
    ];

    const spaces = [
        {
            id: 1,
            title: 'Addit ⚡',
            description: 'Add objects to images using text prompts',
            author: 'nvidia',
            timeAgo: '2 days ago',
            from: 'ZERO',
            mcp: true,
            gradient: 'from-pink-500 to-orange-500'
        },
        {
            id: 2,
            title: 'PartCrafter 🧩',
            description: '3D Mesh Generation via Compositional Latent Diffusion',
            author: 'alexnasa',
            timeAgo: 'about 24 hours ago',
            from: 'ZERO',
            mcp: true,
            gradient: 'from-blue-500 to-purple-600'
        },
        {
            id: 3,
            title: 'Audio Flamingo 3 Chat 🔥',
            description: 'Audio Flamingo 3 demo for multi-turn multi-audio chat',
            author: 'nvidia',
            timeAgo: '12 days ago',
            from: 'A100',
            gradient: 'from-gray-600 to-blue-800'
        },
        {
            id: 4,
            title: 'Voxtral 🧠',
            description: 'Demo space for Mistral latest speech models',
            author: 'MohamedRashad',
            timeAgo: '5 days ago',
            from: 'ZERO',
            mcp: true,
            gradient: 'from-red-500 to-pink-600'
        },
        {
            id: 5,
            title: 'Calligrapher: Freestyle Text Image Customization',
            description: 'Customize text in images using a reference style',
            author: 'Calligrapher2025',
            timeAgo: '11 days ago',
            from: 'ZERO',
            gradient: 'from-purple-600 to-indigo-600'
        },
        {
            id: 6,
            title: 'ZenCtrl Inpaint 🎭',
            description: 'Create scenes with your subject in it with ZenCtrl Inpaint',
            author: 'fotographerai',
            timeAgo: '5 days ago',
            from: 'Running',
            gradient: 'from-purple-500 to-pink-500'
        },
        {
            id: 7,
            title: 'AudioRag Demo 🎵',
            description: 'Search audio files for specific queries',
            author: 'fdaudens',
            timeAgo: '8 days ago',
            from: 'ZERO',
            gradient: 'from-teal-500 to-blue-600'
        },
        {
            id: 8,
            title: 'Owen TTS Demo 📢',
            description: 'Generate speech from text with different voices',
            author: 'Owen',
            timeAgo: '12 days ago',
            from: 'Running',
            gradient: 'from-green-500 to-blue-600'
        }
    ];

    const sortOptions = [
        'Relevance',
        'Recently Created',
        'Recently Updated',
        'Installed Date',
        'By Usage'
    ];

    const [isDropdownOpen, setIsDropdownOpen] = useState(false);



    return (
        <div className="min-h-screen bg-gray-50">
            {/* Header */}
            <header className="bg-white border-b border-gray-200 px-6 py-4">
                <div className="max-w-7xl mx-auto flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <div className="flex items-center gap-2">
                            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                                <Box className="w-5 h-5 text-white" />
                            </div>
                            <div>
                                <h1 className="text-xl font-bold">Spaces</h1>
                                <p className="text-sm text-gray-600">Your App Directory</p>
                            </div>
                        </div>
                    </div>

                    <div className="flex items-center gap-4">
                        <button className="bg-blue-600 text-white px-4 py-2 rounded-lg font-medium hover:bg-blue-700 transition-colors">
                            + New Space
                        </button>
                    </div>
                </div>
            </header>

            {/* Search Bar */}
            <div className="bg-white border-b border-gray-200 px-6 py-4">
                <div className="max-w-7xl mx-auto">
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                        <input
                            type="text"
                            placeholder="Search spaces..."
                            className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                        />
                        <button className="absolute right-3 top-1/2 transform -translate-y-1/2 p-1 cursor-pointer hover:bg-gray-100 rounded-full transition-colors">
                            <Zap className="w-5 h-5 text-gray-400" />
                        </button>
                    </div>
                </div>
            </div>

            <div className="bg-white border-b border-gray-200 px-6 py-4">
                <div className="max-w-7xl mx-auto">
                    <div className="flex items-center gap-4 scrollbar-hide flex-wrap overflow-x-hidden">
                        {categories.map((category, index) => {
                            const IconComponent = category.icon as React.ElementType;
                            return (
                                <button
                                    key={index}
                                    className="flex items-center gap-2 px-4 py-2 text-sm text-gray-600 hover:text-gray-800 hover:bg-gray-200 rounded-lg transition-colors whitespace-nowrap"
                                >
                                    <IconComponent className="w-4 h-4" />
                                    <span>{category.name}</span>
                                </button>
                            );
                        })}
                    </div>
                </div>
            </div>

            <div className="max-w-7xl mx-auto px-6 py-8">
                <div className="mb-8">
                    <div className="flex items-center justify-between mb-6">
                        <div className="flex items-center gap-3">
                            <div className="w-3 h-3 bg-orange-500 rounded-full"></div>
                            <h2 className="text-xl font-bold">Favorites</h2>
                        </div>

                        <div className="flex items-center gap-4">

                            <button className="flex items-center gap-2 px-3 py-2 border border-gray-300 rounded-lg text-sm hover:bg-gray-50 transition-colors">
                                <Filter className="w-4 h-4" />
                                <span>Filters (0)</span>
                            </button>

                            <div className="relative">
                                <button
                                    onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                                    className="flex items-center gap-2 px-3 py-2 border border-gray-300 rounded-lg text-sm hover:bg-gray-50 transition-colors"
                                >
                                    <ArrowUpDown className="w-4 h-4" />
                                    <span>Sort: {selectedFilter}</span>
                                </button>

                                {isDropdownOpen && (
                                    <div className="absolute right-0 top-full mt-1 w-48 bg-white border border-gray-300 rounded-lg shadow-lg z-10">
                                        {sortOptions.map((option) => (
                                            <button
                                                key={option}
                                                onClick={() => {
                                                    setSelectedFilter(option);
                                                    setIsDropdownOpen(false);
                                                }}
                                                className={`w-full text-left px-3 py-2 text-sm hover:bg-gray-50 first:rounded-t-lg last:rounded-b-lg ${selectedFilter === option ? 'bg-blue-50 text-blue-600' : 'text-gray-700'
                                                    }`}
                                            >
                                                {option}
                                            </button>
                                        ))}
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                        {spaces.map((space) => (
                            <SpaceCard key={space.id} space={space} />
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
};

const SpaceCard = ({ space }: { space: any }) => {


    return (

        <div className={`relative overflow-hidden rounded-xl bg-gradient-to-br ${space.gradient} p-6 text-white min-h-[200px] group hover:scale-105 transition-transform duration-200 cursor-pointer`}>
            <div className="flex flex-col h-full justify-between">
                <div>

                    <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2 text-sm">
                            <span className="bg-white/20 backdrop-blur-sm px-2 py-1 rounded text-sm">
                                {space.from}
                            </span>
                            {space.mcp && (
                                <span className="bg-pink-500/80 px-2 py-1 rounded text-xs">🔥 MCP</span>
                            )}
                        </div>

                        <div className='flex justify-end'>
                            <div className="flex items-center gap-1 text-xs">
                                <Heart className="w-4 h-4" />
                            </div>
                        </div>

                    </div>

                    <h3 className="text-xl font-bold mb-2">{space.title}</h3>
                    <p className="text-sm text-white/90 mb-4 line-clamp-2">{space.description}</p>
                </div>

                <div className="flex items-center justify-between text-sm">
                    <div className="flex items-center gap-2">
                        <div className="w-6 h-6 bg-white/20 rounded-full flex items-center justify-center">
                            <Users className="w-3 h-3" />
                        </div>

                        <div className="flex flex-col">
                            <span className="font-medium">{space.author}</span>
                            <div className="flex items-center gap-1 text-xs">
                                <span>{space.timeAgo}</span>
                            </div>
                        </div>


                    </div>
                </div>
            </div>
        </div>
    )
};