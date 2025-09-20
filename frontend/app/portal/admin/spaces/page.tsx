"use client";
import React, { useEffect, useRef, useState } from 'react';
import { Search, Filter, ArrowUpDown, Heart, Users, Zap, Image, Box, Octagon, SquareUserRound, BadgeDollarSign, BookOpenText, BookHeart, BriefcaseBusiness, Drama, Bolt, CloudLightning, ScrollText, Files, Grid2x2Plus, Cog, Trash2Icon, FileCode2 } from 'lucide-react';
import { createPortal } from 'react-dom';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { GAppStateHandle, ModalHandle, useGApp } from '@/hooks';
import { Tabs } from '@skeletonlabs/skeleton-react';
import { deletePackage, InstalledSpace, installPackage, installPackageZip, listInstalledSpaces, Package, Space } from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { staticGradients } from '@/app/utils';
import { useRouter } from 'next/navigation';



export default function Page() {
    return (<>
        <SpacesDirectory />
    </>)
}









const SpacesDirectory = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedFilter, setSelectedFilter] = useState('Relevance');
    const gapp = useGApp();
    const [packageIndex, setPackageIndex] = useState<Record<number, Package>>({});
    const router = useRouter();



    const loader = useSimpleDataLoader<InstalledSpace>({
        loader: listInstalledSpaces,
        ready: gapp.isInitialized,
    });

    useEffect(() => {
        if (loader.data && loader.data.packages) {
            const nextPackageIndex = loader.data.packages.reduce((acc, pkg) => {
                acc[pkg.id] = pkg;
                return acc;
            }, {} as Record<number, Package>);

            setPackageIndex(nextPackageIndex);
        }
    }, [loader.data]);


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
                        router.push('/portal/admin/store');
                    }}
                />
            }
        >






            <BigSearchBar
                searchText={searchTerm}
                setSearchText={setSearchTerm}
            />


            <div className="max-w-7xl mx-auto px-6 py-8 w-full">
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
                        {loader.data?.spaces.map((space) => {

                            const pkg = packageIndex[space.package_id] || { name: "Unknown", description: "Unknown" };
                            const gradient = staticGradients[space.id % staticGradients.length];


                            return <SpaceCard
                                key={space.id}
                                actionHandler={async (action: string) => {

                                    if (action === "delete") {
                                        // Show confirmation modal
                                        gapp.modal.openModal({
                                            title: "Delete Space",
                                            content: (
                                                <div className="space-y-4">
                                                    <div className="flex items-center gap-3">
                                                        <div className="w-10 h-10 bg-red-100 dark:bg-red-900/20 rounded-full flex items-center justify-center">
                                                            <Trash2Icon className="w-5 h-5 text-red-600 dark:text-red-400" />
                                                        </div>
                                                        <div>
                                                            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                                                                Are you sure you want to delete this space?
                                                            </h3>
                                                            <p className="text-sm text-gray-600 dark:text-gray-400">
                                                                This action cannot be undone. All data associated with this space will be permanently removed.
                                                            </p>
                                                        </div>
                                                    </div>
                                                    
                                                    <div className="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
                                                        <div className="flex items-center gap-3">
                                                            <div className={`w-8 h-8 rounded-lg bg-gradient-to-br ${gradient} flex items-center justify-center text-white text-sm font-semibold`}>
                                                                #{space.id}
                                                            </div>
                                                            <div>
                                                                <p className="font-medium text-gray-900 dark:text-white">{pkg.name}</p>
                                                                <p className="text-sm text-gray-600 dark:text-gray-400">{pkg.description || pkg.info}</p>
                                                            </div>
                                                        </div>
                                                    </div>

                                                    <div className="flex gap-3 justify-end">
                                                        <button
                                                            onClick={() => gapp.modal.closeModal()}
                                                            className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-600 hover:bg-gray-200 dark:hover:bg-gray-500 rounded-lg transition-colors"
                                                        >
                                                            Cancel
                                                        </button>
                                                        <button
                                                            onClick={async () => {
                                                                try {
                                                                    await deletePackage(space.id);
                                                                    loader.reload();
                                                                    gapp.modal.closeModal();
                                                                } catch (error) {
                                                                    console.error('Failed to delete space:', error);
                                                                    // You might want to show an error message here
                                                                }
                                                            }}
                                                            className="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg transition-colors"
                                                        >
                                                            Delete Space
                                                        </button>
                                                    </div>
                                                </div>
                                            ),
                                            size: "md"
                                        });
                                    } else if (action === "run") {
                                        router.push(`/portal/admin/exec?nskey=${space.namespace_key}`);
                                    } else if (action === "logs") {
                                        router.push(`/portal/admin/spaces/logs?id=${space.id}`);
                                    } else if (action === "files") {
                                        router.push(`/portal/admin/spaces/files?id=${space.id}`);
                                    } else if (action === "package-files") {
                                        router.push(`/portal/admin/spaces/package-files?packageId=${space.package_id}`);
                                    } else if (action === "kv") {
                                        router.push(`/portal/admin/spaces/kv?id=${space.id}`);
                                    }
                                    
                                }}

                                space={{
                                    id: space.id,
                                    title: pkg.name,
                                    description: pkg.description || pkg.info,
                                    author: "",
                                    timeAgo: "",
                                    gradient: gradient,
                                    from: pkg.slug,
                                    mcp: false,

                                }} />
                        })}
                    </div>
                </div>
            </div>
        </WithAdminBodyLayout>

    );
};

