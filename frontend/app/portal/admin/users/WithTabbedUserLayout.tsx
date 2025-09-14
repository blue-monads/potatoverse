"use client"

import { Tabs } from '@skeletonlabs/skeleton-react';
import { useRouter } from 'next/navigation';

const tabs = [
    {
        label: 'Users',
        value: 'users',
        url: '/portal/admin/users',
    },
    {
        label: 'Invites',
        value: 'invites',
        url: '/portal/admin/users/invites',
    },
]


interface PropsType {
    children: React.ReactNode;
    activeTab: string;
}


const WithTabbedUserLayout = (props: PropsType) => {
    const router = useRouter();
    return (

        <div className='max-w-7xl mx-auto w-full px-2'>
            <Tabs value={props.activeTab}
                onValueChange={(e) => {
                    const currentTab = tabs.find((tab) => tab.value === e.value);
                    if (currentTab) {
                        router.push(currentTab.url);
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

export default WithTabbedUserLayout;