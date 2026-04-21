"use client";
import React, { useState } from 'react';
import { Zap, Filter, Edit, Trash2, Plus, CloudLightning } from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import {
    listEventSubscriptions,
    EventSubscription,
    deleteEventSubscription,
    listQueues,
    MQEvent,
} from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';

export default function Page() {
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const spaceId = searchParams.get('space_id');

    if (!installId) {
        return <div>Install ID not provided</div>;
    }

    return <EventSubscriptionsListingPage
        installId={parseInt(installId)}
        spaceId={spaceId ? parseInt(spaceId) : undefined}
    />;
}

const EventSubscriptionsListingPage = ({ installId, spaceId }: { installId: number; spaceId?: number }) => {
    const router = useRouter();
    const [searchTerm, setSearchTerm] = useState('');
    const [activeView, setActiveView] = useState<'subscriptions' | 'queue'>('subscriptions');

    const loader = useSimpleDataLoader<EventSubscription[]>({
        loader: () => {
            return listEventSubscriptions(installId);
        },
        ready: activeView === 'subscriptions',
        dependencies: [installId, spaceId, activeView],
    });

    const queueLoader = useSimpleDataLoader<MQEvent[]>({
        loader: () => {
            return listQueues(installId);
        },
        ready: activeView === 'queue',
        dependencies: [installId, spaceId, activeView],
    });

    // Filter data based on search term
    const filteredData = loader.data?.filter(sub => {
        const matchesSearch = searchTerm === '' ||
            sub.event_key.toLowerCase().includes(searchTerm.toLowerCase()) ||
            sub.target_type.toLowerCase().includes(searchTerm.toLowerCase()) ||
            sub.target_endpoint.toLowerCase().includes(searchTerm.toLowerCase());

        return matchesSearch;
    }) || [];

    const filteredQueueData = queueLoader.data?.filter(event => {
        const matchesSearch = searchTerm === '' ||
            event.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            event.status.toLowerCase().includes(searchTerm.toLowerCase());

        return matchesSearch;
    }) || [];

    const handleDelete = async (id: number) => {
        try {
            await deleteEventSubscription(installId, id);
            loader.reload();
        } catch (error) {
            console.error('Failed to delete event subscription:', error);
            alert('Failed to delete event subscription: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    const handleEdit = (id: number) => {
        const params = new URLSearchParams();
        params.set('install_id', installId.toString());
        params.set('event_id', id.toString());
        if (spaceId) params.set('space_id', spaceId.toString());
        router.push(`/portal/admin/spaces/tools/events/edit?${params.toString()}`);
    };

    const handleNew = () => {
        const params = new URLSearchParams();
        params.set('install_id', installId.toString());
        if (spaceId) params.set('space_id', spaceId.toString());
        router.push(`/portal/admin/spaces/tools/events/new?${params.toString()}`);
    };

    return (
        <WithAdminBodyLayout
            Icon={CloudLightning}
            name="Event Management"
            description="Manage event subscriptions and view queue history"
            variant="none"
        >


            <div className="max-w-7xl mx-auto px-4 py-4 w-full flex flex-col gap-4">
                {/* Table Header Area with Heading and View Switcher */}
                <div className="flex justify-between items-center mb-4">
                    <h2 className="text-xl font-semibold text-gray-900">
                        {activeView === 'subscriptions' ? 'Event Subscriptions' : 'Event Queue'}
                    </h2>
                    <div className="inline-flex bg-gray-100 rounded-lg p-1">
                        <button
                            onClick={() => setActiveView('subscriptions')}
                            className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${activeView === 'subscriptions'
                                    ? 'bg-white text-gray-900 shadow-sm'
                                    : 'text-gray-600 hover:text-gray-900 hover:bg-gray-200'
                                }`}
                        >
                            Subscriptions
                        </button>
                        <button
                            onClick={() => setActiveView('queue')}
                            className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${activeView === 'queue'
                                    ? 'bg-white text-gray-900 shadow-sm'
                                    : 'text-gray-600 hover:text-gray-900 hover:bg-gray-200'
                                }`}
                        >
                            Queue
                        </button>
                    </div>
                </div>


                {/* Search Bar */}
                <BigSearchBar
                    searchText={searchTerm}
                    setSearchText={setSearchTerm}
                    rightContent={
                        activeView === 'subscriptions' && (
                            <AddButton
                                name="+ New Subscription"
                                onClick={handleNew}
                            />
                        )
                    }
                />



                {/* Table based on active view */}
                {activeView === 'subscriptions' ? (
                    /* Subscriptions Table */
                    <div className="bg-white rounded-lg shadow overflow-hidden">
                        <div className="overflow-x-auto">
                            <table className="min-w-full divide-y divide-gray-200">
                                <thead className="bg-gray-50">
                                    <tr>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Event Key
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Target Type
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Endpoint
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Scope
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Status
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Actions
                                        </th>
                                    </tr>
                                </thead>
                                <tbody className="bg-white divide-y divide-gray-200">
                                    {loader.loading ? (
                                        <tr>
                                            <td colSpan={6} className="px-6 py-4 text-center text-gray-500">
                                                Loading...
                                            </td>
                                        </tr>
                                    ) : filteredData.length === 0 ? (
                                        <tr>
                                            <td colSpan={6} className="px-6 py-4 text-center text-gray-500">
                                                No event subscriptions found
                                            </td>
                                        </tr>
                                    ) : (
                                        filteredData.map((sub) => (
                                            <tr key={sub.id} className="hover:bg-gray-50">
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    <div className="flex items-center gap-2">
                                                        <Zap className="w-4 h-4 text-yellow-500" />
                                                        <div className="text-sm font-medium text-gray-900">
                                                            {sub.event_key || '-'}
                                                        </div>
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                                                        {sub.target_type || '-'}
                                                    </span>
                                                </td>
                                                <td className="px-6 py-4">
                                                    <div className="text-sm text-gray-900 truncate max-w-xs">
                                                        {sub.target_endpoint || '-'}
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${sub.space_id === 0
                                                        ? 'bg-purple-100 text-purple-800'
                                                        : 'bg-green-100 text-green-800'
                                                        }`}>
                                                        {sub.space_id === 0 ? 'Package (Root)' : `Space ${sub.space_id}`}
                                                    </span>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${sub.disabled
                                                        ? 'bg-red-100 text-red-800'
                                                        : 'bg-green-100 text-green-800'
                                                        }`}>
                                                        {sub.disabled ? 'Disabled' : 'Active'}
                                                    </span>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                                    <div className="flex items-center gap-2">
                                                        <button
                                                            onClick={() => handleEdit(sub.id)}
                                                            className="text-blue-600 hover:text-blue-900"
                                                        >
                                                            <Edit className="w-4 h-4" />
                                                        </button>
                                                        <button
                                                            onClick={() => {
                                                                if (confirm('Are you sure you want to delete this event subscription?')) {
                                                                    handleDelete(sub.id);
                                                                }
                                                            }}
                                                            className="text-red-600 hover:text-red-900"
                                                        >
                                                            <Trash2 className="w-4 h-4" />
                                                        </button>
                                                    </div>
                                                </td>
                                            </tr>
                                        ))
                                    )}
                                </tbody>
                            </table>
                        </div>
                    </div>
                ) : (
                    /* Queue Table */
                    <div className="bg-white rounded-lg shadow overflow-hidden">
                        <div className="overflow-x-auto">
                            <table className="min-w-full divide-y divide-gray-200">
                                <thead className="bg-gray-50">
                                    <tr>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Event ID
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Event Name
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Status
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Created At
                                        </th>
                                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                            Payload
                                        </th>
                                    </tr>
                                </thead>
                                <tbody className="bg-white divide-y divide-gray-200">
                                    {queueLoader.loading ? (
                                        <tr>
                                            <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                                                Loading...
                                            </td>
                                        </tr>
                                    ) : filteredQueueData.length === 0 ? (
                                        <tr>
                                            <td colSpan={5} className="px-6 py-4 text-center text-gray-500">
                                                No queue events found
                                            </td>
                                        </tr>
                                    ) : (
                                        filteredQueueData.map((event) => (
                                            <tr key={event.id} className="hover:bg-gray-50">
                                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                                    #{event.id}
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    <div className="text-sm font-medium text-gray-900">
                                                        {event.name}
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${event.status === 'completed'
                                                        ? 'bg-green-100 text-green-800'
                                                        : event.status === 'failed'
                                                            ? 'bg-red-100 text-red-800'
                                                            : event.status === 'processing'
                                                                ? 'bg-blue-100 text-blue-800'
                                                                : 'bg-gray-100 text-gray-800'
                                                        }`}>
                                                        {event.status}
                                                    </span>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                                    {new Date(event.created_at).toLocaleString()}
                                                </td>
                                                <td className="px-6 py-4 text-sm text-gray-500">
                                                    <button
                                                        onClick={() => {
                                                            try {
                                                                const decoded = atob(event.payload);
                                                                alert(decoded);
                                                            } catch (e) {
                                                                alert(event.payload);
                                                            }
                                                        }}
                                                        className="text-blue-600 hover:text-blue-900 flex items-center gap-1"
                                                    >
                                                        <Filter className="w-3 h-3" /> View
                                                    </button>
                                                </td>
                                            </tr>
                                        ))
                                    )}
                                </tbody>
                            </table>
                        </div>
                    </div>
                )}
            </div>
        </WithAdminBodyLayout>
    );
};
