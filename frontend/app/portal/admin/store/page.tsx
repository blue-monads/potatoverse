"use client";
import React, { useEffect, useRef, useState } from 'react';
import { Search, Filter, ArrowUpDown, Heart, Users, Zap, Image, Box, Octagon, SquareUserRound, BadgeDollarSign, BookOpenText, BookHeart, BriefcaseBusiness, Drama, Store, CloudDownload, InfoIcon, Bolt } from 'lucide-react';
import { createPortal } from 'react-dom';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import AddButton from '@/contain/AddButton';
import { GAppStateHandle, ModalHandle, useGApp } from '@/hooks';
import { Tabs } from '@skeletonlabs/skeleton-react';
import { installPackage, installPackageZip, listEPackages } from '@/lib';



export default function Page() {
    return (<>
        <StoreDirectory />
    </>)
}



const StoreDirectory = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedFilter, setSelectedFilter] = useState('Relevance');
    const gapp = useGApp();

    const categories = [
        { name: 'Personal', icon: BookHeart },
        { name: 'AI', icon: Octagon },
        { name: 'Productivity', icon: BriefcaseBusiness },
        { name: 'Entertainment', icon: Drama },
        { name: 'Finance', icon: BadgeDollarSign },
        { name: 'Education', icon: BookOpenText },
        { name: 'Social', icon: SquareUserRound },
    ];

    useEffect(() => {
        const fetchPackages = async () => {
            const resp = await listEPackages();
            console.log(resp);
        }
        fetchPackages();
    }, []);

    const storeItems = [
        {
            id: 1,
            title: 'Addit ⚡',
            description: 'Add objects to images using text prompts',
            author: 'nvidia',
            timeAgo: '2 days ago',
            mcp: true,
            gradient: 'from-pink-500 to-orange-500'
        },
        {
            id: 2,
            title: 'PartCrafter 🧩',
            description: '3D Mesh Generation via Compositional Latent Diffusion',
            author: 'alexnasa',
            timeAgo: 'about 24 hours ago',
            mcp: true,
            gradient: 'from-blue-500 to-purple-600'
        },
        {
            id: 3,
            title: 'Audio Flamingo 3 Chat 🔥',
            description: 'Audio Flamingo 3 demo for multi-turn multi-audio chat',
            author: 'nvidia',
            timeAgo: '12 days ago',
            gradient: 'from-gray-600 to-blue-800'
        },
        {
            id: 4,
            title: 'Voxtral 🧠',
            description: 'Demo space for Mistral latest speech models',
            author: 'MohamedRashad',
            timeAgo: '5 days ago',
            mcp: true,
            gradient: 'from-red-500 to-pink-600'
        },
        {
            id: 5,
            title: 'Calligrapher: Freestyle Text Image Customization',
            description: 'Customize text in images using a reference style',
            author: 'Calligrapher2025',
            timeAgo: '11 days ago',
            gradient: 'from-purple-600 to-indigo-600'
        },
        {
            id: 6,
            title: 'ZenCtrl Inpaint 🎭',
            description: 'Create scenes with your subject in it with ZenCtrl Inpaint',
            author: 'fotographerai',
            timeAgo: '5 days ago',
            gradient: 'from-purple-500 to-pink-500'
        },
        {
            id: 7,
            title: 'AudioRag Demo 🎵',
            description: 'Search audio files for specific queries',
            author: 'fdaudens',
            timeAgo: '8 days ago',
            gradient: 'from-teal-500 to-blue-600'
        },
        {
            id: 8,
            title: 'Owen TTS Demo 📢',
            description: 'Generate speech from text with different voices',
            author: 'Owen',
            timeAgo: '12 days ago',
            gradient: 'from-green-500 to-blue-600'
        }
    ];

    const sortOptions = [
        'Relevance',
        'Recently Updated',
        'By Usage'
    ];

    const [isDropdownOpen, setIsDropdownOpen] = useState(false);



    return (

        <WithAdminBodyLayout
            Icon={Store}
            name='Store'
            description="Your App spaces."
            rightContent={<>
                <AddButton
                    name="+ Import"
                    onClick={() => { 

                        gapp.modal.openModal({
                            title: "Import Package",
                            content: (
                                <ImportSpaceModal gapp={gapp} />
                            ),
                            size: "lg"
                        });


                    }}
                />
            </>}


        >
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
                            <h2 className="text-xl font-bold">{selectedFilter}</h2>
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

                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {storeItems.map((item) => (
                            <StoreItemCard key={item.id} item={item} />
                        ))}
                    </div>
                </div>
            </div>
        </WithAdminBodyLayout>

    );
};

