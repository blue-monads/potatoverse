"use client";
import React, { useEffect, useRef, useState } from 'react';
import { Search, Filter, ArrowUpDown, Heart, Users, Zap, Image, Box, Octagon, SquareUserRound, BadgeDollarSign, BookOpenText, BookHeart, BriefcaseBusiness, Drama, Bolt, CloudLightning, ScrollText, Files, Grid2x2Plus, Cog } from 'lucide-react';
import { createPortal } from 'react-dom';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { GAppStateHandle, ModalHandle, useGApp } from '@/hooks';
import { Tabs } from '@skeletonlabs/skeleton-react';
import { installPackage, installPackageZip } from '@/lib';



export default function Page() {
    return (<>
        <SpacesDirectory />
    </>)
}









const SpacesDirectory = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedFilter, setSelectedFilter] = useState('Relevance');
    const gapp = useGApp();

    const spaces = [
        {
            id: 1,
            title: 'Addit ⚡',
            description: 'Add objects to images using text prompts',
            author: 'nvidia',
            timeAgo: '2 days ago',
            from: 'ZERO',
            mcp: true,
            gradient: 'bg-gray-600'
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

        <WithAdminBodyLayout
            Icon={Box}
            name="Spaces"
            description="Your App Directory"
            rightContent={
                <AddButton
                    name="+ Space"
                    onClick={() => {

                        gapp.modal.openModal({
                            title: "Import Space",
                            content: (
                                <ImportSpaceModal gapp={gapp} />
                            ),
                            size: "lg"
                        });




                    }}
                />
            }
        >






            <BigSearchBar
                searchText={searchTerm}
                setSearchText={setSearchTerm}
            />


            <div className="max-w-7xl mx-auto px-6 py-8">
                <div className="mb-8">
                    <div className="flex items-center justify-between mb-6">
                        <div className="flex items-center gap-3">
                            <div className="w-3 h-3 bg-orange-500 rounded-full"></div>
                            <h2 className="text-xl font-bold">Installed Spaces</h2>
                        </div>

                        <div className="flex items-center gap-4">



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

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        {spaces.map((space) => (
                            <SpaceCard key={space.id} space={space} />
                        ))}
                    </div>
                </div>
            </div>
        </WithAdminBodyLayout>

    );
};

const SpaceCard = ({ space }: { space: any }) => {


    return (

        <div className={`relative overflow-hidden rounded-xl bg-gradient-to-br ${space.gradient} p-6 text-white min-h-[200px] group hover:scale-105 transition-transform duration-200 `}>
            <div className="flex flex-col h-full justify-between">
                <div>

                    <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2 text-sm">
                            <span className="font-semibold">
                                #{space.id}
                            </span>
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

                    <div className="flex gap-2">
                        {/* Run Action and other action drop down */}

                        <button className="flex items-center gap-1 text-xs bg-white/20 backdrop-blur-sm px-3 py-2 rounded-lg hover:bg-white/40 transition-colors cursor-pointer hover:text-blue-600">
                            <CloudLightning className="w-4 h-4" />
                            <span>Run</span>
                        </button>

                        <ActionDropdown />
                    </div>



                </div>
            </div>
        </div>
    )
};


const actionsOptions = [
    { label: "Run in dev mode", icon: <Bolt className="w-4 h-4" /> },
    { label: "Logs", icon: <ScrollText className="w-4 h-4" /> },
    { label: "Files", icon: <Files className="w-4 h-4" /> },
    { label: "KV State", icon: <Grid2x2Plus className="w-4 h-4" /> },
    { label: "Tools", icon: <Box className="w-4 h-4" /> },
    { label: "Users", icon: <SquareUserRound className="w-4 h-4" /> },
    { label: "Settings", icon: <Cog className="w-4 h-4" /> }
]



const ActionDropdown = () => {
    const [isDropdownOpen, setIsDropdownOpen] = useState(false);
    const [buttonRect, setButtonRect] = useState<DOMRect | null>(null);
    const buttonRef = useRef<HTMLButtonElement>(null);

    const handleToggleDropdown = () => {
        if (!isDropdownOpen && buttonRef.current) {
            const rect = buttonRef.current.getBoundingClientRect();
            setButtonRect(rect);
        }
        setIsDropdownOpen(!isDropdownOpen);
    };

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (isDropdownOpen && buttonRef.current && !buttonRef.current.contains(event.target as Node)) {
                setIsDropdownOpen(false);
            }
        };

        const handleScroll = () => {
            if (isDropdownOpen && buttonRef.current) {
                const rect = buttonRef.current.getBoundingClientRect();
                setButtonRect(rect);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        window.addEventListener('scroll', handleScroll, true);
        window.addEventListener('resize', handleScroll);

        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
            window.removeEventListener('scroll', handleScroll, true);
            window.removeEventListener('resize', handleScroll);
        };
    }, [isDropdownOpen]);

    return (
        <>
            <div className="flex items-center gap-4">
                <div className="relative">
                    <button
                        ref={buttonRef}
                        onClick={handleToggleDropdown}
                        className="flex items-center gap-2 px-3 py-2 border border-gray-300 rounded-lg text-sm hover:bg-gray-50 transition-colors hover:text-blue-600 cursor-pointer"
                    >
                        <Bolt className="w-4 h-4" />
                        <span>Actions</span>
                    </button>
                </div>
            </div>

            {/* Render dropdown in a portal */}
            {isDropdownOpen && buttonRect && createPortal(
                <div
                    className="fixed w-48 bg-white border border-gray-300 rounded-lg shadow-lg z-[9999]"
                    style={{
                        top: buttonRect.bottom + 4,
                        left: buttonRect.right - 192,
                    }}
                >
                    {actionsOptions.map((option) => (
                        <button
                            key={option.label}
                            onClick={() => {
                                setIsDropdownOpen(false);
                            }}
                            className="w-full text-left px-3 py-2 text-sm first:rounded-t-lg last:rounded-b-lg text-gray-700 hover:text-blue-600 transition-colors hover:bg-gray-200 cursor-pointer "
                        >
                            <div className="inline-flex items-center gap-2">
                                {option.icon}
                                {option.label}
                            </div>
                        </button>
                    ))}
                </div>,
                document.body
            )}
        </>
    );
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
                    Directly import spaces from a URL or upload a zip file.
                </p>

                <div className='flex gap-2 my-2 min-h-[200px]'>
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
