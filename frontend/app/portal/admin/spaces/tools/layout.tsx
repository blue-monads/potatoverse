"use client"

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Info, FileText, Key, Package, Layers, Users, Calendar, BookOpen } from 'lucide-react';

const navItems = [
    {
        label: 'About',
        value: 'about',
        url: '/portal/admin/spaces/tools/about',
        icon: Info,
    },
    {
        label: 'Files',
        value: 'files',
        url: '/portal/admin/spaces/tools/files',
        icon: FileText,
    },
    {
        label: 'Key-Value',
        value: 'kv',
        url: '/portal/admin/spaces/tools/kv',
        icon: Key,
    },
    {
        label: 'Package Files',
        value: 'package-files',
        url: '/portal/admin/spaces/tools/package-files',
        icon: Package,
    },
    {
        label: 'Capabilities',
        value: 'capabilities',
        url: '/portal/admin/spaces/tools/capabilities',
        icon: Layers,
    },
    {
        label: 'Users',
        value: 'users',
        url: '/portal/admin/spaces/tools/users',
        icon: Users,
    },
    {
        label: 'Events',
        value: 'events',
        url: '/portal/admin/spaces/tools/events',
        icon: Calendar,
    },
    {
        label: 'Docs',
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
    // Extract the last segment of the path (e.g., 'about' from '/portal/admin/spaces/tools/about')
    const activeTab = pathname?.split('/').filter(Boolean).pop();

    const isActive = (value: string) => {
        return activeTab === value;
    };

    return (
        <div className="flex w-full h-full">
            <aside className="w-48 p-2 fixed left-14 top-0 h-full bg-white">
                <nav className="space-y-2 overflow-y-auto mt-24">
                    {navItems.map((item) => {
                        const Icon = item.icon;
                        const active = isActive(item.value);
                        const params = new URLSearchParams(window.location.search);

                        const url = `${item.url}?${params.toString()}`;


                        return (
                            <Link
                                key={item.value}
                                href={url}
                                className={`flex items-center gap-1 px-2 py-2 rounded-lg transition-colors text-sm ${
                                    active
                                        ? 'bg-blue-600 text-white'
                                        : 'text-gray-700 hover:bg-gray-200 hover:text-gray-900'
                                }`}
                            >
                                <Icon className="w-5 h-5" />
                                <span className="font-medium">{item.label}</span>
                            </Link>
                        );
                    })}
                </nav>
            </aside>

            <div className="flex-1 p-2 ml-48">
                {props.children}
            </div>
        </div>
    );
}

export default WithTabbedToolsLayout;