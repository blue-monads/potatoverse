"use client";
import React from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import EventSubscriptionEditor from '../components/EventSubscriptionEditor';
import { getEventSubscription, updateEventSubscription, EventSubscription } from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';

export default function Page() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const spaceId = searchParams.get('space_id');
    const eventId = searchParams.get('event_id');

    if (!installId || !eventId) {
        return <div>Install ID or Event ID not provided</div>;
    }

    const loader = useSimpleDataLoader<EventSubscription>({
        loader: () => getEventSubscription(parseInt(installId), parseInt(eventId)),
        ready: true,
        dependencies: [installId, eventId],
    });

    const handleSave = async (data: any) => {
        try {
            await updateEventSubscription(parseInt(installId), parseInt(eventId), data);
            const params = new URLSearchParams();
            params.set('install_id', installId);
            if (spaceId) params.set('space_id', spaceId);
            router.push(`/portal/admin/spaces/tools/events?${params.toString()}`);
        } catch (error) {
            console.error('Failed to update event subscription:', error);
            alert('Failed to update event subscription: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
            throw error;
        }
    };

    const handleBack = () => {
        const params = new URLSearchParams();
        params.set('install_id', installId);
        if (spaceId) params.set('space_id', spaceId);
        router.push(`/portal/admin/spaces/tools/events?${params.toString()}`);
    };

    if (loader.loading) {
        return (
            <div className="flex items-center justify-center h-64">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
                    <p className="text-gray-500">Loading subscription...</p>
                </div>
            </div>
        );
    }

    if (!loader.data) {
        return <div>Subscription not found</div>;
    }

    return (
        <EventSubscriptionEditor
            onSave={handleSave}
            onBack={handleBack}
            initialData={loader.data}
        />
    );
}