const SpaceCard = ({ space, actionHandler }: { space: any, actionHandler: any }) => {
    const router = useRouter();


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

                        <button

                            className="flex items-center gap-1 text-xs bg-white/20 backdrop-blur-sm px-3 py-2 rounded-lg hover:bg-white/40 transition-colors cursor-pointer hover:text-blue-600"
                            onClick={() => {
                                router.push(`/portal/admin/exec?nskey=${space.from}`);
                            }}

                        >
                            <CloudLightning className="w-4 h-4" />
                            <span>Run</span>
                        </button>

                        <ActionDropdown onClick={actionHandler} />
                    </div>



                </div>
            </div>
        </div>
    )
};


const actionsOptions = [
    { id: "run", label: "Run in dev mode", icon: <Bolt className="w-4 h-4" /> },
    { id: "logs", label: "Logs", icon: <ScrollText className="w-4 h-4" /> },
    { id: "package-files", label: "PackageFiles", icon: <FileCode2 className="w-4 h-4" /> },
    { id: "files", label: "Files", icon: <Files className="w-4 h-4" /> },
    { id: "kv", label: "KV State", icon: <Grid2x2Plus className="w-4 h-4" /> },
    { id: "tools", label: "Tools", icon: <Box className="w-4 h-4" /> },
    { id: "users", label: "Users", icon: <SquareUserRound className="w-4 h-4" /> },
    { id: "settings", label: "Settings", icon: <Cog className="w-4 h-4" /> },
    { id: "delete", label: "Delete", icon: <Trash2Icon className="w-4 h-4" /> }
]


interface ActionDropdownProps {
    onClick: (action: string) => void;
}

const ActionDropdown = (props: ActionDropdownProps) => {
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
        const dropdownRef = document.getElementById("action-dropdown");
        const handleClickOutside = (event: MouseEvent) => {
            if (
                isDropdownOpen &&
                buttonRef.current &&
                !buttonRef.current.contains(event.target as Node) &&
                dropdownRef &&
                !dropdownRef.contains(event.target as Node)
            ) {
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
                    id="action-dropdown"
                    className="fixed w-48 bg-white border border-gray-300 rounded-lg shadow-lg z-[9999]"
                    style={{
                        top: buttonRect.bottom + 4,
                        left: buttonRect.right - 192,
                    }}
                >
                    {actionsOptions.map((option) => (
                        <button
                            key={option.id}
                            onClick={async () => {
                                console.log("clicked", option.id);
                                props.onClick(option.id);
                                setTimeout(() => {
                                    setIsDropdownOpen(false);
                                }, 100);
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



