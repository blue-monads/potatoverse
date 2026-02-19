"use client"

import BigSearchBar from '@/contain/compo/BigSearchBar';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import { Tabs } from '@skeletonlabs/skeleton-react';
import { Settings } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useState } from 'react';

const tabs = [
    {
        label: 'Users',
        value: 'users',
        url: '/portal/admin/setting/users',
    },
    {
        label: 'Invites',
        value: 'invites',
        url: '/portal/admin/setting/users/invites',
    },
    {
        label: 'Groups',
        value: 'groups',
        url: '/portal/admin/setting/users/groups',
    },
]


interface PropsType {
    children: React.ReactNode;
    activeTab: string;
}


const WithTabbedUserLayout = (props: PropsType) => {
    const router = useRouter();
    const [searchTerm, setSearchTerm] = useState('');
    return (


        <WithAdminBodyLayout
            Icon={Settings}
            name='Settings'
            description="Manage configuration, preferences, and other settings for your application."
            rightContent={<>

            </>}

        >

            <BigSearchBar
                setSearchText={setSearchTerm}
                searchText={searchTerm}
                placeholder="Search settings..."
            />



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

        </WithAdminBodyLayout>
    )
}

export default WithTabbedUserLayout;