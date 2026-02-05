"use client";
import React from 'react';
import {
    Folder,
    File,
    ChevronRight,
    Activity,
    FileCode,
    History,
    Layers,
    Search,
    Info,
    BookOpen
} from 'lucide-react';

const packageInfo = {
    name: "Cimple GIS",
    version: "0.1.0",
    description: "Cimple GIS is a GIS (Geographic Information System) framework for Deno.",
    author: "Runelk",
    license: "MIT",
    repository: "https://github.com/runelk/cimple-gis",
    tags: ["gis", "deno", "framework"],
    publishedAt: "12 hours ago",
    isLatest: true,
}

const tabs = [
    { name: 'Overview', icon: Info },
    { name: 'Versions', icon: History },
    { name: 'Files', icon: Folder, active: true },
    { name: 'Spec', icon: BookOpen, },
]


const Page = () => {
    return (
        <div className="flex flex-col gap-6 p-6 max-w-7xl mx-auto w-full">
            {/* Header Section */}
            <div className="flex flex-col gap-2">
                <div className="flex items-center gap-3">
                    <h1 className="text-3xl font-bold text-blue-500">
                        {packageInfo.name}
                    </h1>
                    <span className="text-gray-500 text-lg">@ {packageInfo.version}</span>
                    {packageInfo.isLatest && (
                        <span className="bg-yellow-400 text-black text-xs font-bold px-2 py-0.5 rounded-full uppercase">
                            latest
                        </span>
                    )}
                </div>

                <div className="flex flex-col md:flex-row items-center gap-4 text-base text-gray-600">

                    <div className='flex items-center gap-1'>
                        <div>License</div>
                        <span className="text-gray-700">•</span>
                        <span className="text-gray-400 text-sm ">{packageInfo.license}</span>
                    </div>


                    <div className='flex items-center gap-1'>
                        <div>Published</div>
                        <span className="text-gray-700">•</span>
                        <span className="text-gray-400 text-sm ">
                            {packageInfo.publishedAt}  ({packageInfo.version})
                        </span>


                    </div>



                    <div className='flex items-center gap-1'>
                        <div>Author</div>
                        <span className="text-gray-700">•</span>
                        <span className="text-gray-400 text-sm ">{packageInfo.author}</span>

                    </div>




                    <div className="flex items-center gap-1">
                        {packageInfo.tags.map((tag) => (
                            <span key={tag} className="bg-gray-800 text-gray-400 text-[10px] px-1.5 py-0.5 rounded-full">
                                {tag}
                            </span>
                        ))}
                    </div>


                </div>
            </div>

            {/* Tabs */}
            <div className="flex items-center gap-1 border-b border-gray-800 overflow-x-auto no-scrollbar">
                {tabs.map((tab) => (
                    <button
                        key={tab.name}
                        className={`flex items-center gap-2 px-4 py-3 text-sm font-medium transition-colors whitespace-nowrap ${tab.active
                            ? 'text-blue-400 border-b-2 border-blue-400 bg-blue-400/5'
                            : 'text-gray-450 hover:text-gray-200 hover:bg-white/5'
                            }`}
                    >
                        <tab.icon className="w-4 h-4" />
                        {tab.name}

                    </button>
                ))}
            </div>



        </div>
    );
};

export default Page;