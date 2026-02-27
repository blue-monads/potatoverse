"use client"

import Link from 'next/link';
import { usePathname, useSearchParams } from 'next/navigation';
import { Info, FileText, Key, Package, Layers, Users, Calendar, BookOpen, Clock, Activity, FileCode, History, ShieldCheck, CloudLightning, Folder, User, Settings, ChevronDown, Upload, UploadCloudIcon, DownloadCloud } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { getInstalledPackageInfo, InstalledPackageInfo } from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { useGApp } from '@/hooks';
import { AddButton } from '@/contain/AddButton';

const navItems = [
    {
        label: 'Overview',
        value: 'overview',
        url: '/portal/admin/spaces/tools/overview',
        icon: Info,
    },
    {
        label: 'Versions',
        value: 'versions',
        url: '/portal/admin/spaces/tools/versions',
        icon: History,
    },
    {
        label: 'Files',
        value: 'files',
        url: '/portal/admin/spaces/tools/files',
        icon: Folder,
    },
    {
        label: 'Key-Value',
        value: 'kv',
        url: '/portal/admin/spaces/tools/kv',
        icon: Key,
    },
    {
        label: 'Capabilities',
        value: 'capabilities',
        url: '/portal/admin/spaces/tools/capabilities',
        icon: ShieldCheck,
    },
    {
        label: 'Users',
        value: 'users',
        url: '/portal/admin/spaces/tools/users',
        icon: User,
    },
    {
        label: 'Events',
        value: 'events',
        url: '/portal/admin/spaces/tools/events',
        icon: CloudLightning,
    },
    {
        label: 'Envs',
        value: 'env-vars',
        url: '/portal/admin/spaces/tools/env-vars',
        icon: Settings,
    },
    {
        label: 'Spec',
        value: 'docs',
        url: '/portal/admin/spaces/tools/docs',
        icon: BookOpen,
    },

];

interface PropsType {
    children: React.ReactNode;
}

const WithTabbedToolsLayout = (props: PropsType) => {
    const pathname = usePathname();
    const searchParams = useSearchParams();
    const gapp = useGApp();
    const installId = searchParams.get('install_id');
    const activeTab = pathname?.split('/').filter(Boolean).pop();

    const loader = useSimpleDataLoader<InstalledPackageInfo>({
        loader: () => getInstalledPackageInfo(parseInt(installId!)),
        ready: gapp.isInitialized && !!installId,
    });

    const packageData = loader.data?.installed_package;
    const packageVersions = loader.data?.package_versions || [];
    const activeVersion = packageData ? packageVersions.find(v => v.id === packageData.active_install_id) : null;

    const isActive = (value: string) => {
        return activeTab === value;
    };

    return (
        <div className="flex flex-col w-full h-full bg-surface-50">
            {/* Shared Package Header */}
            {packageData && (
                <div className="bg-white px-6 py-6 border-b border-gray-200">
                    <div className="max-w-7xl mx-auto flex flex-col md:flex-row gap-2 justify-between ">
                        <div className=" flex flex-col gap-2">
                            <div className="flex items-center gap-3">
                                <h1 className="text-3xl font-bold text-blue-600">
                                    {packageData.name}
                                </h1>
                                <span className="text-gray-500 text-lg">
                                    @ {activeVersion?.version || '0.0.0'}
                                </span>
                                <span className="bg-yellow-400 text-black text-[10px] font-bold px-2 py-0.5 rounded-full uppercase">
                                    latest
                                </span>
                            </div>

                            <div className="flex flex-row items-center gap-4 text-sm text-gray-600">
                                <div className='flex items-center gap-1'>
                                    <span className="text-gray-400">License</span>
                                    <span className="text-gray-300">•</span>
                                    <span className="text-gray-700">{activeVersion?.license || 'MIT'}</span>
                                </div>

                                <div className='flex items-center gap-1'>
                                    <span className="text-gray-400">Author</span>
                                    <span className="text-gray-300">•</span>
                                    <span className="text-gray-700">{activeVersion?.author_name || 'Anonymous'}</span>
                                </div>

                                <div className="flex items-center gap-1">
                                    {(activeVersion?.tags?.split(',') || ['deno', 'package']).map((tag) => (
                                        <span key={tag} className="bg-gray-100 text-gray-500 text-[10px] px-1.5 py-0.5 rounded-full border border-gray-200">
                                            {tag.trim()}
                                        </span>
                                    ))}
                                </div>
                            </div>
                        </div>
                        <div>


                            <button
                                className={"btn btn-sm md:btn-base  preset-filled text-white bg-secondary-600 hover:bg-secondary-700"}
                            >
                                <DownloadCloud className="w-3 h-3 md:w-4 md:h-4" />
                                Export
                            </button>

                        </div>

                    </div>

                </div>
            )}

            {/* Tabs Navigation */}
            <div className="flex items-center gap-1 border-b border-gray-200 overflow-x-auto no-scrollbar bg-white sticky top-0 z-10 px-4">
                <div className="max-w-7xl mx-auto flex items-center w-full">
                    {navItems.map((item) => {
                        const Icon = item.icon;
                        const active = isActive(item.value);
                        const params = new URLSearchParams(searchParams.toString());
                        const url = `${item.url}?${params.toString()}`;

                        return (
                            <Link
                                key={item.value}
                                href={url}
                                className={`flex items-center gap-2 px-4 py-3 text-sm font-medium transition-colors whitespace-nowrap border-b-2 ${active
                                    ? 'text-blue-600 border-blue-600 bg-blue-50/50'
                                    : 'text-gray-500 border-transparent hover:text-gray-700 hover:bg-gray-50'
                                    }`}
                            >
                                <Icon className="w-4 h-4" />
                                {item.label}
                            </Link>
                        );
                    })}
                </div>
            </div>

            <div className="flex-1">
                {props.children}
            </div>
        </div>
    );
}

export default WithTabbedToolsLayout;