const StoreItemCard = ({ item }: { item: any }) => {


    return (

        <div className={`relative overflow-hidden rounded-xl bg-gradient-to-br ${item.gradient} p-6 text-white min-h-[200px] group hover:scale-105 transition-transform duration-200 cursor-pointer`}>
            <div className="flex flex-col h-full justify-between">
                <div>

                    <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2 text-sm">

                            {item.mcp && (
                                <span className="bg-pink-500/80 px-2 py-1 rounded text-xs">🔥 MCP</span>
                            )}
                        </div>



                    </div>

                    <h3 className="text-xl font-bold mb-2">{item.title}</h3>
                    <p className="text-sm text-white/90 mb-4 line-clamp-2">{item.description}</p>
                </div>

                <div className="flex items-center justify-between text-sm">
                    <div className="flex items-center gap-2">
                        <div className="w-6 h-6 bg-white/20 rounded-full flex items-center justify-center">
                            <Users className="w-3 h-3" />
                        </div>

                        <div className="flex flex-col">
                            <span className="font-medium">{item.author}</span>
                            <div className="flex items-center gap-1 text-xs">
                                <span>{item.timeAgo}</span>
                            </div>
                        </div>
                    </div>

                    <div className="flex gap-2">

                        <button className="flex items-center gap-1 text-xs bg-white/20 backdrop-blur-sm px-3 py-2 rounded-lg hover:bg-white/40 transition-colors cursor-pointer hover:text-blue-600">
                            <CloudDownload className="w-4 h-4" />
                            <span>Install</span>
                        </button>

                    </div>



                </div>
            </div>
        </div>
    )
};


interface ImportSpaceModalProps {
    gapp: GAppStateHandle;
}

const tabs = [
    { label: "URL", value: "url" },
    { label: "Zip", value: "zip" }
]

const ImportSpaceModal = (props: ImportSpaceModalProps) => {
    const [activeTab, setActiveTab] = useState('url');
    const [mode, setMode] = useState<'enter_input' | 'importing' | 'success' | 'error'>('enter_input');
    const inputRef = useRef<HTMLInputElement>(null);
    const gapp = props.gapp;



    return (<>

        {mode === 'enter_input' && (<>

            <div className="space-y-1">
                <p className="text-gray-600 dark:text-gray-300">
                    Directly import packages from a URL or upload a zip file.
                </p>

                <div className='flex gap-2 my-2 min-h-[100px]'>
                    <Tabs value={activeTab}
                        onValueChange={(e) => {
                            const currentTab = tabs.find((tab) => tab.value === e.value);
                            if (currentTab) {
                                setActiveTab(currentTab.value);
                            }

                        }}>
                        <Tabs.List>
                            {tabs.map((tab) => (
                                <Tabs.Control key={tab.value} value={tab.value}>{tab.label}</Tabs.Control>
                            ))}
                        </Tabs.List>
                        <Tabs.Content>
                            {activeTab === 'url' && <div>
                                <input ref={inputRef} type="text" placeholder="Enter URL" className="w-full p-2 border border-gray-300 rounded-lg" />
                            </div>}
                            {activeTab === 'zip' && <div>
                                <input
                                    ref={inputRef}
                                    type="file"
                                    accept=".zip"
                                    placeholder="Upload ZIP file"
                                    className="w-full p-2 border border-gray-300 rounded-lg"
                                />
                            </div>}
                        </Tabs.Content>
                    </Tabs>


                </div>


                <div className="flex gap-2 justify-end">
                    <button
                        onClick={() => gapp.modal.closeModal()}
                        className="bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded-lg transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={async () => {
                            setMode('importing');
                            if (activeTab === 'url') {
                                const url = inputRef.current?.value;
                                if (!url) {
                                    setMode('error');
                                    return;
                                }
                                const response = await installPackage(url);
                                if (response.status !== 200) {
                                    setMode('error');
                                    return;
                                }

                                setMode('success');

                            } else {
                                const file = inputRef.current?.files?.[0];
                                if (!file) {
                                    setMode('error');
                                    return;
                                }
                                const zip = await file.arrayBuffer();
                                if (!zip) {
                                    setMode('error');
                                    return;
                                }

                                const response = await installPackageZip(zip);
                                if (response.status !== 200) {
                                    setMode('error');
                                    return;
                                }

                                setMode('success');
                            }

                        }}
                        className="bg-primary-500 hover:bg-primary-600 text-white px-4 py-2 rounded-lg transition-colors"
                    >
                        Import
                    </button>
                </div>
            </div>

        </>)}

        {mode === 'importing' && (<>
            <div className="space-y-1">
                <p className="text-gray-600 dark:text-gray-300">
                    Importing space...
                </p>
            </div>
        </>)}

        {mode === 'success' && (<>
            <div className="space-y-1">
                <p className="text-gray-600 dark:text-gray-300">
                    Space imported successfully
                </p>
            </div>
        </>)}

        {mode === 'error' && (<>
            <div className="space-y-1">
                <p className="text-gray-600 dark:text-gray-300">
                    Error importing space
                </p>
            </div>
        </>)}




    </>);


};
