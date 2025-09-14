"use client";
import React, { useEffect, useRef, useState } from 'react';
import { Search, Filter, ArrowUpDown, Heart, Users, Zap, Image, Box, Octagon, SquareUserRound, BadgeDollarSign, BookOpenText, BookHeart, BriefcaseBusiness, Drama, Store, CloudDownload, InfoIcon, Bolt, Loader2 } from 'lucide-react';
import { createPortal } from 'react-dom';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import AddButton from '@/contain/AddButton';
import { GAppStateHandle, ModalHandle, useGApp } from '@/hooks';
import { Tabs } from '@skeletonlabs/skeleton-react';
import { EPackage, getUsers, installPackage, installPackageEmbed, installPackageZip, listEPackages, User } from '@/lib';
import { staticGradients } from '@/app/utils';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';



export default function Page() {
    return (<>
        <StoreDirectory />
    </>)
}



const StoreDirectory = () => {
    const [searchTerm, setSearchTerm] = useState('');    
    const [selectedFilter, setSelectedFilter] = useState('Relevance');
    const [isSortDropdownOpen, setIsSortDropdownOpen] = useState(false);
    const [selectedType, setSelectedType] = useState('Embed');
    const [isTypeDropdownOpen, setIsTypeDropdownOpen] = useState(false);


    
    const gapp = useGApp();
    const [storeItems, setStoreItems] = useState<any[]>([]);

    const categories = [
        { name: 'Personal', icon: BookHeart },
        { name: 'AI', icon: Octagon },
        { name: 'Productivity', icon: BriefcaseBusiness },
        { name: 'Entertainment', icon: Drama },
        { name: 'Finance', icon: BadgeDollarSign },
        { name: 'Education', icon: BookOpenText },
        { name: 'Social', icon: SquareUserRound },
    ];

    const loader = useSimpleDataLoader<EPackage[]>({
        loader: listEPackages,
        ready: gapp.isInitialized,
      });




    const sortOptions = [
        'Relevance',
        'Recently Updated',
        'By Usage'
    ];




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

            <div className="max-w-7xl mx-auto px-6 py-8 w-full">
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
                                    onClick={() => setIsSortDropdownOpen(!isSortDropdownOpen)}
                                    className="flex items-center gap-2 px-3 py-2 border border-gray-300 rounded-lg text-sm hover:bg-gray-50 transition-colors"
                                >
                                    <ArrowUpDown className="w-4 h-4" />
                                    <span>Sort: {selectedFilter}</span>
                                </button>

                                {isSortDropdownOpen && (
                                    <div className="absolute right-0 top-full mt-1 w-48 bg-white border border-gray-300 rounded-lg shadow-lg z-10">
                                        {sortOptions.map((option) => (
                                            <button
                                                key={option}
                                                onClick={() => {
                                                    setSelectedFilter(option);
                                                    setIsSortDropdownOpen(false);
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

                            <div className="relative">
                                <button
                                    onClick={() => setIsTypeDropdownOpen(!isTypeDropdownOpen)}
                                    className="flex items-center gap-2 px-3 py-2 border border-gray-300 rounded-lg text-sm hover:bg-gray-50 transition-colors"
                                >
                                    <ArrowUpDown className="w-4 h-4" />
                                    <span>Source: {selectedType}</span>
                                </button>

                                {isTypeDropdownOpen && (
                                    <div className="absolute right-0 top-full mt-1 w-48 bg-white border border-gray-300 rounded-lg shadow-lg z-10">
                                        {["Embed", "Official", "Community"].map((option) => (
                                            <button
                                                key={option}
                                                onClick={() => {
                                                    setSelectedType(option);
                                                    setIsTypeDropdownOpen(false);
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
                        {loader.data?.map((item, index) => (
                            <StoreItemCard key={item.name} item={item} index={index} />
                        ))}

                        {loader.loading && (
                            <div className="col-span-full">
                                <div className="flex items-center justify-center">
                                    <Loader2 className="w-20 h-20 animate-spin my-10 text-gray-700" />
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </WithAdminBodyLayout>

    );
};

const StoreItemCard = ({ item, index }: { item: any, index: number }) => {
    const gapp = useGApp();
    const gradient = staticGradients[index % staticGradients.length];
    item.gradient = item.gradient || gradient;



    return (

        <div className={`relative overflow-hidden rounded-xl bg-gradient-to-br ${item.gradient || " " } p-6 text-white min-h-[200px] group hover:scale-105 transition-transform duration-200 cursor-pointer`}>
            <div className="flex flex-col h-full justify-between">
                <div>

                    <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2 text-sm">

                            {item.mcp && (
                                <span className="bg-pink-500/80 px-2 py-1 rounded text-xs">🔥 MCP</span>
                            )}
                        </div>



                    </div>

                    <h3 className="text-xl font-bold mb-2">{item.name}</h3>
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

                        <button 
                        className="flex items-center gap-1 text-xs bg-white/20 backdrop-blur-sm px-3 py-2 rounded-lg hover:bg-white/40 transition-colors cursor-pointer hover:text-blue-600"
                        onClick={() => {
                            gapp.modal.openModal({
                                title: "Install Package",
                                content: <InstallPackageModal slug={item.slug} gapp={gapp} />,
                                size: "lg"
                            });
                             

                        }}
                        
                        >
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


const InstallPackageModal = ({ slug, gapp }: { slug: string, gapp: GAppStateHandle }) => {
    const [mode, setMode] = useState<'verify' | 'importing' | 'success' | 'error'>('verify');


    return (<>
        {mode === 'verify' && (<>
            <div className="space-y-1">
                <p className="text-gray-600 dark:text-gray-300">
                    Are you sure you want to install this package?
                </p>
            </div>

            <div className="flex gap-2 justify-end">
                <button
                    onClick={async () => {
                        setMode('importing');
                        const resp = await installPackageEmbed(slug);
                        if (resp.status !== 200) {
                            setMode('error');
                            return;
                        }
                        setMode('success');

                    }}
                    className="bg-primary-500 hover:bg-primary-600 text-white px-4 py-2 rounded-lg transition-colors"
                >
                    Install
                </button>
                <button
                    onClick={() => gapp.modal.closeModal()}
                    className="bg-red-500 hover:bg-red-600 text-white px-4 py-2 rounded-lg transition-colors"
                >
                    Cancel
                </button>
            </div>


        </>)}
        {mode === 'importing' && (<>
            <div className="space-y-1">
                <p className="text-gray-600 dark:text-gray-300">
                    Importing package...
                </p>
            </div>
        </>)}
        {mode === 'success' && (<>
            <div className="space-y-1">
                <p className="text-gray-600 dark:text-gray-300">
                    Package imported successfully
                </p>
            </div>
        </>)}
        {mode === 'error' && (<>
            <div className="space-y-1">
                <p className="text-gray-600 dark:text-gray-300">
                    Error importing package
                </p>
            </div>
        </>)}

    </>);
}