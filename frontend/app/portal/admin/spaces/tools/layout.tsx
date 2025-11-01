"use client"

import { Tabs } from '@skeletonlabs/skeleton-react';
import { usePathname, useRouter } from 'next/navigation';

const tabs = [
    {
        label: 'About',
        value: 'about',
        url: '/portal/admin/spaces/tools/about',
    },

    {
        label: 'Files',
        value: 'files',
        url: '/portal/admin/spaces/tools/files',
    },
    {
        label: 'Key-Value',
        value: 'kv',
        url: '/portal/admin/spaces/tools/kv',
    },
    {
        label: 'Package Files',
        value: 'package-files',
        url: '/portal/admin/spaces/tools/package-files',
    },
    {
        label: 'Capabilities',
        value: 'capabilities',
        url: '/portal/admin/spaces/tools/capabilities',
    },
    {
        label: 'Users',
        value: 'users',
        url: '/portal/admin/spaces/tools/users',
    },

    {
        label: 'Events',
        value: 'events',
        url: '/portal/admin/spaces/tools/events',
    },

    // {
    //     label: 'Data Tables',
    //     value: 'data-tables',
    //     url: '/portal/admin/spaces/tools/data-tables',
    // },

    // {
    //     label: 'Plugins',
    //     value: 'plugins',
    //     url: '/portal/admin/spaces/tools/plugins',
    // },

        // {
    //     label: 'Logs',
    //     value: 'logs',
    //     url: '/portal/admin/spaces/tools/logs',
    // },
]


interface PropsType {
    children: React.ReactNode;
}


const WithTabbedToolsLayout = (props: PropsType) => {
    const router = useRouter();
    const pathname = usePathname();
    // Extract the last segment of the path (e.g., 'about' from '/portal/admin/spaces/tools/about')
    const activeTab = pathname?.split('/').filter(Boolean).pop();
    return (

        <div className='w-full px-2'>
            <Tabs value={activeTab}
                onValueChange={(e) => {
                    const currentTab = tabs.find((tab) => tab.value === e.value);
                    if (currentTab) {
                        const params = new URLSearchParams(window.location.search);
                        router.push(`${currentTab.url}?${params.toString()}`);
                    }

                }}>
                <Tabs.List>
                    {tabs.map((tab) => (
                        <Tabs.Control classes='uppercase' key={tab.value} value={tab.value}>{tab.label}</Tabs.Control>
                    ))}
                </Tabs.List>
                <Tabs.Content>

                    {props.children}

                </Tabs.Content>
            </Tabs>
        </div>
    )
}

export default WithTabbedToolsLayout;