"use client"

import { Tabs } from '@skeletonlabs/skeleton-react';
import { useRouter } from 'next/navigation';

const tabs = [
    {
        label: 'Logs',
        value: 'logs',
        url: '/portal/admin/spaces/tools/logs',
    },
    {
        label: 'Files',
        value: 'files',
        url: '/portal/admin/spaces/tools/files',
    },
    {
        label: 'KV',
        value: 'kv',
        url: '/portal/admin/spaces/tools/kv',
    },
    {
        label: 'Package',
        value: 'package-files',
        url: '/portal/admin/spaces/tools/package-files',
    },
    {
        label: 'Resources',
        value: 'resources',
        url: '/portal/admin/spaces/tools/resources',
    },
    {
        label: 'Users',
        value: 'users',
        url: '/portal/admin/spaces/tools/users',
    },
    {
        label: 'Plugins',
        value: 'plugins',
        url: '/portal/admin/spaces/tools/plugins',
    },
]


interface PropsType {
    children: React.ReactNode;
}


const WithTabbedToolsLayout = (props: PropsType) => {
    const router = useRouter();
    const activeTab = location.pathname.split('/').at(8);
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
                        <Tabs.Control key={tab.value} value={tab.value}>{tab.label}</Tabs.Control>
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